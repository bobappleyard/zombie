package wasm

type Type interface {
	WasmAppender
	typ()
	Matches(t Type) bool
}

type FuncType struct {
	In, Out []Type
}

func (FuncType) typ() {}

func (t FuncType) AppendWasm(buf []byte) []byte {
	buf = append(buf, 0x60)
	buf = appendVector(buf, t.In)
	buf = appendVector(buf, t.Out)
	return buf
}

func (t FuncType) Matches(u Type) bool {
	uf, ok := u.(FuncType)
	if !ok {
		return false
	}

	if len(uf.In) != len(t.In) {
		return false
	}
	if len(uf.Out) != len(t.Out) {
		return false
	}

	for i, a := range t.In {
		b := uf.In[i]
		if !a.Matches(b) {
			return false
		}
	}
	for i, a := range t.Out {
		b := uf.Out[i]
		if !a.Matches(b) {
			return false
		}
	}

	return true
}

type NumberType byte

const (
	Int32 NumberType = 0x7f - iota
	Int64
	Float32
	Float64
)

func (NumberType) typ() {}

func (t NumberType) AppendWasm(buf []byte) []byte {
	return append(buf, byte(t))
}

func (t NumberType) Matches(u Type) bool {
	return t == u
}
