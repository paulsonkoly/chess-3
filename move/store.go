package move

const StoreSize = 2048

type Store struct {
	data    []Move
	allocIx int
	frames  []frame
}

type frame struct {
	ix int
}

func NewStore() *Store {
	data := make([]Move, StoreSize)
	return &Store{data: data}
}

func (s *Store) Clear() {
	s.allocIx = 0
	s.frames = s.frames[:0]
}

func (s *Store) Alloc() *Move {
	s.allocIx++
	return &s.data[s.allocIx-1]
}

func (s *Store) Push() {
	s.frames = append(s.frames, frame{s.allocIx})
}

func (s *Store) Pop() {
	if len(s.frames) == 0 {
		s.allocIx = 0
		return
	}
	frame := s.frames[len(s.frames)-1]
	s.allocIx = frame.ix
	s.frames = s.frames[:len(s.frames)-1]
}

func (s *Store) Frame() []Move {
	start := 0
	if len(s.frames) != 0 {
		frame := s.frames[len(s.frames)-1]
		start = frame.ix
	}
	return s.data[start:s.allocIx]
}
