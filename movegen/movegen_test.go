package movegen_test

import (
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"
	"github.com/stretchr/testify/assert"

	. "github.com/paulsonkoly/chess-3/types"
)

func TestPawnSinglePushMoves(t *testing.T) {
	tests := []struct {
		name  string
		from  board.BitBoard
		color Color
		to    board.BitBoard
	}{
		{
			name:  "pawn push white",
			from:  board.BitBoardFromSquares(B1, E5, D7, C8, G8),
			color: White,
			to:    board.BitBoardFromSquares(B2, E6, D8),
		},
		{
			name:  "pawn push black",
			from:  board.BitBoardFromSquares(B1, E5, D7, C8, G8),
			color: Black,
			to:    board.BitBoardFromSquares(E4, D6, C7, G7),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.to, movegen.PawnSinglePushMoves(tt.from, tt.color))
		})
	}

}

func TestPawnCaptureMoves(t *testing.T) {
	tests := []struct {
		name  string
		from  board.BitBoard
		color Color
		to    board.BitBoard
	}{
		{
			name:  "pawn capture white",
			from:  board.BitBoardFromSquares(B1, E5, D7, G8),
			color: White,
			to:    board.BitBoardFromSquares(A2, C2, D6, F6, C8, E8),
		},
		{
			name:  "pawn capture black",
			from:  board.BitBoardFromSquares(B1, E5, D7, G8),
			color: Black,
			to:    board.BitBoardFromSquares(D4, F4, C6, E6, F7, H7),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.to, movegen.PawnCaptureMoves(tt.from, tt.color))
		})
	}
}

