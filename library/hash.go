package library

import (
	"github.com/pixiv/go-libjpeg/jpeg"
	"log"
	"os"
	"time"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/duplo"
)

type HashResult struct {
	handle *common.Handle
	hash *duplo.Hash
}

func hashImage(input chan *common.Handle, output chan *HashResult) {
	for handle := range input {
		startTime := time.Now()
		imageFile, err := os.Open(handle.GetPath())
		defer imageFile.Close()
		if err != nil {
			ReturnResult(output, handle, nil)
		}
		decodedImage, err := jpeg.Decode(imageFile, &jpeg.DecoderOptions{})
		if err != nil {
			ReturnResult(output, handle, nil)
		}
		endTime := time.Now()
		log.Printf("'%s': Image loaded in %s", handle.GetPath(), endTime.Sub(startTime).String())

		startTime = time.Now()
		hash, _ := duplo.CreateHash(decodedImage)
		endTime = time.Now()
		log.Printf("'%s': Calculated hash in %s", handle.GetPath(), endTime.Sub(startTime).String())
		ReturnResult(output, handle, &hash)
	}

}
