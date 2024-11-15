package compiler

import "github.com/bobappleyard/zombie/util/sexpr"

func neededSlots(e sexpr.Expr) (args, vars int) {
	if e.Kind() != sexpr.List || e.Empty() {
		return 0, 0
	}

	if e.Head().Kind() == sexpr.Symbol {
		switch e.Head().UnsafeText() {
		case "begin", "set!", "if":
			for _, e := range e.Tail().All() {
				a, v := neededSlots(e)
				args = max(args, a)
				vars = max(vars, v)
			}
			return args, vars

		case "let":
			var bdgs, in sexpr.Expr
			e.Tail().Bind(&bdgs, &in)

			args, vars = neededSlots(in)
			vars = max(vars, exprLen(bdgs))
			for i, b := range bdgs.All() {
				var val sexpr.Expr
				b.Bind(nil, &val)

				a, v := neededSlots(val)
				args = max(args, a)
				vars = max(vars, v+i)
			}

			return args, vars

		case "lambda":
			return 0, 0

		}
	}

	return exprLen(e), 0

}

func exprLen(e sexpr.Expr) int {
	i := 0
	for range e.All() {
		i++
	}
	return i
}
