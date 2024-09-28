package data

import (
	"strings"
	"testing"

	"github.com/bobappleyard/zombie/util/assert"
)

func TestSet(t *testing.T) {
	s := NewSet(strings.Compare)

	x, ok := s.Get("a")
	assert.Equal(t, x, "")
	assert.False(t, ok)
	assert.Equal(t, s.Size(), 0)

	s.Put("a")

	x, ok = s.Get("a")
	assert.Equal(t, x, "a")
	assert.True(t, ok)
	assert.Equal(t, s.Size(), 1)

	s.Delete("a")

	x, ok = s.Get("a")
	assert.Equal(t, x, "")
	assert.False(t, ok)

	s.Delete("a")

	x, ok = s.Get("a")
	assert.Equal(t, x, "")
	assert.False(t, ok)
	assert.Equal(t, s.Size(), 0)

	s.Put("a")

	for x := range s.Items() {
		assert.Equal(t, x, "a")
		break
	}

}
