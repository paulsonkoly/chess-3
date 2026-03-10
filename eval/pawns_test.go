package eval

import (
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/stretchr/testify/assert"

	. "github.com/paulsonkoly/chess-3/chess"
)

func TestHoles(t *testing.T) {
	tests := [...]struct {
		name  string
		fen   string
		color Color
		want  BitBoard
	}{
		{
			"startpos white",
			"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			White,
			BitBoardFromSquares(A1, B1, C1, D1, E1, F1, G1, H1, A2, B2, C2, D2, E2, F2, G2, H2),
		},
		{
			"startpos black",
			"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			Black,
			BitBoardFromSquares(A7, B7, C7, D7, E7, F7, G7, H7, A8, B8, C8, D8, E8, F8, G8, H8),
		},
		{
			"complicated pawn structure with white",
			"7k/2ppp3/1p3p2/p5pP/2P3Pp/1P1P1P2/P3P3/4K3 w - - 0 1",
			White,
			BitBoardFromSquares(
				A1, B1, C1, D1, E1, F1, G1, H1,
				A2, B2, C2, D2, E2, F2, G2, H2,
				A3, C3, E3, G3, H3,
				H4,
			),
		},
		{
			"complicated pawn structure with black",
			"7k/2ppp3/1p3p2/p5pP/2P3Pp/1P1P1P2/P3P3/4K3 w - - 0 1",
			Black,
			BitBoardFromSquares(
				H5,
				A6, G6, H6,
				A7, B7, C7, D7, E7, F7, G7, H7,
				A8, B8, C8, D8, E8, F8, G8, H8,
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Must(board.FromFEN(tt.fen))
			pawns := pawns{}
			pawns.fromBoard(b)

			assert.Equal(t, tt.want, pawns.holes(tt.color), "fen %s color %v", tt.fen, tt.color)
		})
	}
}

