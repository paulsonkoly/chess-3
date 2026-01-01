package stack

import . "github.com/paulsonkoly/chess-3/chess"

// SliceArena is a stack frame based slice arena storage.
//
// # Motivation
//
// A recursive search function requires to allocate an array of moves or move
// weights per stack frame. While we can allocate it on the stack, avoiding
// escaping to the heap, we would need to allocate a certain size storage per
// stack frame. Dynamic sized storage cannot live on the stack. This means that
// we would either have to accept escaping to the heap or accept waste with
// overallocation per recursive call.
//
// The SliceArena type addresses this issue by providing a storage type, that
// has a single (heap allocated) backing array that holds all allocations from
// all recursive calls, packed without gaps.
//
// # Example
//
//  // a recursive function
//  func f(s *SliceArena[int], d Depth) {
//  	if d == 0 {
//  		return // exit from recursion
//  	}
//
//  	// frame points to an empty slice dedicated for this stack frame backed by s
//  	frame := s.Push() 
//  	defer s.Pop()
//  	*frame = append(*frame, int(d+1)) // adds numbers to the current frame
//  	*frame = append(*frame, int(d+2))
//  	*frame = append(*frame, int(d+3))
//  
//  	f(s, d-1)
//  }
//  
//  // allocate store
//  s = NewSliceArena[int]()
//  f(5)
//
// This is the storage layout after 2 recursive calls of f()
//
//   ix
//    0 : 11 // d == 10 frame 0 start
//    1 : 12 // d == 10
//    2 : 13 // d == 10 frame 0 end
//    3 : 10 // d == 9 frame 1 start
//    4 : 11 // d == 9
//    5 : 12 // d == 9 frame 1 end
//
// # Warning
//
// The following rules should be respected:
//
//   - The number of consecutive Push calls (recursion level) should not exceed MaxPlies.
//   - SliceArena relies on strictly nested Push/Pop usage. Frames must be popped in reverse order of Push.
//   - The number of appends to a single frame shouldn't exceed MaxMoves.
//
// These restrictions don't impose any practical limitation for a chess engine
// movegen.
//
// Following these rules
//   - avoids allocations altogether in Push / Pop
//   - avoids corruption on frame overrun
//   - avoids a panic when the returned frame would have less capacity than MaxMoves.
type SliceArena[T any] struct {
	data       []T
	frames     [MaxPlies][]T
	frameIndex int
}

// SliceArenaSize is the backing array size for slice arenas.
const SliceArenaSize = 2048

// NewSliceArena creates a new SliceArena.
func NewSliceArena[T any]() *SliceArena[T] {
	return &SliceArena[T]{data: make([]T, 0, SliceArenaSize)}
}

// Push allocates a new frame from s and returns an empty slice with at least
// MaxMoves capacity. The returned slice can accept appends up to MaxMoves
// items, however it should not be re-sliced. It should not be re-assigned
// apart from appends. It should not be used after a corresponding Pop call.
func (s *SliceArena[T]) Push() *[]T {
	if s.frameIndex >= MaxPlies {
		panic("max plies exceeded")
	}

	newFrame := s.data
	if s.frameIndex > 0 {
		newFrame = s.frames[s.frameIndex-1]
		newFrame = newFrame[len(newFrame):]
	}
	if cap(newFrame) < MaxMoves {
		panic("backing array overrun")
	}
	s.frames[s.frameIndex] = newFrame
	s.frameIndex++
	return &s.frames[s.frameIndex-1]
}

// Pop pops the last frame from s.
func (s *SliceArena[T]) Pop() {
	if s.frameIndex <= 0 {
		panic("underpop")
	}
	s.frameIndex--
}

// Clear removes any dangling allocations from s.
func (s *SliceArena[T]) Clear() {
	s.frameIndex = 0
}
