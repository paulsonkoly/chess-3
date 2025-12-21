package heur

import (
	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"
	. "github.com/paulsonkoly/chess-3/types"
)

// MVVLVA is the most valuable victim / least valuable aggressor heuristic. good should be set for good captures.
func MVVLVA(b *board.Board, m move.Move, good bool) Score {
	base := Captures
	if !good {
		base = -Captures - Score(King*King*King)
	}

	return base + rank(b, m)
}

func rank(b *board.Board, m move.Move) Score {
	victim := b.SquaresToPiece[b.CaptureSq(m)]
	aggressor := b.SquaresToPiece[m.From()]
	return Score(m.Promo()*King*King + victim*King - aggressor)
}
