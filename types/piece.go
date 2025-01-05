package types

import "iter"

type Piece byte

const (
	NoPiece = Piece(iota)
	King
	Knight
)

func AllPieces() iter.Seq[Piece] {
	return func(yield func(Piece) bool) {
		for piece := King; piece <= Knight; piece++ {
			if !yield(piece) {
				return
			}
		}
	}
}
