package wasm

type Table byte

const (
	FuncTable   Table = 0x70
	ExternTable Table = 0x6f
)

func (t Table) AppendWasm(buf []byte) []byte {
	return append(buf, byte(t), 0, 0)
}

type Element interface {
	WasmAppender
	element()
}

type FuncElement struct {
	Funcs []Index[Func]
}

func (FuncElement) element() {}

func (e *FuncElement) AppendWasm(buf []byte) []byte {
	buf = append(buf, 1, 0)
	buf = appendVector(buf, e.Funcs)
	return buf
}
