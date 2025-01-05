package movegen_test

import (
	"slices"
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"
	"github.com/stretchr/testify/assert"

	// revive:disable-next-line
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
			b:      board.FromFEN("8/8/8/8/8/4K3/8/k7 w - - 0 1"),
			target: board.Full,
			want: []move.Move{
				K(E3, D2), K(E3, E2), K(E3, F2),
				K(E3, D3), K(E3, F3),
				K(E3, D4), K(E3, E4), K(E3, F4),
			},
		},
		{
			name:   "king in the corner",
			b:      board.FromFEN("8/8/8/8/8/8/K7/7k b - - 0 1"),
			target: board.Full,
			want: []move.Move{
				K(H1, H2), K(H1, G2), K(H1, G1),
			},
		},
		{
			name:   "simple knight move",
			b:      board.FromFEN("8/8/8/8/8/4N3/8/k6K w - - 0 1"),
			target: board.Full,
			want: []move.Move{
				N(E3, C4), N(E3, D5), N(E3, F5), N(E3, G4),
				N(E3, C2), N(E3, D1), N(E3, F1), N(E3, G2),
				K(H1, G1), K(H1, G2), K(H1, H2), 
			},
		},
		{
			name:   "knight in the corner",
			b:      board.FromFEN("k7/8/8/8/8/8/8/K6N w - - 0 1"),
			target: board.Full,
			want: []move.Move{
				N(H1, F2), N(H1, G3),
				K(A1, A2), K(A1, B2), K(A1, B1), 
			},
		},
		{
			name:   "simple bishop move",
			b:      board.FromFEN("k7/8/8/8/8/3B4/8/7K w - - 0 1"),
			target: board.Full,
			want: []move.Move{
        B(D3, C2), B(D3, B1), B(D3, E2), B(D3, F1), B(D3, C4), B(D3, B5),
        B(D3, A6), B(D3, E4), B(D3, F5), B(D3, G6), B(D3, H7), 
				K(H1, G1), K(H1, G2), K(H1, H2), 
			},
		},
		{
			name:   "bishop in the corner",
			b:      board.FromFEN("k7/8/8/8/8/8/8/B6K w - - 0 1"),
			target: board.Full,
			want: []move.Move{
        B(A1, B2), B(A1, C3), B(A1, D4), B(A1, E5), B(A1, F6), B(A1, G7), B(A1, H8),
				K(H1, G1), K(H1, G2), K(H1, H2), 
			},
		},
		{
			name:   "bishop blocked by friendly",
			b:      board.FromFEN("k7/8/8/8/8/2K5/1B6/8 w - - 0 1"),
			target: board.Full,
			want: []move.Move{
        B(B2, A3), B(B2, A1) , B(B2, C1),
        K(C3, B3), K(C3, B4), K(C3, C2), K(C3, C4), K(C3, D2), K(C3, D3), K(C3, D4),
			},
		},
		{
			name:   "simple rook move",
			b:      board.FromFEN("k7/8/8/8/4R3/8/8/7K w - - 0 1"),
			target: board.Full,
			want: []move.Move{
        R(E4, D4), R(E4, C4), R(E4, B4), R(E4, A4), R(E4, H4), R(E4, G4), R(E4, F4),
        R(E4, E5), R(E4, E6), R(E4, E7), R(E4, E8), R(E4, E3), R(E4, E2), R(E4, E1),
				K(H1, G1), K(H1, G2), K(H1, H2), 
			},
		},
		{
			name:   "rook in the corner",
			b:      board.FromFEN("k7/8/8/8/8/8/8/R6K w - - 0 1"),
			target: board.Full,
			want: []move.Move{
        R(A1, A2), R(A1, A3), R(A1, A4), R(A1, A5), R(A1, A6), R(A1, A7), R(A1, A8),
        R(A1, B1), R(A1, C1), R(A1, D1), R(A1, E1), R(A1, F1), R(A1, G1),
				K(H1, G1), K(H1, G2), K(H1, H2), 
			},
		},
		{
			name:   "rook blocked by friendly",
			b:      board.FromFEN("k7/8/8/8/8/2K5/2R5/8 w - - 0 1"),
			target: board.Full,
			want: []move.Move{
        R(C2, B2), R(C2, A2), R(C2, D2), R(C2, E2), R(C2, F2), R(C2, G2), R(C2, H2),
        R(C2, C1),
        K(C3, B3), K(C3, B4), K(C3, B2), K(C3, C4), K(C3, D2), K(C3, D3), K(C3, D4),
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

func B(f, t Square) move.Move {
	return move.Move{From: f, To: t, Piece: Bishop}
}

func R(f, t Square) move.Move {
	return move.Move{From: f, To: t, Piece: Rook}
}

func TestIsAttacked(t *testing.T) {
	tests := []struct {
		name   string
		b      *board.Board
		by     Color
		target board.BitBoard
		want   bool
	}{
		{
			name:   "king not in check",
			b:      board.FromFEN("8/1k6/8/8/8/8/8/RNBQKBNR w - - 0 1"),
			by:     White,
			target: board.BitBoardFromSquares(B7),
			want:   false,
		},
		{
			name:   "king in check by knight",
			b:      board.FromFEN("8/8/8/8/8/2k5/8/RNBQKBNR w - - 0 1"),
			by:     White,
			target: board.BitBoardFromSquares(C3),
			want:   true,
		},
		{
			name:   "king in check by bishop",
			b:      board.FromFEN("8/8/8/8/8/4k3/8/RNBQKBNR w - - 0 1"),
			by:     White,
			target: board.BitBoardFromSquares(E3),
			want:   true,
		},
		{
			name:   "bishop does not attack through a blocking piece",
			b:      board.FromFEN("8/8/8/8/8/4k3/3N4/R1BQKBNR w - - 0 1"),
			by:     White,
			target: board.BitBoardFromSquares(E3),
			want:   false,
		},
		{
			name:   "king in check by rook",
			b:      board.FromFEN("k7/8/8/8/8/8/8/RNBQKBNR w - - 0 1"),
			by:     White,
			target: board.BitBoardFromSquares(A8),
			want:   true,
		},
		{
			name:   "rook does not attack through a blocking piece",
			b:      board.FromFEN("k7/8/8/8/8/N7/8/R1BQKBNR w - - 0 1"),
			by:     White,
			target: board.BitBoardFromSquares(A8),
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, movegen.IsAttacked(tt.b, tt.by, tt.target))
		})
	}
}

