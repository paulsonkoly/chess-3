package main

import (
	"fmt"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/movegen"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

func main() {
	fmt.Println(Perft(board.New(), 3))
}

func Perft(b *board.Board, depth int) int {
	if depth == 0 {
		return 1
	}

	perft := 0
  me := b.STM
	for m := range movegen.Moves(b, board.Full) {
		b.MakeMove(&m)

		kingBB := b.Pieces[King] & b.Colors[me]
		if !movegen.IsAttacked(b, me.Flip(), kingBB) {
			perft += Perft(b, depth-1)
		}

		b.UndoMove(&m)
	}

	return perft
}
