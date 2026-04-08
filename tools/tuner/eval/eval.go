package eval

import (
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/eval"
)

type CoeffSet eval.CoeffSet[float64]

type ColorRelative eval.Eval[float64]

func New() *ColorRelative {
	return (*ColorRelative)(eval.New[float64]())
}

func (cr *ColorRelative) Score(b *board.Board, coeffs *CoeffSet) float64 {
	score := ((*eval.Eval[float64])(cr)).Score(b, (*eval.CoeffSet[float64])(coeffs))
	if b.STM == Black {
		score = -score
	}
	return score
}
