package sexpr

import (
	"iter"
	"unsafe"
)

type Expr struct {
	text      []byte
	structure []node
	pos       int
}

type Kind int

const (
	Symbol Kind = iota
	String
	Number
	List
)

type Builder struct {
	text      []byte
	structure []node
	stack     []int
}

type node struct {
	kind       Kind
	start, end int
}

func (e Expr) Kind() Kind {
	if e.pos != 0 {
		return List
	}
	return e.structure[0].kind
}

func (e Expr) Position() int {
	return e.structure[0].start
}

func (e Expr) Text() string {
	e.needsAtom()
	node := e.structure[0]

	return string(e.text[node.start:node.end])
}

func (e Expr) UnsafeText() string {
	e.needsAtom()
	node := e.structure[0]

	return unsafe.String(&e.text[node.start], node.end-node.start)
}

func (e Expr) Items() iter.Seq2[int, Expr] {
	e.needsList()
	return func(yield func(int, Expr) bool) {
		off := 0
		for p := e; !p.Empty(); p = p.Tail() {
			if !yield(off, p.Head()) {
				return
			}
			off++
		}
	}
}

func (e Expr) Empty() bool {
	e.needsList()
	return e.structure[0].end-max(e.pos, 1) == 0
}

func (e Expr) Head() Expr {
	e.needsList()
	p := max(e.pos, 1)
	return Expr{
		text:      e.text,
		structure: e.structure[p:],
	}
}

func (e Expr) Tail() Expr {
	e.needsList()
	p := max(e.pos, 1)
	node := e.structure[p]
	if node.kind == List {
		p += node.end
	} else {
		p++
	}
	return Expr{
		text:      e.text,
		structure: e.structure,
		pos:       p,
	}
}

func (e Expr) Bind(parts ...*Expr) bool {
	e.needsList()
	i := 0
	for _, p := range e.Items() {
		if i >= len(parts) {
			return false
		}
		if parts[i] == nil {
			i++
			continue
		}
		*parts[i] = p
		i++
	}
	return i == len(parts)
}

func (e Expr) needsList() {
	if e.Kind() != List {
		panic("bad kind")
	}
}

func (e Expr) needsAtom() {
	if e.Kind() == List {
		panic("bad kind")
	}
}

func (b *Builder) Expr() Expr {
	return Expr{text: b.text, structure: b.structure}
}

func (b *Builder) Atom(kind Kind, value string) {
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
	b.structure = append(b.structure, node{kind: List})
}

func (b *Builder) ListEnd() {
	start := b.stack[len(b.stack)-1]
	b.stack = b.stack[:len(b.stack)-1]
	b.structure[start].end = len(b.structure) - start
}

func (b *Builder) Copy(e Expr) {
	if e.Kind() != List {
		b.Atom(e.Kind(), e.UnsafeText())
		return
	}
	b.ListStart()
	for p := e; !p.Empty(); p = p.Tail() {
		b.Copy(p.Head())
	}
	b.ListEnd()
}
