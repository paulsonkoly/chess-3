package types

import "iter"

type Piece byte

const (
	NoPiece = Piece(iota)
  Pawn
	Knight
  Bishop
  Rook
  Queen
	King
)

func AllPieces() iter.Seq[Piece] {
	return func(yield func(Piece) bool) {
		for piece := Pawn; piece <= King; piece++ {
			if !yield(piece) {
				return
			}
		}
	}
}
