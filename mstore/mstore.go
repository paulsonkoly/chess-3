package mstore

import "github.com/paulsonkoly/chess-3/move"

const StoreSize = 2048

type MStore struct {
	data    []move.Move
	allocIx int
	frames  []frame
}

type frame struct {
	ix int
}

func New() *MStore {
	data := make([]move.Move, StoreSize)
	return &MStore{data: data}
}

func (m *MStore) Alloc() *move.Move {
	defer func() { m.allocIx++ }()
	return &m.data[m.allocIx]
}

func (m *MStore) Push() {
	m.frames = append(m.frames, frame{m.allocIx})
}

func (m *MStore) Pop() {
	if len(m.frames) == 0 {
		m.allocIx = 0
		return
	}
	frame := m.frames[len(m.frames)-1]
	m.allocIx = frame.ix
	m.frames = m.frames[:len(m.frames)-1]
}

func (m *MStore) Frame() []move.Move {
	start := 0
	if len(m.frames) != 0 {
		frame := m.frames[len(m.frames)-1]
		start = frame.ix
	}
	return m.data[start:m.allocIx]
}
