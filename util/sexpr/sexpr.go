package sexpr

import (
	"iter"
	"unsafe"
)

type Expr struct {
	text      []byte
	structure []node
	scopes    []node
	pos       int
}

type Kind int

const (
	List Kind = iota
	String
	Number
	Scope
	Symbol
)

type Builder struct {
	text      []byte
	structure []node
	stack     []int
}

type node struct {
	Kind       Kind
	Start, End int
}

func (e Expr) Kind() Kind {
	if e.pos != 0 {
		return List
	}
	k := e.node().Kind
	if k >= Symbol {
		return Symbol
	}
	return k
}

func (e Expr) Position() int {
	return e.structure[e.pos].Start
}

func (e Expr) Scope() string {
	e.needsSymbol()
	n := e.node()
	if n.Kind == Symbol {
		return ""
	}
	node := e.scopes[n.Kind-Symbol-1]

	return string(e.text[node.Start:node.End])
}

func (e Expr) Text() string {
	e.needsAtom()
	node := e.node()

	return string(e.text[node.Start:node.End])
}

func (e Expr) UnsafeScope() string {
	e.needsSymbol()
	n := e.node()
	if n.Kind == Symbol {
		return ""
	}
	node := e.scopes[n.Kind-Symbol-1]

	return unsafe.String(&e.text[node.Start], node.End-node.Start)
}

func (e Expr) UnsafeText() string {
	e.needsAtom()
	node := e.node()

	return unsafe.String(&e.text[node.Start], node.End-node.Start)
}

func (e Expr) All() iter.Seq2[int, Expr] {
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
	return e.node().End-max(e.pos, 1) == 0
}

func (e Expr) Head() Expr {
	e.needsList()
	p := max(e.pos, 1)
	return Expr{
		text:      e.text,
		structure: e.structure[p:],
		scopes:    e.scopes,
	}
}

func (e Expr) Tail() Expr {
	e.needsList()
	p := max(e.pos, 1)
	node := e.structure[p]
	if node.Kind == List {
		p += node.End
	} else {
		p++
	}
	return Expr{
		text:      e.text,
		structure: e.structure,
		scopes:    e.scopes,
		pos:       p,
	}
}

func (e Expr) Bind(parts ...*Expr) bool {
	e.needsList()
	var i int
	var p Expr
	for i, p = range e.All() {
		if i >= len(parts) {
			return false
		}
		if parts[i] == nil {
			continue
		}
		*parts[i] = p
	}
	return i+1 == len(parts)
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

func (e Expr) needsSymbol() {
	if e.Kind() != Symbol {
		panic("bad kind")
	}
}

func (e Expr) node() node {
	return e.structure[0]
}

func (b *Builder) Expr() Expr {
	return Expr{text: b.text, structure: b.structure}
}

func (b *Builder) Atom(kind Kind, value string) {
	start := len(b.text)
	b.text = append(b.text, []byte(value)...)
	b.structure = append(b.structure, node{
		Kind:  kind,
		Start: start,
		End:   len(b.text),
	})
}

func (b *Builder) ListStart() {
	b.stack = append(b.stack, len(b.structure))
	b.structure = append(b.structure, node{
		Kind:  List,
		Start: len(b.text),
	})
}

func (b *Builder) ListEnd() {
	start := b.stack[len(b.stack)-1]
	b.stack = b.stack[:len(b.stack)-1]
	b.structure[start].End = len(b.structure) - start
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
