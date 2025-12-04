package main

import (
	"database/sql"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"runtime"
	"sync"

	_ "modernc.org/sqlite"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/eval"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"
	"github.com/paulsonkoly/chess-3/search"
	"github.com/paulsonkoly/chess-3/transp"
	"github.com/paulsonkoly/chess-3/types"
)

const (
	gamesQueueDepth = 16
	gameLength      = 128
)

type Config struct {
	gameCount     int
	openingDepth  int
	openingMargin int
	softNodes     int
	hardNodes     int
	threads       int
	dbFile        string
}

var config = Config{}

func main() {

	flag.IntVar(&config.gameCount, "gameCount", 1_000_000, "number of games to generate")
	// flag.IntVar(&config.gameCount, "gameCount", 2, "number of games to generate")
	flag.IntVar(&config.openingDepth, "openingDepth", 8, "number of random generated opening moves")
	flag.IntVar(&config.openingMargin, "openingMargin", 300, "margin for what's considered to be balanced opening (cp)")
	flag.IntVar(&config.threads, "threads", runtime.NumCPU()-1, "number of threads")
	flag.IntVar(&config.softNodes, "softNodes", 15_000, "soft node count for search")
	flag.IntVar(&config.hardNodes, "hardNodes", 8_000_000, "hard node count for search")
	flag.StringVar(&config.dbFile, "dbFile", "datagen.db", "file name for the database")
	flag.Parse()

	gamesPerThread := (config.gameCount + config.threads - 1) / config.threads

	workersWG := sync.WaitGroup{}

	games := make(chan *Game, gamesQueueDepth)

	for count, t := 0, 0; count < config.gameCount; count, t = count+gamesPerThread, t+1 {
		gamesPerThread = min(gamesPerThread, config.gameCount-count)
		workersWG.Go(func() { NewGenerator().Games(gamesPerThread, games) })
	}

	writerWG := sync.WaitGroup{}
	writerWG.Go(func() { writer(games) })

	workersWG.Wait()

	close(games)

	writerWG.Wait()
}

type Game struct {
	wdl       byte
	positions []Position
}

type Position struct {
	fen   string
	bm    move.SimpleMove
	score types.Score
}

type Generator struct {
	search    *search.Search
	moveStore *move.Store
}

func NewGenerator() Generator {
	s := search.New(1 * transp.MegaBytes)
	ms := move.NewStore()

	return Generator{search: s, moveStore: ms}
}

func (g Generator) Games(count int, out chan<- *Game) {
	g.search = search.New(1 * transp.MegaBytes)
	for range count {
		g.Game(out)
	}
}

func (g Generator) Game(out chan<- *Game) {
	g.search.Clear()

	b := g.Opening()

	positions := make([]Position, gameLength)

	for {
		score, bm := g.search.Go(b,
			search.WithSoftNodes(config.softNodes),
			search.WithNodes(config.hardNodes))

		if bm == 0 {
			break
		}

		positions = append(positions, Position{fen: b.FEN(), bm: bm, score: score})

		// convert bm back to full move
		move := movegen.FromSimple(b, bm)

		b.MakeMove(&move)
	}

	out <- &Game{positions: positions}
}

func (g *Generator) Opening() *board.Board {
	success := false
	var b *board.Board

	for !success {
		g.moveStore.Clear()
		b = board.StartPos()

		success = true
		for range config.openingDepth {
			g.moveStore.Push()
			movegen.GenMoves(g.moveStore, b)
			moves := g.moveStore.Frame()

			if len(moves) < 1 {
				success = false
				break
			}

			move := &moves[rand.IntN(len(moves))]
			b.MakeMove(move)
			if movegen.InCheck(b, b.STM.Flip()) { // pseudo legality check
				success = false
				break
			}
		}

		if success {
			score := eval.Eval(b, &eval.Coefficients)
			if score < types.Score(config.openingMargin) || score > types.Score(config.openingMargin) {
				success = false
			}
		}
	}

	return b
}

func writer(games <-chan *Game) {
	db, err := sql.Open("sqlite", config.dbFile)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// back to back error count
	b2bErrorCnt := 0

	for game := range games {

		tx, err := db.Begin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create tx %v", err)
			b2bErrorCnt++
			continue
		}

		if _, err := db.Exec("insert into games values ()", game.wdl); err != nil {
			fmt.Fprintf(os.Stderr, "insert into games failed %v", err)
			b2bErrorCnt++
			goto Fin
		}

		for _, pos := range game.positions {
			if _, err := db.Exec("insert into positions values ()", pos.fen); err != nil {
				b2bErrorCnt++
				goto Fin
			}
		}

		b2bErrorCnt = 0

	Fin:

		switch {

		case b2bErrorCnt == 0:
			if err := tx.Commit(); err != nil {
				fmt.Fprintf(os.Stderr, "failed to commit tx %v", err)
				b2bErrorCnt++
			}

		case b2bErrorCnt > 3:
			panic("more than 3 errors happened back to back in the writer, giving up")

		default:
			if err := tx.Rollback(); err != nil {
				fmt.Fprintf(os.Stderr, "failed to rollback tx %v", err)
				b2bErrorCnt++
			}
		}
	}
}
