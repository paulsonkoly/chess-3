package board_test

import (
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/stretchr/testify/assert"

	. "github.com/paulsonkoly/chess-3/chess"
)

func TestIsStalemate(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		want bool
	}{
		{
			name: "king can move",
			fen:  "7k/7p/6pP/4p1P1/4P3/3B4/8/1K6 b - - 0 1",
			want: false,
		},
		{
			name: "king can't move",
			fen:  "7k/7p/6pP/4p1P1/2B1P3/8/8/1K6 b - - 0 1",
			want: true,
		},
		{
			name: "pawn can push",
			fen:  "7K/5P2/7k/8/8/8/8/6r1 w - - 0 1",
			want: false,
		},
		{
			name: "pawn can capture",
			fen:  "7K/8/4pp1k/4P3/8/8/8/6r1 w - - 0 1",
			want: false,
		},
		{
			name: "pinned queen can move",
			fen:  "7K/8/7k/4Q3/3b4/8/8/6r1 w - - 0 1",
			want: false,
		},
		{
			name: "pinned rook can't move",
			fen:  "7K/8/7k/4R3/3b4/8/8/6r1 w - - 0 1",
			want: true,
		},
		{
			name: "pinned rook can move",
			fen:  "q3R2K/8/7k/8/8/8/8/6r1 w - - 0 1",
			want: false,
		},
		{
			name: "pinned knight can't move",
			fen:  "7K/8/7k/4N3/3b4/8/8/6r1 w - - 0 1",
			want: true,
		},
		{
			name: "knight can move",
			fen:  "7K/8/7k/4N3/8/8/8/6r1 w - - 0 1",
			want: false,
		},
		{
			name: "pinned pawn can't move",
			fen:  "7K/8/7k/4P3/3b4/8/8/6r1 w - - 0 1",
			want: true,
		},
		{
			name: "en-passant captureable",
			fen:  "7k/7p/6pP/3B2P1/2pP4/2N5/8/1K6 b - d3 0 1",
			want: false,
		},
		{
			name: "en-passant pinned (diag)",
			fen:  "7k/7p/6pP/3B2P1/2pP4/2B5/8/1K6 b - d3 0 1",
			want: true,
		},
		{
			name: "en-passant pinned (rank)",
			fen:  "1kb4q/6p1/3p2P1/r2Pp1K1/r7/8/8/8 w - e6 0 2",
			want: true,
		},
		{
			name: "pawn can promote",
			fen:  "8/1P6/8/8/8/2rk2b1/8/3K4 w - - 0 1",
			want: false,
		},
		{
			name: "knight not actually pinned",
			fen:  "8/8/8/BB2n2B/R6B/4k3/4p3/K3R3 w - - 0 1",
			want: false,
		},
		{
			name: "edge pawn capture",
			fen:  "8/8/6pp/7P/5k1K/7P/8/8 w - - 0 1",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Must(board.FromFEN(tt.fen))
			assert.Equal(t, tt.want, b.IsStalemate(), "fen: %s", tt.fen)
		})
	}
}

func TestEnPassantStates(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		move move.Move
		want Square
	}{
		{
			name: "En passant possible after pawn move",
			fen:  "4k3/8/8/8/3p4/8/2P5/4K3 w - - 0 1",
			move: move.From(C2) | move.To(C4),
			want: C3,
		},
		{
			name: "En passant not possible due to no pawn",
			fen:  "4k3/8/8/8/8/8/2P5/4K3 w - - 0 1",
			move: move.From(C2) | move.To(C4),
			want: 0,
		},
		{
			name: "En passant not possible due to simple pin",
			fen:  "8/8/1k6/8/3p4/8/2P5/3K2B1 w - - 0 1",
			move: move.From(C2) | move.To(C4),
			want: 0,
		},
		{
			name: "En passant not possible due to tricky pin",
			fen:  "8/8/8/8/k2p3R/8/2P5/3K4 w - - 0 1",
			move: move.From(C2) | move.To(C4),
			want: 0,
		},
		{
			name: "En passant possible in pin that's not affected",
			fen:  "4r3/pkp3b1/1p5p/2P1npp1/P2rp3/6PN/1P2PPBP/1RR3K1 w - - 0 22",
			move: move.From(F2) | move.To(F4),
			want: F3,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			b := Must(board.FromFEN(tt.fen))
			b.MakeMove(tt.move)
			assert.Equal(t, tt.want, b.EnPassant, "fen: %s, move: %s", tt.fen, tt.move)
		})
	}
}

