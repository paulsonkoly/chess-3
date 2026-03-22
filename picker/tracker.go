package picker

import (
	"slices"

	"github.com/paulsonkoly/chess-3/heur"
	"github.com/paulsonkoly/chess-3/move"
)

type tracker struct {
	ix    int
	moves [heur.KillerStride + 1]move.Move // +1 for the hashMove
}

func (t *tracker) Add(m move.Move) {
	t.moves[t.ix] = m
	t.ix++
}

func (t *tracker) Has(m move.Move) bool {
	return slices.Contains(t.moves[:t.ix], m)
}
