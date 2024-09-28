package wasm

type Import interface {
	WasmAppender
	imprt()
}

type Export interface {
	WasmAppender
	export()
}

func appendExport[T any](buf []byte, name string, id byte, index Index[T]) []byte {
	buf = appendString(buf, name)
	buf = append(buf, id)
	buf = index.AppendWasm(buf)

	return buf
}

type MemoryExport struct {
	Name string
	Mem  Index[Memory]
}

func (MemoryExport) export() {}

func (e MemoryExport) AppendWasm(buf []byte) []byte {
	return appendExport(buf, e.Name, 2, e.Mem)
}

type MemoryImport struct {
	Module string
	Name   string
	Type   Memory
}

func (MemoryImport) imprt() {}

func (e MemoryImport) AppendWasm(buf []byte) []byte {
	buf = appendString(buf, e.Module)
	buf = appendString(buf, e.Name)
	buf = append(buf, 2)
	buf = e.Type.AppendWasm(buf)
	return buf
}

type TableExport struct {
	Name  string
	Table Index[Table]
}

func (TableExport) export() {}

func (e TableExport) AppendWasm(buf []byte) []byte {
	return appendExport(buf, e.Name, 1, e.Table)
}

type TableImport struct {
	Module string
	Name   string
}

func (TableImport) imprt() {}

func (e TableImport) AppendWasm(buf []byte) []byte {
	buf = appendString(buf, e.Module)
	buf = appendString(buf, e.Name)
	// just do functions with no further requirements
	buf = append(buf, 1, 0x70, 0, 0)
	return buf
}
