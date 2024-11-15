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
			[]node{{Kind: Number, Start: 0, End: 1}},
		},
		{
			"Int",
			"123",
			[]node{{Kind: Number, Start: 0, End: 3}},
		},
		{
			"Real",
			"123.456",
			[]node{{Kind: Number, Start: 0, End: 7}},
		},
		{
			"Symbol",
			"+",
			[]node{{Kind: Symbol, Start: 0, End: 1}},
		},
		{
			"NumSymbol",
			"1+",
			[]node{{Kind: Symbol, Start: 0, End: 2}},
		},
		{
			"ScopedSymbol",
			"a:b",
			[]node{
				{Kind: Symbol + 1, Start: 2, End: 3},
			},
		},
		{
			"String",
			"\"hello\"",
			[]node{{Kind: String, Start: 0, End: 7}},
		},
		{
			"EscapeString",
			"\"hell\\no\"",
			[]node{{Kind: String, Start: 0, End: 9}},
		},
		{
			"Empty",
			"()",
			[]node{
				{Kind: List, End: 1},
			},
		},
		{
			"List",
			"(a 16 b)",
			[]node{
				{Kind: List, End: 4},
				{Kind: Symbol, Start: 1, End: 2},
				{Kind: Number, Start: 3, End: 5},
				{Kind: Symbol, Start: 6, End: 7},
			},
		},
		{
			"ScopedSymbolList",
			"(a:a a:b)",
			[]node{
				{Kind: List, End: 3},
				{Kind: Symbol + 1, Start: 3, End: 4},
				{Kind: Symbol + 1, Start: 7, End: 8},
			},
		},
		{
			"Tree",
			"(a 16 (b))",
			[]node{
				{Kind: List, End: 5},
				{Kind: Symbol, Start: 1, End: 2},
				{Kind: Number, Start: 3, End: 5},
				{Kind: List, Start: 6, End: 2},
				{Kind: Symbol, Start: 7, End: 8},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			expr, rest, err := Read([]byte(test.in))
			assert.Nil(t, err)
			assert.Equal(t, rest, []byte{})
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
		{
			"ScopeEOF",
			"a:",
			ErrScopeSyntaxError,
		},
		{
			"ScopeWhitespace",
			"a: ",
			ErrScopeSyntaxError,
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

func TestScopeParse(t *testing.T) {
	src := `(zombie.core:define zombie.internal:size zombie.internal.size)`

	expr, _, err := Read([]byte(src))
	assert.Nil(t, err)

	assert.Equal(t, expr.Kind(), List)
	assert.Equal(t, expr.Head().Kind(), Symbol)
	assert.Equal(t, expr.Head().UnsafeScope(), "zombie.core")
	assert.Equal(t, expr.Head().UnsafeText(), "define")

	expr = expr.Tail()
	assert.Equal(t, expr.Kind(), List)
	assert.Equal(t, expr.Head().Kind(), Symbol)
	assert.Equal(t, expr.Head().UnsafeScope(), "zombie.internal")
	assert.Equal(t, expr.Head().UnsafeText(), "size")

	expr = expr.Tail()
	assert.Equal(t, expr.Kind(), List)
	assert.Equal(t, expr.Head().Kind(), Symbol)
	assert.Equal(t, expr.Head().UnsafeScope(), "")
	assert.Equal(t, expr.Head().UnsafeText(), "zombie.internal.size")

}
