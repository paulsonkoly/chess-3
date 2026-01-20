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
	m := move.From(E1) | move.To(G1)

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

	m = move.From(E1) | move.To(F1)

	r = b.MakeMove(m)

	assert.Equal(t, Castles(0), b.Castles)

	b.UndoMove(m, r)

	assert.Equal(t, ShortWhite|LongWhite, b.Castles)

	m = move.From(A1) | move.To(B1)

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
			move: move.From(B2) | move.To(B3),
			want: false,
		},
		{
			name: "move is not pseudo legal due to wrong color",
			fen:  "4k3/8/8/8/8/1p6/8/4K3 w - - 0 1",
			move: move.From(B3) | move.To(B2),
			want: false,
		},
		{
			name: "knight move is pseudo legal",
			fen:  "4k3/8/8/8/2N5/8/8/4K3 w - - 0 1",
			move: move.From(C4) | move.To(D6),
			want: true,
		},
		{
			name: "knight move has wrong dst square",
			fen:  "4k3/8/8/8/2N5/8/8/4K3 w - - 0 1",
			move: move.From(C4) | move.To(D5),
			want: false,
		},
		{
			name: "bishop move is pseudo legal",
			fen:  "4k3/8/8/8/2B5/8/8/4K3 w - - 0 1",
			move: move.From(C4) | move.To(F7),
			want: true,
		},
		{
			name: "bishop move has wrong dst square",
			fen:  "4k3/8/7B/8/2B5/8/8/4K3 w - - 0 1",
			move: move.From(C4) | move.To(G7),
			want: false,
		},
		{
			name: "rook move is pseudo legal",
			fen:  "4k3/8/R7/8/8/8/8/4K3 w - - 0 1",
			move: move.From(A6) | move.To(E6),
			want: true,
		},
		{
			name: "rook move has wrong dst square",
			fen:  "4k3/8/R7/8/8/8/8/4K1R1 w - - 0 1",
			move: move.From(A6) | move.To(G5),
			want: false,
		},
		{
			name: "queen move is pseudo legal",
			fen:  "4k3/8/Q7/8/8/8/8/4K1Q1 w - - 0 1",
			move: move.From(A6) | move.To(E2),
			want: true,
		},
		{
			name: "queen move has wrong dst square",
			fen:  "4k3/8/Q7/8/8/8/8/4K1Q1 w - - 0 1",
			move: move.From(A6) | move.To(E3),
			want: false,
		},
		{
			name: "king move is pseudo legal",
			fen:  "4k3/8/2K5/8/8/8/8/8 w - - 0 1",
			move: move.From(C6) | move.To(C5),
			want: true,
		},
		{
			name: "king move has wrong dst square",
			fen:  "4k3/8/2K5/8/8/8/8/8 w - - 0 1",
			move: move.From(C6) | move.To(E5),
			want: false,
		},
		{
			name: "king stepping into check is pseudo legal",
			fen:  "4k3/8/2K5/8/8/8/8/8 w - - 0 1",
			move: move.From(C6) | move.To(D7),
			want: true,
		},
		{
			name: "white short castle pseudo legal",
			fen:  "4k3/8/8/8/8/8/8/4K2R w K - 0 1",
			move: move.From(E1) | move.To(G1),
			want: true,
		},
		{
			name: "white long castle pseudo legal",
			fen:  "4k3/8/8/8/8/8/8/R3K3 w Q - 0 1",
			move: move.From(E1) | move.To(C1),
			want: true,
		},
		{
			name: "black short castle pseudo legal",
			fen:  "4k2r/8/8/8/8/8/8/4K3 b k - 0 1",
			move: move.From(E8) | move.To(G8),
			want: true,
		},
		{
			name: "black long castle pseudo legal",
			fen:  "r3k3/8/8/8/8/8/8/4K3 b q - 0 1",
			move: move.From(E8) | move.To(C8),
			want: true,
		},
		{
			name: "castle not pseudo legal due to check",
			fen:  "4k3/8/8/b7/8/8/8/4K2R w K - 0 1",
			move: move.From(E1) | move.To(G1),
			want: false,
		},
		{
			name: "castle not pseudo legal due to checking in between",
			fen:  "4k3/8/8/1b6/8/8/8/4K2R w K - 0 1",
			move: move.From(E1) | move.To(G1),
			want: false,
		},
		{
			name: "castle not pseudo legal due to no right",
			fen:  "4k3/8/8/8/8/8/8/4K2R w k - 0 1",
			move: move.From(E1) | move.To(G1),
			want: false,
		},
		{
			name: "castle not pseudo legal due castling through occupied square",
			fen:  "4k3/8/8/8/8/8/8/RN2K3 w Q - 0 1",
			move: move.From(E1) | move.To(C1),
			want: false,
		},
		{
			name: "single pawn push pseudo legal",
			fen:  "4k3/8/8/8/8/8/1P6/4K3 w - - 0 1",
			move: move.From(B2) | move.To(B3),
			want: true,
		},
		{
			name: "single pawn wrong direction",
			fen:  "4k3/8/8/8/8/1P6/8/4K3 w - - 0 1",
			move: move.From(B3) | move.To(B2),
			want: false,
		},
		{
			name: "single pawn push not pseudo legal due to block",
			fen:  "4k3/8/8/8/8/1p6/1P6/4K3 w - - 0 1",
			move: move.From(B2) | move.To(B3),
			want: false,
		},
		{
			name: "single pawn push not pseudo legal due to missing promo",
			fen:  "4k3/1P6/8/8/8/8/8/4K3 w - - 0 1",
			move: move.From(B7) | move.To(B8),
			want: false,
		},
		{
			name: "double pawn push pseudo legal",
			fen:  "4k3/8/8/8/8/8/1P6/4K3 w - - 0 1",
			move: move.From(B2) | move.To(B4),
			want: true,
		},
		{
			name: "double pawn push not pseudo legal due to block/1",
			fen:  "4k3/8/8/8/8/1p6/1P6/4K3 w - - 0 1",
			move: move.From(B2) | move.To(B4),
			want: false,
		},
		{
			name: "double pawn push not pseudo legal due to block/2",
			fen:  "4k3/8/8/8/1p6/8/1P6/4K3 w - - 0 1",
			move: move.From(B2) | move.To(B4),
			want: false,
		},
		{
			name: "double pawn push not pseudo legal due not second rank",
			fen:  "4k3/8/8/8/8/1P6/8/4K3 w - - 0 1",
			move: move.From(B3) | move.To(B5),
			want: false,
		},
		{
			name: "pawn push not pseudo legal pushing too much",
			fen:  "4k3/8/8/8/8/8/1P6/4K3 w - - 0 1",
			move: move.From(B2) | move.To(B5),
			want: false,
		},
		{
			name: "pawn capture pseudo legal",
			fen:  "4k3/8/8/8/1p6/P7/8/4K3 w - - 0 1",
			move: move.From(A3) | move.To(B4),
			want: true,
		},
		{
			name: "pawn capture not pseudo legal due to nothing captured",
			fen:  "4k3/8/8/8/8/P7/8/4K3 w - - 0 1",
			move: move.From(A3) | move.To(B4),
			want: false,
		},
		{
			name: "pawn capture not pseudo legal due to self capture",
			fen:  "4k3/8/8/8/3P4/2P5/8/4K3 w - - 0 1",
			move: move.From(C3) | move.To(D4),
			want: false,
		},
		{
			name: "pawn capture not pseudo legal due to missing promo",
			fen:  "2n1k3/1P6/8/8/8/8/8/4K3 w - - 0 1",
			move: move.From(B7) | move.To(C8),
			want: false,
		},
		{
			name: "pawn capture not pseudo legal due to wrap around the board",
			fen:  "4k3/8/8/8/P6p/8/8/4K3 w - - 0 1",
			move: move.From(A4) | move.To(H4),
			want: false,
		},
		{
			name: "pawn capture not pseudo legal sideways",
			fen:  "4k3/8/8/8/Pp6/8/8/4K3 w - - 0 1",
			move: move.From(A4) | move.To(B4),
			want: false,
		},
		{
			name: "en passant pseudo legal",
			fen:  "4k3/8/8/8/1Pp5/8/8/4K3 b - b3 0 1",
			move: move.From(C4) | move.To(B3),
			want: true,
		},
		{
			name: "en passant not pseudo legal due to missing state",
			fen:  "4k3/8/8/8/1Pp5/8/8/4K3 b - - 0 1",
			move: move.From(C4) | move.To(B3),
			want: false,
		},
		{
			name: "en passant not pseudo legal when en-passant is null value",
			fen:  "4k3/8/8/8/8/8/1p6/4K3 b - - 0 1",
			move: move.From(B2) | move.To(A1) | move.Promo(Bishop),
			want: false,
		},
		{
			name: "promo push pseudo legal",
			fen:  "4k3/1P6/8/8/8/8/8/4K3 w - - 0 1",
			move: move.From(B7) | move.To(B8) | move.Promo(Knight),
			want: true,
		},
		{
			name: "promo capture pseudo legal",
			fen:  "2n1k3/1P6/8/8/8/8/8/4K3 w - - 0 1",
			move: move.From(B7) | move.To(C8) | move.Promo(Knight),
			want: true,
		},
		{
			name: "sliding piece not pseudo legal due to block",
			fen:  "4k3/8/8/8/3p4/8/1B6/4K3 w - - 0 1",
			move: move.From(B2) | move.To(G7),
			want: false,
		},
		{
			name: "when otherwise normal and legal non-pawn move contains promo",
			fen:  "8/1P5k/2p2r2/3pr2q/1Q4p1/5pPp/5P1K/2RR4 w - - 0 1",
			move: move.From(H2) | move.To(H1) | move.Promo(Queen),
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

func TestMoveCounts(t *testing.T) {
	b := Must(board.FromFEN("6k1/1n3ppp/4r3/8/8/3B3P/2R2PP1/6K1 w - - 10 111"))

	r1 := b.MakeMove(move.From(C2) | move.To(C7))
	assert.Equal(t, "6k1/1nR2ppp/4r3/8/8/3B3P/5PP1/6K1 b - - 11 111", b.FEN())
	// black move increments full move counter
	r2 := b.MakeMove(move.From(B7) | move.To(D6))
	assert.Equal(t, "6k1/2R2ppp/3nr3/8/8/3B3P/5PP1/6K1 w - - 12 112", b.FEN())
	// pawn move resets fifty move counter
	r3 := b.MakeMove(move.From(G2) | move.To(G3))
	assert.Equal(t, "6k1/2R2ppp/3nr3/8/8/3B2PP/5P2/6K1 b - - 0 112", b.FEN())

	b.UndoMove(move.From(G2)|move.To(G3), r3)
	assert.Equal(t, "6k1/2R2ppp/3nr3/8/8/3B3P/5PP1/6K1 w - - 12 112", b.FEN())
	b.UndoMove(move.From(B7)|move.To(D6), r2)
	assert.Equal(t, "6k1/1nR2ppp/4r3/8/8/3B3P/5PP1/6K1 b - - 11 111", b.FEN())
	b.UndoMove(move.From(C2)|move.To(C7), r1)
	assert.Equal(t, "6k1/1n3ppp/4r3/8/8/3B3P/2R2PP1/6K1 w - - 10 111", b.FEN())
}
