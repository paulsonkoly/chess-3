package heur_test

import (
	"slices"
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/heur"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"
	"github.com/stretchr/testify/assert"
)

func TestMVVLVA(t *testing.T) {
	b := Must(board.FromFEN("2n1k3/1P4b1/8/R2q3N/2P5/8/8/4K3 w - - 0 1"))

	ms := move.NewStore()
	ms.Push()
	defer ms.Pop()

	movegen.GenNoisy(ms, b)
	moves := ms.Frame()

	for i, m := range moves {
		moves[i].Weight = heur.MVVLVA(b, m.Move, true)
	}

	slices.SortFunc(moves, func(a, b move.Weighted) int {
		switch {
		case a.Weight < b.Weight:
			return 1
		case a.Weight > b.Weight:
			return -1
		default: // ==
			return 0
		}
	})

	listMoves := make([]string, 0, len(moves))
	for _, m := range moves {
		listMoves = append(listMoves, m.String())
	}

	assert.Equal(t, []string{
		"b7c8q", "b7b8q", "b7c8r", "b7b8r", "b7c8b", "b7b8b", "b7c8n", "b7b8n",
		"c4d5", "a5d5", "h5g7",
	}, listMoves)
}
