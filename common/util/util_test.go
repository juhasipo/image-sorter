package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReverseStringArray(t *testing.T) {
	assert.New(t)

	values := []string{"A", "B", "C"}
	Reverse(values)

	expected := []string{"C", "B", "A"}
	assert.Equal(t, values, expected)
}

func TestReverseIntArray(t *testing.T) {
	assert.New(t)

	values := []int{101, 100, 102}
	Reverse(values)

	expected := []int{102, 100, 101}
	assert.Equal(t, values, expected)
}

func TestReverseNil(t *testing.T) {
	assert.New(t)
	Reverse(nil)
}

func TestReverseEmpty(t *testing.T) {
	assert.New(t)

	var values []int

	Reverse(values)

	var expected []int
	assert.Equal(t, values, expected)
}
