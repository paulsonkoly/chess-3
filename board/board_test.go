package board_test

import (
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
	"github.com/stretchr/testify/assert"
)

func TestCastle(t *testing.T) {
	b := board.FromFEN("k7/p7/8/8/8/8/8/R3K2R w KQ - 0 1")
	m := move.Move{SimpleMove: move.SimpleMove{From: E1, To: G1}, Piece: King, Castle: ShortWhite, CRights: CRights(LongWhite, ShortWhite)}

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

	m = move.Move{SimpleMove: move.SimpleMove{From: E1, To: F1}, Piece: King, CRights: CRights(LongWhite, ShortWhite)}

	b.MakeMove(&m)

	assert.Equal(t, CRights(), b.CRights)

	b.UndoMove(&m)

	assert.Equal(t, CRights(ShortWhite, LongWhite), b.CRights)

	m = move.Move{SimpleMove: move.SimpleMove{From: A1, To: B1}, Piece: Rook, CRights: CRights(LongWhite)}

	b.MakeMove(&m)

	assert.Equal(t, CRights(ShortWhite), b.CRights)

	b.UndoMove(&m)

	assert.Equal(t, CRights(ShortWhite, LongWhite), b.CRights)
}

func TestZobrist(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		b *board.Board
	}{
		{
			name: "castle / en-passant / capture",
			b:    board.FromFEN("r3k3/8/8/4p1Pp/8/1p6/3P4/3BK2R w Kq h6 0 1"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.b

			ms := move.NewStore()
			movegen.GenMoves(ms, b, board.Full)

			for _, m := range ms.Frame() {
				mOld := m
				b.MakeMove(&m)

				king := b.Colors[b.STM.Flip()] & b.Pieces[King]

				if movegen.IsAttacked(b, b.STM, king) {
					// illegal (pseudo-leagal) move, skip
					b.UndoMove(&m)
					continue
				}

				assert.Equal(t, b.CalculateHash(), b.Hash(), "move", m)

				b.UndoMove(&m)

				assert.Equal(t, tt.b, b)
				assert.Equal(t, mOld, m, mOld)
			}
		})
	}
}
