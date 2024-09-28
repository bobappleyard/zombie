package wasm

type Module struct {
	Types    []Type
	Imports  []Import
	Funcs    []Index[Type]
	Tables   []Table
	Globals  []GlobalDecl
	Memories []Memory
	Exports  []Export
	Codes    []Func
	Elements []Element
}

type LocalDecl struct {
	Count uint32
	Type  Type
}

type GlobalDecl struct {
	Type Type
	Init Expr
}

func (g GlobalDecl) AppendWasm(buf []byte) []byte {
	buf = g.Type.AppendWasm(buf)
	buf = append(buf, 1)
	buf = append(buf, g.Init.Instructions...)
	return buf
}

func (m *Module) AppendWasm(mod []byte) []byte {
	mod = m.wasmHeader(mod)
	mod = appendSection(mod, 1, m.Types)
	mod = appendSection(mod, 2, m.Imports)
	mod = appendSection(mod, 3, m.Funcs)
	mod = appendSection(mod, 4, m.Tables)
	mod = appendSection(mod, 5, m.Memories)
	mod = appendSection(mod, 6, m.Globals)
	mod = appendSection(mod, 7, m.Exports)
	mod = appendSection(mod, 9, m.Elements)
	mod = appendSection(mod, 10, m.Codes)
	return mod
}

func (m *Module) Type(t Type) Index[Type] {
	typeID := -1
	for i, u := range m.Types {
		if !t.Matches(u) {
			continue
		}
		typeID = i
		break
	}
	if typeID == -1 {
		typeID = len(m.Types)
		m.Types = append(m.Types, t)
	}
	return Index[Type](typeID)
}

func (m *Module) wasmHeader(buf []byte) []byte {
	buf = append(buf, 0)
	buf = append(buf, []byte("asm")...)
	buf = append(buf, 1, 0, 0, 0)
	return buf
}
