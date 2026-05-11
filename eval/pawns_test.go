package eval

import (
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/stretchr/testify/assert"

	. "github.com/paulsonkoly/chess-3/chess"
)

func TestOutposts(t *testing.T) {
	tests := [...]struct {
		name  string
		fen   string
		color Color
		want  BitBoard
	}{
		{"startpos white", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", White, 0},
		{"startpos black", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", Black, 0},
		{
			"complicated pawn structure with white",
			"4k3/3p4/2pp1p2/P6P/8/6P1/8/4K3 w - - 0 1",
			White,
			BitBoardFromSquares(B6, F4, H4, G6),
		},
		{
			"complicated pawn structure with black",
			"4k3/8/p7/6p1/4p3/4P3/1P4PP/4K3 w - - 0 1",
			Black,
			BitBoardFromSquares(B5, D3),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Must(board.FromFEN(tt.fen))
			e := New[Score]()
			e.Score(b, &Coefficients)

			assert.Equal(t, tt.want, e.outposts(tt.color), "fen %s color %v", tt.fen, tt.color)
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
			e := New[Score]()
			e.pawns[White].calc(b, White)
			e.pawns[Black].calc(b, Black)

			assert.Equal(t, tt.want, e.passers(tt.color), "fen %s color %v", tt.fen, tt.color)
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
			e := New[Score]()
			e.pawns[White].calc(b, White)
			e.pawns[Black].calc(b, Black)

			assert.Equal(
				t,
				tt.want,
				e.doubledPawns(b.Colors[tt.color]&b.Pieces[Pawn], tt.color),
				"fen %s color %v",
				tt.fen,
				tt.color,
			)
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
			e := New[Score]()
			e.pawns[White].calc(b, White)
			e.pawns[Black].calc(b, Black)

			assert.Equal(t, tt.want, e.isolatedPawns(
				b.Colors[tt.color]&b.Pieces[Pawn],
				tt.color,
			), "fen %s color %v", tt.fen, tt.color)
		})
	}
}

func TestFrontLine(t *testing.T) {
	tests := [...]struct {
		name  string
		fen   string
		color Color
		want  BitBoard
	}{
		{"startpos white", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", White, SecondRankBB},
		{"startpos black", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", Black, SeventhRankBB},
		{"empty board", "4k3/8/8/8/8/8/8/4K3 w - - 0 1", White, 0},
		{
			"white complex example",
			"4k3/8/6p1/3P2P1/P5P1/P1P1P1P1/4P3/4K3 w - - 0 1",
			White,
			BitBoardFromSquares(A4, C3, D5, E3, G5),
		},
		{
			"black complex example",
			"4k3/8/2p3p1/p1p3p1/2P1p1p1/p7/8/4K3 b - - 0 1",
			Black,
			BitBoardFromSquares(A3, C5, E4, G4),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Must(board.FromFEN(tt.fen))
			pawns := Pawns{}
			pawns.calc(b, tt.color)
			assert.Equal(t, tt.want, pawns.frontline, "fen %s color %v", tt.fen, tt.color)
		})
	}
}

func TestBackMost(t *testing.T) {
	tests := [...]struct {
		name  string
		fen   string
		color Color
		want  BitBoard
	}{
		{"startpos white", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", White, SecondRankBB},
		{"startpos black", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", Black, SeventhRankBB},
		{"empty board", "4k3/8/8/8/8/8/8/4K3 w - - 0 1", White, 0},
		{
			"white complex example",
			"4k3/8/6p1/3P2P1/P5P1/P1P1P1P1/4P3/4K3 w - - 0 1",
			White,
			BitBoardFromSquares(A3, C3, D5, E2, G3),
		},
		{
			"black complex example",
			"4k3/8/p2p2p1/p1pp4/3p4/8/P7/K7 w - - 0 1",
			Black,
			BitBoardFromSquares(A6, C5, D6, G6),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Must(board.FromFEN(tt.fen))
			pawns := Pawns{}
			pawns.calc(b, tt.color)
			assert.Equal(t, tt.want, pawns.backmost, "fen %s color %v", tt.fen, tt.color)
		})
	}
}
