package main

import (
	"fmt"
	"os"
)

func registerBuiltins(e *Env) {
	e.Define("zombie.internal.builtins", map[string]any{
		"print": liftProcedure(func(x any) { fmt.Println(x) }),
		"nil":   nil,
		"eq?":   liftProcedure(func(a, b any) bool { return a == b }),

		"boolean?": liftProcedure(testType[bool]),
		"true":     true,
		"false":    false,

		"number?": liftProcedure(testType[int]),
		"+":       liftProcedure(func(a, b int) int { return a + b }),
		"-":       liftProcedure(func(a, b int) int { return a - b }),
		"*":       liftProcedure(func(a, b int) int { return a * b }),
		"/":       liftProcedure(func(a, b int) int { return a / b }),
		"mod":     liftProcedure(func(a, b int) int { return a % b }),
		"=":       liftProcedure(func(a, b int) bool { return a == b }),
		">":       liftProcedure(func(a, b int) bool { return a > b }),
		"<":       liftProcedure(func(a, b int) bool { return a < b }),

		"vector": builtinProcedure(func(p *process) {
			v := make([]any, p.argc)
			copy(v, p.args[:])
			p.returnValue(v)
		}),
		"make-vector": liftProcedure(func(n int, of any) []any {
			v := make([]any, n)
			for i := range v {
				v[i] = of
			}
			return v
		}),
		"vector?": liftProcedure(testType[[]any]),
		"vector-length": liftProcedure(func(xs []any) int {
			return len(xs)
		}),
		"vector-ref": liftProcedure(func(v []any, idx int) any {
			return v[idx]
		}),
		"vector-set!": liftProcedure(func(v []any, idx int, value any) {
			v[idx] = value
		}),

		"make-struct-type": liftProcedure(func(name string, size int) *StructType {
			return &StructType{
				name: name,
				size: size,
			}
		}),
		"make-struct": builtinProcedure(func(p *process) {
			if p.argc < 1 {
				p.fail(ErrWrongArgCount)
				return
			}
			t := p.args[0].(*StructType)
			if p.argc != t.size+1 {
				p.fail(ErrWrongArgCount)
				return
			}
			o := &Struct{
				of:   t,
				data: make([]any, t.size),
			}
			copy(o.data, p.args[1:])
			p.returnValue(o)
		}),
		"struct-is?": liftProcedure(func(t *StructType, x any) bool {
			o, ok := x.(*Struct)
			return ok && o.of == t
		}),
		"bind-struct": builtinProcedure(func(p *process) {
			if p.argc != 3 {
				p.fail(ErrWrongArgCount)
				return
			}
			t := p.args[0].(*StructType)
			o := p.args[1].(*Struct)
			f := p.args[2].(procedure)
			if o.of != t {
				p.fail(ErrWrongType)
				return
			}
			p.call(f, o.data, true)
		}),

		"buffer?": liftProcedure(testType[[]byte]),
		"make-buffer": liftProcedure(func(n int) []byte {
			return make([]byte, n)
		}),
		"buffer-segment": liftProcedure(func(buf []byte, start, end int) []byte {
			return buf[start:end]
		}),
		"buffer-length": liftProcedure(func(b []byte) int { return len(b) }),
		"buffer-ref": liftProcedure(func(buf []byte, idx int) int {
			return int(buf[idx])
		}),
		"buffer-set!": liftProcedure(func(buf []byte, idx int, value int) {
			buf[idx] = byte(value)
		}),
		"write-buffer!": liftProcedure(func(a, b []byte) {
			copy(a, b)
		}),
		"read-file":      liftProcedure(os.ReadFile),
		"string->buffer": liftProcedure(func(x string) []byte { return []byte(x) }),
		"buffer->string": liftProcedure(func(x []byte) string { return string(x) }),
	})
}

func testType[T any](x any) bool {
	_, ok := x.(T)
	return ok
}
