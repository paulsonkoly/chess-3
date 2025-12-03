package main

import (
	"database/sql"
	"flag"
	"math/rand/v2"
	"runtime"
	"sync"
	_ "modernc.org/sqlite"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/eval"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"
	"github.com/paulsonkoly/chess-3/types"
)

const gamesQueueDepth = 16

type Config struct {
	gameCount     int
	openingDepth  int
	openingMargin int
	threads       int
}

var config = Config{}

func main() {

	flag.IntVar(&config.gameCount, "gameCount", 1_000_000, "number of games to generate")
	flag.IntVar(&config.openingDepth, "openingDepth", 8, "number of random generated opening moves")
	flag.IntVar(&config.openingMargin, "openingMargin", 300, "margin for what's considered to be balanced opening (cp)")
	flag.IntVar(&config.threads, "threads", runtime.NumCPU()-1, "number of threads")
	flag.Parse()

	gamesPerThread := (config.gameCount + config.threads - 1) / config.threads

	workersWG := sync.WaitGroup{}

	games := make(chan *Game, gamesQueueDepth)

	for t := range config.threads {
		if t == config.threads-1 { // last worker thread?
			gamesPerThread = config.gameCount - t*gamesPerThread
		}

		workersWG.Go(func() { generateGames(gamesPerThread, games) })
	}

	writerWG := sync.WaitGroup{}
	writerWG.Go(func() { writer(games) })

	workersWG.Wait()

	close(games)

	writerWG.Wait()
}

type Game struct {
	wdl byte
	pos []Position
}

type Position struct {
	fen  string
	bm   move.SimpleMove
	eval types.Score
}

func generateGames(count int, out chan<- *Game) {
	b := board.StartPos()
	ms := move.NewStore()

	defer ms.Clear()

	for range config.openingDepth {
		ms.Push()
		movegen.GenMoves(ms, b)

		moves := ms.Frame()

		if len(moves) < 1 {
			// we got into a position where there are no moves in the opening, highly unlikely
			return
		}

		chose := rand.IntN(len(moves))
		move := moves[chose]
		ms.Pop()

		b.MakeMove(&move)
	}

	// verify the opening in that it's resulted in a roughly equal position
	staticEval := eval.Eval(b, &eval.Coefficients)
	if staticEval < -types.Score(config.openingMargin) || staticEval > types.Score(config.openingMargin) {
		return
	}
}

func writer(in <-chan *Game) {
	db, err := sql.Open("sqlite", "file.sql")
	if err != nil {
		panic(err)
	}
	for game := range in {
	}
}
