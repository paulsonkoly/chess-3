package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"slices"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/debug"
	"github.com/paulsonkoly/chess-3/eval"
	"github.com/paulsonkoly/chess-3/search"
)

var debugFEN = flag.String("debugFEN", "", "Debug a given fen to a given depth using stockfish perft")
var debugDepth = flag.Int("debugDepth", 3, "Debug a given depth")
var cpuProf = flag.String("cpuProf", "", "cpu profile file name")

func main() {
	flag.Parse()

	if *cpuProf != "" {
		cpu, err := os.Create(*cpuProf)
		if err != nil {
			panic(err)
		}
		err = pprof.StartCPUProfile(cpu)
		if err != nil {
			panic(err)
		}
		defer pprof.StopCPUProfile()
	}

	if *debugFEN != "" {
		b := board.FromFEN(*debugFEN)

		debug.MatchPerft(b, *debugDepth)
		return
	}

	b := board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")

	for depth := range 10 {
		eval, moves := search.AlphaBeta(b, -eval.Inf, eval.Inf, depth+1)
		fmt.Println(search.QDepth)
		slices.Reverse(moves)
		fmt.Println(eval, moves)
	}
}
