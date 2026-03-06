package eval

import (
	"math"

	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
)

type kingAttacks[T ScoreType] struct {
	accum [2][2]T
}

func (ka *kingAttacks[T]) addAttackPieces(color Color, pType Piece, attacks BitBoard, kingNB BitBoard, c *CoeffSet[T]) {
	if kingNB&attacks != 0 {
		ka.accum[0][color] += c.KingAttackPieces[0][pType-Knight]
		ka.accum[1][color] += c.KingAttackPieces[1][pType-Knight]
	}
}

func (ka *kingAttacks[T]) addSafeChecks(color Color, pType Piece, safeChecks BitBoard, c *CoeffSet[T]) {
	ka.accum[0][color] += c.SafeChecks[0][pType-Knight] * T(safeChecks.Count())
	ka.accum[1][color] += c.SafeChecks[1][pType-Knight] * T(safeChecks.Count())
}

func (ka *kingAttacks[T]) addShelter(color Color, penalty T, c *CoeffSet[T]) {
	ka.accum[0][color.Flip()] += c.KingShelter[0] * penalty
	ka.accum[1][color.Flip()] += c.KingShelter[1] * penalty
}

func (ka *kingAttacks[T]) blend(color Color, phase phase[T], c *CoeffSet[T]) T {
	// tapered eval blend mg & eg
	kingAttack := phase.blend(ka.accum[0][color], ka.accum[1][color])
	mag := phase.blend(c.KingAttackMagnitude[0], c.KingAttackMagnitude[1])
	stp := phase.blend(c.KingAttackSteepness[0], c.KingAttackSteepness[1])

	// this condition should stop the tuner going towards degenerate values.
	if mag < 0 || stp < 0 {
		return 0
	}

	// sigmoidal steepness (transition rate) from stp, and magnitude (king attack
	// importance) from mag.
	return T(float64(mag) / (1 + math.Exp(-(float64(stp)/64)*(float64(kingAttack)-200))))
}

func (ka *kingAttacks[T]) score(b *board.Board, phase phase[T], c *CoeffSet[T]) T {
	scores := [2]T{ka.blend(White, phase, c), ka.blend(Black, phase, c)}

	return scores[b.STM] - scores[b.STM.Flip()]
}
