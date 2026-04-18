package main

import (
	"fmt"
	"io"
	"iter"
	"maps"
	"os"
	"path"
	"slices"

	"github.com/bobappleyard/zombie/internal/sexpr"
)

type Env struct {
	path string
	pkgs map[string]*Pkg
}

type Pkg struct {
	owner   *Env
	path    string
	defs    map[string]any
	syntax  map[string]syntaxRules
	init    bool
	exports []string
}

func newEnv(path string) *Env {
	return &Env{
		path: path,
		pkgs: map[string]*Pkg{},
	}
}

func (p *Pkg) Eval(e sexpr.Expr) (err error) {
	defer attachTrace(&err, p.path, e)
	if e.Kind() == sexpr.List && e.Head().Kind() == sexpr.Symbol {
		switch e.Head().UnsafeText() {
		case "import":
			return p.evalImport(e.Tail())

		case "export":
			return p.evalExport(e.Tail())

		case "define":
			return p.evalDefine(e.Tail())

		case "define-syntax":
			return p.evalDefineSyntax(e.Tail())
		}
	}
	_, err = p.evalExpr(e)
	return err
}

func (e *Env) Import(name string) (*Pkg, error) {
	if p, ok := e.pkgs[name]; ok {
		if !p.init {
			return nil, ErrCircularImport
		}
		return p, nil
	}
	src, err := os.ReadFile(path.Join(e.path, name+".zl"))
	if err != nil {
		return nil, err
	}
	p := &Pkg{
		path:  name,
		owner: e,
		defs:  map[string]any{},
	}
	e.pkgs[name] = p
	if name != "zombie" {
		err = p.Import("zombie")
		if err != nil {
			return nil, err
		}
	}
	err = p.evalFile(src)
	if err != nil {
		return nil, err
	}
	p.init = true
	return p, nil
}

func (e *Env) Define(name string, defs map[string]any) {
	e.pkgs[name] = &Pkg{
		init:    true,
		exports: slices.Collect(maps.Keys(defs)),
		defs:    defs,
	}
}

func (p *Pkg) Import(name string) error {
	q, err := p.owner.Import(name)
	if err != nil {
		return err
	}
	for _, v := range q.exports {
		if d, ok := q.syntax[v]; ok {
			p.syntax[v] = d
			continue
		}
		if d, ok := q.defs[v]; ok {
			p.defs[v] = d
			continue
		}
		return fmt.Errorf("%s: %w", v, ErrUnboundVar)
	}
	return nil
}

func (p *Pkg) evalFile(src []byte) error {
	var pos int
	for {
		expr, next, err := sexpr.Read(src, pos)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		expr, _, err = p.expand(expr)
		err = p.Eval(expr)
		if err != nil {
			return err
		}
		pos = next
	}
	return nil
}

func (p *Pkg) expand(expr sexpr.Expr) (sexpr.Expr, bool, error) {
	var hasChanged bool
	for {
		if expr.Kind() != sexpr.List || expr.Empty() {
			break
		}
		if s, ok := p.syntax[expr.Head().UnsafeText()]; ok {
			next, ch, err := s.expand(expr)
			if err != nil {
				return sexpr.Expr{}, false, err
			}
			hasChanged = hasChanged || ch
			expr = next
			continue
		}
		var changed bool
		var next []sexpr.Expr
		for _, e := range expr.All() {
			n, ch, err := p.expand(e)
			if err != nil {
				return sexpr.Expr{}, false, err
			}
			changed = changed || ch
			next = append(next, n)
		}
		if !changed {
			break
		}
		hasChanged = hasChanged || changed
		expr = sexpr.ListOf(next...)
	}
	return expr, hasChanged, nil
}

func (p *Pkg) evalImport(e sexpr.Expr) error {
	var name sexpr.Expr
	if !e.Bind(&name) {
		return ErrBadSyntax
	}
	if name.Kind() != sexpr.Symbol {
		return ErrBadSyntax
	}
	return p.Import(name.Text())
}

func (p *Pkg) evalExport(e sexpr.Expr) error {
	for _, s := range e.All() {
		if s.Kind() != sexpr.Symbol {
			return ErrBadSyntax
		}
		p.exports = append(p.exports, s.Text())
	}
	return nil
}

func (p *Pkg) evalDefine(e sexpr.Expr) error {
	if e.Head().Kind() == sexpr.List {
		return p.evalDefineFunction(e)
	}
	var name, value sexpr.Expr
	if !e.Bind(&name, &value) {
		return ErrBadSyntax
	}
	if name.Kind() != sexpr.Symbol {
		return ErrBadSyntax
	}
	x, err := p.evalExpr(value)
	if err != nil {
		return err
	}
	p.defs[name.Text()] = x
	return nil
}

func (p *Pkg) evalDefineFunction(e sexpr.Expr) error {
	if e.Head().Empty() {
		return ErrBadSyntax
	}
	name := e.Head().Head()
	if name.Kind() != sexpr.Symbol {
		return ErrBadSyntax
	}
	var vars []string
	for _, v := range e.Head().Tail().All() {
		if v.Kind() != sexpr.Symbol {
			return ErrBadSyntax
		}
		vars = append(vars, v.Text())
	}
	var body sexpr.Expr
	if !e.Tail().Empty() && e.Tail().Tail().Empty() {
		body = e.Tail().Head()
	} else {
		var bb sexpr.Builder
		bb.ListStart()
		bb.Symbol("begin")
		for _, e := range e.Tail().All() {
			bb.Copy(e)
		}
		bb.ListEnd()
		body = bb.Expr()
	}
	p.defs[name.Text()] = &lambda{
		scope: &scope{pkg: p},
		vars:  vars,
		body:  body,
	}
	return nil
}

func (p *Pkg) evalDefineSyntax(expr sexpr.Expr) error {
	var name, rules sexpr.Expr
	if !expr.Bind(&name, &rules) {
		return ErrBadSyntax
	}
	if rules.Kind() != sexpr.List || rules.Head().UnsafeText() != "syntax-rules" {
		return ErrBadSyntax
	}
	if rules.Tail().Empty() || rules.Tail().Head().Kind() != sexpr.List {
		return ErrBadSyntax
	}
	kws := slices.Collect(iterRight(rules.Tail().Head().All()))
	rs := syntaxRules{
		kws: kws,
	}
	for _, r := range rules.Tail().Tail().All() {
		err := rs.addArm(r)
		if err != nil {
			return err
		}
	}
	p.syntax[name.Text()] = rs
	return nil
}

func (p *Pkg) evalExpr(e sexpr.Expr) (any, error) {
	v := &process{}
	s := &scope{
		pkg: p,
	}
	v.eval(s, e, false)
	return v.value, v.err
}

func (p *Pkg) getVar(name string) (any, bool) {
	x, ok := p.defs[name]
	if !ok {
		return nil, false
	}
	return x, ok
}

func (p *Pkg) setVar(name string, value any) bool {
	_, ok := p.defs[name]
	if !ok {
		return false
	}
	p.defs[name] = value
	return true
}

func iterRight[T, U any](s iter.Seq2[T, U]) iter.Seq[U] {
	return func(yield func(U) bool) {
		for _, x := range s {
			if !yield(x) {
				return
			}
		}
	}
}
