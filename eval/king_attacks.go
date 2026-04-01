package eval

import (
	"math"

	. "github.com/paulsonkoly/chess-3/chess"
)

type kingAttacks[T ScoreType] struct {
	accum [2]T
}

func (ka *kingAttacks[T]) addAttackPieces(color Color, pType Piece, attacks BitBoard, kingNB BitBoard, c *CoeffSet[T]) {
	if kingNB&attacks != 0 {
		ka.accum[color] += c.KingAttackPieces[pType-Knight]
	}
}

func (ka *kingAttacks[T]) addSafeChecks(color Color, pType Piece, checks BitBoard, c *CoeffSet[T]) {
	ka.accum[color] += c.SafeChecks[pType-Knight] * T(checks.Count())
}

func (ka *kingAttacks[T]) addUnsafeChecks(color Color, pType Piece, checks BitBoard, c *CoeffSet[T]) {
	ka.accum[color] += c.UnsafeChecks[pType-Knight] * T(checks.Count())
}

func (ka *kingAttacks[T]) addPawns(pw *pieceWise, pawns *pawns, c *CoeffSet[T]) {
	for color := range Colors {
		eKing := pw.kingSq[color.Flip()]
		kFile := eKing.File()
		kRank := eKing.Rank()

		frontLine := pawns.frontLine[color]
		backMost := pawns.backMost[color.Flip()]

		var central, front, side Coord
		if kFile >= EFile {
			central, front, side = kFile-1, kFile, kFile+1
		} else {
			central, front, side = kFile+1, kFile, kFile-1
		}

		// central
		{
			fileBB := FileBB(central & 7)

			if storm := frontLine & fileBB; storm != 0 {
				dist := Abs(kRank - storm.LowestSet().Rank())
				ka.accum[color] += c.KingStorm[0][dist]
			}

			if shelter := backMost & fileBB; shelter != 0 {
				dist := Abs(kRank - shelter.LowestSet().Rank())
				ka.accum[color] -= c.KingShelter[0][dist]
			} else {
				ka.accum[color] += c.KingOpenFile[0]
			}
		}

		// front
		{
			fileBB := FileBB(front & 7)

			if storm := frontLine & fileBB; storm != 0 {
				dist := Abs(kRank - storm.LowestSet().Rank())
				ka.accum[color] += c.KingStorm[1][dist]
			}

			if shelter := backMost & fileBB; shelter != 0 {
				dist := Abs(kRank - shelter.LowestSet().Rank())
				ka.accum[color] -= c.KingShelter[1][dist]
			} else {
				ka.accum[color] += c.KingOpenFile[1]
			}
		}

		// side
		{
			if side < 0 || side >= 8 {
				continue
			}
			fileBB := FileBB(side & 7)

			if storm := frontLine & fileBB; storm != 0 {
				dist := Abs(kRank - storm.LowestSet().Rank())
				ka.accum[color] += c.KingStorm[2][dist]
			}

			if shelter := backMost & fileBB; shelter != 0 {
				dist := Abs(kRank - shelter.LowestSet().Rank())
				ka.accum[color] -= c.KingShelter[2][dist]
			} else {
				ka.accum[color] += c.KingOpenFile[2]
			}
		}
	}
}

func (ka *kingAttacks[T]) sigmoidal(color Color) T {
	return sigmoidal(ka.accum[color])
}

// def f(x) = 600.fdiv(1+Math.exp(-0.1*(x)))
//
// 100.times.map { |x| f(x).round }.each_slice(10).to_a
//
// where 600 is the maximal bonus for attack, 0.1 is the steepness of the
// sigmoid, and 0 is the inflection point, implying a -50..50 range for king
// attack score.
var sigm = [...]Score{
	4, 4, 5, 5, 6, 7, 7, 8, 9, 10,
	11, 12, 13, 14, 16, 18, 19, 21, 23, 26,
	28, 31, 34, 38, 41, 46, 50, 55, 60, 65,
	72, 78, 85, 93, 101, 109, 119, 128, 139, 150,
	161, 173, 186, 199, 213, 227, 241, 255, 270, 285,
	300, 315, 330, 345, 359, 373, 387, 401, 414, 427,
	439, 450, 461, 472, 481, 491, 499, 507, 515, 522,
	528, 535, 540, 545, 550, 554, 559, 562, 566, 569,
	572, 574, 577, 579, 581, 582, 584, 586, 587, 588,
	589, 590, 591, 592, 593, 593, 594, 595, 595, 596,
}

func sigmoidal[T ScoreType](n T) T {
	if _, ok := (any(n)).(Score); ok {
		return T(sigm[Clamp(int(n+50), 0, len(sigm)-1)])
	}
	return T(600.0 / (1.0 + math.Exp(-0.1*(float64(n)))))
}
