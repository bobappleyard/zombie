package compiler

import (
	"strings"
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
	getCode(f *procedure, p int)
	setCode(f *procedure, p int, val func(*procedure))
}

type argKind struct{}

// getCode implements bindingKind.
func (a *argKind) getCode(f *procedure, p int) {
	f.LocalGet(argBaseLocal)
	f.I32Load(2, uint32(p)*wordSize)
}

// setCode implements bindingKind.
func (a *argKind) setCode(f *procedure, p int, val func(*procedure)) {
	f.LocalGet(argBaseLocal)
	val(f)
	f.I32Store(2, uint32(p)*wordSize)
}

type localKind struct{}

// getCode implements bindingKind.
func (a *localKind) getCode(f *procedure, p int) {
	f.LocalGet(varBaseLocal)
	f.I32Load(2, uint32(p)*wordSize)
}

// setCode implements bindingKind.
func (a *localKind) setCode(f *procedure, p int, val func(*procedure)) {
	f.LocalGet(varBaseLocal)
	val(f)
	f.I32Store(2, uint32(p)*wordSize)
}

type globalKind struct{}

// getCode implements bindingKind.
func (g *globalKind) getCode(f *procedure, p int) {
	f.GlobalGet(packageDefsGlobal)
	f.I32Load(2, uint32(p)*wordSize)
}

// setCode implements bindingKind.
func (g *globalKind) setCode(f *procedure, p int, val func(*procedure)) {
	f.GlobalGet(packageDefsGlobal)
	val(f)
	f.I32Store(2, uint32(p)*wordSize)
}
