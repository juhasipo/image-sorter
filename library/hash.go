package library

import (
	"time"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/duplo"
	"vincit.fi/image-sorter/imageloader/goimage"
	"vincit.fi/image-sorter/logger"
)

type HashResult struct {
	handle      *common.Handle
	hash        *duplo.Hash
	imageLoader goimage.ImageLoader
}

var hashImageSize = common.SizeOf(duplo.ImageScale, duplo.ImageScale)

func hashImage(input chan *common.Handle, output chan *HashResult, quitChannel chan bool, imageLoader goimage.ImageLoader) {
	for {
		select {
		case <-quitChannel:
			logger.Debug.Printf("Quit")
			return
		case handle := <-input:
			{
				startTime := time.Now()
				decodedImage, err := imageLoader.LoadImageScaled(handle, hashImageSize)
				endTime := time.Now()
				logger.Debug.Printf("'%s': Image loaded in %s", handle.GetPath(), endTime.Sub(startTime).String())

				if err != nil {
					ReturnResult(output, handle, nil)
				}

				startTime = time.Now()
				hash, _ := duplo.CreateHash(decodedImage)
				endTime = time.Now()
				logger.Debug.Printf("'%s': Calculated hash in %s", handle.GetPath(), endTime.Sub(startTime).String())
				ReturnResult(output, handle, &hash)
			}
		}
	}
}
