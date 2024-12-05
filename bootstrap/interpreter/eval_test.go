package main

import (
	"maps"
	"testing"

	"github.com/bobappleyard/zombie/internal/assert"
	"github.com/bobappleyard/zombie/internal/sexpr"
)

func TestEval(t *testing.T) {
	pk := &Pkg{
		defs: map[string]any{},
	}
	defs := map[string]any{
		"x":     1,
		"false": false,
	}
	for _, test := range []struct {
		name string
		in   string
		out  any
	}{
		{
			name: "Number",
			in:   `1`,
			out:  1,
		},
		{
			name: "String",
			in:   `"Hello, world!\n"`,
			out:  "Hello, world!\n",
		},
		{
			name: "Var",
			in:   `x`,
			out:  1,
		},
		{
			name: "If",
			in:   `(if 1 2 3)`,
			out:  2,
		},
		{
			name: "IfFalse",
			in:   `(if false 2 3)`,
			out:  3,
		},
		{
			name: "Begin",
			in:   `(begin 1 2 3)`,
			out:  3,
		},
		{
			name: "BeginCall",
			in:   `(let ((f (lambda (x) x))) (begin (f 1) 2 3))`,
			out:  3,
		},
		{
			name: "Let",
			in:   `(let ((y x) (x 2)) x)`,
			out:  2,
		},
		{
			name: "LetRebind",
			in:   `(let ((y x) (x 2)) y)`,
			out:  1,
		},
		{
			name: "LetScope",
			in:   `(begin (let ((x 2)) 3) x)`,
			out:  1,
		},
		{
			name: "Set",
			in:   `(begin (set! x 2) x)`,
			out:  2,
		},
		{
			name: "Lambda",
			in:   `((lambda (x) x) 5)`,
			out:  5,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			e, _, _ := sexpr.Read([]byte(test.in))
			p := &process{}
			s := &scope{
				pkg:  pk,
				defs: maps.Clone(defs),
			}
			p.eval(s, e, false)
			assert.Equal(t, p.value, test.out)
		})
	}
}
