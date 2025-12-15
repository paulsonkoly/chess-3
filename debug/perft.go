package debug

import (
	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"
)

func Perft(b *board.Board, depth int) int {
	ms := move.NewStore()
	return perft(ms, b, depth)
}

func perft(ms *move.Store, b *board.Board, depth int) int {
	if depth == 0 {
		return 1
	}

	cnt := 0
	me := b.STM

	ms.Push()
	defer ms.Pop()

	movegen.GenMoves(ms, b)

	for _, m := range ms.Frame() {
		b.MakeMove(&m)

		if !movegen.InCheck(b, me) {
			cnt += perft(ms, b, depth-1)
		}

		b.UndoMove(&m)
	}

	return cnt
}
