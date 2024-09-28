package compiler

import (
	"strings"

	"github.com/bobappleyard/zombie/util/wasm"
)

type binding struct {
	name     string
	kind     bindingKind
	position int
}

func (a binding) compare(b binding) int {
	return strings.Compare(a.name, b.name)
}

type bindingKind interface {
	getCode(f *wasm.Func, p int)
	setCode(f *wasm.Func, p int, val func(*wasm.Func))
}

type argKind struct{}

// getCode implements bindingKind.
func (a *argKind) getCode(f *wasm.Func, p int) {
	f.LocalGet(argBaseLocal)
	f.I32Load(2, uint32(p)*wordSize)
}

// setCode implements bindingKind.
func (a *argKind) setCode(f *wasm.Func, p int, val func(*wasm.Func)) {
	f.LocalGet(argBaseLocal)
	val(f)
	f.I32Store(2, uint32(p)*wordSize)
}

type globalKind struct{}

// getCode implements bindingKind.
func (g *globalKind) getCode(f *wasm.Func, p int) {
	panic("unimplemented")
}

// setCode implements bindingKind.
func (g *globalKind) setCode(f *wasm.Func, p int, val func(*wasm.Func)) {
	panic("unimplemented")
}
