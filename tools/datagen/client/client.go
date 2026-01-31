package client

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/search"
	"github.com/paulsonkoly/chess-3/tools/datagen/shim"
	"github.com/paulsonkoly/chess-3/transp"
)

const gameLength = 128

func Run(args []string) {
	var host string
	var port int
	var numThreads int

	cFlags := flag.NewFlagSet("client", flag.ExitOnError)
	cFlags.StringVar(&host, "host", "localhost", "host to connect to")
	cFlags.IntVar(&port, "port", 9001, "port to connect to")
	cFlags.IntVar(&numThreads, "threads", runtime.NumCPU(), "number of worker threads")

	cFlags.Parse(args)

	client, err := shim.NewClient(host, port)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			panic(err)
		}
	}()

	config, err := client.RequestConfig()
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}

	for range numThreads {
		wg.Go(func() {
			generate := NewGenerator()
			generate.Games(config, client)
		})
	}

	wg.Wait()
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

func (g Generator) Games(config shim.Config, client shim.Client) {
	errCnt := 0
	for {
		ok, err := g.Game(config, client)
		if !ok {
			return
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			errCnt++
			if errCnt > 3 {
				fmt.Fprintln(os.Stderr, "max retries exceeded, giving up")
				return
			}
		}
	}
}

func (g Generator) Game(config shim.Config, client shim.Client) (ok bool, err error) {
	g.search.Clear()

	b, err := client.RequestOpening()
	if err != nil {
		return true, err
	}
	if b == nil {
		return false, nil
	}

	positions := make([]shim.Position, 0, gameLength)
	drawCounter := 0
	winCounter := 0
	winSign := chess.Score(1)
	var score chess.Score

	for moveCounter := 0; ; moveCounter++ {
		var bm move.Move
		score, bm, _ = g.search.Go(b,
			search.WithSoftNodes(config.SoftNodes),
			search.WithNodes(config.HardNodes),
			search.WithOutput(nil))

		positions = append(positions, shim.Position{FEN: b.FEN(), BM: bm, Score: score})

		if bm == 0 {
			break
		}

		if config.Draw && moveCounter >= config.DrawAfter && Range(config.DrawMargin).Contains(score) {
			drawCounter++
		} else {
			drawCounter = 0
		}

		if drawCounter >= config.DrawCount {
			break
		}

		if config.Win && moveCounter >= config.WinAfter {
			if winCounter == 0 {
				// determine the side winning on the first detection
				switch {
				case Range(config.WinMargin).IsHigherThan(score):
					winCounter++
				case Range(config.WinMargin).IsLowerThan(score):
					winCounter++
					winSign = -1
				}
			} else {
				if Range(config.WinMargin).IsLowerThan(winSign * score) {
					winCounter++
					winSign *= -1
				} else {
					winCounter = 0
					winSign = 1
				}
			}
		}

		if winCounter >= config.WinCount {
			break
		}

		b.MakeMove(bm)
	}

	// determine the WDL result
	// conver score to white's perspective
	if b.STM == chess.Black {
		score = -score
	}

	var wdl shim.WDL
	switch {
	case Range(config.DrawMargin).Contains(score):
		wdl = shim.Draw

	case Range(config.WinMargin).IsLowerThan(score):
		wdl = shim.WhiteWins

	case Range(config.WinMargin).IsHigherThan(score):
		wdl = shim.BlackWins

	default:
		panic(fmt.Sprintf("cannot determine game outcome %d", score))
	}

	if err := client.RegisterGame(&shim.Game{Positions: positions, WDL: wdl}); err != nil {
		return true, err
	}
	return true, nil
}

type Range int

func (r Range) Contains(s chess.Score) bool {
	return chess.Score(-r) <= s && s <= chess.Score(r)
}

func (r Range) IsLowerThan(s chess.Score) bool {
	return chess.Score(r) < s
}

func (r Range) IsHigherThan(s chess.Score) bool {
	return s < chess.Score(-r)
}