func TestPassers(t *testing.T) {
	tests := [...]struct {
		name  string
		fen   string
		color Color
		want  BitBoard
	}{
		{"startpos white", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", White, 0},
		{"startpos black", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", Black, 0},
		{"empty board", "4k3/8/8/8/8/8/8/4K3 w - - 0 1", White, 0},
		{"single pawn", "4k3/1p6/8/8/8/8/8/4K3 w - - 0 1", Black, BitBoardFromSquares(B7)},
		{"connected pawns, not passers", "4k3/1p6/8/8/8/8/PP6/4K3 w - - 0 1", White, 0},
		{"blocked pawn on edge", "4k3/8/8/p7/P7/8/8/4K3 w - - 0 1", White, 0},
		{"blocked pawn in the middle", "4k3/8/8/3p4/3P4/8/8/4K3 w - - 0 1", White, 0},
		{"pawn on same file already passed", "4k3/8/8/4P3/4p3/8/8/4K3 w - - 0 1", White, BitBoardFromSquares(E5)},
		{"pawn blocked by both enemy frontline & cover", "4k3/1p6/p7/8/P7/8/8/4K3 w - - 0 1", White, 0},
		{"doubled pawns", "4k3/8/8/8/8/1P6/1P6/4K3 w - - 0 1", White, BitBoardFromSquares(B3)},
		{"pawn one step from promotion", "4k3/2P5/8/8/8/8/8/4K3 w - - 0 1", White, BitBoardFromSquares(C7)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Must(board.FromFEN(tt.fen))
			pawns := pawns{}
			pawns.fromBoard(b)

			assert.Equal(t, tt.want, pawns.passers(tt.color), "fen %s color %v", tt.fen, tt.color)
		})
	}
}

func TestDoubledPawns(t *testing.T) {
	tests := [...]struct {
		name  string
		fen   string
		color Color
		want  BitBoard
	}{
		{"startpos white", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", White, 0},
		{"startpos black", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", Black, 0},
		{"empty board", "4k3/8/8/8/8/8/8/4K3 w - - 0 1", White, 0},
		{"single pawn", "4k3/8/8/8/8/8/1P6/4K3 w - - 0 1", White, 0},
		{"two pawns stacked vertically", "4k3/8/8/8/8/1P6/1P6/4K3 w - - 0 1", White, BitBoardFromSquares(B2)},
		{"three pawns stacked vertically", "4k3/8/8/8/1P6/1P6/1P6/4K3 w - - 0 1", White, BitBoardFromSquares(B2, B3)},
		{"two pawns on same file with gap", "4k3/8/8/8/1P6/8/1P6/4K3 w - - 0 1", White, BitBoardFromSquares(B2)},
		{
			"two pawns on same file with enemy pawn in gap",
			"4k3/8/8/8/1P6/1p6/1P6/4K3 w - - 0 1",
			White,
			BitBoardFromSquares(B2),
		},
		{"connected pawns", "4k3/8/8/8/8/2P5/1P6/4K3 w - - 0 1", White, 0},
		{"doubled on a file", "4k3/8/8/8/P7/P7/8/4K3 w - - 0 1", White, BitBoardFromSquares(A3)},
		{"doubled on h file", "4k3/8/7P/7P/8/8/8/4K3 w - - 0 1", White, BitBoardFromSquares(H5)},
		{"two sets of doubled pawns", "4k3/8/8/P7/P7/1P6/1P6/4K3 w - - 0 1", White, BitBoardFromSquares(B2, A4)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Must(board.FromFEN(tt.fen))
			pawns := pawns{}
			pawns.fromBoard(b)

			assert.Equal(t, tt.want, pawns.doubledPawns(tt.color), "fen %s color %v", tt.fen, tt.color)
		})
	}
}

func TestIsolatedPawns(t *testing.T) {
	tests := [...]struct {
		name  string
		fen   string
		color Color
		want  BitBoard
	}{
		{"startpos white", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", White, 0},
		{"startpos black", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", Black, 0},
		{"empty board", "4k3/8/8/8/8/8/8/4K3 w - - 0 1", White, 0},
		{"single isolated pawn on a-file", "4k3/8/8/8/8/P7/8/4K3 w - - 0 1", White, BitBoardFromSquares(A3)},
		{"single isolated pawn on h-file", "4k3/8/8/8/8/8/7P/4K3 w - - 0 1", White, BitBoardFromSquares(H2)},
		{"single isolated pawn on e-file", "4k3/8/8/8/8/4P3/8/4K3 w - - 0 1", White, BitBoardFromSquares(E3)},
		{"connected pawns", "4k3/8/8/4P3/8/8/5P2/4K3 w - - 0 1", White, 0},
		{"pawn connected to enemy pawn", "4k3/8/8/4Pp2/8/8/8/4K3 w - - 0 1", White, BitBoardFromSquares(E5)},
		{"pawn on edge supported by chain", "4k3/8/8/8/P7/1P6/8/4K3 w - - 0 1", White, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Must(board.FromFEN(tt.fen))
			pawns := pawns{}
			pawns.fromBoard(b)
			assert.Equal(t, tt.want, pawns.isolatedPawns(tt.color), "fen %s color %v", tt.fen, tt.color)
		})
	}
}

func TestBackwardPawns(t *testing.T) {
	tests := [...]struct {
		name  string
		fen   string
		color Color
		want  BitBoard
	}{
		{"startpos white", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", White, 0},
		{"startpos black", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", Black, 0},
		{"empty board", "4k3/8/8/8/8/8/8/4K3 w - - 0 1", White, 0},

		{"white isolated pawn blocked", "4k3/8/8/8/1p6/1P6/8/4K3 w - - 0 1", White, BitBoardFromSquares(B3)},
		{"white isolated pawn not blocked", "4k3/8/8/8/8/1P6/8/4K3 w - - 0 1", White, 0},
		{"white isolated pawn push threatened", "4k3/8/8/2p5/8/1P6/8/4K3 w - - 0 1", White, BitBoardFromSquares(B3)},
		{"white phalanx pawn threatened", "4k3/8/8/2p5/8/1PP5/8/4K3 w - - 0 1", White, 0},
		{"white connected backward blocked", "4k3/8/8/8/1Pp5/2P5/8/4K3 w - - 0 1", White, BitBoardFromSquares(C3)},
		{"white connected backward threatened", "4k3/8/8/3p4/1P6/2P5/8/4K3 w - - 0 1", White, BitBoardFromSquares(C3)},
		{"white V threatened", "4k3/8/8/3p4/1P1P4/2P5/8/4K3 w - - 0 1", White, BitBoardFromSquares(C3)},
		{"white V blocked", "4k3/8/8/8/1PpP4/2P5/8/4K3 w - - 0 1", White, BitBoardFromSquares(C3)},

		{"black isolated pawn blocked", "4k3/8/8/8/1p6/1P6/8/4K3 w - - 0 1", Black, BitBoardFromSquares(B4)},
		{"black isolated pawn not blocked", "4k3/8/1p6/8/8/8/8/4K3 w - - 0 1", Black, 0},
		{"black isolated pawn push threatened", "4k3/8/8/2p5/8/1P6/8/4K3 w - - 0 1", Black, BitBoardFromSquares(C5)},
		{"black phalanx pawn threatened", "4k3/8/8/1pp5/8/2P5/8/4K3 w - - 0 1", Black, 0},
		{"black connected backward blocked", "4k3/8/2p5/1pP5/8/8/8/4K3 w - - 0 1", Black, BitBoardFromSquares(C6)},
		{"black connected backward threatened", "4k3/8/8/3p4/4p3/2P5/8/4K3 w - - 0 1", Black, BitBoardFromSquares(D5)},
		{"black V threatened", "4k3/8/2p5/1p1p4/3P4/8/8/4K3 w - - 0 1", Black, BitBoardFromSquares(C6)},
		{"black V blocked", "4k3/8/2p5/1pPp4/8/8/8/4K3 w - - 0 1", Black, BitBoardFromSquares(C6)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Must(board.FromFEN(tt.fen))
			pawns := pawns{}
			pawns.fromBoard(b)

			assert.Equal(t, tt.want, pawns.backwardPawns(tt.color), "fen %s color %v", tt.fen, tt.color)
		})
	}
}
