package compiler

import (
	"strings"
	"testing"

	"github.com/bobappleyard/zombie/util/assert"
	"github.com/bobappleyard/zombie/util/sexpr"
)

func TestFreeVars(t *testing.T) {
	for _, test := range []struct {
		name, expr, vars string
	}{
		{
			name: "Number",
			expr: `1`,
			vars: "",
		},
		{
			name: "String",
			expr: `"v"`,
			vars: "",
		},
		{
			name: "Symbol",
			expr: `v`,
			vars: "v",
		},
		{
			name: "Empty",
			expr: `()`,
			vars: "",
		},
		{
			name: "Begin",
			expr: `(begin a b)`,
			vars: "a,b",
		},
		{
			name: "Call",
			expr: `(f x)`,
			vars: "f,x",
		},
		{
			name: "If",
			expr: `(if a b c)`,
			vars: "a,b,c",
		},
		{
			name: "Set",
			expr: `(set! x y)`,
			vars: "x,y",
		},
		{
			name: "LambdaNoVars",
			expr: `(lambda () x)`,
			vars: "x",
		},
		{
			name: "LambdaVars",
			expr: `(lambda (x) x)`,
			vars: "",
		},
		{
			name: "Let",
			expr: `(let ((x 1)) x)`,
			vars: "",
		},
		{
			name: "LetBoundVar",
			expr: `(let ((x y)) x)`,
			vars: "y",
		},
		{
			name: "LetFreeVar",
			expr: `(let ((x 1)) y)`,
			vars: "y",
		},
		{
			name: "LetSelfBinding",
			expr: `(let ((x x)) y)`,
			vars: "x,y",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			expr, _, _ := sexpr.Read([]byte(test.expr))
			got := freeVars(expr)
			for _, v := range strings.Split(test.vars, ",") {
				if v == "" {
					continue
				}
				if _, ok := got.Get(v); !ok {
					t.Errorf("expected but not found: %q", v)
				}
				got.Delete(v)
			}
			for v := range got.Items() {
				t.Errorf("found but not expected: %q", v)
			}
		})
	}
}

func TestReplaceClosures(t *testing.T) {
	for _, test := range []struct {
		name, in, out string
	}{
		{
			name: "Atoms",
			in:   `(begin "a" 1 x)`,
			out:  `(begin "a" 1 x)`,
		},
		{
			name: "Let",
			in:   `(let ((x 1)) x)`,
			out:  `(let ((x 1)) x)`,
		},
		{
			name: "LambdaToplevel",
			in:   `(lambda () x)`,
			out:  `(lambda () x)`,
		},
		{
			name: "LambdaNested",
			in:   `(lambda (x) (lambda () x))`,
			out:  `(lambda (x) (curry (lambda (x) x) x))`,
		},
		{
			name: "LetLambda",
			in:   `(let ((x 1)) (lambda () x))`,
			out:  `(let ((x 1)) (curry (lambda (x) x) x))`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			in, _, _ := sexpr.Read([]byte(test.in))
			got := sexpr.WriteString(closuresToCurry(in))
			assert.Equal(t, got, test.out)
		})
	}
}
