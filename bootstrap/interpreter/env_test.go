package main

import (
	"testing"

	"github.com/bobappleyard/zombie/internal/assert"
)

type testLogger struct {
	log []string
}

func (l *testLogger) apply(p *process) {
	l.log = append(l.log, p.args[0].(string))
	p.returnValue(nil)
}

func TestToplevel(t *testing.T) {
	for _, test := range []struct {
		name string
		in   string
		out  []string
	}{
		{
			name: "CallLog",
			in:   `(log "hello")`,
			out:  []string{"hello"},
		},
		{
			name: "DefineVar",
			in: `
				(define x "hello")
				(log x)
			`,
			out: []string{"hello"},
		},
		{
			name: "DefineFunc",
			in: `
				(define f (lambda (x) (log x)))
				(f "hello")
			`,
			out: []string{"hello"},
		},
		{
			name: "Import",
			in: `
				(import test)
				(log test-var)
			`,
			out: []string{"hey"},
		},
		{
			name: "Set",
			in: `
				(define x "hello")
				(set! x "hey")
				(log x)
			`,
			out: []string{"hey"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			l := &testLogger{}
			e := &Env{
				pkgs: map[string]*Pkg{
					"test": {
						exports: []string{"test-var"},
						defs: map[string]any{
							"test-var": "hey",
						},
						init: true,
					},
				},
			}
			p := &Pkg{
				owner: e,
				defs: map[string]any{
					"log": l,
				},
			}
			err := p.evalFile([]byte(test.in))
			assert.Nil(t, err)
			assert.Equal(t, l.log, test.out)
		})
	}
}

func TestPackageFile(t *testing.T) {
	logger := &testLogger{}
	e := &Env{
		path: "testdata",
		pkgs: map[string]*Pkg{
			"zombie.test": &Pkg{
				exports: []string{"log"},
				init:    true,
				defs: map[string]any{
					"log": logger,
				},
			},
		},
	}
	p := &Pkg{
		owner: e,
		defs:  map[string]any{},
	}
	err := p.evalFile([]byte(`
	
		(import zombie.test)
		(import package)

		(greet log "hello")
	
	`))
	assert.Nil(t, err)
	assert.Equal(t, logger.log, []string{"hello"})
}
