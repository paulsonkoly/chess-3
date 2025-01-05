package movegen_test

import (
	"slices"
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"
	"github.com/stretchr/testify/assert"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

func TestMoves(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		b      *board.Board
		target board.BitBoard
		want   []move.Move
	}{
		{
			name:   "simple king move",
			b:      board.FromFEN("8/8/8/8/8/4K3/8/8 w - - 0 1"),
			target: board.Full,
			want: []move.Move{
				K(E3, D2), K(E3, E2), K(E3, F2),
				K(E3, D3), K(E3, F3),
				K(E3, D4), K(E3, E4), K(E3, F4),
			},
		},
		{
			name:   "king in the corner",
			b:      board.FromFEN("8/8/8/8/8/8/8/7k b - - 0 1"),
			target: board.Full,
			want: []move.Move{
				K(H1, H2), K(H1, G2), K(H1, G1),
			},
		},
		{
			name:   "simple knight move",
			b:      board.FromFEN("8/8/8/8/8/4N3/8/8 w - - 0 1"),
			target: board.Full,
			want: []move.Move{
				N(E3, C4), N(E3, D5), N(E3, F5), N(E3, G4),
				N(E3, C2), N(E3, D1), N(E3, F1), N(E3, G2),
			},
		},
		{
			name:   "knight in the corner",
			b:      board.FromFEN("8/8/8/8/8/8/8/7n b - - 0 1"),
			target: board.Full,
			want: []move.Move{
				N(H1, F2), N(H1, G3),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			want := tt.want
			ok := make([]bool, len(want))

			for m := range movegen.Moves(tt.b, tt.target) {
				ix := slices.Index(want, m)
				if ix == -1 {
					t.Errorf("unexpected move %s generated", m)
				} else {
					ok[ix] = true
				}
			}

			for ix, v := range ok {
				if !v {
					t.Errorf("move %s not generated", want[ix])
				}
			}
		})
	}
}

func K(f, t Square) move.Move {
	return move.Move{From: f, To: t, Piece: King}
}

func N(f, t Square) move.Move {
	return move.Move{From: f, To: t, Piece: Knight}
}

func TestIsAttacked(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		b      *board.Board
		by     Color
		target board.BitBoard
		want   bool
	}{
		{
			name:   "king not in check",
			b:      board.FromFEN("8/8/4k3/4n3/8/4K3/8/8 w - - 0 1"),
			by:     Black,
			target: board.BitBoardFromSquares(E3),
			want:   false,
		},
		{
			name:   "king in check by knight",
			b:      board.FromFEN("8/8/4k3/4n3/2n5/4K3/8/8 w - - 0 1"),
			by:     Black,
			target: board.BitBoardFromSquares(E3),
			want:   true,
		},
		{
			name:   "king in check by king (illegal)",
			b:      board.FromFEN("8/8/8/4k3/4K3/8/8/8 w - - 0 1"),
			by:     Black,
			target: board.BitBoardFromSquares(E4),
			want:   true,
		},
		{
			name:   "king not in check by own knight",
			b:      board.FromFEN("8/8/8/4k3/8/8/4K3/2N5 w - - 0 1"),
			by:     Black,
			target: board.BitBoardFromSquares(E2),
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, movegen.IsAttacked(tt.b, tt.by, tt.target))
		})
	}
}
