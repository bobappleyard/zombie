package assert

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func Equal[T any](t testing.TB, got, expected T) bool {
	t.Helper()
	diff := cmp.Diff(expected, got,
		cmpopts.EquateErrors(),
		cmp.Exporter(func(t reflect.Type) bool { return true }))
	if diff == "" {
		return true
	}
	t.Error(diff)
	return false
}

func Nil(t testing.TB, got any) bool {
	t.Helper()
	if got == nil {
		return true
	}
	if reflect.ValueOf(got).IsNil() {
		return true
	}
	t.Errorf("got %v, expecting nil", got)
	return false
}

func True(t testing.TB, got bool) bool {
	t.Helper()
	return Equal(t, got, true)
}

func False(t testing.TB, got bool) bool {
	t.Helper()
	return Equal(t, got, false)
}
