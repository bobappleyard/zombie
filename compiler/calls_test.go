package compiler

import (
	"testing"

	"github.com/bobappleyard/zombie/util/assert"
	"github.com/bobappleyard/zombie/util/sexpr"
)

func TestFlattenCall(t *testing.T) {
	for _, test := range []struct {
		name    string
		in, out string
	}{
		{
			name: "Atom",
			in:   "x",
			out:  "x",
		},
		{
			name: "Simple",
			in:   "(f x)",
			out:  "(f x)",
		},
		{
			name: "LambdaArg",
			in:   "(f x (lambda () x))",
			out:  "(f x (lambda () x))",
		},
		{
			name: "NestedOnce",
			in:   "(f (g x))",
			out:  "(let ((v0 (g x))) (f v0))",
		},
		{
			name: "NestedTwice",
			in:   "(f (g x (h y)))",
			out:  "(let ((v0 (let ((v0 (h y))) (g x v0)))) (f v0))",
		},
		{
			name: "Begin",
			in:   "(begin (f (g x)))",
			out:  "(begin (let ((v0 (g x))) (f v0)))",
		},
		{
			name: "Let",
			in:   "(let ((v0 (g x))) (f v0))",
			out:  "(let ((v0 (g x))) (f v0))",
		},
		{
			name: "Lambda",
			in:   "(lambda () (f (g x)))",
			out:  "(lambda () (let ((v0 (g x))) (f v0)))",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			e, _, _ := sexpr.Read([]byte(test.in))
			var dest sexpr.Builder
			flattenCalls(&dest, e)
			out := sexpr.WriteString(dest.Expr())
			assert.Equal(t, out, test.out)
		})
	}

}
