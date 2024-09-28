package data

import (
	"iter"
	"slices"
)

type Set[T any] struct {
	items []T
	cmp   func(T, T) int
}

func NewSet[T any](cmp func(T, T) int) *Set[T] {
	return &Set[T]{cmp: cmp}
}

func (s *Set[T]) Size() int {
	return len(s.items)
}

func (s *Set[T]) Get(x T) (T, bool) {
	idx, ok := slices.BinarySearchFunc(s.items, x, s.cmp)
	if !ok {
		var zero T
		return zero, false
	}
	return s.items[idx], true
}

func (s *Set[T]) Put(x T) {
	idx, ok := slices.BinarySearchFunc(s.items, x, s.cmp)
	if !ok {
		var zero T
		s.items = append(s.items, zero)
		copy(s.items[idx+1:], s.items[idx:])
	}
	s.items[idx] = x
}

func (s *Set[T]) Delete(x T) {
	idx, ok := slices.BinarySearchFunc(s.items, x, s.cmp)
	if !ok {
		return
	}
	copy(s.items[idx:], s.items[idx+1:])
	s.items = s.items[:len(s.items)-1]
}

func (s *Set[T]) Items() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, x := range s.items {
			if !yield(x) {
				return
			}
		}
	}
}
