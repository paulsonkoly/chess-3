package eval

import (
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/stretchr/testify/assert"
)

type callCounter struct {
	counts [evalIDs]int
}

func (c *callCounter) inc(id evalID) {
	c.counts[id]++
}

func (c *callCounter) callsOnce(id evalID) bool {
	return c.counts[id] == 1
}

func (c *callCounter) onlyCalls(id evalID) bool {
	for i := range evalIDs {
		if i == id {
			continue
		}
		if c.counts[i] != 0 {
			return false
		}
	}
	return true
}

func (c *callCounter) reset() {
	for i := range evalIDs {
		c.counts[i] = 0
	}
}

type mockMaterialDispatch[T ScoreType] struct {
	*Eval[T]
	callCounter callCounter
}

func newMockMaterialDispatch[T ScoreType]() *mockMaterialDispatch[T] {
	m := &mockMaterialDispatch[T]{Eval: New[T]()}
	for i := range evalIDs {
		m.matFuncs[i] = func(*Eval[T], *board.Board, *CoeffSet[T]) T {
			m.callCounter.inc(i)
			return 0
		}
	}
	return m
}

func (m *mockMaterialDispatch[T]) run(b *board.Board) *callCounter {
	// reset for fresh run
	m.callCounter.reset()
	// we dont care about returned call of original
	_ = m.Score(b, &CoeffSet[T]{})
	// return the counters for assertion chaining
	return &m.callCounter
}

func TestMaterial(t *testing.T) {
	tests := []struct {
		fen  string
		want evalID
	}{
		{StartPosFEN, evalPositionalID},
	}

	for _, tt := range tests {
		t.Run(tt.fen, func(t *testing.T) {
			mockEval := newMockMaterialDispatch[Score]()
			b := Must(board.FromFEN(tt.fen))
			assert.True(t, mockEval.run(b).callsOnce(tt.want), "fen %s calls once evalId %v", tt.fen, tt.want)
			assert.True(t, mockEval.run(b).onlyCalls(tt.want), "fen %s only calls evalId %v", tt.fen, tt.want)
		})
	}
}
