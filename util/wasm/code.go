package wasm

type Func struct {
	Locals []LocalDecl
	Expr
}

type Expr struct {
	Instructions []byte
}

type FuncExport struct {
	Name string
	Func Index[Func]
}

func (FuncExport) export() {}

func (e FuncExport) AppendWasm(buf []byte) []byte {
	return appendExport(buf, e.Name, 0, e.Func)
}

type FuncImport struct {
	Module string
	Name   string
	Type   Index[Type]
}

func (FuncImport) imprt() {}

func (e FuncImport) AppendWasm(buf []byte) []byte {
	buf = appendString(buf, e.Module)
	buf = appendString(buf, e.Name)
	buf = append(buf, 0)
	buf = e.Type.AppendWasm(buf)
	return buf
}

func (c Func) AppendWasm(buf []byte) []byte {
	var tmp []byte
	tmp = appendVector(tmp, c.Locals)
	tmp = append(tmp, c.Instructions...)

	buf = appendBytes(buf, tmp)
	return buf
}

func (c LocalDecl) AppendWasm(buf []byte) []byte {
	buf = appendUint32(buf, c.Count)
	buf = c.Type.AppendWasm(buf)

	return buf
}

func (c *Expr) op(code byte, args ...uint32) {
	c.Instructions = append(c.Instructions, code)
	for _, arg := range args {
		c.Instructions = appendUint32(c.Instructions, arg)
	}
}

func (c *Expr) sop(code byte, args ...int32) {
	c.Instructions = append(c.Instructions, code)
	for _, arg := range args {
		c.Instructions = appendInt32(c.Instructions, arg)
	}
}

func (c *Expr) brTableOp(code byte, labels []uint32) {
	c.Instructions = append(c.Instructions, code)
	c.Instructions = appendUint32(c.Instructions, uint32(len(labels)-1))
	for _, label := range labels {
		c.Instructions = appendUint32(c.Instructions, label)
	}
}

func (c *Expr) Unreachable()                  { c.op(0x0) }
func (c *Expr) Block()                        { c.op(0x02, 0x40) }
func (c *Expr) Loop()                         { c.op(0x03, 0x40) }
func (c *Expr) If()                           { c.op(0x04, 0x40) }
func (c *Expr) Else()                         { c.op(0x05) }
func (c *Expr) End()                          { c.op(0x0b) }
func (c *Expr) Br(depth uint32)               { c.op(0x0c, depth) }
func (c *Expr) BrIf(depth uint32)             { c.op(0x0d, depth) }
func (c *Expr) BrTable(labels []uint32)       { c.brTableOp(0x0e, labels) }
func (c *Expr) Return()                       { c.op(0x0f) }
func (c *Expr) Call(idx uint32)               { c.op(0x10, idx) }
func (c *Expr) CallIndirect(idx uint32)       { c.op(0x11, idx, 0) }
func (c *Expr) ReturnCall(idx uint32)         { c.op(0x12, idx) }
func (c *Expr) ReturnCallIndirect(idx uint32) { c.op(0x13, idx, 0) }

func (c *Expr) Drop()                { c.op(0x1a) }
func (c *Expr) LocalGet(idx uint32)  { c.op(0x20, idx) }
func (c *Expr) LocalSet(idx uint32)  { c.op(0x21, idx) }
func (c *Expr) GlobalGet(idx uint32) { c.op(0x23, idx) }
func (c *Expr) GlobalSet(idx uint32) { c.op(0x24, idx) }

func (c *Expr) NullFunc() { c.op(0xd0, 0x70) }

func (c *Expr) TableInit(elem, table uint32) { c.op(0xfc, 0x0c, elem, table) }
func (c *Expr) TableGrow(table uint32)       { c.op(0xfc, 0xf, table) }
func (c *Expr) TableGet(table uint32)        { c.op(0x25, table) }

func (c *Expr) I32Load(align, offset uint32)  { c.op(0x28, align, offset) }
func (c *Expr) I32Store(align, offset uint32) { c.op(0x36, align, offset) }
func (c *Expr) MemGrow()                      { c.op(0x40, 0) }

func (c *Expr) I32Const(x int32) { c.sop(0x41, x) }
func (c *Expr) I32Eqz()          { c.op(0x45) }
func (c *Expr) I32Eq()           { c.op(0x46) }
func (c *Expr) I32Ne()           { c.op(0x47) }
func (c *Expr) I32Lts()          { c.op(0x48) }
func (c *Expr) I32Gts()          { c.op(0x4a) }
func (c *Expr) I32Les()          { c.op(0x4c) }
func (c *Expr) I32Ges()          { c.op(0x4e) }
func (c *Expr) I32Add()          { c.op(0x6a) }
func (c *Expr) I32Sub()          { c.op(0x6b) }
func (c *Expr) I32Mul()          { c.op(0x6c) }
func (c *Expr) I32Div()          { c.op(0x6d) }
func (c *Expr) I32And()          { c.op(0x71) }
func (c *Expr) I32Or()           { c.op(0x72) }
func (c *Expr) I32Shl()          { c.op(0x74) }
func (c *Expr) I32Shr()          { c.op(0x76) }
