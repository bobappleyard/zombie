package wasm

import (
	"testing"

	"github.com/bobappleyard/zombie/util/assert"
	"github.com/bytecodealliance/wasmtime-go/v23"
)

func TestTables(t *testing.T) {
	var m1 Module
	m1.Tables = []Table{FuncTable}
	m1.Exports = []Export{TableExport{Name: "table"}}

	var m2 Module
	m2.Imports = []Import{TableImport{Module: "m1", Name: "table"}}

	f := new(Func)
	f.I32Const(2)
	f.End()
	addTestFunc(&m2, *f)

	engine := wasmtime.NewEngine()
	store := wasmtime.NewStore(engine)

	mod1, err := wasmtime.NewModule(engine, m1.AppendWasm(nil))
	if err != nil {
		t.Error(err)
		return
	}
	inst1, err := wasmtime.NewInstance(store, mod1, nil)
	if err != nil {
		t.Error(err)
		return
	}

	mod2, err := wasmtime.NewModule(engine, m2.AppendWasm(nil))
	if err != nil {
		t.Error(err)
		return
	}

	exp := inst1.Exports(store)

	inst2, err := wasmtime.NewInstance(store, mod2, []wasmtime.AsExtern{exp[0]})
	if err != nil {
		t.Error(err)
		return
	}

	test := inst2.GetFunc(store, "test")

	res, err := test.Call(store, int32(1))
	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, res.(int32), 2)
}
