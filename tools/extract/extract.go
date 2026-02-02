package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"runtime/pprof"

	progress "github.com/schollz/progressbar/v3"
	_ "modernc.org/sqlite"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/eval"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/tools/extract/sampling"
)

var (
	dbFn          string
	outFn         string
	filterNoisy   bool
	filterMate    bool
	samplePhase   bool
	sampleOutcome bool
	samplePerGame int
)

func main() {
	var cpuProf string

	flag.StringVar(&dbFn, "database", "database.db", "input database file name")
	flag.StringVar(&outFn, "output", "extract.epd", "output epd file")
	flag.BoolVar(&filterNoisy, "filterNoisy", true, "filter positions with bestmove being noisy")
	flag.BoolVar(&filterMate, "filterMate", true, "filter positions with mate scores")
	flag.BoolVar(&samplePhase, "samplePhase", true, "sample positions for game phase")
	flag.BoolVar(&sampleOutcome, "sampleOutcome", true, "sample positions for outcome")
	flag.IntVar(&samplePerGame, "samplePerGame", 40, "number of maximum positions from a game; (-1) to disable")
	flag.StringVar(&cpuProf, "cpuProf", "", "cpu profile (empty to disable)")

	flag.Parse()

	if cpuProf != "" {
		cpu, err := os.Create(cpuProf)
		if err != nil {
			panic(err)
		}
		err = pprof.StartCPUProfile(cpu)
		if err != nil {
			panic(err)
		}
		defer pprof.StopCPUProfile()
	}

	db, err := sql.Open("sqlite", dbFn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	entries, err := loadAndFilter(db)
	if err != nil {
		panic(err)
	}

	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "No positions loaded (database empty or all filtered)")
		os.Exit(1)
	}

	discretizers := []sampling.Discretizer{}
	if samplePhase {
		discretizers = append(discretizers,
			sampling.NewFeature(eval.MaxPhase+1, func(d any) int {
				if epdE, ok := d.(EPDEntry); ok {
					var b board.Board
					if err := board.ParseFEN(&b, epdE.fen); err != nil {
						panic(err)
					}

					pieceCnt := 0
					for pt := chess.Pawn; pt < chess.King; pt++ {
						pieceCnt += b.Pieces[pt].Count() * eval.Phase[pt]
					}
					pieceCnt = min(pieceCnt, eval.MaxPhase)
					return pieceCnt
				}
				panic("interface conversion")
			}),
		)
	}

	if sampleOutcome {
		discretizers = append(discretizers,
			sampling.NewFeature(2, func(d any) int {
				if epdE, ok := d.(EPDEntry); ok {
					if epdE.wdl == 0 {
						return 0 // draw
					} else { // decisive
						return 1
					}
				}
				panic("interface conversion")
			}),
		)
	}

	combined := sampling.NewCombined(discretizers...)
	downScaled := sampling.NewScale(combined, int(float64(combined.Dim())*0.6))

	counter := sampling.NewCounter(downScaled.Dim())
	bar := progress.NewOptions(len(entries), progress.OptionSetDescription("counting features"))
	for _, entry := range entries {
		counter.Add(downScaled.Value(entry))
		bar.Add(1)
	}
	bar.Close()

	sampler := sampling.NewSampler(counter)
	if err := output(entries, downScaled, sampler); err != nil {
		panic(err)
	}
}

func loadAndFilter(db *sql.DB) ([]EPDEntry, error) {

	var entryCnt int
	if err := db.QueryRow("select count(*) from positions").Scan(&entryCnt); err != nil {
		return nil, err
	}

	bar := progress.NewOptions(entryCnt, progress.OptionSetDescription("loading"))
	defer bar.Close()

	// cut the entry count in half, speculating that that's how many positions we
	// are going to filter.
	entries := make([]EPDEntry, 0, entryCnt/2)

	games, err := db.Query("select id, wdl from games")
	if err != nil {
		return nil, err
	}
	defer games.Close()

	posStm, err := db.Prepare("select fen, best_move, eval from positions where game_id=?")
	if err != nil {
		return nil, err
	}
	defer posStm.Close()

	for games.Next() {
		var (
			gameId int
			wdl    int
		)
		if err := games.Scan(&gameId, &wdl); err != nil {
			return nil, err
		}

		gameEntries, err := loadGamePositions(posStm, gameId, wdl, bar)
		if err != nil {
			return nil, err
		}

		if samplePerGame != -1 && samplePerGame < len(gameEntries) {
			rand.Shuffle(len(gameEntries), func(i, j int) {
				gameEntries[i], gameEntries[j] = gameEntries[j], gameEntries[i]
			})

			gameEntries = gameEntries[:samplePerGame]
		}

		entries = append(entries, gameEntries...)
	}

	if err := games.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

func loadGamePositions(posStm *sql.Stmt, gameId, wdl int, bar *progress.ProgressBar) ([]EPDEntry, error) {
	positions, err := posStm.Query(gameId)
	if err != nil {
		return nil, err
	}
	defer positions.Close()

	cap := samplePerGame
	if cap < 0 {
		cap = 64
	}
	gameEntries := make([]EPDEntry, 0, cap)

	var b board.Board

	for positions.Next() {
		var (
			fen   []byte
			bm    move.Move
			score chess.Score
		)

		if err := positions.Scan(&fen, &bm, &score); err != nil {
			return nil, err
		}

		bar.Add(1)

		if filterMate && score.IsMate() {
			continue
		}

		if filterNoisy {
			// filter noisy best moves
			if bm.Promo() != chess.NoPiece {
				continue
			}

			if err := board.ParseFEN(&b, fen); err != nil {
				return nil, err
			}

			if b.SquaresToPiece[bm.To()] != chess.NoPiece {
				continue
			}
		}

		fenCopy := make([]byte, len(fen))
		copy(fenCopy, fen)

		gameEntries = append(gameEntries, EPDEntry{fen: fenCopy, wdl: wdl})
	}
	if err := positions.Err(); err != nil {
		return nil, err
	}

	return gameEntries, nil
}

func output(entries []EPDEntry, disc sampling.Discretizer, sampler sampling.Sampler) error {
	outF, err := os.Create(outFn)
	if err != nil {
		return err
	}
	defer outF.Close()
	outB := bufio.NewWriter(outF)

	bar := progress.NewOptions(len(entries), progress.OptionSetDescription("output"))
	defer bar.Close()

	for _, entry := range entries {
		r := rand.Float64()

		if r < sampler.KeepProb(disc.Value(entry)) {
			fmt.Fprintf(outB, "%s; %.1f\n", entry.fen, wdlToEPD(entry.wdl))
		}
		bar.Add(1)
	}

	if err := outB.Flush(); err != nil {
		return err
	}

	return nil
}

type EPDEntry struct {
	fen []byte
	wdl int
}

func wdlToEPD(n int) float64 {
	switch n {
	case 0:
		return 0.5
	case 1:
		return 1.0
	case 2:
		return 0.0
	}
	panic(fmt.Sprintf("unexpected wdl %d", n))
}
