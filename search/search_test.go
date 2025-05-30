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

func TestAlphaBeta(t *testing.T) {
	tests := []struct {
		name string
		b    *board.Board
		d    Depth
	}{
		{name: "position fen qqqqkqqq/8/8/8/8/8/8/QQQQKQQQ w - - 0 1",
			b: Must(board.FromFEN("qqqqkqqq/8/8/8/8/8/8/QQQQKQQQ w - - 0 1")),
			d: 9,
		},
	}
	sst := search.NewState(1)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sst.Clear()
			assert.NotPanics(t,
				func() {
					search.Search(tt.b, tt.d, sst)
				},
			)
		})
	}
}

func TestQuiescence(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		b    *board.Board
		want Score
	}{
		{name: "initial position",
			b:    Must(board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")),
			want: 0,
		},
		{name: "white a pawn up white to move",
			b:    Must(board.FromFEN("8/8/2k5/8/8/2KP4/8/8 w - - 0 1")),
			want: 100,
		},
		{name: "white a pawn up black to move",
			b:    Must(board.FromFEN("8/8/2k5/8/8/2KP4/8/8 b - - 0 1")),
			want: -100,
		},
		{name: "stalemate",
			b:    Must(board.FromFEN("2k5/2P5/2K5/8/8/8/8/8 b - - 0 1")),
			want: 0,
		},
		{name: "white checkmating",
			b:    Must(board.FromFEN("2k5/1PP5/2K5/8/8/8/8/8 b - - 0 1")),
			want: -Inf,
		},
		{name: "black checkmating",
			b:    Must(board.FromFEN("8/8/8/8/8/5k2/6pP/6BK w - - 0 1")),
			want: -Inf,
		},
	}

	sst := search.NewState(1)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sst.Clear()
			got := search.Quiescence(tt.b, -Inf-1, Inf+1, 0, 0, sst)
			assert.InDelta(t, int(tt.want), int(got), 50.0)
		})
	}
}

func TestDraws(t *testing.T) {
	tests := []struct {
		name  string
		b     *board.Board
		moves []move.Move
		want  int
	}{
		{name: "threefold",
			b:     Must(board.FromFEN("r5k1/p1R5/1p5R/2p5/8/2P4P/P1P3PK/r7 w - - 3 36")),
			moves: []move.Move{R(H6, G6), K(G8, H8), R(G6, H6), K(H8, G8), R(H6, G6), K(G8, H8), R(G6, H6), K(H8, G8)},
			want:  0,
		},
		{name: "fifty",
			b:     Must(board.FromFEN("8/5R1P/8/3Q4/7k/8/1P6/6K1 b - - 98 161")),
			moves: []move.Move{K(H4, H3), Q(D5, D3)},
			want:  0,
		},
	}

	sst := search.NewState(1)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, m := range tt.moves {
				tt.b.MakeMove(&m)
			}

			sst.Clear()

			score, _ := search.Search(tt.b, 3, sst)

			assert.InDelta(t, tt.want, int(score), 200)
		})
	}
}

func K(f, t Square) move.Move {
	return move.Move{SimpleMove: move.FromSquares(f, t), Piece: King}
}

func N(f, t Square) move.Move {
	return move.Move{SimpleMove: move.FromSquares(f, t), Piece: Knight}
}

func B(f, t Square) move.Move {
	return move.Move{SimpleMove: move.FromSquares(f, t), Piece: Bishop}
}

func R(f, t Square) move.Move {
	return move.Move{SimpleMove: move.FromSquares(f, t), Piece: Rook}
}

func Q(f, t Square) move.Move {
	return move.Move{SimpleMove: move.FromSquares(f, t), Piece: Queen}
}

func P(f, t Square) move.Move {
	return move.Move{SimpleMove: move.FromSquares(f, t), Piece: Pawn}
}

func PP(f, t Square, p Piece) move.Move {
	sm := move.FromSquares(f, t)
	sm.SetPromo(p)
	return move.Move{SimpleMove: sm, Piece: Pawn}
}
