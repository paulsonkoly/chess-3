package main

import (
	"flag"
	"fmt"
	"slices"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/debug"
	"github.com/paulsonkoly/chess-3/eval"
	"github.com/paulsonkoly/chess-3/search"
)

var debugFEN = flag.String("debugFEN", "", "Debug a given fen to a given depth using stockfish perft")
var debugDepth = flag.Int("debugDepth", 3, "Debug a given depth")

func main() {
	flag.Parse()

	if *debugFEN != "" {
		b := board.FromFEN(*debugFEN)

		debug.MatchPerft(b, *debugDepth)
		return
	}

	b := board.FromFEN("6k1/8/1P4K1/8/8/8/8/8 w - - 0 1")

	for depth := range 15 {
		eval, moves := search.AlphaBeta(b, -eval.Inf, eval.Inf, depth)
		slices.Reverse(moves)
		fmt.Println(eval, moves)
	}
}
