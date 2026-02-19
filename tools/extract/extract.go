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
	"github.com/paulsonkoly/chess-3/heur"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/tools/extract/sampling"
)

var (
	dbFn             string
	outFn            string
	filterNoisy      bool
	filterMate       bool
	filterOutlier    bool
	filterInCheck    bool
	filterMgDecisive bool
	samplePhase      bool
	sampleOutcome    bool
	sampleImbalance  bool
	samplePerGame    int
)

type EPDEntry struct {
	board *board.Board
	wdl   int
}

var pool []EPDEntry

func main() {
	var cpuProf string

	flag.StringVar(&dbFn, "database", "database.db", "input database file name")
	flag.StringVar(&outFn, "output", "extract.epd", "output epd file")
	flag.BoolVar(&filterNoisy, "filterNoisy", false, "filter positions with bestmove being noisy")
	flag.BoolVar(&filterMate, "filterMate", true, "filter positions with mate scores")
	flag.BoolVar(&filterOutlier, "filterOutlier", true, "filter positions with eval mismatching wdl by margin")
	flag.BoolVar(&filterInCheck, "filterInCheck", false, "filter in check positions")
	flag.BoolVar(&filterMgDecisive, "filterMgDecisive", false, "filter decisive middle games with absolute score under threshold")
	flag.BoolVar(&samplePhase, "samplePhase", true, "sample positions for game phase")
	flag.BoolVar(&sampleOutcome, "sampleOutcome", true, "sample positions for outcome")
	flag.BoolVar(&sampleImbalance, "sampleImbalance", true, "sample positions for material imbalance")
	flag.IntVar(&samplePerGame, "samplePerGame", 100, "number of maximum positions from a game; (-1) to disable")
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

	if err := loadAndFilter(); err != nil {
		panic(err)
	}

	if len(pool) == 0 {
		fmt.Fprintln(os.Stderr, "No positions loaded (database empty or all filtered)")
		os.Exit(1)
	}

	discretizers := []sampling.Discretizer{}

	if sampleImbalance {
		discretizers = append(discretizers,
			sampling.NewFeature(2, func(d any) int {
				if epdE, ok := d.(EPDEntry); ok {
					b := epdE.board

					whitePieces := 0
					for pt := chess.Pawn; pt < chess.King; pt++ {
						whitePieces += (b.Colors[chess.White] & b.Pieces[pt]).Count() * (int(heur.PieceValues[pt]) / 100)
					}
					blackPieces := 0
					for pt := chess.Pawn; pt < chess.King; pt++ {
						blackPieces += (b.Colors[chess.Black] & b.Pieces[pt]).Count() * (int(heur.PieceValues[pt]) / 100)
					}
					return chess.Clamp(chess.Abs(whitePieces-blackPieces), 0, 1)
				}
				panic("interface conversion")
			}),
		)

	}

	if samplePhase {
		discretizers = append(discretizers,
			sampling.NewFeature(eval.MaxPhase/3+1, func(d any) int {
				if epdE, ok := d.(EPDEntry); ok {
					b := epdE.board

					pieceCnt := 0
					for pt := chess.Pawn; pt < chess.King; pt++ {
						pieceCnt += b.Pieces[pt].Count() * eval.Phase[pt]
					}
					pieceCnt /= 3
					pieceCnt = min(pieceCnt, eval.MaxPhase/3)
					return pieceCnt
				}
				panic("interface conversion")
			}),
		)
	}

	if sampleOutcome {
		discretizers = append(discretizers,
			sampling.NewFeature(3, func(d any) int {
				if epdE, ok := d.(EPDEntry); ok {
					return int(epdE.wdl)
				}
				panic("interface conversion")
			}),
		)
	}

	combined := sampling.NewCombined(discretizers...)

	counter := sampling.NewCounter(combined.Dim())
	bar := progress.NewOptions(len(pool), progress.OptionSetDescription("counting features"))
	for _, entry := range pool {
		counter.Add(combined.Value(entry))
		bar.Add(1)
	}
	bar.Close()

	sampler := sampling.NewUniformSampler(counter)
	if err := output(pool, combined, sampler); err != nil {
		panic(err)
	}
}

func loadAndFilter() error {
	db, err := sql.Open("sqlite", dbFn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var entryCnt int
	if err := db.QueryRow("select count(*) from positions").Scan(&entryCnt); err != nil {
		return err
	}

	bar := progress.NewOptions(entryCnt, progress.OptionSetDescription("loading"))
	defer bar.Close()

	// cut the entry count in half, speculating that that's how many positions we
	// are going to filter.
	pool = make([]EPDEntry, 0, entryCnt)

	games, err := db.Query("select id, wdl from games")
	if err != nil {
		return err
	}
	defer games.Close()

	posStm, err := db.Prepare("select fen, best_move, eval from positions where game_id=?")
	if err != nil {
		return err
	}
	defer posStm.Close()

	for games.Next() {
		var (
			gameId int
			wdl    int
		)
		if err := games.Scan(&gameId, &wdl); err != nil {
			return err
		}

		gameEntries, err := loadGamePositions(posStm, gameId, wdl, bar)
		if err != nil {
			return err
		}

		if samplePerGame != -1 && samplePerGame < len(gameEntries) {
			rand.Shuffle(len(gameEntries), func(i, j int) {
				gameEntries[i], gameEntries[j] = gameEntries[j], gameEntries[i]
			})

			gameEntries = gameEntries[:samplePerGame]
		}

		pool = append(pool, gameEntries...)
	}

	if err := games.Err(); err != nil {
		return err
	}

	return nil
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

	for positions.Next() {
		var (
			fen   string
			bm    move.Move
			score chess.Score
		)

		if err := positions.Scan(&fen, &bm, &score); err != nil {
			return nil, err
		}

		bar.Add(1)

		b, err := board.FromFEN(fen)
		if err != nil {
			return nil, err
		}
		b.ResetFifty()

		if filterOutlier {
			if b.STM == chess.Black {
				score = -score
			}

			if (score < -600 && wdl == WhiteWon) || (score > 600 && wdl == BlackWon) {
				continue
			}
		}

		if filterMate && score.IsMate() {
			continue
		}

		if filterNoisy {
			// filter noisy best moves
			if bm.Promo() != chess.NoPiece {
				continue
			}

			if b.SquaresToPiece[bm.To()] != chess.NoPiece {
				continue
			}
		}

		if filterMgDecisive && wdl != Draw {
			pieceCnt := 0
			for pt := chess.Pawn; pt < chess.King; pt++ {
				pieceCnt += b.Pieces[pt].Count() * eval.Phase[pt]
			}
			pieceCnt /= 3

			if pieceCnt > 4 && chess.Abs(score) < 200 {
				continue
			}
		}

		if filterInCheck && b.InCheck(b.STM) {
			continue
		}

		gameEntries = append(gameEntries, EPDEntry{board: b, wdl: wdl})
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
			fmt.Fprintf(outB, "%s; %.1f\n", entry.board.FEN(), wdlToEPD(entry.wdl))
		}
		bar.Add(1)
	}

	if err := outB.Flush(); err != nil {
		return err
	}

	return nil
}

const (
	Draw     = 0
	WhiteWon = 1
	BlackWon = 2
)

func wdlToEPD(n int) float64 {
	switch n {
	case Draw:
		return 0.5
	case WhiteWon:
		return 1.0
	case BlackWon:
		return 0.0
	}
	panic(fmt.Sprintf("unexpected wdl %d", n))
}
