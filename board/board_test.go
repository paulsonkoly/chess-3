package board_test

import (
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
	"github.com/stretchr/testify/assert"
)

func TestCastle(t *testing.T) {
	b := board.FromFEN("k7/p7/8/8/8/8/8/R3K2R w KQ - 0 1")
	m := move.Move{From: E1, To: G1, Piece: King, Castle: ShortWhite, CRights: 0}

	b.MakeMove(&m)

	assert.Equal(t, CRights(), b.CRights)
	assert.Equal(t, b.SquaresToPiece[E1], NoPiece)
	assert.Equal(t, b.SquaresToPiece[F1], Rook)
	assert.Equal(t, b.SquaresToPiece[G1], King)
	assert.Equal(t, b.SquaresToPiece[H1], NoPiece)

	b.UndoMove(&m)

	assert.Equal(t, CRights(ShortWhite, LongWhite), b.CRights)
	assert.Equal(t, b.SquaresToPiece[E1], King)
	assert.Equal(t, b.SquaresToPiece[F1], NoPiece)
	assert.Equal(t, b.SquaresToPiece[G1], NoPiece)
	assert.Equal(t, b.SquaresToPiece[H1], Rook)

	m = move.Move{From: E1, To: F1, Piece: King, CRights: 0}

	b.MakeMove(&m)

	assert.Equal(t, CRights(), b.CRights)

	b.UndoMove(&m)

	assert.Equal(t, CRights(ShortWhite, LongWhite), b.CRights)

	m = move.Move{From: A1, To: B1, Piece: Rook, CRights: CRights(ShortWhite)}

	b.MakeMove(&m)

	assert.Equal(t, CRights(ShortWhite), b.CRights)

	b.UndoMove(&m)

	assert.Equal(t, CRights(ShortWhite, LongWhite), b.CRights)
}
