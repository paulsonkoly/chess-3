// Package stack provides different stack based storages.
//
// - [Stack] generic stack storage.
// - [SliceArena] storage useful for movegens.
package stack

import . "github.com/paulsonkoly/chess-3/types"

// Stack is a generic stack with a maximal depth of MaxPlies.
type Stack[T any] struct {
	data []T
	ix   int
}

// New creates a new stack.
func New[T any]() Stack[T] {
	return Stack[T]{data: make([]T, MaxPlies)}
}

// Push pushes v on s. If there are MaxPlies number of consecutive pushes it
// panics.
func (s *Stack[T]) Push(v T) {
	if s.ix >= MaxPlies {
		panic("stack overflow")
	}
	s.data[s.ix] = v
	s.ix++
}

// Pop pops the last push off the stack. If there is no last push it panics.
func (s *Stack[T]) Pop() {
	if s.ix <= 0 {
		panic("stack underflow")
	}
	s.ix--
}

// Top is the top n elements pushed to the stack. If there are no n elements in
// the stack it returns less. The returned slice references the stack store.
func (s Stack[T]) Top(n int) []T {
	start := max(0, s.ix-n)
	return s.data[start:s.ix]
}

// Reset resets the stack to be empty.
func (s *Stack[T]) Reset() {
	s.ix = 0
}
