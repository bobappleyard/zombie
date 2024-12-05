package main

import (
	"io"
	"os"
	"testing"

	"github.com/bobappleyard/zombie/internal/assert"
)

func TestBuiltins(t *testing.T) {
	for _, test := range []struct {
		name string
		in   string
		out  string
	}{
		{
			name: "Print",
			in: `
				(print "hello")
			`,
			out: "hello\n",
		},
		{
			name: "Add",
			in: `
				(print (+ 1 1))
			`,
			out: "2\n",
		},
		{
			name: "IsNumber",
			in: `
				(print (number? 1))
			`,
			out: "true\n",
		},
		{
			name: "Arithmetic",
			in: `
				(print (- (/ (+ 1 (* 7 2)) 5) 1))
			`,
			out: "2\n",
		},
		{
			name: "Pair",
			in: `
				(define %pair (make-type "pair" 2))
				(define %head (cell-accessor %pair 0))
				(define %tail (cell-accessor %pair 1))
				(define make-pair (lambda (a d)
									(let ((p (make %pair)))
									(begin
										(cell-set! %head p a)
										(cell-set! %tail p d)
										p))))
				(define pair? (lambda (x) (is? %pair x)))
				(define head (lambda (x) (cell-ref %head x)))
				(define tail (lambda (x) (cell-ref %tail x)))

				(print (head (make-pair 1 2)))
			`,
			out: "1\n",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			e := &Env{
				path: "testdata",
				pkgs: map[string]*Pkg{},
			}
			registerBuiltins(e)
			out, err := captureOutput(func() error {
				p := &Pkg{
					owner: e,
					defs:  map[string]any{},
				}
				p.Import("zombie.internal.builtins")
				return p.evalFile([]byte(test.in))
			})
			assert.Nil(t, err)
			assert.Equal(t, out, test.out)
		})
	}
}

func captureOutput(f func() error) (string, error) {
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = orig }()
	err := f()
	w.Close()
	out, _ := io.ReadAll(r)
	return string(out), err
}
