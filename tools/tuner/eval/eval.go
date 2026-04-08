package eval

import (
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/eval"
)

type CoeffSet eval.CoeffSet[float64]

// ColorRelative is a float64 version of eval.
type ColorRelative eval.Eval[float64]

func New() *ColorRelative {
	return (*ColorRelative)(eval.New[float64]())
}

// Score is the evaluation score relative to color not stm.
func (cr *ColorRelative) Score(b *board.Board, coeffs *CoeffSet) float64 {
	score := ((*eval.Eval[float64])(cr)).Score(b, (*eval.CoeffSet[float64])(coeffs))
	if b.STM == Black {
		score = -score
	}
	return score
}
