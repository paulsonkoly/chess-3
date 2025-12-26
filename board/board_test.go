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
		name string // description of this test case
		// Named input parameters for target function.
		b       *board.Board
		invalid bool // validity test result
	}{
		{
			name:    "startpos",
			b:       Must(board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")),
			invalid: false,
		},
		{
			name:    "8 queens",
			b:       Must(board.FromFEN("3k4/8/8/8/8/3K4/QQQQQQQQ/8 w - - 0 1")),
			invalid: false,
		},
		{
			name:    "9 queens",
			b:       Must(board.FromFEN("3k4/8/8/8/8/3K4/QQQQQQQQ/3Q4 w - - 0 1")),
			invalid: false,
		},
		{
			name:    "10 queens",
			b:       Must(board.FromFEN("3k4/8/8/8/8/3K4/QQQQQQQQ/3QQ3 w - - 0 1")),
			invalid: true,
		},
		{
			name:    "2 queens 8 pawns",
			b:       Must(board.FromFEN("k7/8/8/8/PPPPPPPP/KQ6/Q7/8 w - - 0 1")),
			invalid: true,
		},
		{
			name:    "4 pawns 5 queens 3 knights",
			b:       Must(board.FromFEN("2k5/8/8/8/8/PPPP4/KQQ2NNN/QQQ5 w - - 0 1")),
			invalid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.invalid, tt.b.InvalidPieceCount())
		})
	}
}
