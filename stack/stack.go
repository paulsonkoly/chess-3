package stack

import (
	. "github.com/paulsonkoly/chess-3/chess"
)

type Stack[T any] struct {
	data [MaxPlies]T
	sp   int
}

func New[T any]() *Stack[T] {
	return &Stack[T]{}
}

func (s *Stack[T]) Reset() {
	s.sp = 0
}

func (s *Stack[T]) Push(v T) {
	s.data[s.sp] = v
	s.sp++
}

func (s *Stack[T]) Pop() {
	s.sp--
}

func (s *Stack[T]) Top(n int) (T, bool) {
	if s.sp > n {
		return s.data[s.sp-n-1], true
	}
	return s.data[0], false
}

func (s *Stack[T]) Size() int {
	return s.sp
}
