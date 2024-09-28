package data

import (
	"testing"

	"github.com/bobappleyard/zombie/util/assert"
)

func TestMatrix_AddRow(t *testing.T) {
	var m SparseMatrix[int]

	assertLookup := func(row, col, expect int) {
		t.Helper()
		x, ok := m.LookupValue(row, col)
		assert.True(t, ok)
		assert.Equal(t, x, expect)
	}

	m.AddRow([]SparseMatrixElement[int]{{0, 0}, {1, 10}, {3, 30}})
	m.AddRow([]SparseMatrixElement[int]{{0, 0}, {2, 10}, {3, 30}})
	m.AddRow([]SparseMatrixElement[int]{{5, 50}, {4, 40}, {9, 60}})
	m.AddRow([]SparseMatrixElement[int]{{0, 0}})

	assertLookup(0, 1, 10)
	assertLookup(1, 2, 10)
	assertLookup(0, 3, 30)

	assert.Equal(t, m.LookupRow(0), []SparseMatrixElement[int]{{0, 0}, {1, 10}, {3, 30}})
	assert.Equal(t, m.LookupRow(2), []SparseMatrixElement[int]{{5, 50}, {4, 40}, {9, 60}})

	t.Log(m)
}

func TestMatrix_LookupValue(t *testing.T) {
	m := SparseMatrix[string]{
		entries: []matrixEntry[string]{{value: "hello"}},
		rows:    []matrixRow{{}},
	}

	v, ok := m.LookupValue(0, 0)
	assert.True(t, ok)
	assert.Equal(t, v, "hello")

	assertFail := func(row, col int) {
		t.Helper()
		_, ok := m.LookupValue(row, col)
		assert.False(t, ok)
	}

	assertFail(-1, -1)
	assertFail(1, 0)
	assertFail(0, 1)
}

func TestMatrix_LookupRow(t *testing.T) {
	m := SparseMatrix[string]{
		entries: []matrixEntry[string]{
			{value: "hello", row: 0, next: -1},
			{},
			{value: "goodbye", row: 1, next: -1},
			{value: "say", row: 1, next: 1},
		},
		rows: []matrixRow{
			{offset: 0, start: 0},
			{offset: 1, start: 2},
		},
	}

	assert.Equal(t, m.LookupRow(0), []SparseMatrixElement[string]{{Col: 0, Value: "hello"}})
	assert.Equal(t, m.LookupRow(1), []SparseMatrixElement[string]{{Col: 2, Value: "say"}, {Col: 1, Value: "goodbye"}})
	assert.Equal(t, m.LookupRow(-1), nil)
	assert.Equal(t, m.LookupRow(2), nil)
}
