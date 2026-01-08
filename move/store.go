package move

// StoreSize is the maximal amount of moves the store can hold (across all frames).
const StoreSize = 2048

// Store is a stack based storage for chess moves. It is assumed that you want
// to allocate "frames" in the store in a stack like manner, pushing new
// frames, and popping from the top. Once a frame is pushed Alloc allows for
// allocating individual moves within the store.
type Store struct {
	data    []Weighted
	allocIx int
	frames  []frame
}

type frame struct {
	ix int
}

// NewStore allocates a new move store.
func NewStore() *Store {
	data := make([]Weighted, StoreSize)
	return &Store{data: data}
}

// Clear deletes everything in the store.
func (s *Store) Clear() {
	s.allocIx = 0
	s.frames = s.frames[:0]
}

// Alloc allocates a single move from the top frame. Push should be called first.
//
// This method panics if s has run out of space, and more than StoreSize moves
// have been allocated.
func (s *Store) Alloc(m Move) *Weighted {
	s.allocIx++
	ptr := &s.data[s.allocIx-1]
	*ptr = Weighted{Move: m} // reset to zero value
	return ptr
}

// Push allocates a new frame on the top of the store.
func (s *Store) Push() {
	s.frames = append(s.frames, frame{s.allocIx})
}

// Pop pops the last frame from the pop.
func (s *Store) Pop() {
	if len(s.frames) == 0 {
		s.allocIx = 0
		return
	}
	frame := s.frames[len(s.frames)-1]
	s.allocIx = frame.ix
	s.frames = s.frames[:len(s.frames)-1]
}

// Frame returns the top frame of the store.
func (s *Store) Frame() []Weighted {
	start := 0
	if len(s.frames) != 0 {
		frame := s.frames[len(s.frames)-1]
		start = frame.ix
	}
	return s.data[start:s.allocIx]
}
