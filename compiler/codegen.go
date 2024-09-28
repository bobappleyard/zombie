package compiler

import (
	"strconv"
	"strings"

	"github.com/bobappleyard/zombie/util/data"
	"github.com/bobappleyard/zombie/util/sexpr"
	"github.com/bobappleyard/zombie/util/wasm"
)

const (
	wordSize = 4

	argBaseLocal  = 0
	argCountLocal = 1
	callBaseLocal = 2
	varBaseLocal  = 3

	procedureCallFunc = 0
	isTrueFunc        = 1

	funcBaseGlobal = 0
)

type pkg struct {
	funcs []*wasm.Func
}

type compiler struct {
	pkg      *pkg
	dest     *wasm.Func
	base     int
	tail     bool
	bindings data.Set[binding]
}

// visitBegin implements visitor.
func (e *compiler) visitBegin(exprs sexpr.Expr) {
	for i, expr := range exprs.Items() {
		if i != 0 {
			e.dest.Drop()
		}
		visitExpression(e, expr)
	}
}

// visitCall implements visitor.
func (e *compiler) visitCall(call sexpr.Expr) {
	n := exprLen(call)

	for i, expr := range call.Items() {
		e.dest.LocalGet(callBaseLocal)
		// note that if the call is in tail position, these will be evaluated in tail position as
		// well. this is not technically accurate, however all nested calls have been factored out
		// into let bindings, so this is not a problem in practice
		visitExpression(e, expr)
		e.dest.I32Store(2, uint32(i)*wordSize)
	}

	if e.tail {

		// new arg base
		e.dest.LocalGet(argBaseLocal)
		e.dest.LocalGet(argCountLocal)
		e.dest.I32Add()
		e.dest.I32Const(int32(n))
		e.dest.I32Sub()
		e.dest.LocalSet(argBaseLocal)

		// shuffle arguments to new base
		for i := n - 1; i >= 0; i-- {
			e.dest.LocalGet(argBaseLocal)
			e.dest.LocalGet(callBaseLocal)
			e.dest.I32Load(2, uint32(i)*wordSize)
			e.dest.I32Store(2, uint32(i)*wordSize)
		}

		e.dest.LocalGet(argBaseLocal)
		e.dest.I32Const(int32(n))
		e.dest.ReturnCall(procedureCallFunc)
	} else {
		e.dest.LocalGet(callBaseLocal)
		e.dest.I32Const(int32(n))
		e.dest.Call(procedureCallFunc)
	}
}

// visitEmpty implements visitor.
func (e *compiler) visitEmpty() {
	panic("unimplemented")
}

// visitIf implements visitor.
func (e *compiler) visitIf(cond sexpr.Expr, then sexpr.Expr, els sexpr.Expr) {
	visitExpression(e, cond)
	e.dest.Call(isTrueFunc)
	e.dest.If()
	visitExpression(e, then)
	e.dest.Else()
	visitExpression(e, els)
	e.dest.End()
}

// visitLambda implements visitor.
func (e *compiler) visitLambda(vars sexpr.Expr, body sexpr.Expr) {

	// prepare env
	bindings := data.NewSet(binding.compare)
	for outer := range e.bindings.Items() {
		if _, ok := outer.kind.(*globalKind); !ok {
			continue
		}
		bindings.Put(outer)
	}
	for i, v := range vars.Items() {
		bindings.Put(binding{
			name:     v.UnsafeText(),
			kind:     &argKind{},
			position: i + 1,
		})
	}

	f := &wasm.Func{
		Locals: []wasm.LocalDecl{{
			Count: 2,
			Type:  wasm.Int32,
		}},
	}

	// prolog
	args, locals := neededSlots(body)

	f.LocalGet(argCountLocal)
	f.I32Const(int32(exprLen(vars)))
	f.I32Ne()
	f.If()
	f.Unreachable()
	f.End()
	f.LocalGet(argBaseLocal)
	f.I32Const(int32(locals))
	f.I32Sub()
	f.LocalSet(varBaseLocal)
	f.LocalGet(varBaseLocal)
	f.I32Const(int32(args))
	f.I32Sub()
	f.LocalSet(callBaseLocal)

	//compile the function
	c := &compiler{
		pkg:      e.pkg,
		dest:     f,
		tail:     true,
		base:     0,
		bindings: *bindings,
	}

	visitExpression(c, body)
	f.End()

	// get the table position
	e.dest.GlobalGet(funcBaseGlobal)
	e.dest.I32Const(int32(len(e.pkg.funcs)))
	e.dest.I32Add()

	e.pkg.funcs = append(e.pkg.funcs, f)

	// encode a procedure reference
	e.dest.I32Const(3)
	e.dest.I32Shl()
	e.dest.I32Const(5)
	e.dest.I32Or()

}

// visitLet implements visitor.
func (e *compiler) visitLet(bdgs sexpr.Expr, in sexpr.Expr) {
	bindings := data.NewSet(binding.compare)
	for outer := range e.bindings.Items() {
		bindings.Put(outer)
	}

	for i, b := range bdgs.Items() {
		var n, v sexpr.Expr
		b.Bind(&n, &v)

		inner := &compiler{
			dest:     e.dest,
			base:     e.base + i,
			tail:     false,
			bindings: e.bindings,
		}
		inner.base += i
		e.dest.LocalGet(varBaseLocal)
		visitExpression(inner, v)
		e.dest.I32Store(2, uint32(inner.base)*wordSize)

		bindings.Put(binding{
			name:     v.UnsafeText(),
			position: i,
		})
	}

	f := &compiler{
		dest:     e.dest,
		base:     e.base + exprLen(bdgs),
		tail:     e.tail,
		bindings: *bindings,
	}
	visitExpression(f, in)
}

// visitNumber implements visitor.
func (e *compiler) visitNumber(s string) {
	if strings.Contains(s, ".") {
		panic("unimplemented")
	}
	x, _ := strconv.ParseInt(s, 10, 30)
	if x < 0 {
		x = -x
		x <<= 2
		x |= 3
	} else {
		x <<= 2
		x |= 2
	}
	e.dest.I32Const(int32(x))
}

// visitSet implements visitor.
func (e *compiler) visitSet(dest sexpr.Expr, val sexpr.Expr) {
	b, _ := e.bindings.Get(binding{name: dest.UnsafeText()})
	b.kind.setCode(e.dest, b.position, func(f *wasm.Func) {
		visitExpression(e, val)
	})
}

// visitString implements visitor.
func (e *compiler) visitString(x string) {
	panic("unimplemented")
}

// visitSymbol implements visitor.
func (e *compiler) visitSymbol(x string) {
	b, _ := e.bindings.Get(binding{name: x})
	b.kind.getCode(e.dest, b.position)
}
