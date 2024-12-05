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

type node struct {
	kind       Kind
	start, end int
}

func (e Expr) Kind() Kind {
	if e.pos != 0 {
		return List
	}
	return e.node().kind
}

func (e Expr) Position() int {
	return e.structure[e.pos].start
}

func (e Expr) Text() string {
	e.needsAtom()
	node := e.node()

	return string(e.text[node.start:node.end])
}

func (e Expr) UnsafeText() string {
	e.needsAtom()
	node := e.node()

	return unsafe.String(&e.text[node.start], node.end-node.start)
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
	return e.node().end-max(e.pos, 1) == 0
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
