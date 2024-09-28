package compiler

import (
	"fmt"

	"github.com/bobappleyard/zombie/util/sexpr"
)

func flattenCalls(dest *sexpr.Builder, e sexpr.Expr) {
	if e.Kind() != sexpr.List || e.Empty() {
		dest.Copy(e)
		return
	}

	if e.Head().Kind() == sexpr.Symbol {
		switch e.Head().UnsafeText() {
		case "begin", "set!", "if":
			dest.ListStart()
			dest.Atom(sexpr.Symbol, e.Head().UnsafeText())
			for _, e := range e.Tail().Items() {
				flattenCalls(dest, e)
			}
			dest.ListEnd()
			return

		case "let":
			var bdgs, in sexpr.Expr
			e.Tail().Bind(&bdgs, &in)

			dest.ListStart()
			dest.Atom(sexpr.Symbol, "let")
			dest.ListStart()
			for _, b := range bdgs.Items() {
				var name, value sexpr.Expr
				b.Bind(&name, &value)
				dest.ListStart()
				dest.Atom(sexpr.Symbol, name.UnsafeText())
				flattenCalls(dest, value)
				dest.ListEnd()
			}
			dest.ListEnd()
			flattenCalls(dest, in)
			dest.ListEnd()
			return

		case "lambda":
			var vs, x sexpr.Expr
			e.Tail().Bind(&vs, &x)

			dest.ListStart()
			dest.Atom(sexpr.Symbol, "lambda")
			dest.Copy(vs)
			flattenCalls(dest, x)
			dest.ListEnd()
			return
		}
	}

	isNestedCall := func(e sexpr.Expr) bool {
		if e.Kind() != sexpr.List {
			return false
		}
		if s := e.Head(); s.Kind() == sexpr.Symbol && s.UnsafeText() == "lambda" {
			return false
		}
		return true
	}

	var nested []sexpr.Expr
	for _, e := range e.Items() {
		if isNestedCall(e) {
			nested = append(nested, e)
		}
	}

	if len(nested) != 0 {
		dest.ListStart()
		dest.Atom(sexpr.Symbol, "let")
		dest.ListStart()
		for i, b := range nested {
			dest.ListStart()
			dest.Atom(sexpr.Symbol, fmt.Sprintf("v%d", i))
			flattenCalls(dest, b)
			dest.ListEnd()
		}
		dest.ListEnd()
	}

	i := 0
	dest.ListStart()
	for _, e := range e.Items() {
		if isNestedCall(e) {
			dest.Atom(sexpr.Symbol, fmt.Sprintf("v%d", i))
			i++
		} else {
			flattenCalls(dest, e)
		}
	}
	dest.ListEnd()

	if len(nested) != 0 {
		dest.ListEnd()
	}
}
