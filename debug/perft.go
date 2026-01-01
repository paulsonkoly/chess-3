package debug

import (
	"fmt"

	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"
	"github.com/paulsonkoly/chess-3/stack"
)

func Perft(b *board.Board, depth Depth, split bool) int {
	ms := stack.NewSliceArena[move.Weighted]()
	return perft(ms, b, depth, split)
}

func perft(ms *stack.SliceArena[move.Weighted], b *board.Board, depth Depth, split bool) int {
	if depth == 0 {
		return 1
	}

	cnt := 0
	me := b.STM

	moves := ms.Push()
	defer ms.Pop()

	movegen.GenMoves(moves, b)

	for _, m := range *moves {
		r := b.MakeMove(m.Move)

		if !b.InCheck(me) {
			v := perft(ms, b, depth-1, false)
			if split {
				fmt.Println(m, v, b.FEN())
			}
			cnt += v
		}

		b.UndoMove(m.Move, r)
	}

	return cnt
}
