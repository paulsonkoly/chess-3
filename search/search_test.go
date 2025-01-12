package search_test

import (
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/eval"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/search"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
	"github.com/stretchr/testify/assert"
)

func TestQuiescence(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		b    *board.Board
		want int
	}{
		{name: "initial position",
			b:    board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"),
			want: 0,
		},
		{name: "white a pawn up white to move",
			b:    board.FromFEN("8/8/2k5/8/8/2KP4/8/8 w - - 0 1"),
			want: 100,
		},
		{name: "white a pawn up black to move",
			b:    board.FromFEN("8/8/2k5/8/8/2KP4/8/8 b - - 0 1"),
			want: -100,
		},
		{name: "stalemate",
			b:    board.FromFEN("2k5/2P5/2K5/8/8/8/8/8 b - - 0 1"),
			want: 0,
		},
		{name: "white checkmating",
			b:    board.FromFEN("2k5/1PP5/2K5/8/8/8/8/8 b - - 0 1"),
			want: -eval.Inf,
		},
		{name: "black checkmating",
			b:    board.FromFEN("8/8/8/8/8/5k2/6pP/6BK w - - 0 1"),
			want: -eval.Inf,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := search.Quiescence(tt.b, -eval.Inf, eval.Inf, 0, nil)
			assert.InDelta(t, tt.want, got, 50.0)
		})
	}
}

func TestAlphabeta(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		b     *board.Board
		depth int
		want  int
		move  move.Move
	}{
		{name: "mate in 1",
			b:     board.FromFEN("knK5/p7/1P6/8/8/8/8/5B2 w - - 0 1"),
			depth: 1,
			want:  eval.Inf,
			move:  P(B6, B7),
		},
		{name: "mate in 2 (morphy)",
			b:     board.FromFEN("kbK5/pp6/1P6/8/8/8/8/R7 w - - 0 1"),
			depth: 3,
			want:  eval.Inf,
			move:  R(A1, A6),
		},
		{name: "mate in 3 (Ra6, Ra8, Qxa8#)",
			b:     board.FromFEN("1k6/1P1p4/3r4/3P4/6p1/6Pp/7P/q5BK b - - 0 1"),
			depth: 5,
			want:  eval.Inf,
			move:  R(D6, A6),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, moves := search.AlphaBeta(tt.b, -eval.Inf, eval.Inf, tt.depth, nil)
			assert.Equal(t, tt.want, got)
			assert.Greater(t, len(moves), 0)
			move := moves[len(moves)-1]
			move.Weight = 0
			assert.Equal(t, tt.move, move, moves)
		})
	}
}

func K(f, t Square) move.Move {
	return move.Move{From: f, To: t, Piece: King}
}

func N(f, t Square) move.Move {
	return move.Move{From: f, To: t, Piece: Knight}
}

func B(f, t Square) move.Move {
	return move.Move{From: f, To: t, Piece: Bishop}
}

func R(f, t Square) move.Move {
	return move.Move{From: f, To: t, Piece: Rook}
}

func Q(f, t Square) move.Move {
	return move.Move{From: f, To: t, Piece: Queen}
}

func P(f, t Square) move.Move {
	return move.Move{From: f, To: t, Piece: Pawn}
}

func PP(f, t Square, p Piece) move.Move {
	return move.Move{From: f, To: t, Piece: Pawn, Promo: p}
}
