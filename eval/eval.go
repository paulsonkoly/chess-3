package eval

import (
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
)

// ScoreType defines the evaluation result type. The engine uses int16 for
// score type, as defined in types. The tuner uses float64.
type ScoreType interface{ Score | float64 }

type Phase byte

const (
	MG = Phase(iota)
	EG

	Phases
)

type Eval[T ScoreType] struct {
	sp [Colors][Phases]T
}

func New[T ScoreType]() *Eval[T] {
	return &Eval[T]{}
}

func (e *Eval[T]) Clear() {
}

// Blend is game phase.
var Blend = [...]int{0, 0, 1, 1, 2, 4, 0}

const (
	// MaxBlend is the sum of pieces on the starting position each piece counted as
	// the corresponding Phase value.
	MaxBlend = 24
)

func (e *Eval[T]) Score(b *board.Board, c *CoeffSet[T]) T {
	e.sp = [2][2]T{}

	var phase int
	for color := range Colors {
		for pType := Pawn; pType <= King; pType++ {
			for pieces := b.Colors[color] & b.Pieces[pType]; pieces != 0; pieces &= pieces - 1 {
				if pType != King {
					e.sp[color][MG] += c.PieceValues[MG][pType]
					e.sp[color][EG] += c.PieceValues[EG][pType]

					phase += Blend[pType]
				}

				sq := pieces.LowestSet()
				if color == White {
					sq ^= 56 // upside down
				}

				ix := pType - 1
				e.sp[color][MG] += c.PSqT[2*ix][sq]
				e.sp[color][EG] += c.PSqT[2*ix+1][sq]
			}
		}
	}

	mgScore := e.sp[b.STM][MG] - e.sp[b.STM.Flip()][MG]
	egScore := e.sp[b.STM][EG] - e.sp[b.STM.Flip()][EG]

	mgPhase := min(phase, MaxBlend)
	egPhase := MaxBlend - mgPhase

	if _, ok := (any(mgScore)).(Score); ok {
		v := int(mgScore)*mgPhase + int(egScore)*egPhase
		return T(v / MaxBlend)
	}

	v := mgScore*T(mgPhase) + egScore*T(egPhase)
	return v / MaxBlend
}
