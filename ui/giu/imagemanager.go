package gtk

import (
	"github.com/AllenDang/giu"
	"image"
	"runtime"
	"sync"
	"time"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
	"vincit.fi/image-sorter/ui/giu/guiapi"
)

type ImageManager struct {
	currentDebounceEntry *debounceEntry
	currentImageEntry    *imageEntry
	mainImageMutex       sync.Mutex
	thumbnailMutex       sync.Mutex
	imageCache           api.ImageStore
	currentTexture       *guiapi.TexturedImage
	thumbnailCache       map[apitype.ImageId]*guiapi.TexturedImage
}

func NewImageManager(imageCache api.ImageStore) *ImageManager {
	return &ImageManager{
		currentDebounceEntry: &debounceEntry{
			imageFile: nil,
			timer:     nil,
			cancelled: true,
		},
		currentImageEntry: &imageEntry{
			currentImage: nil,
			width:        0,
			height:       0,
			zoomMode:     guiapi.ZoomFit,
		},
		mainImageMutex: sync.Mutex{},
		thumbnailMutex: sync.Mutex{},
		thumbnailCache: map[apitype.ImageId]*guiapi.TexturedImage{},
		imageCache:     imageCache,
		currentTexture: nil,
	}
}

type imageEntry struct {
	currentImage *apitype.ImageFile
	width        float32
	height       float32
	zoomMode     guiapi.ZoomMode
	zoomFactor   float32
}

type debounceEntry struct {
	imageFile  *apitype.ImageFile
	timer      *time.Timer
	cancelled  bool
	width      float32
	height     float32
	zoomMode   guiapi.ZoomMode
	zoomFactor float32
	mux        sync.Mutex
}

func (s *debounceEntry) Cancel() {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.cancelled = true
	if s.timer != nil {
		s.timer.Stop()
	}
	s.imageFile = nil
	s.timer = nil
}

const debounceTime = time.Millisecond * 150
const useThumbnailAsPlaceholder = false

func (s *ImageManager) GetThumbnailTexture(imageFile *apitype.ImageFile) *guiapi.TexturedImage {
	s.thumbnailMutex.Lock()
	defer s.thumbnailMutex.Unlock()

	imageId := imageFile.Id()
	if texture, ok := s.thumbnailCache[imageId]; ok {
		return texture
	} else {
		newEntry := guiapi.NewEmptyTexturedImage()
		s.thumbnailCache[imageId] = newEntry
		go func() {
			thumbnail, err := s.imageCache.GetThumbnail(imageId)
			if err != nil {
				logger.Error.Print(err)
				return
			}

			loadedTexture, err := giu.NewTextureFromRgba(thumbnail.(*image.RGBA))
			if err != nil {
				logger.Error.Print(err)
				return
			}
			loadedEntry := guiapi.NewTexturedImage(imageFile, loadedTexture)
			newEntry.SetLoaded(loadedEntry)
		}()
		return newEntry
	}
}

func (s *ImageManager) CurrentImage() *apitype.ImageFile {
	return s.currentImageEntry.currentImage
}

func (s *ImageManager) CurrentImageTexture() *guiapi.TexturedImage {
	if s.currentTexture != nil {
		return s.currentTexture
	} else {
		return guiapi.NewEmptyTexturedImage()
	}
}

func (s *ImageManager) SetSize(width float32, height float32, zoomMode guiapi.ZoomMode, zoomFactor float32) {
	if s.currentDebounceEntry != nil {
		s.SetCurrentImage(s.currentImageEntry.currentImage, width, height, zoomMode, zoomFactor)
	}
}
func (s *ImageManager) SetCurrentImage(newImage *apitype.ImageFile, width float32, height float32, zoomMode guiapi.ZoomMode, zoomFactor float32) {
	if newImage != nil {
		s.mainImageMutex.Lock()
		oldEntry := s.currentDebounceEntry
		imageChanged := newImage.Id() != oldEntry.imageFile.Id()
		s.mainImageMutex.Unlock()
		if oldEntry != nil {
			oldEntry.Cancel()
		}

		newEntry := &debounceEntry{
			imageFile:  newImage,
			width:      width,
			height:     height,
			zoomMode:   zoomMode,
			zoomFactor: zoomFactor,
		}

		s.mainImageMutex.Lock()
		s.currentDebounceEntry = newEntry
		if s.currentTexture != nil {
			if useThumbnailAsPlaceholder {
				s.thumbnailMutex.Lock()
				if thumbnail, ok := s.thumbnailCache[newImage.Id()]; ok {
					s.currentTexture.SetLoaded(thumbnail)
				} else {
					s.currentTexture.IsLoading = imageChanged
				}
				s.thumbnailMutex.Unlock()
			} else {
				s.currentTexture.IsLoading = imageChanged
			}
		}

		s.currentDebounceEntry.timer = s.createDelayedSendFunc(newEntry, debounceTime)
		s.mainImageMutex.Unlock()
	}
}

func (s *ImageManager) createDelayedSendFunc(entry *debounceEntry, duration time.Duration) *time.Timer {
	if logger.IsLogLevel(logger.TRACE) {
		logger.Trace.Printf("Creating new debounce entry for image %d", entry.imageFile.Id())
	}
	return time.AfterFunc(duration, func() {
		entry.mux.Lock()
		cancelled := entry.cancelled
		width, height, zoomMode, zoomFactor := entry.width, entry.height, entry.zoomMode, entry.zoomFactor
		imageFile := entry.imageFile
		entry.mux.Unlock()

		if !cancelled {
			if logger.IsLogLevel(logger.TRACE) {
				logger.Trace.Printf("Actually loading debounced image %d", entry.imageFile.Id())
			}

			var requiredW float32
			var requiredH float32
			if zoomMode == guiapi.ZoomFit {
				// If zoom to fit, then just load what was asked
				requiredW = width
				requiredH = height
			} else if zoomFactor < 1.0 {
				// If zoomed out, load using the image size and zoom factor
				requiredW = float32(imageFile.Width()) * zoomFactor
				requiredH = float32(imageFile.Height()) * zoomFactor
			} else {
				// If zoomed in, just load the max size image
				requiredW = float32(imageFile.Width())
				requiredH = float32(imageFile.Height())
			}

			img, err := s.imageCache.GetScaled(imageFile.Id(), apitype.SizeOf(int(requiredW), int(requiredH)))
			if err != nil {
				logger.Error.Print(err)
			}
			texture, err := giu.NewTextureFromRgba(img.(*image.RGBA))

			if err != nil {
				logger.Error.Print(err)
			}
			texturedImage := guiapi.NewTexturedImage(imageFile, texture)

			s.mainImageMutex.Lock()
			s.currentTexture = texturedImage
			s.currentImageEntry.currentImage = imageFile
			s.currentImageEntry.width = width
			s.currentImageEntry.height = height
			s.currentImageEntry.zoomMode = zoomMode
			s.currentTexture.IsLoading = false
			s.mainImageMutex.Unlock()

			logger.Trace.Printf("Loading image %d completed", imageFile.Id())
			runtime.GC()
		} else {
			if logger.IsLogLevel(logger.TRACE) {
				logger.Trace.Printf("Skip sending cancelled image %d", imageFile.Id())
			}
		}
	})
}
