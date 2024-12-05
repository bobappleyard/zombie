package sexpr

import (
	"fmt"
	"testing"

	"github.com/bobappleyard/zombie/internal/assert"
)

func TestBuild(t *testing.T) {
	var b Builder
	b.ListStart()
	b.Symbol("a")
	b.Symbol("b")
	b.ListStart()
	b.Symbol("c")
	b.ListEnd()
	b.ListEnd()

	e := b.Expr()
	assert.Equal(t, e.Kind(), List)
	assert.False(t, e.Empty())
	assert.Equal(t, e.Head().UnsafeText(), "a")
	assert.Equal(t, e.Position(), 0)

	e = e.Tail()
	assert.Equal(t, e.Kind(), List)
	assert.False(t, e.Empty())
	assert.Equal(t, e.Head().UnsafeText(), "b")
	assert.Equal(t, e.Position(), 1)

	e = e.Tail()
	assert.Equal(t, e.Kind(), List)
	assert.False(t, e.Empty())

	f := e.Head()
	assert.Equal(t, f.Kind(), List)
	assert.False(t, f.Empty())
	assert.Equal(t, f.Head().UnsafeText(), "c")
	assert.Equal(t, f.Position(), 2)

	e = e.Tail()
	f = f.Tail()
	assert.True(t, e.Empty())
	assert.True(t, f.Empty())
}

func TestHeadTail(t *testing.T) {
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

func TestIter(t *testing.T) {
	e, _, _ := Read([]byte(`(1 2 3)`))

	for i, x := range e.All() {
		assert.Equal(t, x.UnsafeText(), fmt.Sprint(i+1))
		if i > 1 {
			break
		}
	}
}

func TestBind(t *testing.T) {
	e, _, _ := Read([]byte(`(1 2 3)`))

	var w, x, y, z Expr
	assert.False(t, e.Bind())
	assert.False(t, e.Bind(&x))
	assert.False(t, e.Bind(&x, &y))
	assert.True(t, e.Bind(&x, &y, &z))

	assert.Equal(t, x.UnsafeText(), "1")
	assert.Equal(t, y.UnsafeText(), "2")
	assert.Equal(t, z.UnsafeText(), "3")

	assert.False(t, e.Bind(&w, &x, &y, &z))
}
