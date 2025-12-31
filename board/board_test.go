package board_test

import (
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"

	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/stretchr/testify/assert"
)

func TestCastle(t *testing.T) {
	b := Must(board.FromFEN("k7/p7/8/8/8/8/8/R3K2R w KQ - 0 1"))
	m := move.New(E1, G1)

	r := b.MakeMove(m)

	assert.Equal(t, Castles(0), b.Castles)
	assert.Equal(t, NoPiece, b.SquaresToPiece[E1])
	assert.Equal(t, Rook, b.SquaresToPiece[F1])
	assert.Equal(t, King, b.SquaresToPiece[G1])
	assert.Equal(t, NoPiece, b.SquaresToPiece[H1])

	b.UndoMove(m, r)

	assert.Equal(t, ShortWhite|LongWhite, b.Castles)
	assert.Equal(t, King, b.SquaresToPiece[E1])
	assert.Equal(t, NoPiece, b.SquaresToPiece[F1])
	assert.Equal(t, NoPiece, b.SquaresToPiece[G1])
	assert.Equal(t, Rook, b.SquaresToPiece[H1])

	m = move.New(E1, F1)

	r = b.MakeMove(m)

	assert.Equal(t, Castles(0), b.Castles)

	b.UndoMove(m, r)

	assert.Equal(t, ShortWhite|LongWhite, b.Castles)

	m = move.New(A1, B1)

	r = b.MakeMove(m)

	assert.Equal(t, ShortWhite, b.Castles)

	b.UndoMove(m, r)

	assert.Equal(t, ShortWhite|LongWhite, b.Castles)
}

func TestInvalidPieceCount(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		want bool
	}{
		{
			name: "startpos",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			want: false,
		},
		{
			name: "8 queens",
			fen:  "3k4/8/8/8/8/3K4/QQQQQQQQ/8 w - - 0 1",
			want: false,
		},
		{
			name: "9 queens",
			fen:  "3k4/8/8/8/8/3K4/QQQQQQQQ/3Q4 w - - 0 1",
			want: false,
		},
		{
			name: "10 queens",
			fen:  "3k4/8/8/8/8/3K4/QQQQQQQQ/3QQ3 w - - 0 1",
			want: true,
		},
		{
			name: "2 queens 8 pawns",
			fen:  "k7/8/8/8/PPPPPPPP/KQ6/Q7/8 w - - 0 1",
			want: true,
		},
		{
			name: "4 pawns 5 queens 3 knights",
			fen:  "2k5/8/8/8/8/PPPP4/KQQ2NNN/QQQ5 w - - 0 1",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Must(board.FromFEN(tt.fen))
			assert.Equal(t, tt.want, b.InvalidPieceCount(), "fen: %s", tt.fen)
		})
	}
}

