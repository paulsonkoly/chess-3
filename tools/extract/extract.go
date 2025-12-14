package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"runtime/pprof"

	progress "github.com/schollz/progressbar/v3"
	_ "modernc.org/sqlite"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/eval"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/types"
)

const PhaseBucketCount = 24

var (
	dbFn          string
	outFn         string
	filterNoisy   bool
	samplePerGame int
)

func main() {
	var cpuProf string

	flag.StringVar(&dbFn, "database", "database.db", "input database file name")
	flag.StringVar(&outFn, "output", "extract.epd", "output epd file")
	flag.BoolVar(&filterNoisy, "filterNoisy", true, "filter positions with bestmove being noisy")
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

	entries, err := loadPositions(db)
	if err != nil {
		panic(err)
	}

	buckets, err := countPhases(entries)
	if err != nil {
		panic(err)
	}

	fmt.Println(buckets)

	k := math.MaxFloat64
	for _, bucket := range buckets {
		rat := (float64(bucket) / float64(len(entries))) / (1.0 / 24.0)
		if rat < k {
			k = rat
		}
	}

	keepProb := [PhaseBucketCount]float64{}
	for ix, bucket := range buckets {
		dist := float64(bucket) / float64(len(entries))
		keepProb[ix] = k * (1.0 / 24.0) / dist
	}

	fmt.Println(keepProb)

	if err := output(entries, keepProb[:]); err != nil {
		panic(err)
	}
}

func loadPositions(db *sql.DB) ([]EPDEntry, error) {

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

	posStm, err := db.Prepare("select fen, best_move from positions where game_id=?")
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
			fen []byte
			bm  move.SimpleMove
		)

		if err := positions.Scan(&fen, &bm); err != nil {
			return nil, err
		}

		bar.Add(1)

		if filterNoisy {
			// filter noisy best moves
			if bm.Promo() != types.NoPiece {
				continue
			}

			if err := board.ParseFEN(&b, fen); err != nil {
				return nil, err
			}

			if b.SquaresToPiece[bm.To()] != types.NoPiece {
				continue
			}
		}

		gameEntries = append(gameEntries, EPDEntry{fen: fen, wdl: wdl})
	}
	if err := positions.Err(); err != nil {
		return nil, err
	}

	return gameEntries, nil
}

func countPhases(entries []EPDEntry) ([]int, error) {
	buckets := [PhaseBucketCount]int{}
	bar := progress.NewOptions(len(entries), progress.OptionSetDescription("sorting phases"))
	defer bar.Close()

	var b board.Board

	for ix, entry := range entries {
		if err := board.ParseFEN(&b, entry.fen); err != nil {
			return nil, err
		}

		bucketIx := 0
		for pt := types.Pawn; pt < types.King; pt++ {
			bucketIx += b.Pieces[pt].Count() * eval.Phase[pt]
		}

		bucketIx = min(bucketIx, PhaseBucketCount-1)

		entries[ix].bucketIx = bucketIx

		buckets[bucketIx]++
		bar.Add(1)
	}

	return buckets[:], nil
}

func output(entries []EPDEntry, keepProb []float64) error {
	outF, err := os.Create(outFn)
	if err != nil {
		return err
	}
	defer outF.Close()
	outB := bufio.NewWriter(outF)

	outBuckets := [PhaseBucketCount]int{}

	bar := progress.NewOptions(len(entries), progress.OptionSetDescription("output"))
	defer bar.Close()

	for _, entry := range entries {
		r := rand.Float64()

		if r < keepProb[entry.bucketIx] {
			fmt.Fprintf(outB, "%s; %.1f\n", entry.fen, wdlToEPD(entry.wdl))
			outBuckets[entry.bucketIx]++
		}
		bar.Add(1)
	}
	fmt.Println(outBuckets)

	if err := outB.Flush(); err != nil {
		return err
	}

	return nil
}

type EPDEntry struct {
	fen      []byte
	wdl      int
	bucketIx int
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
