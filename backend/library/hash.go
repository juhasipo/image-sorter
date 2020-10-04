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
	handle      *apitype.Handle
	hash        *duplo.Hash
	imageLoader api.ImageLoader
}

func ReturnResult(channel chan *HashResult, handle *apitype.Handle, hash *duplo.Hash) {
	channel <- &HashResult{
		handle: handle,
		hash:   hash,
	}
}

var hashImageSize = apitype.SizeOf(duplo.ImageScale, duplo.ImageScale)

func hashImage(input chan *apitype.Handle, output chan *HashResult, quitChannel chan bool, imageLoader api.ImageLoader) {
	for {
		select {
		case <-quitChannel:
			logger.Debug.Printf("Quit hashing")
			return
		case handle := <-input:
			{
				if decodedImage, err := openImageForHashing(imageLoader, handle); err != nil {
					ReturnResult(output, handle, nil)
				} else {
					hash := generateHash(decodedImage, handle)
					ReturnResult(output, handle, &hash)
				}
			}
		}
	}
}

func openImageForHashing(imageLoader api.ImageLoader, handle *apitype.Handle) (image.Image, error) {
	startTime := time.Now()
	decodedImage, err := imageLoader.LoadImageScaled(handle, hashImageSize)
	endTime := time.Now()
	logger.Trace.Printf("'%s': Image loaded in %s", handle.GetPath(), endTime.Sub(startTime).String())
	return decodedImage, err
}

func generateHash(img image.Image, handle *apitype.Handle) duplo.Hash {
	startTime := time.Now()
	hash, _ := duplo.CreateHash(img)
	endTime := time.Now()
	logger.Trace.Printf("'%s': Calculated hash in %s", handle.GetPath(), endTime.Sub(startTime).String())
	return hash
}
