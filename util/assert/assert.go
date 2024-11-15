package assert

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func Equal[T any](t testing.TB, got, expected T) {
	t.Helper()
	diff := cmp.Diff(expected, got, cmpopts.EquateErrors())
	if diff == "" {
		return
	}
	t.Error(diff)
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
