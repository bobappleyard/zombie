package compiler

import (
	"testing"

	"github.com/bobappleyard/zombie/util/assert"
	"github.com/bobappleyard/zombie/util/sexpr"
)

func TestSlotsNeeded(t *testing.T) {
	for _, test := range []struct {
		name       string
		expr       string
		args, vars int
	}{
		{
			name: "Symbol",
			expr: "a",
			args: 0,
			vars: 0,
		},
		{
			name: "Lambda",
			expr: "(lambda (x) x)",
			args: 0,
			vars: 0,
		},
		{
			name: "Call",
			expr: "(f 1 2 3)",
			args: 4,
			vars: 0,
		},
		{
			name: "Begin",
			expr: "(begin (f 1) 2)",
			args: 2,
			vars: 0,
		},
		{
			name: "Let",
			expr: "(let ((x 1) (y 2)) y)",
			args: 0,
			vars: 2,
		},
		{
			name: "NestedLet",
			expr: "(let ((x 1) (y (let ((z 2) (w 3)) w))) x)",
			args: 0,
			vars: 3,
		},
		{
			name: "NestedLetCalls",
			expr: "(let ((x (f 1)) (y (let ((z (f 2)) (w (f 3 4))) w))) x)",
			args: 3,
			vars: 3,
		},
		{
			name: "LetBody",
			expr: "(let ((x 1) (y 2)) (let ((x 3) (y 4) (z 5)) z))",
			args: 0,
			vars: 3,
		},
		{
			name: "LetBodyCall",
			expr: "(let ((x 1) (y 2)) (f 2))",
			args: 2,
			vars: 2,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			e, _, _ := sexpr.Read([]byte(test.expr))
			args, vars := neededSlots(e)
			assert.Equal(t, args, test.args)
			assert.Equal(t, vars, test.vars)
		})
	}
}
