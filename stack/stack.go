// Package stack provides different stack based storages.
//
// - [SliceArena] storage useful for movegens.
package stack

import . "github.com/paulsonkoly/chess-3/types"

// Stack is a generic stack.
type Stack[T any] struct {
	data []T
	ix   int
}

func New[T any]() Stack[T] {
	return Stack[T]{data: make([]T, MaxPlies)}
}

func (s *Stack[T]) Push(v T) {
	if s.ix >= MaxPlies {
		panic("stack overflow")
	}
	s.data[s.ix] = v
	s.ix++
}

func (s *Stack[T]) Pop() {
	if s.ix <= 0 {
		panic("stack underflow")
	}
	s.ix--
}

func (s Stack[T]) Top(n int) []T {
	start := max(0, s.ix-n)
	return s.data[start:s.ix]
}

func (s *Stack[T]) Reset() {
	s.ix = 0
}
