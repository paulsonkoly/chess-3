package heur_test

import (
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/heur"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/stretchr/testify/assert"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

func TestSEE(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		b    *board.Board
		m    move.Move
		want Score
	}{
		{name: "3Q4/3q4/1B2N3/5N2/2KPk3/3r4/2n1nb2/3R4 b - - 0 1",
			b:    board.FromFEN("3Q4/3q4/1B2N3/5N2/2KPk3/3r4/2n1nb2/3R4 b - - 0 1"),
			m:    N(C2, D4),
			want: -200,
		},
		{name: "7k/2b5/8/8/2N5/1R6/8/7K w - - 0 4",
			b:    board.FromFEN("7k/2b5/8/8/2N5/1R6/8/7K w - - 0 4"),
			m:    R(B3, B6),
			want: -200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := heur.SEE(tt.b, &tt.m)
			assert.Equal(t, tt.want, got)
		})
	}
}

func K(f, t Square) move.Move {
  return move.Move{SimpleMove: move.SimpleMove {From: f, To: t}, Piece: King}
}

func N(f, t Square) move.Move {
  return move.Move{SimpleMove: move.SimpleMove {From: f, To: t}, Piece: Knight}
}

func B(f, t Square) move.Move {
  return move.Move{SimpleMove: move.SimpleMove {From: f, To: t}, Piece: Bishop}
}

func R(f, t Square) move.Move {
  return move.Move{SimpleMove: move.SimpleMove {From: f, To: t}, Piece: Rook}
}

func Q(f, t Square) move.Move {
  return move.Move{SimpleMove: move.SimpleMove {From: f, To: t}, Piece: Queen}
}

func P(f, t Square) move.Move {
  return move.Move{SimpleMove: move.SimpleMove {From: f, To: t}, Piece: Pawn}
}

func PP(f, t Square, p Piece) move.Move {
  return move.Move{SimpleMove: move.SimpleMove {From: f, To: t, Promo: p}, Piece: Pawn}
}