func TestMoves(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		b    *board.Board
		want []string
	}{
		{
			name: "simple king move",
			b:    Must(board.FromFEN("8/8/8/8/8/4K3/8/k7 w - - 0 1")),
			want: []string{"e3d2", "e3e2", "e3f2", "e3d3", "e3f3", "e3d4", "e3e4", "e3f4"},
		},
		{
			name: "king in the corner",
			b:    Must(board.FromFEN("8/8/8/8/8/8/K7/7k b - - 0 1")),
			want: []string{"h1h2", "h1g2", "h1g1"},
		},
		{
			name: "simple knight move",
			b:    Must(board.FromFEN("8/8/8/8/8/4N3/8/k6K w - - 0 1")),
			want: []string{"e3c4", "e3d5", "e3f5", "e3g4", "e3c2", "e3d1", "e3f1", "e3g2", "h1g1", "h1g2", "h1h2"},
		},
		{
			name: "knight in the corner",
			b:    Must(board.FromFEN("k7/8/8/8/8/8/8/K6N w - - 0 1")),
			want: []string{"h1f2", "h1g3", "a1a2", "a1b2", "a1b1"},
		},
		{
			name: "simple bishop move",
			b:    Must(board.FromFEN("k7/8/8/8/8/3B4/8/7K w - - 0 1")),
			want: []string{
				"d3c2", "d3b1", "d3e2", "d3f1", "d3c4", "d3b5", "d3a6", "d3e4", "d3f5", "d3g6", "d3h7", "h1g1", "h1g2", "h1h2",
			},
		},
		{
			name: "bishop in the corner",
			b:    Must(board.FromFEN("k7/8/8/8/8/8/8/B6K w - - 0 1")),
			want: []string{"a1b2", "a1c3", "a1d4", "a1e5", "a1f6", "a1g7", "a1h8", "h1g1", "h1g2", "h1h2"},
		},
		{
			name: "bishop blocked by friendly",
			b:    Must(board.FromFEN("k7/8/8/8/8/2K5/1B6/8 w - - 0 1")),
			want: []string{"b2a3", "b2a1", "b2c1", "c3b3", "c3b4", "c3c2", "c3c4", "c3d2", "c3d3", "c3d4"},
		},
		{
			name: "simple rook move",
			b:    Must(board.FromFEN("k7/8/8/8/4R3/8/8/7K w - - 0 1")),
			want: []string{
				"e4d4", "e4c4", "e4b4", "e4a4", "e4h4", "e4g4", "e4f4", "e4e5", "e4e6", "e4e7", "e4e8", "e4e3", "e4e2", "e4e1",
				"h1g1", "h1g2", "h1h2",
			},
		},
		{
			name: "rook in the corner",
			b:    Must(board.FromFEN("k7/8/8/8/8/8/8/R6K w - - 0 1")),
			want: []string{
				"a1a2", "a1a3", "a1a4", "a1a5", "a1a6", "a1a7", "a1a8", "a1b1", "a1c1", "a1d1", "a1e1", "a1f1", "a1g1", "h1g1",
				"h1g2", "h1h2",
			},
		},
		{
			name: "rook blocked by friendly",
			b:    Must(board.FromFEN("k7/8/8/8/8/2K5/2R5/8 w - - 0 1")),
			want: []string{
				"c2b2", "c2a2", "c2d2", "c2e2", "c2f2", "c2g2", "c2h2", "c2c1", "c3b3", "c3b4", "c3b2", "c3c4", "c3d2", "c3d3",
				"c3d4",
			},
		},
		{
			name: "simple queen move",
			b:    Must(board.FromFEN("k7/8/8/8/4Q3/8/8/7K w - - 0 1")),
			want: []string{
				"e4d4", "e4c4", "e4b4", "e4a4", "e4h4", "e4g4", "e4f4", "e4e5", "e4e6", "e4e7", "e4e8", "e4e3", "e4e2", "e4e1",
				"e4f5", "e4g6", "e4h7", "e4f3", "e4g2", "e4d5", "e4c6", "e4b7", "e4a8", "e4d3", "e4c2", "e4b1", "h1g1", "h1g2",
				"h1h2",
			},
		},
		{
			name: "queen in the corner",
			b:    Must(board.FromFEN("k7/8/8/8/8/8/8/Q6K w - - 0 1")),
			want: []string{
				"a1a2", "a1a3", "a1a4", "a1a5",
				"a1a6", "a1a7", "a1a8",
				"a1b1", "a1c1", "a1d1", "a1e1",
				"a1f1", "a1g1",
				"a1b2", "a1c3", "a1d4", "a1e5",
				"a1f6", "a1g7", "a1h8",
				"h1g1", "h1g2", "h1h2",
			},
		},
		{
			name: "queen blocked by friendly",
			b:    Must(board.FromFEN("8/8/8/2k5/8/2K5/2Q5/8 w - - 0 1")),
			want: []string{
				"c2b1", "c2c1", "c2d1", "c2a2", "c2b2", "c2d2", "c2e2", "c2f2", "c2g2", "c2h2", "c2b3", "c2d3", "c2a4", "c2e4",
				"c2f5", "c2g6", "c2h7", "c3b2", "c3d2", "c3b3", "c3d3",
			},
		},
		{
			name: "single pawn push forward as white",
			b:    Must(board.FromFEN("4k3/8/8/8/4P3/8/8/K7 w - - 0 1")),
			want: []string{"e4e5", "a1b1", "a1b2", "a1a2"},
		},
		{
			name: "single pawn push forward as black",
			b:    Must(board.FromFEN("7K/8/8/8/4p3/8/8/k7 b - - 0 1")),
			want: []string{"e4e3", "a1b1", "a1b2", "a1a2"},
		},
		{
			name: "blocked pawn push forward as white",
			b:    Must(board.FromFEN("8/8/8/4k3/4P3/8/8/K7 w - - 0 1")),
			want: []string{"a1b1", "a1b2", "a1a2"},
		},
		{
			name: "blocked pawn push forward as black",
			b:    Must(board.FromFEN("8/8/8/8/4p3/4K3/8/k7 b - - 0 1")),
			want: []string{"a1b1", "a1b2", "a1a2"},
		},
		{
			name: "double pawn push forward as white",
			b:    Must(board.FromFEN("7k/8/8/8/8/8/4P3/K7 w - - 0 1")),
			want: []string{"e2e3", "e2e4", "a1b1", "a1b2", "a1a2"},
		},
		{
			name: "double pawn push forward as black",
			b:    Must(board.FromFEN("7K/4p3/8/8/8/8/8/k7 b - - 0 1")),
			want: []string{"e7e6", "e7e5", "a1b1", "a1b2", "a1a2"},
		},
		{
			name: "double pawn push blocked by a piece directly in front",
			b:    Must(board.FromFEN("8/8/8/8/8/4k3/4P3/K7 w - - 0 1")),
			want: []string{"a1b1", "a1b2", "a1a2"},
		},
		{
			name: "double pawn push blocked by a piece 2 squares in front",
			b:    Must(board.FromFEN("8/8/8/8/4k3/8/4P3/K7 w - - 0 1")),
			want: []string{"e2e3", "a1b1", "a1b2", "a1a2"},
		},
		{
			name: "pawn capture",
			b:    Must(board.FromFEN("7k/8/8/8/3n4/4P3/8/K7 w - - 0 1")),
			want: []string{"e3e4", "e3d4", "a1b1", "a1b2", "a1a2"},
		},
		{
			name: "pawn capture on AFile testing for wrap to HFile",
			b:    Must(board.FromFEN("7k/8/8/8/1n5n/P7/8/K7 w - - 0 1")),
			want: []string{"a3a4", "a3b4", "a1b1", "a1b2"},
		},
		{
			name: "pawn promotion (push)",
			b:    Must(board.FromFEN("7k/4P3/8/8/8/8/8/K7 w - - 0 1")),
			want: []string{"e7e8q", "e7e8r", "e7e8b", "e7e8n", "a1b1", "a1b2", "a1a2"},
		},
		{
			name: "pawn promotion blocked (push)",
			b:    Must(board.FromFEN("8/8/8/8/8/8/6p1/k5K1 b - - 0 1")),
			want: []string{"a1b1", "a1b2", "a1a2"},
		},
		{
			name: "pawn promotion (capture)",
			b:    Must(board.FromFEN("3nn2k/4P3/8/8/8/8/8/K7 w - - 0 1")),
			want: []string{"e7d8q", "e7d8r", "e7d8b", "e7d8n", "a1b1", "a1b2", "a1a2"},
		},
		{
			name: "pawn promotion (capture or push)",
			b:    Must(board.FromFEN("3n3k/4P3/8/8/8/8/8/K7 w - - 0 1")),
			want: []string{"e7d8q", "e7d8r", "e7d8b", "e7d8n", "e7e8q", "e7e8r", "e7e8b", "e7e8n", "a1b1", "a1b2", "a1a2"},
		},
		{
			name: "en passant",
			b:    Must(board.FromFEN("7k/8/8/2Pp4/8/8/8/K7 w - d6 0 1")),
			want: []string{"c5c6", "c5d6", "a1b1", "a1b2", "a1a2"},
		},
		{
			name: "regression #1",
			b:    Must(board.FromFEN("rnbqkbnr/1ppppppp/8/p7/8/7P/PPPPPPP1/RNBQKBNR w - - 0 1")),
			want: []string{
				"a2a3", "b2b3", "c2c3", "d2d3", "e2e3", "f2f3", "g2g3", "h3h4", "a2a4", "b2b4", "c2c4", "d2d4", "e2e4", "f2f4",
				"g2g4", "b1a3", "b1c3", "g1f3", "h1h2",
			},
		},
		{
			name: "regression #2",
			b:    Must(board.FromFEN("rnbqkbnr/1ppppppp/8/p7/1P6/8/P1PPPPPP/RNBQKBNR w - - 0 1")),
			want: []string{
				"a2a3", "c2c3", "d2d3", "e2e3", "f2f3", "g2g3", "h2h3", "b4b5", "a2a4", "c2c4", "d2d4", "e2e4", "f2f4", "g2g4",
				"h2h4", "b4a5", "b1a3", "b1c3", "g1f3", "g1h3", "c1b2", "c1a3",
			},
		},
		{
			name: "regression #3",
			b:    Must(board.FromFEN("rnbqkbnr/1ppppppp/8/p7/8/N6N/PPPPPPPP/R1BQKB1R b - - 0 1")),
			want: []string{
				"a5a4", "b7b6", "c7c6", "d7d6", "e7e6", "f7f6", "g7g6", "h7h6", "b7b5", "c7c5", "d7d5", "e7e5", "f7f5", "g7g5",
				"h7h5", "b8a6", "b8c6", "g8f6", "g8h6", "a8a6", "a8a7",
			},
		},
		{
			name: "regression #4",
			b:    Must(board.FromFEN("rnbqkbnr/1ppppppp/p7/1N6/8/8/PPPPPPPP/R1BQKBNR b - - 0 1")),
			want: []string{
				"a6a5", "b7b6", "c7c6", "d7d6", "e7e6", "f7f6", "g7g6", "h7h6", "c7c5", "d7d5", "e7e5", "f7f5", "g7g5", "h7h5",
				"a6b5", "b8c6", "g8f6", "g8h6", "a8a7",
			},
		},
		{
			name: "regression #5",
			b:    Must(board.FromFEN("rnbq3r/pp1Pbpkp/2p3p1/6P1/2B5/8/PPP1Nn1P/RNBQ1K1R b - - 0 1")),
			want: []string{
				"c6c5", "a7a6", "b7b6", "f7f6", "h7h6", "a7a5", "b7b5", "f7f5", "h7h5", "f2d1", "f2h1", "f2d3", "f2h3", "f2e4",
				"f2g4", "b8a6", "b8d7", "e7a3", "e7b4", "e7c5", "e7g5", "e7d6", "e7f6", "e7f8", "c8d7", "h8e8", "h8f8", "h8g8",
				"d8a5", "d8b6", "d8c7", "d8d7", "d8e8", "d8f8", "d8g8", "g7g8", "g7f8",
			},
		},
		{
			name: "regression #6",
			b:    Must(board.FromFEN("rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPPBN1PP/R3KN1n w Q - 0 1")),
			want: []string{
				"a2a3", "b2b3", "c2c3", "g2g3", "h2h3", "a2a4", "b2b4", "g2g4", "h2h4", "d7c8q", "d7c8r", "d7c8b", "d7c8n",
				"f1e3", "f1g3", "e2c1", "e2g1", "e2c3", "e2g3", "e2d4", "e2f4", "d2c1", "d2c3", "d2e3", "d2b4", "d2f4", "d2a5",
				"d2g5", "d2h6", "c4b3", "c4d3", "c4b5", "c4d5", "c4a6", "c4e6", "c4f7", "a1b1", "a1c1", "a1d1", "e1d1", "e1c1",
			},
		},
	}

	ms := move.NewStore()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.b

			ms.Push()
			defer ms.Pop()

			movegen.GenMoves(ms, b)

			filter := ms.Frame()[:0]
			for _, m := range ms.Frame() {
				r := b.MakeMove(m.Move)

				if movegen.InCheck(b, b.STM.Flip()) {
					b.UndoMove(m.Move, r)
					continue
				}
				b.UndoMove(m.Move, r)

				filter = append(filter, m)
			}

			uciStrs := make([]string, 0, len(filter))

			for _, m := range filter {
				uciStrs = append(uciStrs, m.String())
			}

			assert.ElementsMatch(t, uciStrs, tt.want)
		})
	}
}

