package compiler

import (
	"fmt"
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

	firstFunc = 1

	funcBaseGlobal    = 0
	packageDefsGlobal = 1
)

type pkg struct {
	funcs   []wasm.Func
	initFn  procedure
	globals data.Set[binding]
}

type procedure struct {
	wasm.Expr

	bindings data.Set[binding]
}

func (p procedure) toWasmFunc() wasm.Func {
	return wasm.Func{
		Locals: []wasm.LocalDecl{{Count: 2, Type: wasm.Int32}},
		Expr:   p.Expr,
	}
}

func newPkg() *pkg {
	p := &pkg{
		globals: *data.NewSet(binding.compare),
	}
	return p
}

func (p *pkg) compileToplevel(e sexpr.Expr) {
	for v := range freeVars(e).All() {
		if _, ok := p.globals.Get(binding{name: v}); ok {
			continue
		}
		p.globals.Put(binding{
			name:     v,
			kind:     &globalKind{},
			position: p.globals.Size(),
		})
	}
	c := &compiler{
		pkg:      p,
		bindings: *p.globals.Clone(),
		dest:     &p.initFn,
	}
	var b sexpr.Builder
	flattenCalls(&b, e)

	fmt.Print("\n\n------------------------\n\n")

	visitExpression(c, b.Expr())
	// p.initFn.Drop()
}

func (p *pkg) module() wasm.Module {
	// p.initFn.I32Const(0)
	p.initFn.End()

	pref := wasm.Func{}
	pref.I32Const(32 * 1024)
	pref.GlobalSet(packageDefsGlobal)
	pref.NullFunc()
	pref.I32Const(int32(len(p.funcs)))
	pref.TableGrow(0)
	pref.GlobalSet(funcBaseGlobal)
	pref.GlobalGet(funcBaseGlobal)
	pref.I32Const(0)
	pref.I32Const(int32(len(p.funcs)))
	pref.TableInit(0, 0)

	// HACK: need to work out how much to actually reserve
	pref.LocalGet(argBaseLocal)
	pref.LocalSet(callBaseLocal)

	p.initFn.Instructions = append(pref.Instructions, p.initFn.Instructions...)

	g := wasm.GlobalDecl{Type: wasm.Int32}
	g.Init.I32Const(0)
	g.Init.End()

	elements := make([]wasm.Index[wasm.Func], len(p.funcs))
	for i := range elements {
		elements[i] = wasm.Index[wasm.Func](i + firstFunc)
	}

	return wasm.Module{
		Globals: []wasm.GlobalDecl{g, g},
		Types: []wasm.Type{wasm.FuncType{
			In:  []wasm.Type{wasm.Int32, wasm.Int32},
			Out: []wasm.Type{wasm.Int32},
		}},
		Codes: append(p.funcs, p.initFn.toWasmFunc()),
		Funcs: make([]wasm.Index[wasm.Type], len(p.funcs)+1),
		Elements: []wasm.Element{
			&wasm.FuncElement{Funcs: elements},
		},
		Imports: []wasm.Import{
			wasm.FuncImport{Module: "runtime", Name: "call-procedure", Type: 0},
			wasm.MemoryImport{Module: "runtime", Name: "memory", Type: wasm.MinMemory{Min: 0}},
			wasm.TableImport{Module: "runtime", Name: "funcs"},
		},
		Exports: []wasm.Export{wasm.FuncExport{
			Name: "test",
			Func: wasm.Index[wasm.Func](len(p.funcs) + firstFunc),
		}},
	}
}

type compiler struct {
	pkg      *pkg
	dest     *procedure
	base     int
	tail     bool
	bindings data.Set[binding]
}

// visitBegin implements visitor.
func (e *compiler) visitBegin(exprs sexpr.Expr) {
	for i, expr := range exprs.All() {
		if i != 0 {
			e.dest.Drop()
		}
		visitExpression(e, expr)
	}
}

// visitCall implements visitor.
func (e *compiler) visitCall(call sexpr.Expr) {
	n := exprLen(call)

	for i, expr := range call.All() {
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
		e.dest.I32Const(wordSize)
		e.dest.I32Mul()
		e.dest.I32Add()
		e.dest.I32Const(int32(n) * wordSize)
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
	condC := &compiler{
		pkg:      e.pkg,
		dest:     e.dest,
		base:     e.base,
		tail:     false,
		bindings: e.bindings,
	}
	visitExpression(condC, cond)
	e.dest.Call(isTrueFunc)
	e.dest.If()
	visitExpression(e, then)
	e.dest.Else()
	visitExpression(e, els)
	e.dest.End()
}

// visitLambda implements visitor.
func (e *compiler) visitLambda(vars sexpr.Expr, body sexpr.Expr) {
	// get the table position
	e.dest.GlobalGet(funcBaseGlobal)
	e.dest.I32Const(e.compileProcedure(vars, body))
	e.dest.I32Add()

	// encode a procedure reference
	e.dest.I32Const(4)
	e.dest.I32Shl()
	e.dest.I32Const(5)
	e.dest.I32Or()
}

func (e *compiler) compileProcedure(vars sexpr.Expr, body sexpr.Expr) int32 {
	// prepare env
	bindings := data.NewSet(binding.compare)
	for outer := range e.bindings.All() {
		if _, ok := outer.kind.(*globalKind); !ok {
			continue
		}
		bindings.Put(outer)
	}
	for i, v := range vars.All() {
		bindings.Put(binding{
			name:     v.UnsafeText(),
			kind:     &argKind{},
			position: i,
		})
	}

	//compile the function
	c := &compiler{
		pkg:      e.pkg,
		dest:     &procedure{},
		tail:     true,
		base:     0,
		bindings: *bindings,
	}

	// prolog
	args, locals := neededSlots(body)
	fmt.Println(args, locals)

	c.dest.LocalGet(argCountLocal)
	c.dest.I32Const(int32(exprLen(vars)))
	c.dest.I32Ne()
	c.dest.If()
	c.dest.Unreachable()
	c.dest.End()
	c.dest.LocalGet(argBaseLocal)
	c.dest.I32Const(int32(locals) * wordSize)
	c.dest.I32Sub()
	c.dest.LocalSet(varBaseLocal)
	c.dest.LocalGet(varBaseLocal)
	c.dest.I32Const(int32(args) * wordSize)
	c.dest.I32Sub()
	c.dest.LocalSet(callBaseLocal)
	visitExpression(c, body)
	c.dest.End()

	id := len(e.pkg.funcs)
	e.pkg.funcs = append(e.pkg.funcs, c.dest.toWasmFunc())

	return int32(id)
}

// visitLet implements visitor.
func (e *compiler) visitLet(bdgs sexpr.Expr, in sexpr.Expr) {
	bindings := data.NewSet(binding.compare)
	for outer := range e.bindings.All() {
		bindings.Put(outer)
	}

	for i, b := range bdgs.All() {
		var n, v sexpr.Expr
		b.Bind(&n, &v)

		inner := &compiler{
			pkg:      e.pkg,
			dest:     e.dest,
			base:     e.base + i,
			tail:     false,
			bindings: e.bindings,
		}
		e.dest.LocalGet(varBaseLocal)
		visitExpression(inner, v)
		e.dest.I32Store(2, uint32(inner.base)*wordSize)

		bindings.Put(binding{
			name:     n.UnsafeText(),
			kind:     &localKind{},
			position: i,
		})
	}

	f := &compiler{
		pkg:      e.pkg,
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
	b.kind.setCode(e.dest, b.position, func(f *procedure) {
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
