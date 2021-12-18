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
	err         error
}

func ReturnResult(channel chan *HashResult, imageId apitype.ImageId, hash *duplo.Hash, err error) {
	channel <- &HashResult{
		imageId: imageId,
		hash:    hash,
		err:     err,
	}
}

var hashImageSize = apitype.SizeOf(duplo.ImageScale, duplo.ImageScale)

func hashImage(input chan *apitype.ImageFile, output chan *HashResult, quitChannel chan bool, imageLoader api.ImageLoader) {
	for {
		select {
		case <-quitChannel:
			logger.Debug.Printf("Quit hashing")
			return
		case imageFile := <-input:
			{
				if decodedImage, err := openImageForHashing(imageLoader, imageFile); err != nil {
					ReturnResult(output, imageFile.Id(), nil, err)
				} else {
					hash := generateHash(decodedImage, imageFile)
					ReturnResult(output, imageFile.Id(), &hash, nil)
				}
			}
		}
	}
}

func openImageForHashing(imageLoader api.ImageLoader, imageFile *apitype.ImageFile) (image.Image, error) {
	startTime := time.Now()
	decodedImage, err := imageLoader.LoadImageScaled(imageFile.Id(), hashImageSize)
	endTime := time.Now()
	logger.Trace.Printf("'%s': Image loaded in %s", imageFile.Path(), endTime.Sub(startTime).String())
	return decodedImage, err
}

func generateHash(img image.Image, imageFile *apitype.ImageFile) duplo.Hash {
	startTime := time.Now()
	hash, _ := duplo.CreateHash(img)
	endTime := time.Now()
	logger.Trace.Printf("'%s': Calculated hash in %s", imageFile.Path(), endTime.Sub(startTime).String())
	return hash
}
