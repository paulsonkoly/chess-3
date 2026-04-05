package eval2

import (
	. "github.com/paulsonkoly/chess-3/chess"
)
import "github.com/paulsonkoly/chess-3/board"

// MaxBlend is the sum of pieces on the starting position each piece counted as
// the corresponding Phase value.
const MaxBlend = 24

// Blend is game phase.
var Blend = [...]int{0, 0, 1, 1, 2, 4, 0}

func (e *Eval[T]) taperedScore(b *board.Board) T {
	mgScore := e.sp[b.STM][MG] - e.sp[b.STM.Flip()][MG]
	egScore := e.sp[b.STM][EG] - e.sp[b.STM.Flip()][EG]

	var phase int
	for pType := Pawn; pType <= Queen; pType++ {
		phase += (e.pieceCounts[White][pType] + e.pieceCounts[Black][pType]) * Blend[pType]
	}

	mgPhase := min(phase, MaxBlend)
	egPhase := MaxBlend - mgPhase

	if _, ok := (any(mgScore)).(Score); ok {
		v := int(mgScore)*mgPhase + int(egScore)*egPhase
		return T(v / MaxBlend)
	}

	v := mgScore*T(mgPhase) + egScore*T(egPhase)
	return v / MaxBlend
}
