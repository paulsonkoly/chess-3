package stack_test

import (
	"testing"

	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/stack"
	"github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {
	s := stack.New[int]()

	v, ok := s.Top(0)
	assert.Equal(t, 0, v, "top of empty stack")
	assert.False(t, ok, "top of empty stack")
	assert.Panics(t, s.Pop, "pop of empty stack")

	s.Push(3)
	v, ok = s.Top(0)
	assert.Equal(t, 3, v, "top of stack <3>")
	assert.True(t, ok, "top of stack <3>")

	s.Push(7)
	v, ok = s.Top(0)
	assert.Equal(t, 7, v, "top of stack <3, 7>")
	assert.True(t, ok, "top of stack <3, 7>")

	s.Pop()
	v, ok = s.Top(0)
	assert.Equal(t, 3, v, "top of stack <3>")
	assert.True(t, ok, "top of stack <3>")

	s.Reset()
	v, ok = s.Top(0)
	assert.Equal(t, 0, v, "top of empty stack")
	assert.False(t, ok, "top of empty stack")

	for i := range MaxPlies {
		s.Push(i)
	}
	assert.Panics(t, func() { s.Push(0) }, "overflow")
}
