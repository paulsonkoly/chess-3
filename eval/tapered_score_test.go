package eval

import (
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/stretchr/testify/assert"
)

func TestTaperedScore(t *testing.T) {
	tests := [...]struct {
		name   string
		fen    string
		scores [Colors][Phases]Score
		want   Score
	}{
		{
			"empty board",
			"4k3/8/8/8/8/8/8/4K3 w - - 0 1",
			[Colors][Phases]Score{{20, 10}, {50, 60}},
			-50, // stm white, full EG: 10 - 60
		},
		{
			"full board",
			"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			[Colors][Phases]Score{{20, 10}, {50, 60}},
			-30, // stm white, full MG: 20 - 50
		},
		{
			"half full board",
			"4k3/8/8/8/8/8/PPPPPPPP/RNBQKBNR b KQ - 0 1",
			[Colors][Phases]Score{{20, 10}, {50, 60}},
			40, // stm black, half MG/EG: ((60 - 10) + (50 - 20))/2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Must(board.FromFEN(tt.fen))
			assert.Equal(t, tt.want, evalForScore(tt.scores).taperedScore(b))
			assert.InDelta(t, float64(tt.want), evalForFloat64(tt.scores).taperedScore(b), 0.5)
		})
	}
}

func evalForScore(scores [Colors][Phases]Score) *Eval[Score] {
	e := New[Score]()
	e.sp = scores
	return e
}

func evalForFloat64(scores [Colors][Phases]Score) *Eval[float64] {
	e := New[float64]()
	e.sp[White][MG] += float64(scores[White][MG])
	e.sp[White][EG] += float64(scores[White][EG])
	e.sp[Black][MG] += float64(scores[Black][MG])
	e.sp[Black][EG] += float64(scores[Black][EG])
	return e
}

func TestEngameScore(t *testing.T) {
	tests := [...]struct {
		name   string
		fen    string
		scores [Colors][Phases]Score
		want   Score
	}{
		{
			"empty board",
			"4k3/8/8/8/8/8/8/4K3 w - - 0 1",
			[Colors][Phases]Score{{20, 10}, {50, 60}},
			-50, // 10-60
		},
		{
			"full board",
			"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			[Colors][Phases]Score{{20, 10}, {50, 60}},
			-50,
		},
		{
			"half full board",
			"4k3/8/8/8/8/8/PPPPPPPP/RNBQKBNR b KQ - 0 1",
			[Colors][Phases]Score{{20, 10}, {50, 60}},
			50, // 60-10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Must(board.FromFEN(tt.fen))
			e := New[Score]()
			e.sp = tt.scores
			assert.Equal(t, tt.want, e.endgameScore(b))
		})
	}
}
