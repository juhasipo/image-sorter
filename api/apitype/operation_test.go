package apitype

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOperation_NextOperation(t *testing.T) {
	a := assert.New(t)
	a.Equal(CATEGORIZE, UNCATEGORIZE.NextOperation())
	a.Equal(UNCATEGORIZE, CATEGORIZE.NextOperation())
}
