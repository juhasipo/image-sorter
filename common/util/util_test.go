package util

import (
	"github.com/stretchr/testify/assert"
	"math"
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

func TestMaxInt_None(t *testing.T) {
	a := assert.New(t)
	a.Equal(math.MinInt32, MaxInt())
}

func TestMaxInt_One(t *testing.T) {
	a := assert.New(t)
	a.Equal(10, MaxInt(10))
}

func TestMaxInt_Many(t *testing.T) {
	a := assert.New(t)
	a.Equal(5, MaxInt(1, 2, 3, 4, 5))
	a.Equal(5, MaxInt(-1, -2, 5, 4, 1))
	a.Equal(-1, MaxInt(-1, -2, -3, -4, -5))
	a.Equal(5, MaxInt(5, 1, 2, 4, 3))
}
