package apitype

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOperation_NextOperation(t *testing.T) {
	a := assert.New(t)
	a.Equal(MOVE, NONE.NextOperation())
	a.Equal(NONE, MOVE.NextOperation())
}
