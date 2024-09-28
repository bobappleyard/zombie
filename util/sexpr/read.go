package sexpr

import (
	"errors"
	"io"
)

var (
	ErrUnmatchedParentheses = errors.New("unmatched parentheses")
	ErrNumberSyntaxError    = errors.New("number syntax error")
	ErrStringSyntaxError    = errors.New("string syntax error")
)

type reader struct {
	text      []byte
	structure []node
	pos       int
}

func Read(src []byte) (Expr, []byte, error) {
	p := reader{
		text: src,
		pos:  -1,
	}
	err := p.parse()
	if err != nil {
		return Expr{}, nil, err
	}
	res := Expr{
		text:      src[:p.pos+1],
		structure: p.structure,
	}
	return res, src[p.pos+1:], nil
}

func (p *reader) parse() error {
	for {
		p.pos++

		if p.pos >= len(p.text) {
			if len(p.structure) == 0 {
				return io.EOF
			}
			return io.ErrUnexpectedEOF
		}

		switch p.text[p.pos] {
		case ' ', '\n', '\t', '\r':
			// do nothing

		case '(':
			return p.parseList()

		case ')':
			return ErrUnmatchedParentheses

		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return p.parseNumber()

		case '"':
			return p.parseString()

		default:
			return p.parseSymbol(p.pos)
		}
	}
}

func (p *reader) parseList() error {
	nodePos := len(p.structure)
	p.structure = append(p.structure, node{kind: List, start: p.pos})

	for {
		err := p.parse()
		if err == ErrUnmatchedParentheses {
			p.structure[nodePos].end = len(p.structure) - nodePos
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func (p *reader) parseNumber() error {
	start := p.pos
	hasDigits := true
	integer := true

loop:
	for {
		p.pos++

		if p.pos >= len(p.text) {
			break loop
		}

		switch p.text[p.pos] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			hasDigits = true

		case '.':
			if !integer {
				return ErrNumberSyntaxError
			}
			integer = false
			hasDigits = false

		case ' ', '\n', '\t', '\r', '(', ')':
			break loop

		default:
			return p.parseSymbol(start)
		}
	}

	if !hasDigits {
		return io.ErrUnexpectedEOF
	}

	p.addNode(Number, start)
	return nil

}

func (p *reader) parseString() error {
	start := p.pos

	for {
		p.pos++

		if p.pos >= len(p.text) {
			return io.ErrUnexpectedEOF
		}

		switch p.text[p.pos] {
		case '\\':
			p.pos++

		case '\n':
			return ErrStringSyntaxError

		case '"':
			p.pos++
			p.addNode(String, start)
			return nil
		}
	}
}

func (p *reader) parseSymbol(start int) error {
loop:
	for {
		p.pos++

		if p.pos >= len(p.text) {
			break loop
		}

		switch p.text[p.pos] {
		case ' ', '\n', '\t', '\r', '(', ')':
			break loop
		}
	}

	p.addNode(Symbol, start)
	return nil
}

func (p *reader) addNode(kind Kind, start int) {
	p.structure = append(p.structure, node{
		kind:  kind,
		start: start,
		end:   p.pos,
	})
	p.pos--
}
