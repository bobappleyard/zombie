package sexpr

import (
	"testing"

	"github.com/bobappleyard/zombie/internal/assert"
)

func TestWrite(t *testing.T) {
	var b Builder
	b.ListStart()
	b.Symbol("b")
	b.ListEnd()

	e := b.Expr()

	assert.Equal(t, WriteString(e), "(b)")
}
