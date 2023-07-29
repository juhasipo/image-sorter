package internal

import (
	"github.com/AllenDang/giu"
	"image"
	"runtime"
	"sync"
	"time"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
	"vincit.fi/image-sorter/ui/giu/internal/guiapi"
)

type ImageManager struct {
	currentDebounceEntry *debounceEntry
	activeImageEntry     imageEntry
	loadedImageEntry     imageEntry
	mainImageMutex       sync.Mutex
	thumbnailMutex       sync.Mutex
	imageCache           api.ImageStore
	loadedImageTexture   *guiapi.TexturedImage
	thumbnailCache       map[apitype.ImageId]*guiapi.TexturedImage
}

func NewImageManager(imageCache api.ImageStore) *ImageManager {
	return &ImageManager{
		currentDebounceEntry: &debounceEntry{
			imageFile: nil,
			timer:     nil,
			cancelled: true,
		},
		activeImageEntry: imageEntry{
			currentImage: nil,
			width:        0,
			height:       0,
			zoomMode:     guiapi.ZoomFit,
		},
		loadedImageEntry: imageEntry{
			currentImage: nil,
			width:        0,
			height:       0,
			zoomMode:     guiapi.ZoomFit,
		},
		mainImageMutex:     sync.Mutex{},
		thumbnailMutex:     sync.Mutex{},
		thumbnailCache:     map[apitype.ImageId]*guiapi.TexturedImage{},
		imageCache:         imageCache,
		loadedImageTexture: nil,
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

			giu.NewTextureFromRgba(thumbnail.(*image.RGBA), func(loadedTexture *giu.Texture) {
				loadedEntry := guiapi.NewTexturedImage(imageFile, loadedTexture)
				newEntry.SetLoaded(loadedEntry)
			})

		}()
		return newEntry
	}
}

func (s *ImageManager) ActiveImageId() apitype.ImageId {
	if s.activeImageEntry.currentImage != nil {
		return s.activeImageEntry.currentImage.Id()
	} else {
		return apitype.NoImage
	}
}

func (s *ImageManager) LoadedImage() *apitype.ImageFile {
	return s.loadedImageEntry.currentImage
}

func (s *ImageManager) LoadedImageTexture() *guiapi.TexturedImage {
	if s.loadedImageTexture != nil {
		return s.loadedImageTexture
	} else {
		return guiapi.NewEmptyTexturedImage()
	}
}

func (s *ImageManager) SetSize(width float32, height float32, zoomStatus *ZoomStatus) {
	if s.currentDebounceEntry != nil {
		s.SetCurrentImage(s.loadedImageEntry.currentImage, width, height, zoomStatus)
	}
}
func (s *ImageManager) SetCurrentImage(newImage *apitype.ImageFile, width float32, height float32, zoomStatus *ZoomStatus) {
	if newImage != nil {
		s.mainImageMutex.Lock()
		s.activeImageEntry.currentImage = newImage
		s.activeImageEntry.width = width
		s.activeImageEntry.height = height
		s.activeImageEntry.zoomMode = zoomStatus.ZoomMode()
		s.activeImageEntry.zoomFactor = zoomStatus.ZoomLevel()
		oldDebounceEntry := s.currentDebounceEntry
		s.mainImageMutex.Unlock()

		imageChanged := true
		if oldDebounceEntry != nil {
			imageChanged = newImage.Id() != oldDebounceEntry.imageFile.Id()
			oldDebounceEntry.Cancel()
		}

		newEntry := &debounceEntry{
			imageFile:  newImage,
			width:      width,
			height:     height,
			zoomMode:   zoomStatus.ZoomMode(),
			zoomFactor: zoomStatus.ZoomLevel(),
		}

		s.mainImageMutex.Lock()
		s.currentDebounceEntry = newEntry
		if s.loadedImageTexture != nil {
			if useThumbnailAsPlaceholder {
				s.thumbnailMutex.Lock()
				if thumbnail, ok := s.thumbnailCache[newImage.Id()]; ok {
					s.loadedImageTexture.SetLoaded(thumbnail)
				} else {
					s.loadedImageTexture.IsLoading = imageChanged
				}
				s.thumbnailMutex.Unlock()
			} else {
				s.loadedImageTexture.IsLoading = imageChanged
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
			giu.NewTextureFromRgba(img.(*image.RGBA), func(texture *giu.Texture) {
				texturedImage := guiapi.NewTexturedImage(imageFile, texture)

				s.mainImageMutex.Lock()
				s.loadedImageEntry.currentImage = imageFile
				s.loadedImageEntry.width = width
				s.loadedImageEntry.height = height
				s.loadedImageEntry.zoomMode = zoomMode

				s.loadedImageTexture = texturedImage
				s.mainImageMutex.Unlock()

				logger.Trace.Printf("Loading image %d completed", imageFile.Id())
				runtime.GC()
			})
		} else {
			if logger.IsLogLevel(logger.TRACE) {
				logger.Trace.Printf("Skip sending cancelled image %d", imageFile.Id())
			}
		}
	})
}
