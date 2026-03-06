package eval

import (
	. "github.com/paulsonkoly/chess-3/chess"
)

// MaxPhase is the sum of pieces on the starting position each piece counted as
// the corresponding Phase value.
const MaxPhase = 24

// Phase is game phase.
var Phase = [...]int{0, 0, 1, 1, 2, 4, 0}

type phase[T ScoreType] struct {
	phase int
}

func (p *phase[T]) addPieces(pType Piece, count int) {
	p.phase += Phase[pType] * count
}

// blend is linear interpolate of mg and eg according to phase.
func (p *phase[T]) blend(mg T, eg T) T {
	mgPhase := min(p.phase, MaxPhase)
	egPhase := MaxPhase - mgPhase

	if _, ok := (any(mg)).(Score); ok {
		v := int(mg)*mgPhase + int(eg)*egPhase
		return T(v / MaxPhase)
	}

	v := mg*T(mgPhase) + eg*T(egPhase)
	return v / MaxPhase
}
