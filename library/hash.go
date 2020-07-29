package library

import (
	"image"
	"time"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/duplo"
	"vincit.fi/image-sorter/imageloader"
	"vincit.fi/image-sorter/logger"
)

type HashResult struct {
	handle      *common.Handle
	hash        *duplo.Hash
	imageLoader imageloader.ImageLoader
}

func ReturnResult(channel chan *HashResult, handle *common.Handle, hash *duplo.Hash) {
	channel <- &HashResult{
		handle: handle,
		hash:   hash,
	}
}

var hashImageSize = common.SizeOf(duplo.ImageScale, duplo.ImageScale)

func hashImage(input chan *common.Handle, output chan *HashResult, quitChannel chan bool, imageLoader imageloader.ImageLoader) {
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

func openImageForHashing(imageLoader imageloader.ImageLoader, handle *common.Handle) (image.Image, error) {
	startTime := time.Now()
	decodedImage, err := imageLoader.LoadImageScaled(handle, hashImageSize)
	endTime := time.Now()
	logger.Trace.Printf("'%s': Image loaded in %s", handle.GetPath(), endTime.Sub(startTime).String())
	return decodedImage, err
}

func generateHash(img image.Image, handle *common.Handle) duplo.Hash {
	startTime := time.Now()
	hash, _ := duplo.CreateHash(img)
	endTime := time.Now()
	logger.Trace.Printf("'%s': Calculated hash in %s", handle.GetPath(), endTime.Sub(startTime).String())
	return hash
}
