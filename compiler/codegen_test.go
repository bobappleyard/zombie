package compiler

import (
	"testing"

	"github.com/bobappleyard/zombie/util/assert"
	"github.com/bobappleyard/zombie/util/sexpr"
	"github.com/bobappleyard/zombie/util/wasm"
	"github.com/bytecodealliance/wasmtime-go/v23"
)

func TestCodegen(t *testing.T) {
	engine := wasmtime.NewEngine()
	store := wasmtime.NewStore(engine)

	e, _, _ := sexpr.Read([]byte(`

	(let ((f (lambda (x) (test x)))) (f 1))

	`))

	p := newPkg()

	p.compileToplevel(e)
	m := p.module()

	module, err := wasmtime.NewModule(engine, m.AppendWasm(nil))
	if err != nil {
		t.Fatal(err)
	}

	mem, err := wasmtime.NewMemory(store, wasmtime.NewMemoryType(1, false, 0))
	if err != nil {
		t.Fatal(err)
	}

	table, err := wasmtime.NewTable(
		store,
		wasmtime.NewTableType(wasmtime.NewValType(wasmtime.KindFuncref), 0, false, 0),
		wasmtime.ValFuncref(nil),
	)
	if err != nil {
		t.Fatal(err)
	}

	rtm, err := runtimeModule(engine, store, mem, table)
	if err != nil {
		t.Fatal(err)
	}
	fi := rtm.GetFunc(store, "call-procedure")

	instance, err := wasmtime.NewInstance(store, module, []wasmtime.AsExtern{fi, mem, table})
	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, instance)

	f := instance.GetFunc(store, "test")
	x, err := f.Call(store, int32(16*1024), int32(2))
	if err != nil {
		t.Error(err)
	}

	if x != nil {
		t.Error(x.(int32))
	}
}

func runtimeModule(e *wasmtime.Engine, s *wasmtime.Store, mem *wasmtime.Memory, table *wasmtime.Table) (*wasmtime.Instance, error) {
	m := wasm.Module{
		Types: []wasm.Type{
			wasm.FuncType{
				In:  []wasm.Type{wasm.Int32, wasm.Int32},
				Out: []wasm.Type{wasm.Int32},
			},
			wasm.FuncType{
				In:  []wasm.Type{wasm.Int32},
				Out: []wasm.Type{wasm.Int32},
			},
		},
		Codes: []wasm.Func{callProcedure(), isTrue()},
		Funcs: []wasm.Index[wasm.Type]{0, 1},
		Imports: []wasm.Import{
			wasm.MemoryImport{Type: wasm.MinMemory{}},
			wasm.TableImport{},
		},
		Exports: []wasm.Export{
			wasm.FuncExport{Name: "call-procedure", Func: 0},
			wasm.FuncExport{Name: "true?", Func: 1},
		},
	}

	mod, err := wasmtime.NewModule(e, m.AppendWasm(nil))
	if err != nil {
		return nil, err
	}

	return wasmtime.NewInstance(s, mod, []wasmtime.AsExtern{mem, table})
}

func callProcedure() wasm.Func {
	f := wasm.Func{
		Locals: []wasm.LocalDecl{{Count: 1, Type: wasm.Int32}},
	}
	f.LocalGet(0)
	f.I32Load(2, 0)
	f.LocalSet(2)
	f.LocalGet(2)
	f.I32Const(0xf)
	f.I32And()
	f.I32Const(5)
	f.I32Eq()
	f.If()
	f.LocalGet(0)
	f.I32Const(wordSize)
	f.I32Add()
	f.LocalGet(1)
	f.I32Const(1)
	f.I32Sub()
	f.LocalGet(2)
	f.I32Const(4)
	f.I32Shr()
	f.ReturnCallIndirect(0)
	f.Else()
	f.Unreachable()
	f.End()
	f.I32Const(0)
	f.End()

	return f
}

func isTrue() wasm.Func {
	f := wasm.Func{}
	f.LocalGet(0)
	f.I32Const(9)
	f.I32Ne()
	f.End()
	return f
}
