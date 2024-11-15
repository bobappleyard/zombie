package compiler

import (
	"strings"

	"github.com/bobappleyard/zombie/util/data"
	"github.com/bobappleyard/zombie/util/sexpr"
)

func closuresToCurry(expr sexpr.Expr) sexpr.Expr {
	globals := freeVars(expr)
	dest := new(sexpr.Builder)
	visitExpression(&closureVisitor{globals, dest}, expr)
	return dest.Expr()
}

func freeVars(expr sexpr.Expr) *data.Set[string] {
	v := &freeVarsVisitor{freeVars: data.NewSet(strings.Compare)}
	visitExpression(v, expr)
	return v.freeVars
}

type closureVisitor struct {
	globals *data.Set[string]
	dest    *sexpr.Builder
}

// visitBegin implements visitor.
func (c *closureVisitor) visitBegin(exprs sexpr.Expr) {
	c.dest.ListStart()
	c.dest.Atom(sexpr.Symbol, "begin")
	for _, expr := range exprs.All() {
		visitExpression(c, expr)
	}
	c.dest.ListEnd()
}

// visitCall implements visitor.
func (c *closureVisitor) visitCall(call sexpr.Expr) {
	c.dest.ListStart()
	for _, expr := range call.All() {
		visitExpression(c, expr)
	}
	c.dest.ListEnd()
}

// visitEmpty implements visitor.
func (c *closureVisitor) visitEmpty() {
	c.dest.ListStart()
	c.dest.ListEnd()
}

// visitIf implements visitor.
func (c *closureVisitor) visitIf(cond sexpr.Expr, then sexpr.Expr, els sexpr.Expr) {
	c.dest.ListStart()
	c.dest.Atom(sexpr.Symbol, "if")
	visitExpression(c, cond)
	visitExpression(c, then)
	visitExpression(c, els)
	c.dest.ListEnd()
}

// visitLambda implements visitor.
func (c *closureVisitor) visitLambda(vars sexpr.Expr, body sexpr.Expr) {
	var captured []string
	args := data.NewSet(strings.Compare)
	for _, v := range vars.All() {
		args.Put(v.UnsafeText())
	}
	for v := range freeVars(body).All() {
		if _, ok := c.globals.Get(v); ok {
			continue
		}
		if _, ok := args.Get(v); ok {
			continue
		}
		captured = append(captured, v)
	}

	if len(captured) != 0 {
		c.dest.ListStart()
		c.dest.Atom(sexpr.Symbol, "curry")
	}

	c.dest.ListStart()
	c.dest.Atom(sexpr.Symbol, "lambda")
	c.dest.ListStart()
	for _, v := range captured {
		c.dest.Atom(sexpr.Symbol, v)
	}
	for _, v := range vars.All() {
		c.dest.Copy(v)
	}
	c.dest.ListEnd()
	visitExpression(c, body)
	c.dest.ListEnd()

	if len(captured) != 0 {
		for _, v := range captured {
			c.dest.Atom(sexpr.Symbol, v)
		}
		c.dest.ListEnd()
	}
}

// visitLet implements visitor.
func (c *closureVisitor) visitLet(bdgs sexpr.Expr, in sexpr.Expr) {
	c.dest.ListStart()
	c.dest.Atom(sexpr.Symbol, "let")
	c.dest.ListStart()
	for _, b := range bdgs.All() {
		var n, v sexpr.Expr
		b.Bind(&n, &v)

		c.dest.ListStart()
		c.dest.Copy(n)
		visitExpression(c, v)
		c.dest.ListEnd()
	}
	c.dest.ListEnd()
	visitExpression(c, in)
	c.dest.ListEnd()
}

// visitNumber implements visitor.
func (c *closureVisitor) visitNumber(x string) {
	c.dest.Atom(sexpr.Number, x)
}

// visitSet implements visitor.
func (c *closureVisitor) visitSet(dest sexpr.Expr, val sexpr.Expr) {
	c.dest.ListStart()
	c.dest.Atom(sexpr.Symbol, "set!")
	c.dest.Copy(dest)
	visitExpression(c, val)
	c.dest.ListEnd()
}

// visitString implements visitor.
func (c *closureVisitor) visitString(x string) {
	c.dest.Atom(sexpr.String, x)
}

// visitSymbol implements visitor.
func (c *closureVisitor) visitSymbol(x string) {
	c.dest.Atom(sexpr.String, x)
}

type freeVarsVisitor struct {
	freeVars *data.Set[string]
}

// visitBegin implements visitor.
func (f *freeVarsVisitor) visitBegin(exprs sexpr.Expr) {
	for _, e := range exprs.All() {
		visitExpression(f, e)
	}
}

// visitCall implements visitor.
func (f *freeVarsVisitor) visitCall(call sexpr.Expr) {
	for _, e := range call.All() {
		visitExpression(f, e)
	}
}

// visitEmpty implements visitor.
func (f *freeVarsVisitor) visitEmpty() {
}

// visitIf implements visitor.
func (f *freeVarsVisitor) visitIf(cond sexpr.Expr, then sexpr.Expr, els sexpr.Expr) {
	visitExpression(f, cond)
	visitExpression(f, then)
	visitExpression(f, els)
}

// visitLambda implements visitor.
func (f *freeVarsVisitor) visitLambda(vars sexpr.Expr, body sexpr.Expr) {
	inner := freeVars(body)
	for _, v := range vars.All() {
		inner.Delete(v.UnsafeText())
	}
	for v := range inner.All() {
		f.freeVars.Put(v)
	}
}

// visitLet implements visitor.
func (f *freeVarsVisitor) visitLet(bdgs sexpr.Expr, in sexpr.Expr) {
	inner := freeVars(in)
	for _, b := range bdgs.All() {
		var n, v sexpr.Expr
		b.Bind(&n, &v)

		inner.Delete(n.UnsafeText())
		visitExpression(f, v)
	}
	for v := range inner.All() {
		f.freeVars.Put(v)
	}
}

// visitNumber implements visitor.
func (f *freeVarsVisitor) visitNumber(x string) {
}

// visitSet implements visitor.
func (f *freeVarsVisitor) visitSet(dest sexpr.Expr, val sexpr.Expr) {
	visitExpression(f, dest)
	visitExpression(f, val)
}

// visitString implements visitor.
func (f *freeVarsVisitor) visitString(x string) {
}

// visitSymbol implements visitor.
func (f *freeVarsVisitor) visitSymbol(x string) {
	f.freeVars.Put(x)
}
