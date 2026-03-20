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
