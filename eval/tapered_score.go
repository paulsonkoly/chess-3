package eval

import (
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
)

const (
	// MaxBlend is the sum of pieces on the starting position each piece counted as
	// the corresponding Phase value.
	MaxBlend       = 24
	MaxScaleFactor = 128
)

// Blend is game phase.
var Blend = [...]int{0, 0, 1, 1, 2, 4, 0}

func (e *Eval[T]) taperedScore(b *board.Board) T {
	mgScore := e.sp[b.STM][MG] - e.sp[b.STM.Flip()][MG]

	// drawishness
	var t T
	if _, ok := ((any)(t)).(Score); ok {
		e.sp[White][EG] = T(int(e.sp[White][EG]) * int(e.scaleFactor[White]) / MaxScaleFactor)
		e.sp[Black][EG] = T(int(e.sp[Black][EG]) * int(e.scaleFactor[Black]) / MaxScaleFactor)
	} else {
		e.sp[White][EG] = e.sp[White][EG] * e.scaleFactor[White] / MaxScaleFactor
		e.sp[Black][EG] = e.sp[Black][EG] * e.scaleFactor[Black] / MaxScaleFactor
	}
	egScore := e.sp[b.STM][EG] - e.sp[b.STM.Flip()][EG]

	var phase int
	for pType := Pawn; pType <= Queen; pType++ {
		phase += b.Pieces[pType].Count() * Blend[pType]
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

func (e *Eval[T]) endgameScore(b *board.Board) T {
	return e.sp[b.STM][EG] - e.sp[b.STM.Flip()][EG]
}