func TestIsPseudoLegal(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		move move.Move
		want bool
	}{
		{
			name: "move is not pseudo legal due to no piece",
			fen:  "4k3/8/8/8/8/8/8/4K3 w - - 0 1",
			move: move.New(B2, B3),
			want: false,
		},
		{
			name: "move is not pseudo legal due to wrong color",
			fen:  "4k3/8/8/8/8/1p6/8/4K3 w - - 0 1",
			move: move.New(B3, B2),
			want: false,
		},
		{
			name: "knight move is pseudo legal",
			fen:  "4k3/8/8/8/2N5/8/8/4K3 w - - 0 1",
			move: move.New(C4, D6),
			want: true,
		},
		{
			name: "knight move has wrong dst square",
			fen:  "4k3/8/8/8/2N5/8/8/4K3 w - - 0 1",
			move: move.New(C4, D5),
			want: false,
		},
		{
			name: "bishop move is pseudo legal",
			fen:  "4k3/8/8/8/2B5/8/8/4K3 w - - 0 1",
			move: move.New(C4, F7),
			want: true,
		},
		{
			name: "bishop move has wrong dst square",
			fen:  "4k3/8/7B/8/2B5/8/8/4K3 w - - 0 1",
			move: move.New(C4, G7),
			want: false,
		},
		{
			name: "rook move is pseudo legal",
			fen:  "4k3/8/R7/8/8/8/8/4K3 w - - 0 1",
			move: move.New(A6, E6),
			want: true,
		},
		{
			name: "rook move has wrong dst square",
			fen:  "4k3/8/R7/8/8/8/8/4K1R1 w - - 0 1",
			move: move.New(A6, G5),
			want: false,
		},
		{
			name: "queen move is pseudo legal",
			fen:  "4k3/8/Q7/8/8/8/8/4K1Q1 w - - 0 1",
			move: move.New(A6, E2),
			want: true,
		},
		{
			name: "queen move has wrong dst square",
			fen:  "4k3/8/Q7/8/8/8/8/4K1Q1 w - - 0 1",
			move: move.New(A6, E3),
			want: false,
		},
		{
			name: "king move is pseudo legal",
			fen:  "4k3/8/2K5/8/8/8/8/8 w - - 0 1",
			move: move.New(C6, C5),
			want: true,
		},
		{
			name: "king move has wrong dst square",
			fen:  "4k3/8/2K5/8/8/8/8/8 w - - 0 1",
			move: move.New(C6, E5),
			want: false,
		},
		{
			name: "king stepping into check is pseudo legal",
			fen:  "4k3/8/2K5/8/8/8/8/8 w - - 0 1",
			move: move.New(C6, D7),
			want: true,
		},
		{
			name: "white short castle pseudo legal",
			fen:  "4k3/8/8/8/8/8/8/4K2R w K - 0 1",
			move: move.New(E1, G1),
			want: true,
		},
		{
			name: "white long castle pseudo legal",
			fen:  "4k3/8/8/8/8/8/8/R3K3 w Q - 0 1",
			move: move.New(E1, C1),
			want: true,
		},
		{
			name: "black short castle pseudo legal",
			fen:  "4k2r/8/8/8/8/8/8/4K3 b k - 0 1",
			move: move.New(E8, G8),
			want: true,
		},
		{
			name: "black long castle pseudo legal",
			fen:  "r3k3/8/8/8/8/8/8/4K3 b q - 0 1",
			move: move.New(E8, C8),
			want: true,
		},
		{
			name: "castle not pseudo legal due to check",
			fen:  "4k3/8/8/b7/8/8/8/4K2R w K - 0 1",
			move: move.New(E1, G1),
			want: false,
		},
		{
			name: "castle not pseudo legal due to checking in between",
			fen:  "4k3/8/8/1b6/8/8/8/4K2R w K - 0 1",
			move: move.New(E1, G1),
			want: false,
		},
		{
			name: "castle not pseudo legal due to no right",
			fen:  "4k3/8/8/8/8/8/8/4K2R w k - 0 1",
			move: move.New(E1, G1),
			want: false,
		},
		{
			name: "castle not pseudo legal due castling through occupied square",
			fen:  "4k3/8/8/8/8/8/8/RN2K3 w Q - 0 1",
			move: move.New(E1, C1),
			want: false,
		},
		{
			name: "single pawn push pseudo legal",
			fen:  "4k3/8/8/8/8/8/1P6/4K3 w - - 0 1",
			move: move.New(B2, B3),
			want: true,
		},
		{
			name: "single pawn wrong direction",
			fen:  "4k3/8/8/8/8/1P6/8/4K3 w - - 0 1",
			move: move.New(B3, B2),
			want: false,
		},
		{
			name: "single pawn push not pseudo legal due to block",
			fen:  "4k3/8/8/8/8/1p6/1P6/4K3 w - - 0 1",
			move: move.New(B2, B3),
			want: false,
		},
		{
			name: "single pawn push not pseudo legal due to missing promo",
			fen:  "4k3/1P6/8/8/8/8/8/4K3 w - - 0 1",
			move: move.New(B7, B8),
			want: false,
		},
		{
			name: "double pawn push pseudo legal",
			fen:  "4k3/8/8/8/8/8/1P6/4K3 w - - 0 1",
			move: move.New(B2, B4),
			want: true,
		},
		{
			name: "double pawn push not pseudo legal due to block/1",
			fen:  "4k3/8/8/8/8/1p6/1P6/4K3 w - - 0 1",
			move: move.New(B2, B4),
			want: false,
		},
		{
			name: "double pawn push not pseudo legal due to block/2",
			fen:  "4k3/8/8/8/1p6/8/1P6/4K3 w - - 0 1",
			move: move.New(B2, B4),
			want: false,
		},
		{
			name: "double pawn push not pseudo legal due not second rank",
			fen:  "4k3/8/8/8/8/1P6/8/4K3 w - - 0 1",
			move: move.New(B3, B5),
			want: false,
		},
		{
			name: "pawn push not pseudo legal pushing too much",
			fen:  "4k3/8/8/8/8/8/1P6/4K3 w - - 0 1",
			move: move.New(B2, B5),
			want: false,
		},
		{
			name: "pawn capture pseudo legal",
			fen:  "4k3/8/8/8/1p6/P7/8/4K3 w - - 0 1",
			move: move.New(A3, B4),
			want: true,
		},
		{
			name: "pawn capture not pseudo legal due to nothing captured",
			fen:  "4k3/8/8/8/8/P7/8/4K3 w - - 0 1",
			move: move.New(A3, B4),
			want: false,
		},
		{
			name: "pawn capture not pseudo legal due to self capture",
			fen:  "4k3/8/8/8/3P4/2P5/8/4K3 w - - 0 1",
			move: move.New(C3, D4),
			want: false,
		},
		{
			name: "pawn capture not pseudo legal due to missing promo",
			fen:  "2n1k3/1P6/8/8/8/8/8/4K3 w - - 0 1",
			move: move.New(B7, C8),
			want: false,
		},
		{
			name: "pawn capture not pseudo legal due to wrap around the board",
			fen:  "4k3/8/8/8/P6p/8/8/4K3 w - - 0 1",
			move: move.New(A4, H4),
			want: false,
		},
		{
			name: "pawn capture not pseudo legal sideways",
			fen:  "4k3/8/8/8/Pp6/8/8/4K3 w - - 0 1",
			move: move.New(A4, B4),
			want: false,
		},
		{
			name: "en passant pseudo legal",
			fen:  "4k3/8/8/8/1Pp5/8/8/4K3 b - b3 0 1",
			move: move.New(C4, B3),
			want: true,
		},
		{
			name: "en passant not pseudo legal due to missing state",
			fen:  "4k3/8/8/8/1Pp5/8/8/4K3 b - - 0 1",
			move: move.New(C4, B3),
			want: false,
		},
		{
			name: "en passant not pseudo legal when en-passant is null value",
			fen:  "4k3/8/8/8/8/8/1p6/4K3 b - - 0 1",
			move: move.New(B2, A1, move.WithPromo(Bishop)),
			want: false,
		},
		{
			name: "promo push pseudo legal",
			fen:  "4k3/1P6/8/8/8/8/8/4K3 w - - 0 1",
			move: move.New(B7, B8, move.WithPromo(Knight)),
			want: true,
		},
		{
			name: "promo capture pseudo legal",
			fen:  "2n1k3/1P6/8/8/8/8/8/4K3 w - - 0 1",
			move: move.New(B7, C8, move.WithPromo(Knight)),
			want: true,
		},
		{
			name: "sliding piece not pseudo legal due to block",
			fen:  "4k3/8/8/8/3p4/8/1B6/4K3 w - - 0 1",
			move: move.New(B2, G7),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Must(board.FromFEN(tt.fen))
			assert.Equal(t, tt.want, b.IsPseudoLegal(tt.move), "fen: %s, move: %s", tt.fen, tt.move)
		})
	}
}
