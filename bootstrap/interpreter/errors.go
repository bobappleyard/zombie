package main

import (
	"errors"
	"fmt"

	"github.com/bobappleyard/zombie/internal/sexpr"
)

var (
	ErrBadSyntax           = errors.New("syntax error")
	ErrUnboundVar          = errors.New("unbound variable")
	ErrWrongArgCount       = errors.New("wrong number of arguments")
	ErrWrongType           = errors.New("wrong type")
	ErrCircularImport      = errors.New("circular import")
	ErrInvalidCellAccessor = errors.New("invalid cell accessor")
)

type WithTrace struct {
	err  error
	file string
	line int
}

func (e *WithTrace) Error() string {
	return fmt.Sprintf("%s\n\tat %s line %d", e.err, e.file, e.line)
}

func attachTrace(err *error, file string, expr sexpr.Expr) {
	if *err != nil {
		line := expr.Line()
		if e, ok := (*err).(*WithTrace); ok {
			if e.file == file && e.line == line {
				return
			}
		}
		*err = &WithTrace{
			err:  *err,
			file: file,
			line: line,
		}
	}
}
