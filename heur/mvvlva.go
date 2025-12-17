package heur

import (
	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"
	. "github.com/paulsonkoly/chess-3/types"
)

// MVVLVA is the most valuable victim / least valuable aggressor heuristic. good should be set for good captures.
func MVVLVA(b *board.Board, m *move.Move, good bool) Score {
	r := rank(b, m)

	if good {
		return Captures + r
	}

	// reverse ranking for bad captures, so that values are less than or equal to
	// -Captures, but still in increasing in terms of how good they are.
	return -Captures - (Score(King*King) - r)
}

func rank(b *board.Board, m *move.Move) Score {
	if m.EPP != NoPiece {
		return Score(Pawn*King - Pawn)
	}
	// this isn't the most correct, maybe we want 3xpieceType buckets.
	victim := max(m.Promo(), b.SquaresToPiece[m.To()])

	return Score(victim*King - m.Piece)
}
