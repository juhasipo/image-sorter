package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSet(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		t.Run("Add", func(t *testing.T) {
			a := assert.New(t)
			set := NewSet[string]()
			set.Add("Foo")
			a.True(set.Contains("Foo"))
		})

		t.Run("Remove", func(t *testing.T) {
			a := assert.New(t)
			set := NewSet[string]()
			set.Add("Foo")
			set.Add("Bar")

			set.Remove("Bar")
			set.Remove("Fizz")

			a.True(set.Contains("Foo"))
			a.False(set.Contains("Bar"))
			a.False(set.Contains("Fizz"))
		})

		t.Run("Contains", func(t *testing.T) {
			a := assert.New(t)
			set := NewSet[string]()
			set.Add("Foo")
			set.Add("Bar")

			a.True(set.Contains("Foo"))
			a.True(set.Contains("Bar"))
			a.False(set.Contains("Fizz"))
		})
	})

	t.Run("int", func(t *testing.T) {
		t.Run("Add", func(t *testing.T) {
			a := assert.New(t)
			set := NewSet[int]()
			set.Add(1)
			a.True(set.Contains(1))
		})

		t.Run("Remove", func(t *testing.T) {
			a := assert.New(t)
			set := NewSet[int]()
			set.Add(1)
			set.Add(2)

			set.Remove(2)
			set.Remove(3)

			a.True(set.Contains(1))
			a.False(set.Contains(2))
			a.False(set.Contains(3))
		})

		t.Run("Contains", func(t *testing.T) {
			a := assert.New(t)
			set := NewSet[int]()
			set.Add(1)
			set.Add(2)

			a.True(set.Contains(1))
			a.True(set.Contains(2))
			a.False(set.Contains(3))
		})
	})

}
