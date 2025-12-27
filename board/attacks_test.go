package board_test

import (
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/stretchr/testify/assert"

	. "github.com/paulsonkoly/chess-3/chess"
)

func TestIsCheckMate(t *testing.T) {
	tests := []struct {
		name string
		b    *board.Board
		want bool
	}{
		{
			name: "king not in check",
			b:    Must(board.FromFEN("5k2/8/8/8/8/8/8/KR6 w - - 0 1")),
			want: false,
		},
		{
			name: "smothered mate",
			b:    Must(board.FromFEN("kr6/ppN5/8/8/8/8/8/K7 b - - 0 1")),
			want: true,
		},
		{
			name: "king in check not checkmate king can move",
			b:    Must(board.FromFEN("5k2/8/8/8/8/8/8/K4Q2 b - - 0 1")),
			want: false,
		},
		{
			name: "king in check not checkmate capture the checker",
			b:    Must(board.FromFEN("4rkr1/4p1p1/8/1b6/8/8/8/K4Q2 b - - 0 1")),
			want: false,
		},
		{
			name: "king in check not checkmate block the checker",
			b:    Must(board.FromFEN("4rkr1/4p1p1/8/8/3n4/8/8/K4Q2 b - - 0 1")),
			want: false,
		},
		{
			name: "king in double check king can move",
			b:    Must(board.FromFEN("4bk1Q/6p1/8/2B5/8/8/8/K7 b - - 0 1")),
			want: false,
		},
		{
			name: "king in double check king can't move",
			b:    Must(board.FromFEN("4bkr1/6p1/8/2B5/8/8/8/K4Q2 b - - 0 1")),
			want: true,
		},
		{
			name: "en-passant captureable",
			b:    Must(board.FromFEN("8/7B/2bbb3/2bkb3/2bnPp2/8/8/K7 b - e3 0 1")),
			want: false,
		},
		{
			name: "pinned bishop by rook can't capture",
			b:    Must(board.FromFEN("8/7k/5n2/r3B2K/8/6bb/8/2q5 w - - 0 1")),
			want: true,
		},
		{
			name: "pinned rook by bishop can't capture",
			b:    Must(board.FromFEN("8/8/7K/8/7k/5N2/5rQ1/4B3 b - - 0 1")),
			want: true,
		},
		{
			name: "pinned bishop by rook can't block",
			b:    Must(board.FromFEN("8/8/7k/8/r6K/8/4n2B/5b1r w - - 0 1")),
			want: true,
		},
		{
			name: "pinned rook by bishop can't block",
			b:    Must(board.FromFEN("2q5/2b5/8/7k/8/7K/6R1/5b2 w - - 0 1")),
			want: true,
		},
		{
			name: "pawn promotion blocks check white to move",
			b:    Must(board.FromFEN("K4rk1/RP6/8/8/8/8/8/8 w - - 0 1")),
			want: false, // Pawn promotes to block
		},
		{
			name: "pawn promotion blocks black to move",
			b:    Must(board.FromFEN("8/8/8/8/8/8/6pr/KR5k b - - 0 1")),
			want: false, // Pawn promotes to block
		},
		{
			name: "pinned defender cannot capture",
			b:    Must(board.FromFEN("6k1/8/8/4b3/8/r7/1B6/KR6 w - - 0 1")),
			want: true, // Bishop is pinned by hypothetical bishop/rook
		},
		{
			name: "knight check with pinned pawn",
			b:    Must(board.FromFEN("r5k1/8/8/8/8/1n6/PP6/KR6 w - - 0 1")),
			want: true, // Pawn is pinned and cannot capture
		},
		{
			name: "regression 1",
			b:    Must(board.FromFEN("1k1r4/pp3R2/6pp/4p3/2B5/7Q/PPP2B2/2Kq4 w - - 1 1")),
			want: true,
		},
		{
			name: "regression 2",
			b:    Must(board.FromFEN("1kbr4/Qp3R2/3q2pp/4p3/2B5/8/PPP2B2/2K5 b - - 0 1")),
			want: true,
		},
		{
			name: "regression 3 / single pawn push blocks",
			b:    Must(board.FromFEN("rnbqkbnr/ppppp1pp/8/5p1Q/4P3/8/PPPP1PPP/RNB1KBNR b KQkq - 0 1")),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.b.IsCheckmate())
		})
	}
}

