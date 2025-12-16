package heur

import (
	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"
	. "github.com/paulsonkoly/chess-3/types"
)

func MVVLVA(b *board.Board, m *move.Move) Score {
	victim := b.SquaresToPiece[m.To()]
	aggressor := m.Piece

	return Captures + Score(victim*King - aggressor)
}
