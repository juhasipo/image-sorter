package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestImageContainer_String(t *testing.T) {
	a := assert.New(t)

	t.Run("Valid", func(t *testing.T) {
		container := ImageContainerNew(HandleNew("foo", "bar"), nil)
		a.Equal("ImageContainer{Handle{bar}}", container.String())
	})
	t.Run("Nil Handle", func(t *testing.T) {
		container := ImageContainerNew(nil, nil)
		a.Equal("ImageContainer{Handle<nil>}", container.String())
	})
	t.Run("Nil", func(t *testing.T) {
		var container *ImageContainer
		a.Equal("ImageContainer<nil>", container.String())
	})

}
