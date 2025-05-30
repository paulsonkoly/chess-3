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

func TestPawnSinglePushMoves(t * testing.T) {
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
		want []move.Move
	}{
		{
			name: "simple king move",
			b:    Must(board.FromFEN("8/8/8/8/8/4K3/8/k7 w - - 0 1")),
			want: []move.Move{
				K(E3, D2), K(E3, E2), K(E3, F2),
				K(E3, D3), K(E3, F3),
				K(E3, D4), K(E3, E4), K(E3, F4),
			},
		},
		{
			name: "king in the corner",
			b:    Must(board.FromFEN("8/8/8/8/8/8/K7/7k b - - 0 1")),
			want: []move.Move{
				K(H1, H2), K(H1, G2), K(H1, G1),
			},
		},
		{
			name: "simple knight move",
			b:    Must(board.FromFEN("8/8/8/8/8/4N3/8/k6K w - - 0 1")),
			want: []move.Move{
				N(E3, C4), N(E3, D5), N(E3, F5), N(E3, G4),
				N(E3, C2), N(E3, D1), N(E3, F1), N(E3, G2),
				K(H1, G1), K(H1, G2), K(H1, H2),
			},
		},
		{
			name: "knight in the corner",
			b:    Must(board.FromFEN("k7/8/8/8/8/8/8/K6N w - - 0 1")),
			want: []move.Move{
				N(H1, F2), N(H1, G3),
				K(A1, A2), K(A1, B2), K(A1, B1),
			},
		},
		{
			name: "simple bishop move",
			b:    Must(board.FromFEN("k7/8/8/8/8/3B4/8/7K w - - 0 1")),
			want: []move.Move{
				B(D3, C2), B(D3, B1), B(D3, E2), B(D3, F1), B(D3, C4), B(D3, B5),
				B(D3, A6), B(D3, E4), B(D3, F5), B(D3, G6), B(D3, H7),
				K(H1, G1), K(H1, G2), K(H1, H2),
			},
		},
		{
			name: "bishop in the corner",
			b:    Must(board.FromFEN("k7/8/8/8/8/8/8/B6K w - - 0 1")),
			want: []move.Move{
				B(A1, B2), B(A1, C3), B(A1, D4), B(A1, E5), B(A1, F6), B(A1, G7), B(A1, H8),
				K(H1, G1), K(H1, G2), K(H1, H2),
			},
		},
		{
			name: "bishop blocked by friendly",
			b:    Must(board.FromFEN("k7/8/8/8/8/2K5/1B6/8 w - - 0 1")),
			want: []move.Move{
				B(B2, A3), B(B2, A1), B(B2, C1),
				K(C3, B3), K(C3, B4), K(C3, C2), K(C3, C4), K(C3, D2), K(C3, D3), K(C3, D4),
			},
		},
		{
			name: "simple rook move",
			b:    Must(board.FromFEN("k7/8/8/8/4R3/8/8/7K w - - 0 1")),
			want: []move.Move{
				R(E4, D4), R(E4, C4), R(E4, B4), R(E4, A4), R(E4, H4), R(E4, G4), R(E4, F4),
				R(E4, E5), R(E4, E6), R(E4, E7), R(E4, E8), R(E4, E3), R(E4, E2), R(E4, E1),
				K(H1, G1), K(H1, G2), K(H1, H2),
			},
		},
		{
			name: "rook in the corner",
			b:    Must(board.FromFEN("k7/8/8/8/8/8/8/R6K w - - 0 1")),
			want: []move.Move{
				R(A1, A2), R(A1, A3), R(A1, A4), R(A1, A5), R(A1, A6), R(A1, A7), R(A1, A8),
				R(A1, B1), R(A1, C1), R(A1, D1), R(A1, E1), R(A1, F1), R(A1, G1),
				K(H1, G1), K(H1, G2), K(H1, H2),
			},
		},
		{
			name: "rook blocked by friendly",
			b:    Must(board.FromFEN("k7/8/8/8/8/2K5/2R5/8 w - - 0 1")),
			want: []move.Move{
				R(C2, B2), R(C2, A2), R(C2, D2), R(C2, E2), R(C2, F2), R(C2, G2), R(C2, H2),
				R(C2, C1),
				K(C3, B3), K(C3, B4), K(C3, B2), K(C3, C4), K(C3, D2), K(C3, D3), K(C3, D4),
			},
		},
		{
			name: "simple queen move",
			b:    Must(board.FromFEN("k7/8/8/8/4Q3/8/8/7K w - - 0 1")),
			want: []move.Move{
				Q(E4, D4), Q(E4, C4), Q(E4, B4), Q(E4, A4), Q(E4, H4), Q(E4, G4), Q(E4, F4),
				Q(E4, E5), Q(E4, E6), Q(E4, E7), Q(E4, E8), Q(E4, E3), Q(E4, E2), Q(E4, E1),
				Q(E4, F5), Q(E4, G6), Q(E4, H7), Q(E4, F3), Q(E4, G2),
				Q(E4, D5), Q(E4, C6), Q(E4, B7), Q(E4, A8), Q(E4, D3), Q(E4, C2), Q(E4, B1),
				K(H1, G1), K(H1, G2), K(H1, H2),
			},
		},
		{
			name: "queen in the corner",
			b:    Must(board.FromFEN("k7/8/8/8/8/8/8/Q6K w - - 0 1")),
			want: []move.Move{
				Q(A1, A2), Q(A1, A3), Q(A1, A4), Q(A1, A5), Q(A1, A6), Q(A1, A7), Q(A1, A8),
				Q(A1, B1), Q(A1, C1), Q(A1, D1), Q(A1, E1), Q(A1, F1), Q(A1, G1),
				Q(A1, B2), Q(A1, C3), Q(A1, D4), Q(A1, E5), Q(A1, F6), Q(A1, G7), Q(A1, H8),
				K(H1, G1), K(H1, G2), K(H1, H2),
			},
		},
		{
			name: "queen blocked by friendly",
			b:    Must(board.FromFEN("8/8/8/2k5/8/2K5/2Q5/8 w - - 0 1")),
			want: []move.Move{
				Q(C2, B1), Q(C2, C1), Q(C2, D1), Q(C2, A2), Q(C2, B2), Q(C2, D2), Q(C2, E2),
				Q(C2, F2), Q(C2, G2), Q(C2, H2), Q(C2, B3), Q(C2, D3), Q(C2, A4), Q(C2, E4),
				Q(C2, F5), Q(C2, G6), Q(C2, H7),
				K(C3, B2), K(C3, D2), K(C3, B3), K(C3, D3),
			},
		},
		{
			name: "single pawn push forward as white",
			b:    Must(board.FromFEN("4k3/8/8/8/4P3/8/8/K7 w - - 0 1")),
			want: []move.Move{
				P(E4, E5),
				K(A1, B1), K(A1, B2), K(A1, A2),
			},
		},
		{
			name: "single pawn push forward as black",
			b:    Must(board.FromFEN("7K/8/8/8/4p3/8/8/k7 b - - 0 1")),
			want: []move.Move{
				P(E4, E3),
				K(A1, B1), K(A1, B2), K(A1, A2),
			},
		},
		{
			name: "blocked pawn push forward as white",
			b:    Must(board.FromFEN("8/8/8/4k3/4P3/8/8/K7 w - - 0 1")),
			want: []move.Move{
				K(A1, B1), K(A1, B2), K(A1, A2),
			},
		},
		{
			name: "blocked pawn push forward as black",
			b:    Must(board.FromFEN("8/8/8/8/4p3/4K3/8/k7 b - - 0 1")),
			want: []move.Move{
				K(A1, B1), K(A1, B2), K(A1, A2),
			},
		},
		{
			name: "double pawn push forward as white",
			b:    Must(board.FromFEN("7k/8/8/8/8/8/4P3/K7 w - - 0 1")),
			want: []move.Move{
				P(E2, E3), P(E2, E4),
				K(A1, B1), K(A1, B2), K(A1, A2),
			},
		},
		{
			name: "double pawn push forward as black",
			b:    Must(board.FromFEN("7K/4p3/8/8/8/8/8/k7 b - - 0 1")),
			want: []move.Move{
				P(E7, E6), P(E7, E5),
				K(A1, B1), K(A1, B2), K(A1, A2),
			},
		},
		{
			name: "double pawn push blocked by a piece directly in front",
			b:    Must(board.FromFEN("8/8/8/8/8/4k3/4P3/K7 w - - 0 1")),
			want: []move.Move{
				K(A1, B1), K(A1, B2), K(A1, A2),
			},
		},
		{
			name: "double pawn push blocked by a piece 2 squares in front",
			b:    Must(board.FromFEN("8/8/8/8/4k3/8/4P3/K7 w - - 0 1")),
			want: []move.Move{
				P(E2, E3),
				K(A1, B1), K(A1, B2), K(A1, A2),
			},
		},
		{
			name: "pawn capture",
			b:    Must(board.FromFEN("7k/8/8/8/3n4/4P3/8/K7 w - - 0 1")),
			want: []move.Move{
				P(E3, E4), P(E3, D4),
				K(A1, B1), K(A1, B2), K(A1, A2),
			},
		},
		{
			name: "pawn capture on AFile testing for wrap to HFile",
			b:    Must(board.FromFEN("7k/8/8/8/1n5n/P7/8/K7 w - - 0 1")),
			want: []move.Move{
				P(A3, A4), P(A3, B4),
				K(A1, B1), K(A1, B2),
			},
		},
		{
			name: "pawn promotion (push)",
			b:    Must(board.FromFEN("7k/4P3/8/8/8/8/8/K7 w - - 0 1")),
			want: []move.Move{
				PP(E7, E8, Queen), PP(E7, E8, Rook), PP(E7, E8, Bishop), PP(E7, E8, Knight),
				K(A1, B1), K(A1, B2), K(A1, A2),
			},
		},
		{
			name: "pawn promotion blocked (push)",
			b:    Must(board.FromFEN("8/8/8/8/8/8/6p1/k5K1 b - - 0 1")),
			want: []move.Move{
				K(A1, B1), K(A1, B2), K(A1, A2),
			},
		},
		{
			name: "pawn promotion (capture)",
			b:    Must(board.FromFEN("3nn2k/4P3/8/8/8/8/8/K7 w - - 0 1")),
			want: []move.Move{
				PP(E7, D8, Queen), PP(E7, D8, Rook), PP(E7, D8, Bishop), PP(E7, D8, Knight),
				K(A1, B1), K(A1, B2), K(A1, A2),
			},
		},
		{
			name: "pawn promotion (capture or push)",
			b:    Must(board.FromFEN("3n3k/4P3/8/8/8/8/8/K7 w - - 0 1")),
			want: []move.Move{
				PP(E7, D8, Queen), PP(E7, D8, Rook), PP(E7, D8, Bishop), PP(E7, D8, Knight),
				PP(E7, E8, Queen), PP(E7, E8, Rook), PP(E7, E8, Bishop), PP(E7, E8, Knight),
				K(A1, B1), K(A1, B2), K(A1, A2),
			},
		},
		{
			name: "en passant",
			b:    Must(board.FromFEN("7k/8/8/2Pp4/8/8/8/K7 w - d6 0 1")),
			want: []move.Move{
				P(C5, C6), P(C5, D6),
				K(A1, B1), K(A1, B2), K(A1, A2),
			},
		},

		{
			name: "regression #1",
			b:    Must(board.FromFEN("rnbqkbnr/1ppppppp/8/p7/8/7P/PPPPPPP1/RNBQKBNR w - - 0 1")),
			want: []move.Move{
				P(A2, A3), P(B2, B3), P(C2, C3), P(D2, D3), P(E2, E3), P(F2, F3), P(G2, G3), P(H3, H4),
				P(A2, A4), P(B2, B4), P(C2, C4), P(D2, D4), P(E2, E4), P(F2, F4), P(G2, G4),
				N(B1, A3), N(B1, C3), N(G1, F3),
				R(H1, H2),
			},
		},
		{
			name: "regression #2",
			b:    Must(board.FromFEN("rnbqkbnr/1ppppppp/8/p7/1P6/8/P1PPPPPP/RNBQKBNR w - - 0 1")),
			want: []move.Move{
				P(A2, A3), P(C2, C3), P(D2, D3), P(E2, E3), P(F2, F3), P(G2, G3), P(H2, H3), P(B4, B5), P(A2, A4),
				P(C2, C4), P(D2, D4), P(E2, E4), P(F2, F4), P(G2, G4), P(H2, H4), P(B4, A5),
				N(B1, A3), N(B1, C3), N(G1, F3), N(G1, H3),
				B(C1, B2), B(C1, A3),
			},
		},
		{
			name: "regression #3",
			b:    Must(board.FromFEN("rnbqkbnr/1ppppppp/8/p7/8/N6N/PPPPPPPP/R1BQKB1R b - - 0 1")),
			want: []move.Move{
				P(A5, A4), P(B7, B6), P(C7, C6), P(D7, D6), P(E7, E6), P(F7, F6), P(G7, G6), P(H7, H6), P(B7, B5),
				P(C7, C5), P(D7, D5), P(E7, E5), P(F7, F5), P(G7, G5), P(H7, H5),
				N(B8, A6), N(B8, C6), N(G8, F6), N(G8, H6),
				R(A8, A6), R(A8, A7),
			},
		},
		{
			name: "regression #4",
			b:    Must(board.FromFEN("rnbqkbnr/1ppppppp/p7/1N6/8/8/PPPPPPPP/R1BQKBNR b - - 0 1")),
			want: []move.Move{
				P(A6, A5), P(B7, B6), P(C7, C6), P(D7, D6), P(E7, E6), P(F7, F6), P(G7, G6), P(H7, H6),
				P(C7, C5), P(D7, D5), P(E7, E5), P(F7, F5), P(G7, G5), P(H7, H5), P(A6, B5),
				N(B8, C6), N(G8, F6), N(G8, H6), R(A8, A7),
			},
		},
		{
			name: "regression #5",
			b:    Must(board.FromFEN("rnbq3r/pp1Pbpkp/2p3p1/6P1/2B5/8/PPP1Nn1P/RNBQ1K1R b - - 0 1")),
			want: []move.Move{
				P(C6, C5), P(A7, A6), P(B7, B6), P(F7, F6), P(H7, H6), P(A7, A5), P(B7, B5), P(F7, F5), P(H7, H5),
				N(F2, D1), N(F2, H1), N(F2, D3), N(F2, H3), N(F2, E4), N(F2, G4), N(B8, A6), N(B8, D7),
				B(E7, A3), B(E7, B4), B(E7, C5), B(E7, G5), B(E7, D6), B(E7, F6), B(E7, F8), B(C8, D7),
				R(H8, E8), R(H8, F8), R(H8, G8),
				Q(D8, A5), Q(D8, B6), Q(D8, C7), Q(D8, D7), Q(D8, E8), Q(D8, F8), Q(D8, G8),
				K(G7, G8), K(G7, F8),
			},
		},
		{
			name: "regression #6",
			b:    Must(board.FromFEN("rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPPBN1PP/R3KN1n w Q - 0 1")),
			want: []move.Move{
				P(A2, A3), P(B2, B3), P(C2, C3), P(G2, G3), P(H2, H3), P(A2, A4), P(B2, B4), P(G2, G4), P(H2, H4),
				PP(D7, C8, Queen), PP(D7, C8, Rook), PP(D7, C8, Bishop), PP(D7, C8, Knight),
				N(F1, E3), N(F1, G3), N(E2, C1), N(E2, G1), N(E2, C3), N(E2, G3), N(E2, D4), N(E2, F4),
				B(D2, C1), B(D2, C3), B(D2, E3), B(D2, B4), B(D2, F4), B(D2, A5), B(D2, G5), B(D2, H6),
				B(C4, B3), B(C4, D3), B(C4, B5), B(C4, D5), B(C4, A6), B(C4, E6), B(C4, F7),
				R(A1, B1), R(A1, C1), R(A1, D1),
				K(E1, D1), K(E1, C1),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := tt.want
			ok := make([]bool, len(want))
			b := tt.b

			ms := move.NewStore()

			movegen.GenMoves(ms, b)

			for _, m := range ms.Frame() {
				b.MakeMove(&m)

				king := b.Colors[b.STM.Flip()] & b.Pieces[King]
				occ := b.Colors[White] | b.Colors[Black]

				if movegen.IsAttacked(b, b.STM, occ, king) {
					// illegal (pseudo-leagal) move, skip
					b.UndoMove(&m)
					continue
				}

				b.UndoMove(&m)

				m.Captured = 0
				m.EPP = 0
				m.EPSq = 0
				m.Castle = 0
				m.CRights = 0
				m.FiftyCnt = 0
				ix := slices.Index(want, m)
				if ix == -1 {
					t.Errorf("unexpected move %s%s generated", m.Piece, m)
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

func TestGenForcing(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		b    *board.Board
		want []move.Move
	}{
		{
			name: "king captures",
			b:    Must(board.FromFEN("8/8/7k/8/8/8/pK6/B7 w - - 0 1")),
			want: []move.Move{
				K(B2, A2),
			},
		},
		{
			name: "knight captures",
			b:    Must(board.FromFEN("8/6k1/8/8/8/p7/8/1N1K4 w - - 0 1")),
			want: []move.Move{
				N(B1, A3),
			},
		},
		{
			name: "bishop captures",
			b:    Must(board.FromFEN("K7/1B6/p6k/8/8/8/8/8 w - - 0 1")),
			want: []move.Move{
				B(B7, A6),
			},
		},
		{
			name: "pawn captures",
			b:    Must(board.FromFEN("6k1/8/8/8/8/p2PP2P/PP3P2/KB6 w - - 0 1")),
			want: []move.Move{
				P(B2, A3),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := tt.want
			ok := make([]bool, len(want))
			b := tt.b

			ms := move.NewStore()

			movegen.GenForcing(ms, b)

			for _, m := range ms.Frame() {
				b.MakeMove(&m)

				king := b.Colors[b.STM.Flip()] & b.Pieces[King]
				occ := b.Colors[White] | b.Colors[Black]

				if movegen.IsAttacked(b, b.STM, occ, king) {
					// illegal (pseudo-leagal) move, skip
					b.UndoMove(&m)
					continue
				}

				b.UndoMove(&m)

				m.Captured = 0
				m.EPP = 0
				m.EPSq = 0
				m.Castle = 0
				m.CRights = 0
				m.FiftyCnt = 0
				ix := slices.Index(want, m)
				if ix == -1 {
					t.Errorf("unexpected move %s%s generated", m.Piece, m)
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
		move   move.SimpleMove
		bAfter *board.Board
	}{
		{
			name:   "En passant possible after pawn move",
			b:      Must(board.FromFEN("4k3/8/8/8/3p4/8/2P5/4K3 w - - 0 1")),
			move:   move.FromSquares(C2, C4),
			bAfter: Must(board.FromFEN("4k3/8/8/8/2Pp4/8/8/4K3 b - c3 0 1")),
		},
		{
			name:   "En passant not possible due to no pawn",
			b:      Must(board.FromFEN("4k3/8/8/8/8/8/2P5/4K3 w - - 0 1")),
			move:   move.FromSquares(C2, C4),
			bAfter: Must(board.FromFEN("4k3/8/8/8/2P5/8/8/4K3 b - - 0 1")),
		},
		{
			name:   "En passant not possible due to simple pin",
			b:      Must(board.FromFEN("8/8/1k6/8/3p4/8/2P5/3K2B1 w - - 0 1")),
			move:   move.FromSquares(C2, C4),
			bAfter: Must(board.FromFEN("8/8/1k6/8/2Pp4/8/8/3K2B1 b - - 0 1")),
		},
		{
			name:   "En passant not possible due to tricky pin",
			b:      Must(board.FromFEN("8/8/8/8/k2p3R/8/2P5/3K4 w - - 0 1")),
			move:   move.FromSquares(C2, C4),
			bAfter: Must(board.FromFEN("8/8/8/8/k1Pp3R/8/8/3K4 b - - 0 1")),
		},
		{
			name:   "En passant possible in pin that's not affected",
			b:      Must(board.FromFEN("4r3/pkp3b1/1p5p/2P1npp1/P2rp3/6PN/1P2PPBP/1RR3K1 w - - 0 22")),
			move:   move.FromSquares(F2, F4),
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

			b.MakeMove(&m)
			assert.Equal(t, tt.bAfter.EnPassant, b.EnPassant)
			assert.Equal(t, tt.bAfter.Hash(), b.Hash())
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
