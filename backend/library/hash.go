package library

import (
	"image"
	"time"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
	"vincit.fi/image-sorter/duplo"
)

type HashResult struct {
	imageId     apitype.ImageId
	hash        *duplo.Hash
	imageLoader api.ImageLoader
}

func ReturnResult(channel chan *HashResult, imageId apitype.ImageId, hash *duplo.Hash) {
	channel <- &HashResult{
		imageId: imageId,
		hash:    hash,
	}
}

var hashImageSize = apitype.SizeOf(duplo.ImageScale, duplo.ImageScale)

func hashImage(input chan *apitype.ImageFile, output chan *HashResult, quitChannel chan bool, imageLoader api.ImageLoader) {
	for {
		select {
		case <-quitChannel:
			logger.Debug.Printf("Quit hashing")
			return
		case handle := <-input:
			{
				if decodedImage, err := openImageForHashing(imageLoader, handle); err != nil {
					ReturnResult(output, handle.GetId(), nil)
				} else {
					hash := generateHash(decodedImage, handle)
					ReturnResult(output, handle.GetId(), &hash)
				}
			}
		}
	}
}

func openImageForHashing(imageLoader api.ImageLoader, handle *apitype.ImageFile) (image.Image, error) {
	startTime := time.Now()
	decodedImage, err := imageLoader.LoadImageScaled(handle.GetId(), hashImageSize)
	endTime := time.Now()
	logger.Trace.Printf("'%s': Image loaded in %s", handle.GetPath(), endTime.Sub(startTime).String())
	return decodedImage, err
}

func generateHash(img image.Image, handle *apitype.ImageFile) duplo.Hash {
	startTime := time.Now()
	hash, _ := duplo.CreateHash(img)
	endTime := time.Now()
	logger.Trace.Printf("'%s': Calculated hash in %s", handle.GetPath(), endTime.Sub(startTime).String())
	return hash
}
