package eval2

import (
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
)

func (e *Eval[T]) isKNBvK(b *board.Board) bool {
	if b.Pieces[Pawn]|b.Pieces[Rook]|b.Pieces[Queen] != 0 {
		return false
	}
	return (e.pieceCounts[White][Bishop] == 1 && e.pieceCounts[White][Knight] == 1 &&
		e.pieceCounts[Black][Bishop] == 0 && e.pieceCounts[Black][Knight] == 0) ||
		(e.pieceCounts[White][Bishop] == 0 && e.pieceCounts[White][Knight] == 0 &&
			e.pieceCounts[Black][Bishop] == 1 && e.pieceCounts[Black][Knight] == 1)
}

// KBCorners are knight-bishop checkmate corners based on parity of square.
var KBCorners = [2][2]Square{{A1, H8}, {H1, A8}}

func (e *Eval[T]) KNBvK(b *board.Board, c *CoeffSet[T]) {
	bishopSq := b.Pieces[Bishop].LowestSet()
	knightSq := b.Pieces[Knight].LowestSet()

	victim := White
	if b.Pieces[Bishop]&b.Colors[White] != 0 {
		victim = Black
	}
	victimKSq := (b.Pieces[King] & b.Colors[victim]).LowestSet()
	attackKSq := (b.Pieces[King] & b.Colors[victim.Flip()]).LowestSet()

	e.addPSqT(victim, King, victimKSq, c)
	e.addPSqT(victim.Flip(), King, attackKSq, c)
	e.addPSqT(victim.Flip(), Knight, knightSq, c)
	e.addPSqT(victim.Flip(), Bishop, bishopSq, c)
	e.sp[victim.Flip()][EG] += c.PieceValues[EG][Knight]
	e.sp[victim.Flip()][EG] += c.PieceValues[EG][Bishop]

	parity := (bishopSq.File() + bishopSq.Rank()) & 1

	cornerDist := min(Chebishev(victimKSq, KBCorners[parity][0]), Chebishev(victimKSq, KBCorners[parity][1]))
	cornerDist = 7 - cornerDist
	cornerDist *= cornerDist

	e.sp[victim.Flip()][EG] += T(cornerDist) * 30
}
