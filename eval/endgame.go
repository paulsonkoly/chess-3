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

	parity := bishopSq.Parity()
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

// kbpvk is the bishop+pawn endgame where a rook pawn draws if the opponent's
// king can reach the queening square and the bishop is on the opposite color
// complex.
func (e *Eval[T]) kbpvk(b *board.Board, c *CoeffSet[T]) T {
	// if not a Rook-pawn then normal positional eval
	e.scaleFactor = [2]T{MaxScaleFactor, MaxScaleFactor}
	var flank Coord
	switch {

	case b.Pieces[Pawn]&^AFileBB == 0:
		flank = 0

	case b.Pieces[Pawn]&^HFileBB == 0:
		flank = 1

	default:
		return e.positional(b, c)
	}

	strongSide := Black
	if b.Counts[White][Bishop] == 1 {
		strongSide = White
	}
	weakSide := strongSide.Flip()

	queeningSq := KBCorners[flank^Coord(weakSide)][weakSide]
	bishopSq := b.Pieces[Bishop].LowestSet()

	// is the bishop controlling the queningSquare?
	if bishopSq.Parity() == queeningSq.Parity() {
		return e.positional(b, c)
	}

	victimK := b.Colors[weakSide] & b.Pieces[King]
	victimKSq := victimK.LowestSet()

	dist := Chebyshev(queeningSq, victimKSq)
	if dist <= 1 {
		return 0
	}
	e.scaleFactor[strongSide] = MaxScaleFactor * T(dist-1) / T(dist)
	e.scaleFactor[weakSide] = MaxScaleFactor

	return e.positional(b, c)
}
