package sexpr

import (
	"testing"

	"github.com/bobappleyard/zombie/util/assert"
)

func TestBuild(t *testing.T) {
	var b Builder

	b.ListStart()
	b.Atom(Symbol, "a")
	b.Atom(Symbol, "b")
	b.ListStart()
	b.Atom(Symbol, "c")
	b.ListEnd()
	b.ListEnd()

	e := b.Expr()
	assert.Equal(t, e.Kind(), List)
	assert.False(t, e.Empty())
	assert.Equal(t, e.Head().UnsafeText(), "a")

	e = e.Tail()
	assert.Equal(t, e.Kind(), List)
	assert.False(t, e.Empty())
	assert.Equal(t, e.Head().UnsafeText(), "b")

	e = e.Tail()
	assert.Equal(t, e.Kind(), List)
	assert.False(t, e.Empty())

	f := e.Head()
	assert.Equal(t, f.Kind(), List)
	assert.False(t, f.Empty())
	assert.Equal(t, f.Head().UnsafeText(), "c")

	e = e.Tail()
	f = f.Tail()
	assert.True(t, e.Empty())
	assert.True(t, f.Empty())
}

func TestListOps(t *testing.T) {
	e, _, _ := Read([]byte(`(1 2 3)`))

	assert.False(t, e.Empty())
	assert.Equal(t, e.Head().Kind(), Number)
	assert.Equal(t, e.Head().UnsafeText(), "1")

	e = e.Tail()
	assert.False(t, e.Empty())
	assert.Equal(t, e.Head().Kind(), Number)
	assert.Equal(t, e.Head().UnsafeText(), "2")

	e = e.Tail()
	assert.False(t, e.Empty())
	assert.Equal(t, e.Head().Kind(), Number)
	assert.Equal(t, e.Head().UnsafeText(), "3")

	e = e.Tail()
	assert.True(t, e.Empty())
}
