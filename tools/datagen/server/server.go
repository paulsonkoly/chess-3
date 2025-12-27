package server

import (
	"database/sql"
	"flag"
	"fmt"
	"math/rand/v2"
	"net"
	"os"
	"sync"
	"time"

	progress "github.com/schollz/progressbar/v3"
	_ "modernc.org/sqlite"

	"github.com/paulsonkoly/chess-3/tools/datagen/shim"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"
	"github.com/paulsonkoly/chess-3/search"
	"github.com/paulsonkoly/chess-3/transp"
)

const (
	OpeningQueueDepth = 32 // OpeningQueueDepth is the depth of openings buffered channel.
	GameQueueDepth    = 16
)

type ServerConfig struct {
	softNodes     int    // softNodes is the search soft node count.
	hardNodes     int    // hardNodes is the search hard node count.
	gameCount     int    // gameCount is number of games to generate.
	dbFile        string // dbFile is the file name for the output database.
	openingDepth  int    // openingDepth is the number of random opening moves.
	openingMargin int    // openingMargin is the score margin the openings exit score has to be limited by.
}

var serverConfig ServerConfig

func Run(args []string) {
	var config shim.Config

	var host string // host is the host name server is listening on.
	var port int    // port is the bound port number for the server to listen on.

	sFlags := flag.NewFlagSet("server", flag.ExitOnError)
	sFlags.IntVar(&serverConfig.gameCount, "gameCount", 1_000_000, "number of games to generate")
	sFlags.StringVar(&serverConfig.dbFile, "dbFile", "datagen.db", "file name for the database")
	sFlags.StringVar(&host, "host", "localhost", "host to listen on")
	sFlags.IntVar(&port, "port", 9001, "port to listen on")
	sFlags.IntVar(&serverConfig.openingDepth, "openingDepth", 8, "number of random generated opening moves")
	sFlags.IntVar(&serverConfig.openingMargin, "openingMargin", 300, "margin for what's considered to be balanced opening (cp)")

	sFlags.IntVar(&config.SoftNodes, "softNodes", 15_000, "soft node count for search")
	sFlags.IntVar(&config.HardNodes, "hardNodes", 8_000_000, "hard node count for search")
	sFlags.BoolVar(&config.Draw, "draw", true, "enable draw adjudication")
	sFlags.IntVar(&config.DrawAfter, "drawAfter", 40, "enables draw adjudication after this many moves")
	sFlags.IntVar(&config.DrawMargin, "drawMargin", 20, "position considered draw with this margin in adjudication (cp)")
	sFlags.IntVar(&config.DrawCount, "drawCount", 4, "number of positions drawn back to back for adjudication")
	sFlags.BoolVar(&config.Win, "win", true, "enable win adjudication")
	sFlags.IntVar(&config.WinAfter, "winAfter", 40, "enables win adjudication after this many moves")
	sFlags.IntVar(&config.WinMargin, "winMargin", 600, "positions considered win with this margin in adjudication (cp)")
	sFlags.IntVar(&config.WinCount, "winCount", 4, "number of positions won back to back for adjudication")

	sFlags.Parse(args)

	serverConfig.softNodes = config.SoftNodes
	serverConfig.hardNodes = config.HardNodes

	openings := make(chan *board.Board, OpeningQueueDepth)
	games := make(chan shim.Game, GameQueueDepth)

	wg := sync.WaitGroup{}

	wg.Go(func() {
		defer close(openings)
		generateOpenings(openings)
	})

	wg.Go(func() {
		defer close(games)
		writer(games)
	})

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		panic(err)
	}

	srv := shim.NewServer(&config, openings, games)
	go func() {
		if err := srv.Serve(lis); err != nil {
			panic(err)
		}
	}()

	wg.Wait()

	srv.Stop()
}

func generateOpenings(openings chan<- *board.Board) {

	generate := OpeningGenerator{
		ms:     move.NewStore(),
		search: search.New(1 * transp.MegaBytes),
		rnd:    rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0x82a1_73b1_69cc_df15)),
	}

	for range serverConfig.gameCount {
		openings <- generate.Opening()
	}
}

type OpeningGenerator struct {
	ms     *move.Store
	search *search.Search
	rnd    *rand.Rand
}

// TODO make sure the opening is unique
func (og *OpeningGenerator) Opening() *board.Board {
	var b *board.Board

	ms := og.ms

Retry:
	for {
		ms.Clear()
		b = board.StartPos()

		for range serverConfig.openingDepth {
			ms.Push()
			movegen.GenMoves(ms, b)
			moves := ms.Frame()

			if len(moves) < 1 {
				goto Retry
			}

			move := &moves[og.rnd.IntN(len(moves))]
			b.MakeMove(move.Move)
			if b.InCheck(b.STM.Flip()) { // pseudo legality check
				goto Retry
			}
		}

		score, _ := og.search.Go(b,
			search.WithSoftNodes(serverConfig.softNodes),
			search.WithNodes(serverConfig.hardNodes),
			search.WithInfo(false))

		if Range(serverConfig.openingMargin).Contains(score) {
			return b
		}
	}
}

type Range int

func (r Range) Contains(s chess.Score) bool {
	return chess.Score(-r) <= s && s <= chess.Score(r)
}

func writer(games <-chan shim.Game) {
	pb := progress.NewOptions(serverConfig.gameCount,
		progress.OptionSetPredictTime(true),
		progress.OptionSetItsString("games"),
		progress.OptionShowIts(),
	)

	db, err := sql.Open("sqlite", serverConfig.dbFile)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if _, err := db.Exec("pragma foreign_keys = on"); err != nil {
		panic(fmt.Sprintf("failed to enable foreign keys: %v", err))
	}

	// back to back error count
	b2bErrorCnt := 0

	for range serverConfig.gameCount {
		game := <-games

		tx, err := db.Begin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create tx %v", err)
			b2bErrorCnt++
			continue
		}

		var gameId int64
		res, err := tx.Exec("insert into games (wdl) values (?)", game.WDL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "insert into games failed %v\n", err)
			b2bErrorCnt++
			goto Fin
		}
		gameId, err = res.LastInsertId()
		if err != nil {
			fmt.Fprintf(os.Stderr, "insert into games failed to return id %v\n", err)
			b2bErrorCnt++
			goto Fin
		}

		for _, pos := range game.Positions {
			if _, err := tx.Exec("insert into positions (game_id, fen, best_move, eval) values (?, ?, ?, ?)",
				gameId,
				pos.FEN,
				pos.BM,
				pos.Score); err != nil {
				fmt.Fprintf(os.Stderr, "insert into positions failed %v\n", err)
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

		pb.Add(1)
	}
}
