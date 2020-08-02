package library

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/imageloader"
)

func TestGenerateHash(t *testing.T) {
	a := assert.New(t)
	imageLoader := imageloader.NewImageLoader()
	handle := common.NewHandle("../testassets", "vertical.jpg")

	if decodedImage, err := openImageForHashing(imageLoader, handle); err != nil {
		a.Fail(err.Error())
	} else {
		hash := generateHash(decodedImage, handle)
		a.NotNil(hash)
	}
}
