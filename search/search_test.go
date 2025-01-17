package search_test

import (
	"testing"

	"github.com/paulsonkoly/chess-3/board"
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
		want Score
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
			want: -Inf,
		},
		{name: "black checkmating",
			b:    board.FromFEN("8/8/8/8/8/5k2/6pP/6BK w - - 0 1"),
			want: -Inf,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := search.Quiescence(tt.b, -Inf-1, Inf+1, 0, nil)
			assert.InDelta(t, int(tt.want), int(got), 50.0)
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
