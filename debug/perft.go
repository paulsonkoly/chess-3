package debug

import (
	"fmt"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"
	. "github.com/paulsonkoly/chess-3/chess"
)

func Perft(b *board.Board, depth Depth, split bool) int {
	ms := move.NewStore()
	return perft(ms, b, depth, split)
}

func perft(ms *move.Store, b *board.Board, depth Depth, split bool) int {
	if depth == 0 {
		return 1
	}

	cnt := 0
	me := b.STM

	ms.Push()
	defer ms.Pop()

	movegen.GenMoves(ms, b)

	for _, m := range ms.Frame() {
		r := b.MakeMove(m.Move)

		if !movegen.InCheck(b, me) {
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
