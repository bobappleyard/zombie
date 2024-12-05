package main

import (
	"io"
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
	defs    map[string]any
	init    bool
	exports []string
}

func newEnv(path string) *Env {
	return &Env{
		path: path,
		pkgs: map[string]*Pkg{},
	}
}

func (p *Pkg) Eval(e sexpr.Expr) error {
	if e.Kind() == sexpr.List && e.Head().Kind() == sexpr.Symbol {
		switch e.Head().UnsafeText() {
		case "import":
			return p.evalImport(e.Tail())

		case "export":
			return p.evalExport(e.Tail())

		case "define":
			return p.evalDefine(e.Tail())
		}
	}
	_, err := p.evalExpr(e)
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
		owner: e,
		defs:  map[string]any{},
	}
	e.pkgs[name] = p
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
		p.defs[v] = q.defs[v]
	}
	return nil
}

func (p *Pkg) evalFile(src []byte) error {
	for {
		expr, rest, err := sexpr.Read(src)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		err = p.Eval(expr)
		if err != nil {
			return err
		}
		src = rest
	}
	return nil
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
	return x, ok
}

func (p *Pkg) setVar(name string, value any) bool {
	if _, ok := p.defs[name]; !ok {
		return false
	}
	p.defs[name] = value
	return true
}
