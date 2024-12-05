package sexpr

type Builder struct {
	text      []byte
	structure []node
	stack     []int
}

func (b *Builder) Expr() Expr {
	return Expr{text: b.text, structure: b.structure}
}

func (b *Builder) String(value string) {
	b.atom(String, value)
}

func (b *Builder) Number(value string) {
	b.atom(Number, value)
}

func (b *Builder) Symbol(name string) {
	b.atom(Symbol, name)
}

func (b *Builder) atom(kind Kind, value string) {
	start := len(b.text)
	b.text = append(b.text, []byte(value)...)
	b.structure = append(b.structure, node{
		kind:  kind,
		start: start,
		end:   len(b.text),
	})
}

func (b *Builder) ListStart() {
	b.stack = append(b.stack, len(b.structure))
	b.structure = append(b.structure, node{
		kind:  List,
		start: len(b.text),
	})
}

func (b *Builder) ListEnd() {
	start := b.stack[len(b.stack)-1]
	b.stack = b.stack[:len(b.stack)-1]
	b.structure[start].end = len(b.structure) - start
}

func (b *Builder) Copy(e Expr) {
	if e.Kind() == List {
		b.ListStart()
		for p := e; !p.Empty(); p = p.Tail() {
			b.Copy(p.Head())
		}
		b.ListEnd()
		return
	}
	b.atom(e.Kind(), e.UnsafeText())
}
