package wasm

import (
	"testing"

	"github.com/bobappleyard/zombie/internal/assert"
	"github.com/bytecodealliance/wasmtime-go/v27"
)

func TestConst(t *testing.T) {
	var m Module

	c := new(Func)
	c.I32Const(12)
	c.End()

	addTestFunc(&m, *c)
	testModule(t, m, 0, 12)
}

func TestLogic(t *testing.T) {
	var m Module

	c := new(Func)
	c.Locals = []LocalDecl{{1, Int32}}

	c.LocalGet(0)
	c.If()
	c.I32Const(21)
	c.LocalSet(1)
	c.Else()
	c.I32Const(1)
	c.LocalSet(1)
	c.End()
	c.LocalGet(1)
	c.End()

	addTestFunc(&m, *c)

	testModule(t, m, 0, 1)
	testModule(t, m, 1, 21)
}

func TestCall(t *testing.T) {
	var m Module

	f := new(Func)
	f.LocalGet(0)
	f.I32Const(1)
	f.I32Add()
	f.End()

	m.Funcs = append(m.Funcs, m.Type(FuncType{In: []Type{Int32}, Out: []Type{Int32}}))
	m.Codes = append(m.Codes, *f)

	g := new(Func)
	g.LocalGet(0)
	g.Call(0)
	g.End()

	addTestFunc(&m, *g)

	testModule(t, m, 10, 11)
}

func TestReturnCall(t *testing.T) {
	var m Module

	f := new(Func)
	f.LocalGet(0)
	f.I32Const(1)
	f.I32Add()
	f.End()

	m.Funcs = append(m.Funcs, m.Type(FuncType{In: []Type{Int32}, Out: []Type{Int32}}))
	m.Codes = append(m.Codes, *f)

	g := new(Func)
	g.LocalGet(0)
	g.ReturnCall(0)
	g.End()

	addTestFunc(&m, *g)

	testModule(t, m, 10, 11)
}

func TestReturnCallStack(t *testing.T) {
	var m Module

	f := new(Func)
	f.Locals = append(f.Locals, LocalDecl{1, Int32})
	f.LocalGet(0)
	f.I32Const(1)
	f.I32Add()
	f.LocalSet(1)
	f.LocalGet(1)
	// this is a large enough value to overflow the stack
	f.I32Const(40000)
	f.I32Lts()
	f.If()
	f.LocalGet(1)
	f.ReturnCall(0)
	f.End()
	f.LocalGet(1)
	f.End()

	addTestFunc(&m, *f)

	testModule(t, m, 1, 40000)
}

func TestCallIndirect(t *testing.T) {
	var m Module
	m.Types = []Type{FuncType{In: []Type{Int32}, Out: []Type{Int32}}}

	f := new(Func)
	f.LocalGet(0)
	f.I32Const(1)
	f.I32Add()
	f.End()

	m.Funcs = append(m.Funcs, m.Type(FuncType{In: []Type{Int32}, Out: []Type{Int32}}))
	m.Codes = append(m.Codes, *f)

	g := new(Func)
	g.Locals = []LocalDecl{{1, Int32}}

	// TableGrow: [fillWith, growAmount] -> [oldSize]
	g.NullFunc()
	g.I32Const(10)
	g.TableGrow(0)
	g.Drop()

	// TableGrow: [fillWith, growAmount] -> [oldSize]
	g.NullFunc()
	g.I32Const(1)
	g.TableGrow(0)
	g.LocalSet(1)

	// TableInit: [destPos, srcPos, size] -> []
	g.LocalGet(1)
	g.I32Const(0)
	g.I32Const(1)
	g.TableInit(0, 0)

	g.LocalGet(0)
	g.LocalGet(1)
	g.CallIndirect(0)
	g.End()

	addTestFunc(&m, *g)

	m.Tables = []Table{FuncTable}
	m.Elements = []Element{&FuncElement{Funcs: []Index[Func]{0}}}

	testModule(t, m, 5, 6)
}

func TestLoop(t *testing.T) {
	var m Module
	c := new(Func)

	// var acc = 0
	c.Locals = []LocalDecl{{1, Int32}}

	c.Loop()

	// acc = acc + n
	c.LocalGet(0)
	c.LocalGet(1)
	c.I32Add()
	c.LocalSet(1)

	// n = n - 1
	c.LocalGet(0)
	c.I32Const(1)
	c.I32Sub()
	c.LocalSet(0)

	// for n > 0
	c.LocalGet(0)
	c.BrIf(0)

	c.End()

	c.LocalGet(1)
	c.End()

	addTestFunc(&m, *c)

	testModule(t, m, 3, 6)
}

func TestMemory(t *testing.T) {
	var m Module
	m.Memories = []Memory{MinMemory{0}}

	c := new(Func)
	c.Locals = []LocalDecl{{1, Int32}}

	// [amt] -> [old]
	c.I32Const(1)
	c.MemGrow()
	c.LocalSet(1)

	// [addr, val] -> []
	c.I32Const(1024)
	c.I32Const(45)
	c.I32Store(2, 0)

	// [addr] -> [val]
	c.I32Const(1024)
	c.I32Load(2, 0)

	c.End()

	addTestFunc(&m, *c)

	testModule(t, m, 0, 45)
}

func TestBrTable(t *testing.T) {
	var m Module

	f := new(Func)
	f.Block()
	f.Block()
	f.Block()
	f.LocalGet(0)
	f.BrTable([]uint32{0, 1, 2})
	f.End()
	f.I32Const(4)
	f.Return()
	f.End()
	f.I32Const(8)
	f.Return()
	f.End()
	f.Unreachable()
	f.End()

	addTestFunc(&m, *f)

	testModule(t, m, 1, 8)
}

func TestGlobal(t *testing.T) {
	var m Module

	g := GlobalDecl{Type: Int32}
	g.Init.I32Const(5)
	g.Init.End()

	m.Globals = append(m.Globals, g)

	f := new(Func)
	f.GlobalGet(0)
	f.End()

	addTestFunc(&m, *f)

	testModule(t, m, 1, 5)
}

func TestGC(t *testing.T) {
	var m Module

	m.Types = append(m.Types, StructType{Fields: []Field{{Int32}}})

	f := new(Func)
	f.StructNew(0)
	f.LocalGet(0)

	addTestFunc(&m, *f)

	testModule(t, m, 1, 1)
}

func addTestFunc(m *Module, f Func) {
	t := m.Type(FuncType{In: []Type{Int32}, Out: []Type{Int32}})
	idx := len(m.Funcs)
	m.Funcs = append(m.Funcs, t)
	m.Codes = append(m.Codes, f)
	m.Exports = append(m.Exports, FuncExport{Name: "test", Func: Index[Func](idx)})
}

func testModule(t *testing.T, m Module, in, out int32) {
	t.Helper()

	engine := wasmtime.NewEngine()
	defer engine.Close()
	store := wasmtime.NewStore(engine)
	mod, err := wasmtime.NewModule(engine, m.AppendWasm(nil))
	if err != nil {
		t.Error("creating module:", err)
		return
	}

	instance, err := wasmtime.NewInstance(store, mod, nil)
	if err != nil {
		t.Error("creating module:", err)
		return
	}

	f := instance.GetFunc(store, "test")
	res, err := f.Call(store, in)
	if err != nil {
		t.Error("creating module:", err)
		return
	}

	assert.Equal(t, res.(int32), out)
}
