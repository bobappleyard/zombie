package wasm

type Memory interface {
	WasmAppender
	memory()
}

type MinMemory struct {
	Min uint32
}

func (MinMemory) memory() {}

func (m MinMemory) AppendWasm(buf []byte) []byte {
	buf = append(buf, 0)
	buf = appendUint32(buf, m.Min)
	return buf
}