func TestIsAttacked(t *testing.T) {
	tests := []struct {
		name   string
		fen    string
		by     Color
		target BitBoard
		want   bool
	}{
		{
			name:   "king not in check",
			fen:    "8/1k6/8/8/8/8/8/RNBQKBNR w - - 0 1",
			by:     White,
			target: BitBoardFromSquares(B7),
			want:   false,
		},
		{
			name:   "king in check by knight",
			fen:    "8/8/8/8/8/2k5/8/RNBQKBNR w - - 0 1",
			by:     White,
			target: BitBoardFromSquares(C3),
			want:   true,
		},
		{
			name:   "king in check by bishop",
			fen:    "8/8/8/8/8/4k3/8/RNBQKBNR w - - 0 1",
			by:     White,
			target: BitBoardFromSquares(E3),
			want:   true,
		},
		{
			name:   "bishop does not attack through a blocking piece",
			fen:    "8/8/8/8/8/4k3/3N4/R1BQKBNR w - - 0 1",
			by:     White,
			target: BitBoardFromSquares(E3),
			want:   false,
		},
		{
			name:   "king in check by rook",
			fen:    "k7/8/8/8/8/8/8/RNBQKBNR w - - 0 1",
			by:     White,
			target: BitBoardFromSquares(A8),
			want:   true,
		},
		{
			name:   "rook does not attack through a blocking piece",
			fen:    "k7/8/8/8/8/N7/8/R1BQKBNR w - - 0 1",
			by:     White,
			target: BitBoardFromSquares(A8),
			want:   false,
		},
		{
			name:   "king in check by queen",
			fen:    "8/8/8/8/3k4/8/8/RNBQKBNR w - - 0 1",
			by:     White,
			target: BitBoardFromSquares(D4),
			want:   true,
		},
		{
			name:   "queen does not attack through a blocking piece",
			fen:    "8/8/8/8/6k1/8/4N3/RNBQKB1R w - - 0 1",
			by:     White,
			target: BitBoardFromSquares(G4),
			want:   false,
		},
		{
			name:   "king in check by pawn",
			fen:    "8/8/8/8/5k2/4P3/8/K7 w - - 0 1",
			by:     White,
			target: BitBoardFromSquares(F4),
			want:   true,
		},
		{
			name:   "king not in check by pawn wrap",
			fen:    "8/8/8/8/7k/P7/8/K7 w - - 0 1",
			by:     White,
			target: BitBoardFromSquares(H4),
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Must(board.FromFEN(tt.fen))
			occ := b.Colors[White] | b.Colors[Black]
			assert.Equal(t, tt.want, b.IsAttacked(tt.by, occ, tt.target),
				"fen: %s by: %s target %8.8x", tt.fen, tt.by, tt.target)
		})
	}
}

func TestCheckers(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		want BitBoard
	}{
		{
			name: "king not in check",
			fen:  "8/1k6/8/8/8/8/8/RNBQKBNR b - - 0 1",
			want: 0,
		},
		{
			name: "king in check by knight",
			fen:  "8/8/8/8/8/2k5/8/RNBQKBNR b - - 0 1",
			want: BitBoardFromSquares(B1),
		},
		{
			name: "king in check by bishop",
			fen:  "8/8/8/8/8/4k3/8/RNBQKBNR b - - 0 1",
			want: BitBoardFromSquares(C1),
		},
		{
			name: "bishop does not attack through a blocking piece",
			fen:  "8/8/8/8/8/4k3/3N4/R1BQKBNR b - - 0 1",
			want: 0,
		},
		{
			name: "king in check by rook",
			fen:  "k7/8/8/8/8/8/8/RNBQKBNR b - - 0 1",
			want: BitBoardFromSquares(A1),
		},
		{
			name: "rook does not attack through a blocking piece",
			fen:  "k7/8/8/8/8/N7/8/R1BQKBNR b - - 0 1",
			want: 0,
		},
		{
			name: "king in check by queen",
			fen:  "8/8/8/8/3k4/8/8/RNBQKBNR b - - 0 1",
			want: BitBoardFromSquares(D1),
		},
		{
			name: "queen does not attack through a blocking piece",
			fen:  "8/8/8/8/6k1/8/4N3/RNBQKB1R b - - 0 1",
			want: 0,
		},
		{
			name: "king in check by pawn",
			fen:  "8/8/8/8/5k2/4P3/8/K7 b - - 0 1",
			want: BitBoardFromSquares(E3),
		},
		{
			name: "king not in check by pawn wrap",
			fen:  "8/8/8/8/7k/P7/8/K7 b - - 0 1",
			want: 0,
		},
		{
			name: "king in double check",
			fen:  "8/8/8/8/5R1k/8/8/K3B3 b - - 0 1",
			want: BitBoardFromSquares(E1, F4),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Must(board.FromFEN(tt.fen))
			assert.Equal(t, tt.want, b.Checkers(), "fen: %s", tt.fen)
		})
	}
}
