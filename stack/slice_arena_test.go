package stack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPushPopBasic(t *testing.T) {
	s := NewSliceArena[int]()

	f1 := s.Push()
	*f1 = append(*f1, 1, 2, 3)

	assert.Equal(t, []int{1, 2, 3}, *f1)

	f2 := s.Push()
	*f2 = append(*f2, 4, 5)

	assert.Equal(t, []int{4, 5}, *f2)

	s.Pop()

	assert.Len(t, s.frames, 1)
	assert.Equal(t, []int{1, 2, 3}, s.frames[0])
}

func TestFrameIsolation(t *testing.T) {
	s := NewSliceArena[int]()

	f1 := s.Push()
	*f1 = append(*f1, 1, 2, 3)

	f2 := s.Push()
	*f2 = append(*f2, 4, 5, 6)

	// modify f2 and ensure f1 is unchanged
	(*f2)[0] = 99

	assert.Equal(t, 1, (*f1)[0])
}

func TestRecursiveUsage(t *testing.T) {
	s := NewSliceArena[int]()

	var rec func(d int)
	rec = func(d int) {
		if d == 0 {
			return
		}

		frame := s.Push()
		defer s.Pop()

		*frame = append(*frame, d)
		*frame = append(*frame, d+1)
		*frame = append(*frame, d+2)

		rec(d - 1)
	}

	rec(10)

	assert.Empty(t, s.frames)
}

func TestNoUnexpectedReallocation(t *testing.T) {
	s := NewSliceArena[int]()

	f1 := s.Push()
	*f1 = append(*f1, 1, 2, 3, 4, 5)

	// capture backing array address
	base := &(*f1)[0]

	f2 := s.Push()
	*f2 = append(*f2, 6)

	assert.Equal(t, base, &(*f1)[0])
}
