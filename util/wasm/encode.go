package wasm

type WasmAppender interface {
	AppendWasm(buf []byte) []byte
}

type Index[T any] uint32

func (i Index[T]) AppendWasm(buf []byte) []byte {
	return appendUint32(buf, uint32(i))
}

func appendInt32(buf []byte, x int32) []byte {
	if x == 0 {
		return append(buf, 0)
	}
	for {
		b := byte(x & 0x7F)
		sign := byte(x & 0x40)
		x >>= 7
		if sign == 0 && x != 0 {
			b |= 0x80
		}
		if x != -1 && (x != 0 || sign != 0) {
			b |= 0x80
		}
		buf = append(buf, b)
		if b&0x80 == 0 {
			break
		}
	}
	return buf
}

func appendUint32(buf []byte, x uint32) []byte {
	if x == 0 {
		return append(buf, 0)
	}
	for x != 0x00 {
		b := byte(x & 0x7F)
		x >>= 7
		if x != 0x00 {
			b |= 0x80
		}
		buf = append(buf, b)
	}
	return buf
}

func appendVector[T WasmAppender](buf []byte, xs []T) []byte {
	buf = appendUint32(buf, uint32(len(xs)))
	for _, x := range xs {
		buf = x.AppendWasm(buf)
	}
	return buf
}

func appendBytes(buf, bs []byte) []byte {
	buf = appendUint32(buf, uint32(len(bs)))
	buf = append(buf, bs...)
	return buf
}

func appendString(buf []byte, s string) []byte {
	return appendBytes(buf, []byte(s))
}

func appendSection[T WasmAppender](buf []byte, id byte, sec []T) []byte {
	if len(sec) == 0 {
		return buf
	}
	buf = append(buf, id)
	buf = appendBytes(buf, appendVector(nil, sec))
	return buf
}
