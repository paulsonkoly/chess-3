// Package stack provides a generic stack implementation with maximal capacity
// of MaxPlies.
package stack

import (
	. "github.com/paulsonkoly/chess-3/chess"
)

// Stack[T] is a generic stack structure.
type Stack[T any] struct {
	data [MaxPlies]T
	sp   int
}

// New creates a new stack for type T.
func New[T any]() *Stack[T] {
	return &Stack[T]{}
}

// Reset clear out the stack.
func (s *Stack[T]) Reset() {
	s.sp = 0
}

// Push pushes one element to the stack.
func (s *Stack[T]) Push(v T) {
	s.data[s.sp] = v
	s.sp++
}

// Pop pops the last element from the stack.
func (s *Stack[T]) Pop() {
	s.sp--
}

// Top returns the nth element counting from the end. It returns ok false if
// there are no n elements in the stack.
func (s *Stack[T]) Top(n int) (T, bool) {
	if s.sp > n {
		return s.data[s.sp-n-1], true
	}
	return s.data[0], false
}
