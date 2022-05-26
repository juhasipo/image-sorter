package util

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestReverseStringArray(t *testing.T) {
	assert.New(t)

	t.Run("int", func(t *testing.T) {
		t.Run("empty", func(t *testing.T) {
			var values []int
			Reverse(values)

			var expected []int
			assert.Equal(t, values, expected)
		})
		t.Run("one", func(t *testing.T) {
			values := []int{101}
			Reverse(values)

			expected := []int{101}
			assert.Equal(t, values, expected)
		})
		t.Run("many", func(t *testing.T) {

			values := []int{101, 100, 102}
			Reverse(values)

			expected := []int{102, 100, 101}
			assert.Equal(t, values, expected)
		})
	})

	t.Run("string", func(t *testing.T) {
		t.Run("empty", func(t *testing.T) {
			var values []string
			Reverse(values)

			var expected []string
			assert.Equal(t, values, expected)
		})
		t.Run("one", func(t *testing.T) {
			values := []string{"A"}
			Reverse(values)

			expected := []string{"A"}
			assert.Equal(t, values, expected)
		})
		t.Run("many", func(t *testing.T) {
			values := []string{"A", "B", "C"}
			Reverse(values)

			expected := []string{"C", "B", "A"}
			assert.Equal(t, values, expected)
		})
	})
}

func TestMaxInt(t *testing.T) {
	a := assert.New(t)

	t.Run("none", func(t *testing.T) {
		a.Equal(math.MinInt32, MaxInt())
	})
	t.Run("one", func(t *testing.T) {
		a.Equal(10, MaxInt(10))
	})
	t.Run("many", func(t *testing.T) {
		t.Run("positive", func(t *testing.T) {
			a.Equal(5, MaxInt(1, 2, 3, 4, 5))
		})
		t.Run("negative and positive", func(t *testing.T) {
			a.Equal(5, MaxInt(-1, -2, 5, 4, 1))
		})
		t.Run("negative", func(t *testing.T) {
			a.Equal(-1, MaxInt(-1, -2, -3, -4, -5))
		})
		t.Run("not in order", func(t *testing.T) {
			a.Equal(5, MaxInt(5, 1, 2, 4, 3))
		})
	})
}
