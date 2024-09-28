package compiler

import (
	"testing"

	"github.com/bobappleyard/zombie/util/sexpr"
	"github.com/bobappleyard/zombie/util/wasm"
	"github.com/bytecodealliance/wasmtime-go/v23"
)

func TestCodegen(t *testing.T) {
	engine := wasmtime.NewEngine()
	store := wasmtime.NewStore(engine)

	e, _, _ := sexpr.Read([]byte(`(lambda () 2)`))

	c := &compiler{
		pkg:  &pkg{},
		dest: &wasm.Func{},
	}

	visitExpression(c, e)
	c.dest.End()

	g := wasm.GlobalDecl{Type: wasm.Int32}
	g.Init.I32Const(0)
	g.Init.End()

	m := wasm.Module{
		Globals: []wasm.GlobalDecl{g},
		Codes:   []wasm.Func{*c.dest},
		Types: []wasm.Type{wasm.FuncType{
			In:  []wasm.Type{wasm.Int32, wasm.Int32},
			Out: []wasm.Type{wasm.Int32},
		}},
		Funcs:   []wasm.Index[wasm.Type]{0},
		Exports: []wasm.Export{wasm.FuncExport{Name: "test", Func: 0}},
	}

	module, err := wasmtime.NewModule(engine, m.AppendWasm(nil))
	if err != nil {
		t.Fatal(err)
	}

	instance, err := wasmtime.NewInstance(store, module, nil)
	if err != nil {
		t.Fatal(err)
	}

	f := instance.GetFunc(store, "test")
	x, err := f.Call(store, 16*1024, 2)
	if err != nil {
		t.Error(err)
	}

	t.Error(x.(int32))
}
