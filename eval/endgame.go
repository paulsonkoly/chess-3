package eval

import (
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/eval/kpvk"
)

// KBCorners are knight-bishop checkmate corners based on parity of square.
var KBCorners = [2][2]Square{{A1, H8}, {H1, A8}}

func (e *Eval[T]) knbvk(b *board.Board, c *CoeffSet[T]) T {
	e.sp = [Colors][Phases]T{}

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

	cornerDist := min(Chebyshev(victimKSq, KBCorners[parity][0]), Chebyshev(victimKSq, KBCorners[parity][1]))
	cornerDist = 7 - cornerDist
	cornerDist *= cornerDist

	e.sp[victim.Flip()][EG] += T(cornerDist) * 30

	return e.endgameScore(b)
}

func evalKPvK[T ScoreType](e *Eval[T], b *board.Board, c *CoeffSet[T]) T {
	if !kpvk.Winning(b) {
		return 0
	}
	pawn := b.Pieces[Pawn]
	strongSide := White
	if b.Colors[Black]&pawn != 0 {
		strongSide = Black
	}
	weakSide := strongSide.Flip()

	e.sp = [Colors][2]T{}

	pawnSq := pawn.LowestSet()

	// we incrementally increase the score based on pawn rank, encouraging pawn
	// advancement, the score should be kept below queen value so we queen at the
	// end, but large enough so we simplify into this if it's a winning endgame.
	//
	// we purposefully leave king distances out, because if we are here we know
	// the king can support the pawn. (this might need reconsidering in the
	// future).
	pawnRank := pawnSq.Rank().FromPerspectiveOf(weakSide)
	e.sp[strongSide][EG] = c.PieceValues[EG][Queen] - T(pawnRank)*15

	return e.endgameScore(b)
}
