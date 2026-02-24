package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"runtime/pprof"

	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/tools/extract/sampling"
)

var (
	dbFn            string
	outFn           string
	filterNoisy     bool
	filterMate      bool
	filterOutlier   bool
	filterInCheck   bool
	samplePhase     bool
	sampleOutcome   bool
	sampleImbalance bool
	samplePerGame   int
)

func main() {
	var cpuProf string

	flag.StringVar(&dbFn, "database", "database.db", "input database file name")
	flag.StringVar(&outFn, "output", "extract.epd", "output epd file")
	flag.BoolVar(&filterNoisy, "filterNoisy", true, "filter positions with bestmove being noisy")
	flag.BoolVar(&filterMate, "filterMate", true, "filter positions with mate scores")
	flag.BoolVar(&filterOutlier, "filterOutlier", false, "filter positions with eval mismatching wdl by margin")
	flag.BoolVar(&filterInCheck, "filterInCheck", true, "filter in check positions")
	flag.BoolVar(&samplePhase, "samplePhase", true, "sample positions for game phase")
	flag.BoolVar(&sampleOutcome, "sampleOutcome", true, "sample positions for outcome")
	flag.BoolVar(&sampleImbalance, "sampleImbalance", false, "sample positions for material imbalance")
	flag.IntVar(&samplePerGame, "samplePerGame", 15, "number of maximum positions from a game; (-1) to disable")
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

	db, err := sql.Open("sqlite3", dbFn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if _, err := db.Exec("PRAGMA cache_size = 100000"); err != nil {
		panic(err)
	}
	if _, err := db.Exec("PRAGMA temp_store = file"); err != nil {
		panic(err)
	}

	dp := discretizerPipe()
	counter := sampling.NewCounter(dp.Dim())

	ids, err := load(db, dp, &counter)
	if err != nil {
		panic(err)
	}

	fmt.Println(counter)

	sampler := sampling.NewUniformSampler(counter)
	if err := output(db, ids, dp, sampler); err != nil {
		panic(err)
	}
}

type epd struct {
	board *board.Board
	wdl   int
}

var pieceValues = [...]int{0, 1, 3, 3, 5, 9}

func discretizerPipe() sampling.Discretizer {
	discretizers := []sampling.Discretizer{}

	if sampleImbalance {
		discretizers = append(discretizers,
			sampling.NewFeature(2, func(d any) int {
				if epdE, ok := d.(epd); ok {
					b := epdE.board

					whitePieces := 0
					for pt := chess.Pawn; pt < chess.King; pt++ {
						whitePieces += (b.Colors[chess.White] & b.Pieces[pt]).Count() * (pieceValues[pt])
					}
					blackPieces := 0
					for pt := chess.Pawn; pt < chess.King; pt++ {
						blackPieces += (b.Colors[chess.Black] & b.Pieces[pt]).Count() * (pieceValues[pt])
					}
					return chess.Clamp(chess.Abs(whitePieces-blackPieces), 0, 1)
				}
				panic("interface conversion")
			}),
		)
	}

	if samplePhase {
		discretizers = append(discretizers,
			sampling.NewFeature(78/18, func(d any) int {
				if epdE, ok := d.(epd); ok {
					b := epdE.board

					pieceCnt := 0
					for pt := chess.Pawn; pt < chess.King; pt++ {
						pieceCnt += b.Pieces[pt].Count() * (pieceValues[pt])
					}
					pieceCnt /= 18
					// merge extreme top (partial) bucket into the previous because there
					// are not enough samples there. Even with a divisor of 78 the top
					// buckets are very sparse in the other features
					pieceCnt = min(pieceCnt, 78/18-1)
					return pieceCnt
				}
				panic("interface conversion")
			}),
		)
	}

	if sampleOutcome {
		discretizers = append(discretizers,
			sampling.NewFeature(3, func(d any) int {
				if epdE, ok := d.(epd); ok {
					return int(epdE.wdl)
				}
				panic("interface conversion")
			}),
		)
	}

	return sampling.NewCombined(discretizers...)
}

func load(db *sql.DB, dp sampling.Discretizer, cntr *sampling.Counter) ([]int, error) {
	var posCnt int
	if err := db.QueryRow("select count(*) from positions").Scan(&posCnt); err != nil {
		return nil, err
	}

	result := make([]int, 0, posCnt)

	bar := progressbar.NewOptions(posCnt, progressbar.OptionSetDescription("counting features"))
	defer bar.Close()

	games, err := db.Query("select id, wdl from games")
	if err != nil {
		return nil, err
	}
	defer games.Close()

	posStm, err := db.Prepare("select id, fen, best_move, eval from positions where game_id=?")
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

		startIx := len(result)
		m, err := loadPositions(posStm, bar, gameId, wdl, &result)
		if err != nil {
			return nil, err
		}
		endIx := len(result)

		samplePerGame = max(samplePerGame, -1)
		if samplePerGame != -1 && samplePerGame < endIx-startIx {
			rand.Shuffle(endIx-startIx, func(i, j int) {
				result[i+startIx], result[j+startIx] = result[j+startIx], result[i+startIx]
			})

			endIx = startIx + samplePerGame
			result = result[:endIx]
		}

		for _, id := range result[startIx:endIx] {
			cntr.Add(dp.Value(m[id]))
		}
	}

	return result, games.Err()
}

func loadPositions(
	posStm *sql.Stmt,
	bar *progressbar.ProgressBar,
	gameId int,
	wdl int,
	result *[]int,
) (map[int]epd, error) {
	positions, err := posStm.Query(gameId)
	if err != nil {
		return nil, err
	}
	defer positions.Close()

	m := make(map[int]epd)

	for ; positions.Next(); bar.Add(1) {
		var (
			id    int
			fen   []byte
			bm    move.Move
			score chess.Score
		)

		if err := positions.Scan(&id, &fen, &bm, &score); err != nil {
			return nil, err
		}

		b := board.Board{}
		if err := board.ParseFEN(&b, fen); err != nil {
			return nil, err
		}

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

		if filterNoisy && (bm.Promo() != chess.NoPiece || b.SquaresToPiece[bm.To()] != chess.NoPiece) {
			continue
		}

		if filterInCheck && b.InCheck(b.STM) {
			continue
		}

		*result = append(*result, id)
		m[id] = epd{&b, wdl}
	}

	return m, positions.Err()
}

func output(db *sql.DB, ids []int, disc sampling.Discretizer, sampler sampling.Sampler) error {
	outF, err := os.Create(outFn)
	if err != nil {
		return err
	}
	defer outF.Close()
	outB := bufio.NewWriter(outF)

	bar := progressbar.NewOptions(len(ids), progressbar.OptionSetDescription("output"))
	defer bar.Close()

	for _, id := range ids {
		var (
			wdl int
			fen []byte
		)

		if err := db.QueryRow(
			"select wdl, fen from games inner join positions on positions.game_id=games.id where positions.id=?",
			id).Scan(&wdl, &fen); err != nil {
			return err
		}

		var b board.Board
		if err := board.ParseFEN(&b, fen); err != nil {
			return err
		}
		b.ResetFifty()

		r := rand.Float64()

		if r < sampler.KeepProb(disc.Value(epd{&b, wdl})) {
			fmt.Fprintf(outB, "%s; %.1f\n", b.FEN(), wdlToEPD(wdl))
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
