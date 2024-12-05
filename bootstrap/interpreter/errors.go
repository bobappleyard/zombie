package main

import "errors"

var (
	ErrBadSyntax           = errors.New("syntax error")
	ErrUnboundVar          = errors.New("unbound variable")
	ErrWrongArgCount       = errors.New("wrong number of arguments")
	ErrWrongType           = errors.New("wrong type")
	ErrCircularImport      = errors.New("circular import")
	ErrInvalidCellAccessor = errors.New("invalid cell accessor")
)
