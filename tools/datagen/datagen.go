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

type Margin int

func (m Margin) RangeContains(s types.Score) bool { return types.Score(-m) <= s && s <= types.Score(m) }
func (m Margin) FallsAbove(s types.Score) bool    { return s < types.Score(-m) }
func (m Margin) FallsUnder(s types.Score) bool    { return types.Score(m) < s }

type Config struct {
	gameCount     int    // gameCount is number of games to generate.
	openingDepth  int    // openingDepth is the number of random opening moves.
	openingMargin Margin // openingMargin is the score margin the openings exit score has to be limited by.
	softNodes     int    // softNodes is the search soft node count.
	hardNodes     int    // hardNodes is the search hard node count.
	draw          bool   // draw enables draw adjudication.
	drawAfter     int    // draw after determines how many moves have to be played before considering draw adjudication.
	drawMargin    Margin // drawScore determines the margin for draw adjudication.
	drawCount     int    // drawCount is the minimum number of back to back positions for draw adjudication.
	win           bool   // win enables win adjudication.
	winAfter      int    // winAfter determines how many moves have to be played before considering win adjudication.
	winMargin     Margin // winScore determines the margin for win adjudication.
	winCount      int    // winCount is the minimum number of back to back positions for win adjudication.
	threads       int    // threads is the number of threads running game generator.
	dbFile        string // dbFile is the file name for the output database.
}

var config = Config{}

func main() {

	// flag.IntVar(&config.gameCount, "gameCount", 1_000_000, "number of games to generate")
	flag.IntVar(&config.gameCount, "gameCount", 4, "number of games to generate")
	flag.IntVar(&config.openingDepth, "openingDepth", 8, "number of random generated opening moves")
	openingMargin := flag.Int("openingMargin", 300, "margin for what's considered to be balanced opening (cp)")
	flag.IntVar(&config.threads, "threads", runtime.NumCPU()-1, "number of threads")
	flag.IntVar(&config.softNodes, "softNodes", 15_000, "soft node count for search")
	flag.IntVar(&config.hardNodes, "hardNodes", 8_000_000, "hard node count for search")
	flag.BoolVar(&config.draw, "draw", true, "enable draw adjudication")
	flag.IntVar(&config.drawAfter, "drawAfter", 40, "enables draw adjudication after this many moves")
	drawScore := flag.Int("drawScore", 20, "position considered draw with this margin in adjudication (cp)")
	flag.IntVar(&config.drawCount, "drawCount", 4, "number of positions drawn back to back for adjudication")
	flag.BoolVar(&config.win, "win", true, "enable win adjudication")
	winScore := flag.Int("winScore", 600, "positions considered win with this margin in adjudication (cp)")
	flag.IntVar(&config.winCount, "winCount", 4, "number of positions won back to back for adjudication")
	flag.StringVar(&config.dbFile, "dbFile", "datagen.db", "file name for the database")
	flag.Parse()

	config.openingMargin = Margin(*openingMargin)
	config.drawMargin = Margin(*drawScore)
	config.winMargin = Margin(*winScore)

	// convert full move counters to half move counters
	config.drawAfter *= 2
	config.drawCount *= 2
	config.winAfter *= 2
	config.winCount *= 2

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

	positions := make([]Position, 0, gameLength)
	drawCounter := 0
	winCounter := 0
	winSign := types.Score(1)

	for moveCounter := 0; ; moveCounter++ {
		score, bm := g.search.Go(b,
			search.WithSoftNodes(config.softNodes),
			search.WithNodes(config.hardNodes),
			search.WithInfo(false))

		positions = append(positions, Position{fen: b.FEN(), bm: bm, score: score})

		if bm == 0 {
			break
		}

		if config.draw && moveCounter >= config.drawAfter && config.drawMargin.RangeContains(score) {
			drawCounter++
		} else {
			drawCounter = 0
		}

		if drawCounter >= config.drawCount {
			break
		}

		if config.win {
			if winCounter == 0 {
				// determine the side winning on the first detection
				switch {
				case config.winMargin.FallsAbove(score):
					winCounter++
				case config.winMargin.FallsUnder(score):
					winCounter++
					winSign = -1
				}
			} else {
				if config.winMargin.FallsUnder(winSign * score) {
					winCounter++
					winSign *= -1
				} else {
					winCounter = 0
					winSign = 1
				}
			}
		}

		if winCounter >= config.winCount {
			break
		}

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
			score, _ := g.search.Go(b,
				search.WithSoftNodes(config.softNodes),
				search.WithNodes(config.hardNodes),
				search.WithInfo(false))

			if !config.openingMargin.RangeContains(score) {
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
