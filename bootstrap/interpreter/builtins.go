package main

import "fmt"

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

		"make-type": liftProcedure(func(name string, size int) *CustomType {
			return &CustomType{
				name: name,
				size: size,
			}
		}),
		"is?": liftProcedure(func(t *CustomType, x any) bool {
			if x, ok := x.(*CustomObject); ok {
				return x.of == t
			}
			return false
		}),
		"cell-accessor": liftProcedure(func(t *CustomType, idx int) (*CellAccessor, error) {
			if idx < 0 || idx >= t.size {
				return nil, ErrInvalidCellAccessor
			}
			return &CellAccessor{
				of:  t,
				idx: idx,
			}, nil
		}),
		"make": liftProcedure(func(t *CustomType) *CustomObject {
			return &CustomObject{of: t, data: make([]any, t.size)}
		}),
		"cell-ref": liftProcedure(func(a *CellAccessor, x *CustomObject) (any, error) {
			if x.of != a.of {
				return nil, ErrWrongType
			}
			return x.data[a.idx], nil
		}),
		"cell-set!": liftProcedure(func(a *CellAccessor, x *CustomObject, value any) error {
			if x.of != a.of {
				return ErrWrongType
			}
			x.data[a.idx] = value
			return nil
		}),
	})
}

func testType[T any](x any) bool {
	_, ok := x.(T)
	return ok
}
