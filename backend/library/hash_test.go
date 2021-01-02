package library

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/imageloader"
)

const testAssetsDir = "../../testassets"

func TestGenerateHash(t *testing.T) {
	a := assert.New(t)
	imageLoader := imageloader.NewImageLoader()
	handle := apitype.NewHandle(-1, testAssetsDir, "vertical.jpg")

	if decodedImage, err := openImageForHashing(imageLoader, handle); err != nil {
		a.Fail(err.Error())
	} else {
		hash := generateHash(decodedImage, handle)
		a.NotNil(hash)
	}
}
