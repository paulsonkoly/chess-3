package eval

import (
	"math"

	. "github.com/paulsonkoly/chess-3/chess"
)

type kingAttacks[T ScoreType] struct {
	attacks  [2][9]T
	defences [2][9]T
	accum    [2]T
}

func (ka *kingAttacks[T]) addAttackingPiece(color Color, pType Piece, sqrs BitBoard, kingSq Square, c *CoeffSet[T]) {
	for ; sqrs != 0; sqrs &= sqrs - 1 {
		sq := sqrs.LowestSet()
		ka.attacks[color][residenceIx(kingSq, sq)] += c.KingAttackPieces[pType-Pawn]
	}
}

func (ka *kingAttacks[T]) addDefendingPiece(color Color, pType Piece, sqrs BitBoard, kingSq Square, c *CoeffSet[T]) {
	for ; sqrs != 0; sqrs &= sqrs - 1 {
		sq := sqrs.LowestSet()
		ka.defences[color.Flip()][residenceIx(kingSq, sq)] += c.KingDefendingPieces[pType-Pawn]
	}
}

// maps sq to a 0 to 8 index around the ring of kingSq. The order is not
// significant as long as it is consistent. The middle square where
// the king resides is never used.
func residenceIx(kingSq, sq Square) int {
	return int((sq.Rank()-kingSq.Rank()+1)*3 + (sq.File() - kingSq.File() + 1))
}

func (ka *kingAttacks[T]) addSafeChecks(color Color, pType Piece, safeChecks BitBoard, c *CoeffSet[T]) {
	ka.accum[color] += c.SafeChecks[pType-Knight] * T(safeChecks.Count())
}

func (ka *kingAttacks[T]) addShelter(color Color, penalty T, c *CoeffSet[T]) {
	ka.accum[color.Flip()] += c.KingShelter[0] * penalty
}

func (ka *kingAttacks[T]) sigmoidal(color Color) T {
	sum := T(0)
	for i := range ka.attacks[color] {
		sum += max(0, ka.attacks[color][i]-ka.defences[color][i])
	}
	return sigmoidal(ka.accum[color] + sum)
}

// def f(x) = 600.fdiv(1+Math.exp(-0.2*(x-50)))
//
// 100.times.map { |x| f(x).round }.each_slice(10).to_a
//
// where 600 is the maximal bonus for attack, 0.2 is the steepness of the
// sigmoid, and 50 is the inflection point, implying a 0-100 range for king
// attack score.
var sigm = [...]Score{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 1, 1, 1, 1, 1,
	1, 2, 2, 3, 3, 4, 5, 6, 7, 9,
	11, 13, 16, 19, 23, 28, 34, 41, 50, 60,
	72, 85, 101, 119, 139, 161, 186, 213, 241, 270,
	300, 330, 359, 387, 414, 439, 461, 481, 499, 515,
	528, 540, 550, 559, 566, 572, 577, 581, 584, 587,
	589, 591, 593, 594, 595, 596, 597, 597, 598, 598,
	599, 599, 599, 599, 599, 599, 600, 600, 600, 600,
	600, 600, 600, 600, 600, 600, 600, 600, 600, 600,
}

func sigmoidal[T ScoreType](n T) T {
	if _, ok := (any(n)).(Score); ok {
		return T(sigm[Clamp(int(n), 0, len(sigm)-1)])
	}
	return T(600.0 / (1.0 + math.Exp(-0.2*(float64(n)-50.0))))
}
