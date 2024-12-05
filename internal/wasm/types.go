package wasm

type Type interface {
	WasmAppender
	typ()
	Matches(t Type) bool
}

type StructType struct {
	Fields []Field
}

func (StructType) typ() {}

type Field struct {
	Type Type
}

func (t StructType) AppendWasm(buf []byte) []byte {
	buf = append(buf, 0x5f)
	buf = appendVector(buf, t.Fields)
	return buf
}

func (t Field) AppendWasm(buf []byte) []byte {
	buf = t.Type.AppendWasm(buf)
	buf = append(buf, 1)
	return buf
}

func (t StructType) Matches(u Type) bool {
	us, ok := u.(StructType)
	if !ok {
		return false
	}
	if len(t.Fields) != len(us.Fields) {
		return false
	}
	for i, a := range t.Fields {
		b := us.Fields[i]
		if !a.Type.Matches(b.Type) {
			return false
		}
	}
	return true
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
	return arraysMatch(t.In, uf.In) && arraysMatch(t.Out, uf.Out)
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

func arraysMatch[T Type](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i, a := range a {
		b := b[i]
		if !a.Matches(b) {
			return false
		}
	}
	return true
}
