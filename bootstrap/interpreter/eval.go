package main

import (
	"fmt"
	"strconv"

	"github.com/bobappleyard/zombie/internal/sexpr"
)

type process struct {
	value any
	err   error
	f     procedure
	argc  int
	args  [256]any
}

type scope struct {
	parent *scope
	pkg    *Pkg
	defs   map[string]any
}

func (s *scope) capture() *scope {
	return &scope{
		parent: s,
		pkg:    s.pkg,
		defs:   map[string]any{},
	}
}

func (s *scope) getVar(name string) (any, bool) {
	if v, ok := s.defs[name]; ok {
		return v, true
	}
	if s.parent == nil {
		return s.pkg.getVar(name)
	}
	return s.parent.getVar(name)
}

func (s *scope) setVar(name string, value any) bool {
	if _, ok := s.defs[name]; ok {
		s.defs[name] = value
		return true
	}
	if s.parent == nil {
		return s.pkg.setVar(name, value)
	}
	return s.parent.setVar(name, value)
}

func (p *process) call(f procedure, args []any, tail bool) {
	p.f = f
	copy(p.args[:], args)
	p.argc = len(args)
	p.value = nil
	p.err = nil
	if tail {
		return
	}
	for p.f != nil && p.err == nil {
		f = p.f
		p.f = nil
		f.apply(p)
	}
}

func (p *process) returnValue(x any) {
	p.value = x
}

func (p *process) fail(e error) {
	p.err = e
}

func (p *process) eval(s *scope, e sexpr.Expr, tail bool) {
	defer attachTrace(&p.err, s.pkg.path, e)
	switch e.Kind() {
	case sexpr.Number:
		x, err := strconv.Atoi(e.UnsafeText())
		p.value = x
		p.err = err
		return

	case sexpr.String:
		x, err := strconv.Unquote(e.UnsafeText())
		p.value = x
		p.err = err
		return

	case sexpr.Symbol:
		x, ok := s.getVar(e.UnsafeText())
		if !ok {
			p.err = fmt.Errorf("%s: %w", e.UnsafeText(), ErrUnboundVar)
			return
		}
		p.value = x
		return
	}
	if e.Empty() {
		p.err = ErrBadSyntax
		return
	}
	if e.Head().Kind() == sexpr.Symbol {
		switch e.Head().UnsafeText() {
		case "if":
			p.evalIf(s, e.Tail(), tail)
			return

		case "begin":
			p.evalBegin(s, e.Tail(), tail)
			return

		case "let":
			p.evalLet(s, e.Tail(), tail)
			return

		case "set!":
			p.evalSet(s, e.Tail())
			return

		case "lambda":
			p.evalLambda(s, e.Tail(), tail)
			return
		}
	}
	p.evalCall(s, e, tail)
}

func (p *process) evalIf(s *scope, e sexpr.Expr, tail bool) {
	var cond, ifTrue, ifFalse sexpr.Expr
	if !e.Bind(&cond, &ifTrue, &ifFalse) {
		p.err = ErrBadSyntax
		return
	}
	p.eval(s, cond, false)
	if p.value == false {
		p.eval(s, ifFalse, tail)
		return
	}
	p.eval(s, ifTrue, tail)
}

func (p *process) evalBegin(s *scope, e sexpr.Expr, tail bool) {
	if e.Empty() {
		p.value = nil
	}
	for !e.Tail().Empty() {
		p.eval(s, e.Head(), false)
		if p.err != nil {
			return
		}
		e = e.Tail()
	}
	p.eval(s, e.Head(), tail)
}

func (p *process) evalLet(s *scope, e sexpr.Expr, tail bool) {
	var bdgs, body sexpr.Expr
	if !e.Bind(&bdgs, &body) {
		p.err = ErrBadSyntax
		return
	}
	inner := s.capture()
	for _, bdg := range bdgs.All() {
		var n, v sexpr.Expr
		if !bdg.Bind(&n, &v) || n.Kind() != sexpr.Symbol {
			p.err = ErrBadSyntax
			return
		}
		p.eval(s, v, false)
		if p.err != nil {
			return
		}
		inner.defs[n.UnsafeText()] = p.value
	}
	p.eval(inner, body, tail)
}

func (p *process) evalSet(s *scope, e sexpr.Expr) {
	var n, v sexpr.Expr
	if !e.Bind(&n, &v) || n.Kind() != sexpr.Symbol {
		p.err = ErrBadSyntax
		return
	}
	p.eval(s, v, false)
	if !s.setVar(n.Text(), p.value) {
		p.err = fmt.Errorf("%s: %w", e.UnsafeText(), ErrUnboundVar)
	}
}

func (p *process) evalLambda(s *scope, e sexpr.Expr, tail bool) {
	var vs, body sexpr.Expr
	if !e.Bind(&vs, &body) || vs.Kind() != sexpr.List {
		p.err = ErrBadSyntax
		return
	}
	var vars []string
	for _, v := range vs.All() {
		if v.Kind() != sexpr.Symbol {
			p.err = ErrBadSyntax
			return
		}
		vars = append(vars, v.Text())
	}
	p.value = &lambda{
		scope: s,
		vars:  vars,
		body:  body,
	}
}

func (p *process) evalCall(s *scope, e sexpr.Expr, tail bool) {
	p.eval(s, e.Head(), false)
	if p.err != nil {
		return
	}
	f := p.value.(procedure)
	var args []any
	for _, e := range e.Tail().All() {
		p.eval(s, e, false)
		if p.err != nil {
			return
		}
		args = append(args, p.value)
	}
	p.call(f, args, tail)
}
