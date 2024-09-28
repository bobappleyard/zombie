package assert

import (
	"go/ast"
	"reflect"
	"strings"
	"testing"

	"github.com/kylelemons/godebug/diff"
)

func Equal[T any](t testing.TB, got, expected T) {
	t.Helper()
	if !reflect.DeepEqual(got, expected) {
		var a, b strings.Builder
		ast.Fprint(&a, nil, got, ast.NotNilFilter)
		ast.Fprint(&b, nil, expected, ast.NotNilFilter)

		t.Error("\n" + diff.Diff(b.String(), a.String()))
	}
}

func Nil(t testing.TB, got any) {
	t.Helper()
	if got == nil {
		return
	}
	if reflect.ValueOf(got).IsNil() {
		return
	}
	t.Errorf("got %v, expecting nil", got)
}

func True(t testing.TB, got bool) {
	t.Helper()
	Equal(t, got, true)
}

func False(t testing.TB, got bool) {
	t.Helper()
	Equal(t, got, false)
}