func TestIsStalemate(t *testing.T) {
	tests := []struct {
		name string
		b    *board.Board
		want bool
	}{
		{
			name: "king can move",
			b:    Must(board.FromFEN("7k/7p/6pP/4p1P1/4P3/3B4/8/1K6 b - - 0 1")),
			want: false,
		},
		{
			name: "king can't move",
			b:    Must(board.FromFEN("7k/7p/6pP/4p1P1/2B1P3/8/8/1K6 b - - 0 1")),
			want: true,
		},
		{
			name: "pawn can push",
			b:    Must(board.FromFEN("7K/5P2/7k/8/8/8/8/6r1 w - - 0 1")),
			want: false,
		},
		{
			name: "pawn can capture",
			b:    Must(board.FromFEN("7K/8/4pp1k/4P3/8/8/8/6r1 w - - 0 1")),
			want: false,
		},
		{
			name: "pinned queen can move",
			b:    Must(board.FromFEN("7K/8/7k/4Q3/3b4/8/8/6r1 w - - 0 1")),
			want: false,
		},
		{
			name: "pinned rook can't move",
			b:    Must(board.FromFEN("7K/8/7k/4R3/3b4/8/8/6r1 w - - 0 1")),
			want: true,
		},
		{
			name: "pinned rook can move",
			b:    Must(board.FromFEN("q3R2K/8/7k/8/8/8/8/6r1 w - - 0 1")),
			want: false,
		},
		{
			name: "pinned knight can't move",
			b:    Must(board.FromFEN("7K/8/7k/4N3/3b4/8/8/6r1 w - - 0 1")),
			want: true,
		},
		{
			name: "knight can move",
			b:    Must(board.FromFEN("7K/8/7k/4N3/8/8/8/6r1 w - - 0 1")),
			want: false,
		},
		{
			name: "pinned pawn can't move",
			b:    Must(board.FromFEN("7K/8/7k/4P3/3b4/8/8/6r1 w - - 0 1")),
			want: true,
		},
		{
			name: "en-passant captureable",
			b:    Must(board.FromFEN("7k/7p/6pP/3B2P1/2pP4/2N5/8/1K6 b - d3 0 1")),
			want: false,
		},
		{
			name: "en-passant pinned (diag)",
			b:    Must(board.FromFEN("7k/7p/6pP/3B2P1/2pP4/2B5/8/1K6 b - d3 0 1")),
			want: true,
		},
		{
			name: "en-passant pinned (rank)",
			b:    Must(board.FromFEN("1kb4q/6p1/3p2P1/r2Pp1K1/r7/8/8/8 w - e6 0 2")),
			want: true,
		},
		{
			name: "pawn can promote",
			b:    Must(board.FromFEN("8/1P6/8/8/8/2rk2b1/8/3K4 w - - 0 1")),
			want: false,
		},
		{
			name: "knight not actually pinned",
			b:    Must(board.FromFEN("8/8/8/BB2n2B/R6B/4k3/4p3/K3R3 w - - 0 1")),
			want: false,
		},
		{
			name: "edge pawn capture",
			b:    Must(board.FromFEN("8/8/6pp/7P/5k1K/7P/8/8 w - - 0 1")),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.b.IsStalemate())
		})
	}
}

