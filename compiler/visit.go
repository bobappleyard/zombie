package compiler

import "github.com/bobappleyard/zombie/util/sexpr"

type visitor interface {
	visitNumber(x string)
	visitString(x string)
	visitSymbol(x string)
	visitEmpty()
	visitCall(call sexpr.Expr)
	visitBegin(exprs sexpr.Expr)
	visitLet(bdgs sexpr.Expr, in sexpr.Expr)
	visitSet(dest, val sexpr.Expr)
	visitIf(cond, then, els sexpr.Expr)
	visitLambda(vars, body sexpr.Expr)
}

func visitExpression(v visitor, e sexpr.Expr) {
	switch e.Kind() {
	case sexpr.Number:
		v.visitNumber(e.UnsafeText())

	case sexpr.String:
		v.visitString(e.UnsafeText())

	case sexpr.Symbol:
		v.visitSymbol(e.UnsafeText())

	case sexpr.List:
		if e.Empty() {
			v.visitEmpty()
			return
		}

		if e.Head().Kind() == sexpr.Symbol {
			switch e.Head().UnsafeText() {
			case "begin":
				v.visitBegin(e.Tail())
				return

			case "let":
				var bdgs, in sexpr.Expr
				e.Tail().Bind(&bdgs, &in)
				v.visitLet(bdgs, in)
				return

			case "set!":
				var dest, val sexpr.Expr
				e.Tail().Bind(&dest, &val)
				v.visitSet(dest, val)
				return

			case "if":
				var cond, then, els sexpr.Expr
				e.Tail().Bind(&cond, &then, &els)
				v.visitIf(cond, then, els)
				return

			case "lambda":
				var vars, body sexpr.Expr
				e.Tail().Bind(&vars, &body)
				v.visitLambda(vars, body)
				return
			}
		}

		v.visitCall(e)
	}

}
