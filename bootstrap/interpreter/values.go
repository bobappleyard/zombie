package main

import (
	"fmt"
	"reflect"

	"github.com/bobappleyard/zombie/internal/sexpr"
)

type procedure interface {
	apply(p *process)
}

type lambda struct {
	scope *scope
	vars  []string
	body  sexpr.Expr
}

func (l *lambda) apply(p *process) {
	s := l.scope.capture()
	if p.argc != len(l.vars) {
		p.err = ErrWrongArgCount
		return
	}
	for i, a := range l.vars {
		s.defs[a] = p.args[i]
	}
	p.eval(s, l.body, true)
}

type builtinProcedure func(p *process)

func (f builtinProcedure) apply(p *process) {
	f(p)
}

var errType = reflect.TypeOf(new(error)).Elem()

func callBuiltin(fv reflect.Value, ft reflect.Type, p *process) []reflect.Value {
	defer func() {
		msg := recover()
		if msg == nil {
			return
		}
		if msg, ok := msg.(error); ok {
			p.err = msg
			return
		}
		p.err = fmt.Errorf("%s", msg)
	}()
	if p.argc != ft.NumIn() {
		p.fail(ErrWrongArgCount)
		return nil
	}
	in := make([]reflect.Value, ft.NumIn())
	for i := range in {
		in[i] = reflect.ValueOf(p.args[i])
	}
	return fv.Call(in)
}

func liftProcedure(f any) procedure {
	fv := reflect.ValueOf(f)
	ft := fv.Type()

	if ft.Kind() != reflect.Func {
		panic("unsupported type")
	}

	switch ft.NumOut() {
	case 0:
		return builtinProcedure(func(p *process) {
			callBuiltin(fv, ft, p)
			if p.err != nil {
				return
			}
			p.returnValue(nil)
		})

	case 1:
		if ft.Out(0) == errType {
			return builtinProcedure(func(p *process) {
				out := callBuiltin(fv, ft, p)
				if p.err != nil {
					return
				}
				if !out[0].IsNil() {
					p.err = out[0].Interface().(error)
				}
				p.returnValue(nil)
			})
		}
		return builtinProcedure(func(p *process) {
			out := callBuiltin(fv, ft, p)
			if p.err != nil {
				return
			}
			p.returnValue(out[0].Interface())
		})

	case 2:
		return builtinProcedure(func(p *process) {
			out := callBuiltin(fv, ft, p)
			if p.err != nil {
				return
			}
			if !out[1].IsNil() {
				p.err = out[1].Interface().(error)
			}
			p.returnValue(out[0].Interface())
		})
	}

	panic("unsupported type")
}

type StructType struct {
	name string
	size int
}

type Struct struct {
	of   *StructType
	data []any
}

type CellAccessor struct {
	of  *StructType
	idx int
}

func (x *Struct) String() string {
	return fmt.Sprintf("#<%s>", x.of.name)
}

func (x *CellAccessor) String() string {
	return "#<accessor>"
}

func (x *StructType) String() string {
	return "#<type>"
}
