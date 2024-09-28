package sexpr

import (
	"io"
	"testing"

	"github.com/bobappleyard/zombie/util/assert"
)

func TestParser(t *testing.T) {
	for _, test := range []struct {
		name string
		in   string
		out  []node
	}{
		{
			"Digit",
			"1",
			[]node{{kind: Number, start: 0, end: 1}},
		},
		{
			"Int",
			"123",
			[]node{{kind: Number, start: 0, end: 3}},
		},
		{
			"Real",
			"123.456",
			[]node{{kind: Number, start: 0, end: 7}},
		},
		{
			"Symbol",
			"+",
			[]node{{kind: Symbol, start: 0, end: 1}},
		},
		{
			"NumSymbol",
			"1+",
			[]node{{kind: Symbol, start: 0, end: 2}},
		},
		{
			"String",
			"\"hello\"",
			[]node{{kind: String, start: 0, end: 7}},
		},
		{
			"EscapeString",
			"\"hell\\no\"",
			[]node{{kind: String, start: 0, end: 9}},
		},
		{
			"Empty",
			"()",
			[]node{
				{kind: List, end: 1},
			},
		},
		{
			"List",
			"(a 16 b)",
			[]node{
				{kind: List, end: 4},
				{kind: Symbol, start: 1, end: 2},
				{kind: Number, start: 3, end: 5},
				{kind: Symbol, start: 6, end: 7},
			},
		},
		{
			"Tree",
			"(a 16 (b))",
			[]node{
				{kind: List, end: 5},
				{kind: Symbol, start: 1, end: 2},
				{kind: Number, start: 3, end: 5},
				{kind: List, start: 6, end: 2},
				{kind: Symbol, start: 7, end: 8},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			expr, rest, err := Read([]byte(test.in))
			assert.Nil(t, err)
			assert.Equal(t, rest, []byte{})
			t.Log(expr.structure)
			assert.Equal(t, expr.structure, test.out)
		})
	}
}

func TestParserErrors(t *testing.T) {
	for _, test := range []struct {
		name string
		in   string
		err  error
	}{
		{
			"Empty",
			"",
			io.EOF,
		},
		{
			"ListEOF",
			"(",
			io.ErrUnexpectedEOF,
		},
		{
			"NumberDoubleDot",
			"12..3",
			ErrNumberSyntaxError,
		},
		{
			"NumberTrailingDot",
			"12.",
			io.ErrUnexpectedEOF,
		},
		{
			"UnterminatedString",
			"\"abc",
			io.ErrUnexpectedEOF,
		},
		{
			"StringWithNewline",
			"\"a\n",
			ErrStringSyntaxError,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := Read([]byte(test.in))
			assert.Equal(t, err, test.err)
		})
	}
}

func TestParseIter(t *testing.T) {
	src := `
	
		(define (for ls f)
		  (loop (break)
		    (if (null? ls)
			  (break void)
			  (begin
			    (define x (head ls))
			    (set! ls (tail ls))
				(f x)))))
	
	`

	expr, _, err := Read([]byte(src))
	assert.Nil(t, err)
	assert.Equal(t, expr.Kind(), List)

}
