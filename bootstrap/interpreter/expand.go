package main

import (
	"github.com/bobappleyard/zombie/internal/sexpr"
)

type syntaxRules struct {
	kws  []sexpr.Expr
	arms []syntaxRulesArm
}

type syntaxRulesArm struct {
	match, replace sexpr.Expr
}

type syntaxRulesMatch struct {
	items map[string]sexpr.Expr
}

func (s *syntaxRules) addArm(r sexpr.Expr) error {
	var match, replace sexpr.Expr
	if !r.Bind(&match, &replace) {
		return ErrBadSyntax
	}
	s.arms = append(s.arms, syntaxRulesArm{
		match:   match,
		replace: replace,
	})
	return nil
}

func (s syntaxRules) expand(expr sexpr.Expr) (sexpr.Expr, bool, error) {
	for _, a := range s.arms {
		m := a.matchExpr(expr)
		if m == nil {
			continue
		}
		return a.expandMatch(m)
	}
	return sexpr.Expr{}, false, ErrBadSyntax
}

func (s syntaxRulesArm) matchExpr(expr sexpr.Expr) *syntaxRulesMatch {
	if s.match.Kind() == sexpr.Symbol {
		return &syntaxRulesMatch{items: map[string]sexpr.Expr{
			s.match.Text(): expr,
		}}
	}
	return nil
}

func (s syntaxRulesArm) expandMatch(m *syntaxRulesMatch) (sexpr.Expr, bool, error) {
	return s.replace, true, nil
}
