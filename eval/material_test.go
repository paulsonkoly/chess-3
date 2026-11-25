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

type dispatchRecorder[T ScoreType] struct {
	*Eval[T]
	callCounter callCounter
}

func newDispatchRecorder[T ScoreType]() *dispatchRecorder[T] {
	m := &dispatchRecorder[T]{Eval: New[T]()}
	for i := range evalIDs {
		m.matFuncs[i] = func(*Eval[T], *board.Board, *CoeffSet[T]) T {
			m.callCounter.inc(i)
			return 0
		}
	}
	return m
}

func (m *dispatchRecorder[T]) run(b *board.Board) *callCounter {
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
		// evalInsufficientID
		{"4k3/8/8/8/8/8/8/4K3 w - - 0 1", evalInsufficientID},
		{"4k3/8/8/8/8/8/8/1N2K3 w - - 0 1", evalInsufficientID},
		{"4k3/8/8/8/8/8/8/2B1K3 w - - 0 1", evalInsufficientID},
		{"4k3/8/8/8/8/8/8/1NN1K3 w - - 0 1", evalInsufficientID},
		{"b3k3/8/8/8/8/8/8/2B1K3 w - - 0 1", evalInsufficientID},

		// evalKPvKID
		{"4k3/8/8/8/8/8/4P3/4K3 w - - 0 1", evalKPvKID},
		{"4k3/p7/8/8/8/8/8/4K3 w - - 0 1", evalKPvKID},
		{"4k3/8/8/8/3P4/8/8/4K3 w - - 0 1", evalKPvKID},

		// evalKBNvKID
		{"4k3/8/8/8/8/8/8/4KBN1 w - - 0 1", evalKBNvKID},
		{"4kbn1/8/8/8/8/8/8/4K3 w - - 0 1", evalKBNvKID},

		// evalKBPvKID
		{"4k3/8/8/8/8/8/4P3/2B1K3 w - - 0 1", evalKBPvKID},
		{"4kb2/p7/8/8/8/8/8/4K3 w - - 0 1", evalKBPvKID},
		{"1k6/1B6/7P/8/8/2K5/8/8 w - - 0 1", evalKBPvKID}, // wrong-coloured rook pawn

		// evalOCBID
		{"b3k3/p7/8/8/8/8/P7/2B1K3 w - - 0 1", evalOCBID},
		{"b3k3/8/8/8/8/8/PPP5/2B1K3 w - - 0 1", evalOCBID},

		// evalOCBKnightsID
		{"bn2k3/p7/8/8/8/8/P7/1NB1K3 w - - 0 1", evalOCBKnightsID},

		// evalOCBRooksID
		{"b3k2r/p7/8/8/8/8/P7/2B1K2R w - - 0 1", evalOCBRooksID},

		// evalKMvKPID
		{"4k3/p7/8/8/8/8/8/1N2K3 w - - 0 1", evalKMvKPID},
		{"4kn2/8/8/8/8/8/P7/4K3 w - - 0 1", evalKMvKPID},
		{"4k3/p7/8/8/8/8/8/2B1K3 w - - 0 1", evalKMvKPID},
		{"4kb2/8/8/8/8/8/P7/4K3 w - - 0 1", evalKMvKPID},

		// evalKRvKNID
		{"1n2k3/8/8/8/8/8/8/4K2R w - - 0 1", evalKRvKNID},
		{"4k2r/8/8/8/8/8/8/1N2K3 w - - 0 1", evalKRvKNID},

		// evalKRvKBID
		{"2b1k3/8/8/8/8/8/8/4K2R w - - 0 1", evalKRvKBID},
		{"4k2r/8/8/8/8/8/8/2B1K3 w - - 0 1", evalKRvKBID},

		// evalKRNvKRID
		{"4k2r/8/8/8/8/8/8/1N2K2R w - - 0 1", evalKRNvKRID},
		{"1n2k2r/8/8/8/8/8/8/4K2R w - - 0 1", evalKRNvKRID},

		// evalKRBvKRID
		{"4k2r/8/8/8/8/8/8/2B1K2R w - - 0 1", evalKRBvKRID},
		{"2b1k2r/8/8/8/8/8/8/4K2R w - - 0 1", evalKRBvKRID},

		// evalPositionalID
		{StartPosFEN, evalPositionalID},
		{"4kb2/p7/8/8/8/8/P7/2B1K3 w - - 0 1", evalPositionalID},   // same-coloured bishops, not OCB
		{"b3k3/8/8/8/8/8/PPPP4/2B1K3 w - - 0 1", evalPositionalID}, // OCB but pawnDiff > 3
		{"4k3/8/8/8/8/8/8/2BBK3 w - - 0 1", evalPositionalID},      // KBBvK: minor-value diff > 3
		{"4k3/8/8/8/8/8/PPP5/2B1K3 w - - 0 1", evalPositionalID},   // bishop + multiple pawns is not KBPvK
	}

	for _, tt := range tests {
		t.Run(tt.fen, func(t *testing.T) {
			mockEval := newDispatchRecorder[Score]()
			b := Must(board.FromFEN(tt.fen))
			assert.True(t, mockEval.run(b).callsOnce(tt.want), "fen %s calls once evalId %v", tt.fen, tt.want)
			assert.True(t, mockEval.run(b).onlyCalls(tt.want), "fen %s only calls evalId %v", tt.fen, tt.want)
		})
	}
}
