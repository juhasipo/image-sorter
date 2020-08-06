package component

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_findZoomIndexForValue(t *testing.T) {
	a := assert.New(t)

	t.Run("0 returns the first", func(t *testing.T) {
		a.Equal(0, findZoomIndexForValue(0))
	})
	t.Run("smaller than first returns the first", func(t *testing.T) {
		a.Equal(0, findZoomIndexForValue(1))
	})
	t.Run("first level returns the first", func(t *testing.T) {
		a.Equal(0, findZoomIndexForValue(5))
	})
	t.Run("almost second level returns the first", func(t *testing.T) {
		a.Equal(0, findZoomIndexForValue(9))
	})
	t.Run("second level returns the second", func(t *testing.T) {
		a.Equal(1, findZoomIndexForValue(10))
	})
	t.Run("third level returns the third", func(t *testing.T) {
		a.Equal(2, findZoomIndexForValue(25))
	})
	t.Run("100", func(t *testing.T) {
		a.Equal(8, findZoomIndexForValue(100))
	})
	t.Run("second to the last returns the last", func(t *testing.T) {
		a.Equal(17, findZoomIndexForValue(500))
	})
	t.Run("last returns the last", func(t *testing.T) {
		a.Equal(18, findZoomIndexForValue(1000))
	})
	t.Run("after the last returns the last", func(t *testing.T) {
		a.Equal(18, findZoomIndexForValue(9999))
	})
}