func TestGenForcing(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		b    *board.Board
		want []string
	}{
		{
			name: "king captures",
			b:    Must(board.FromFEN("8/8/7k/8/8/8/pK6/B7 w - - 0 1")),
			want: []string{"b2a2"},
		},
		{
			name: "knight captures",
			b:    Must(board.FromFEN("8/6k1/8/8/8/p7/8/1N1K4 w - - 0 1")),
			want: []string{"b1a3"},
		},
		{
			name: "bishop captures",
			b:    Must(board.FromFEN("K7/1B6/p6k/8/8/8/8/8 w - - 0 1")),
			want: []string{"b7a6"},
		},
		{
			name: "pawn captures",
			b:    Must(board.FromFEN("6k1/8/8/8/8/p2PP2P/PP3P2/KB6 w - - 0 1")),
			want: []string{"b2a3"},
		},
	}

	ms := move.NewStore()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.b

			ms.Push()
			defer ms.Pop()

			movegen.GenForcing(ms, b)

			filter := ms.Frame()[:0]
			for _, m := range ms.Frame() {
				r := b.MakeMove(m.Move)

				if movegen.InCheck(b, b.STM.Flip()) {
					b.UndoMove(m.Move, r)
					continue
				}
				b.UndoMove(m.Move, r)

				filter = append(filter, m)
			}

			uciStrs := make([]string, 0, len(filter))

			for _, m := range filter {
				uciStrs = append(uciStrs, m.String())
			}

			assert.ElementsMatch(t, uciStrs, tt.want)
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
			b:    Must(board.FromFEN("7K/5P2/7k/8/8/8/8/6r1 w - - 0 1")),
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
			assert.Equal(t, tt.want, movegen.IsStalemate(tt.b))
		})
	}
}

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
			name: "somthered mate",
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
			assert.Equal(t, tt.want, movegen.IsCheckmate(tt.b))
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
		ms := move.NewStore()

		t.Run(tt.name, func(t *testing.T) {
			b := tt.b
			m := movegen.FromSimple(b, tt.move)

			ms.Push()
			movegen.GenMoves(ms, b)
			assert.Contains(t, ms.Frame(), m)
			ms.Pop()

			b.MakeMove(m.Move)
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
		target board.BitBoard
		want   bool
	}{
		{
			name:   "king not in check",
			b:      Must(board.FromFEN("8/1k6/8/8/8/8/8/RNBQKBNR w - - 0 1")),
			by:     White,
			target: board.BitBoardFromSquares(B7),
			want:   false,
		},
		{
			name:   "king in check by knight",
			b:      Must(board.FromFEN("8/8/8/8/8/2k5/8/RNBQKBNR w - - 0 1")),
			by:     White,
			target: board.BitBoardFromSquares(C3),
			want:   true,
		},
		{
			name:   "king in check by bishop",
			b:      Must(board.FromFEN("8/8/8/8/8/4k3/8/RNBQKBNR w - - 0 1")),
			by:     White,
			target: board.BitBoardFromSquares(E3),
			want:   true,
		},
		{
			name:   "bishop does not attack through a blocking piece",
			b:      Must(board.FromFEN("8/8/8/8/8/4k3/3N4/R1BQKBNR w - - 0 1")),
			by:     White,
			target: board.BitBoardFromSquares(E3),
			want:   false,
		},
		{
			name:   "king in check by rook",
			b:      Must(board.FromFEN("k7/8/8/8/8/8/8/RNBQKBNR w - - 0 1")),
			by:     White,
			target: board.BitBoardFromSquares(A8),
			want:   true,
		},
		{
			name:   "rook does not attack through a blocking piece",
			b:      Must(board.FromFEN("k7/8/8/8/8/N7/8/R1BQKBNR w - - 0 1")),
			by:     White,
			target: board.BitBoardFromSquares(A8),
			want:   false,
		},
		{
			name:   "king in check by queen",
			b:      Must(board.FromFEN("8/8/8/8/3k4/8/8/RNBQKBNR w - - 0 1")),
			by:     White,
			target: board.BitBoardFromSquares(D4),
			want:   true,
		},
		{
			name:   "queen does not attack through a blocking piece",
			b:      Must(board.FromFEN("8/8/8/8/6k1/8/4N3/RNBQKB1R w - - 0 1")),
			by:     White,
			target: board.BitBoardFromSquares(G4),
			want:   false,
		},
		{
			name:   "king in check by pawn",
			b:      Must(board.FromFEN("8/8/8/8/5k2/4P3/8/K7 w - - 0 1")),
			by:     White,
			target: board.BitBoardFromSquares(F4),
			want:   true,
		},
		{
			name:   "king not in check by pawn wrap",
			b:      Must(board.FromFEN("8/8/8/8/7k/P7/8/K7 w - - 0 1")),
			by:     White,
			target: board.BitBoardFromSquares(H4),
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			occ := tt.b.Colors[White] | tt.b.Colors[Black]
			assert.Equal(t, tt.want, movegen.IsAttacked(tt.b, tt.by, occ, tt.target))
		})
	}
}
