package eval

import (
	"math"

	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
)

type kingAttacks[T ScoreType] struct {
	accum [2]T
}

func (ka *kingAttacks[T]) addAttackingPiece(color Color, pType Piece, sqrs BitBoard, c *CoeffSet[T]) {
	ka.accum[color] += T(sqrs.Count()) * c.KingAttackPieces[pType-Pawn]
}

func (ka *kingAttacks[T]) addDefendingPiece(color Color, pType Piece, sqrs BitBoard, c *CoeffSet[T]) {
	ka.accum[color.Flip()] -= T(sqrs.Count()) * c.KingDefendingPieces[pType-Pawn]
}

func (ka *kingAttacks[T]) addSafeChecks(color Color, pType Piece, safeChecks BitBoard, c *CoeffSet[T]) {
	ka.accum[color] += c.SafeChecks[pType-Knight] * T(safeChecks.Count())
}

func (ka *kingAttacks[T]) addShelter(color Color, penalty T, c *CoeffSet[T]) {
	ka.accum[color.Flip()] += c.KingShelter[0] * penalty
}

func (ka *kingAttacks[T]) overall(b *board.Board, color Color, phase byte, c *CoeffSet[T]) T {
	queenIx := 0
	if b.Pieces[Queen]&b.Colors[color] == 0 {
		queenIx = 1
	}
	sgm := sigmoidal(ka.accum[color])
	if _, ok := (any)(sgm).(Score); ok {
		return T((int)(sgm) * int(c.KingAttackMagnitude[phase][queenIx]) / 64)
	}
	return sgm * c.KingAttackMagnitude[phase][queenIx] / 64
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