func TestEnPassantStates(t *testing.T) {
	tests := []struct {
		name   string
		b      *board.Board
		move   move.Move
		bAfter *board.Board
	}{
		{
			name:   "En passant possible after pawn move",
			b:      Must(board.FromFEN("4k3/8/8/8/3p4/8/2P5/4K3 w - - 0 1")),
			move:   move.New(C2, C4),
			bAfter: Must(board.FromFEN("4k3/8/8/8/2Pp4/8/8/4K3 b - c3 0 1")),
		},
		{
			name:   "En passant not possible due to no pawn",
			b:      Must(board.FromFEN("4k3/8/8/8/8/8/2P5/4K3 w - - 0 1")),
			move:   move.New(C2, C4),
			bAfter: Must(board.FromFEN("4k3/8/8/8/2P5/8/8/4K3 b - - 0 1")),
		},
		{
			name:   "En passant not possible due to simple pin",
			b:      Must(board.FromFEN("8/8/1k6/8/3p4/8/2P5/3K2B1 w - - 0 1")),
			move:   move.New(C2, C4),
			bAfter: Must(board.FromFEN("8/8/1k6/8/2Pp4/8/8/3K2B1 b - - 0 1")),
		},
		{
			name:   "En passant not possible due to tricky pin",
			b:      Must(board.FromFEN("8/8/8/8/k2p3R/8/2P5/3K4 w - - 0 1")),
			move:   move.New(C2, C4),
			bAfter: Must(board.FromFEN("8/8/8/8/k1Pp3R/8/8/3K4 b - - 0 1")),
		},
		{
			name:   "En passant possible in pin that's not affected",
			b:      Must(board.FromFEN("4r3/pkp3b1/1p5p/2P1npp1/P2rp3/6PN/1P2PPBP/1RR3K1 w - - 0 22")),
			move:   move.New(F2, F4),
			bAfter: Must(board.FromFEN("4r3/pkp3b1/1p5p/2P1npp1/P2rpP2/6PN/1P2P1BP/1RR3K1 b - f3 0 23")),
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			b := tt.b
			b.MakeMove(tt.move)
			assert.Equal(t, tt.bAfter.EnPassant, b.EnPassant)
			assert.Equal(t, tt.bAfter.Hash(), b.Hash())
		})
	}
}

func TestIsAttacked(t *testing.T) {
	tests := []struct {
		name   string
		b      *board.Board
		by     Color
		target BitBoard
		want   bool
	}{
		{
			name:   "king not in check",
			b:      Must(board.FromFEN("8/1k6/8/8/8/8/8/RNBQKBNR w - - 0 1")),
			by:     White,
			target: BitBoardFromSquares(B7),
			want:   false,
		},
		{
			name:   "king in check by knight",
			b:      Must(board.FromFEN("8/8/8/8/8/2k5/8/RNBQKBNR w - - 0 1")),
			by:     White,
			target: BitBoardFromSquares(C3),
			want:   true,
		},
		{
			name:   "king in check by bishop",
			b:      Must(board.FromFEN("8/8/8/8/8/4k3/8/RNBQKBNR w - - 0 1")),
			by:     White,
			target: BitBoardFromSquares(E3),
			want:   true,
		},
		{
			name:   "bishop does not attack through a blocking piece",
			b:      Must(board.FromFEN("8/8/8/8/8/4k3/3N4/R1BQKBNR w - - 0 1")),
			by:     White,
			target: BitBoardFromSquares(E3),
			want:   false,
		},
		{
			name:   "king in check by rook",
			b:      Must(board.FromFEN("k7/8/8/8/8/8/8/RNBQKBNR w - - 0 1")),
			by:     White,
			target: BitBoardFromSquares(A8),
			want:   true,
		},
		{
			name:   "rook does not attack through a blocking piece",
			b:      Must(board.FromFEN("k7/8/8/8/8/N7/8/R1BQKBNR w - - 0 1")),
			by:     White,
			target: BitBoardFromSquares(A8),
			want:   false,
		},
		{
			name:   "king in check by queen",
			b:      Must(board.FromFEN("8/8/8/8/3k4/8/8/RNBQKBNR w - - 0 1")),
			by:     White,
			target: BitBoardFromSquares(D4),
			want:   true,
		},
		{
			name:   "queen does not attack through a blocking piece",
			b:      Must(board.FromFEN("8/8/8/8/6k1/8/4N3/RNBQKB1R w - - 0 1")),
			by:     White,
			target: BitBoardFromSquares(G4),
			want:   false,
		},
		{
			name:   "king in check by pawn",
			b:      Must(board.FromFEN("8/8/8/8/5k2/4P3/8/K7 w - - 0 1")),
			by:     White,
			target: BitBoardFromSquares(F4),
			want:   true,
		},
		{
			name:   "king not in check by pawn wrap",
			b:      Must(board.FromFEN("8/8/8/8/7k/P7/8/K7 w - - 0 1")),
			by:     White,
			target: BitBoardFromSquares(H4),
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			occ := tt.b.Colors[White] | tt.b.Colors[Black]
			assert.Equal(t, tt.want, tt.b.IsAttacked(tt.by, occ, tt.target))
		})
	}
}
