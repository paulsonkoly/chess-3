package main

import (
	"database/sql"
	"flag"
	"fmt"
	"math"
	"math/rand/v2"
	"os"

	_ "modernc.org/sqlite"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/eval"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/types"
)

const PhaseBucketCount = 24

func main() {
	var dbFn string
	var outFn string
	var filterNoisy bool
	var samplePerGame int

	flag.StringVar(&dbFn, "database", "database.db", "input database file name")
	flag.StringVar(&outFn, "output file", "extract.epd", "output epd file")
	flag.BoolVar(&filterNoisy, "filterNoisy", true, "filter positions with bestmove being noisy")
	flag.IntVar(&samplePerGame, "samplePerGame", 40, "number of maximum positions from a game; (-1) to disable")

	flag.Parse()

	db, err := sql.Open("sqlite", dbFn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	games, err := db.Query("select id, wdl from games")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := games.Close(); err != nil {
			panic(err)
		}
	}()

	entries := make([]EPDEntry, 0)

	fmt.Println("loading...")

	cnt := 0
	for games.Next() {
		var (
			game_id int
			wdl     int
		)
		if err := games.Scan(&game_id, &wdl); err != nil {
			panic(err)
		}

		positions, err := db.Query("select fen, best_move from positions where game_id=?", game_id)
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := positions.Close(); err != nil {
				panic(err)
			}
		}()

		gameEntries := make([]EPDEntry, 0)

		for positions.Next() {
			var (
				fen string
				bm  move.SimpleMove
			)

			if err := positions.Scan(&fen, &bm); err != nil {
				panic(err)
			}

			if filterNoisy {
				// filter noisy best moves
				if bm.Promo() != types.NoPiece {
					continue
				}

				b, err := board.FromFEN(fen)
				if err != nil {
					panic(err)
				}

				if b.SquaresToPiece[bm.To()] != types.NoPiece {
					continue
				}
			}

			gameEntries = append(gameEntries, EPDEntry{fen: fen, wdl: wdl})

			if cnt % 10_000 == 0 {
				fmt.Println(cnt)
			}
			cnt++
		}

		if samplePerGame != -1 && samplePerGame < len(gameEntries) {
			rand.Shuffle(len(gameEntries), func(i, j int) {
				gameEntries[i], gameEntries[j] = gameEntries[j], gameEntries[i]
			})

			gameEntries = gameEntries[:samplePerGame]
		}

		entries = append(entries, gameEntries...)
	}

	fmt.Println("loaded")

	buckets := [PhaseBucketCount]int{}

	for _, entry := range entries {
		b, err := board.FromFEN(entry.fen)
		if err != nil {
			panic(err)
		}

		bucketIx := 0
		for pt := types.Pawn; pt < types.King; pt++ {
			bucketIx += b.Pieces[pt].Count() * eval.Phase[pt]
		}

		bucketIx = min(bucketIx , PhaseBucketCount-1)

		buckets[bucketIx]++
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

	outF, err := os.Create(outFn)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := outF.Close(); err != nil {
			panic(err)
		}
	}()

	outBuckets := [PhaseBucketCount]int{}

	for _, entry := range entries {
		r := rand.Float64()
		b, err := board.FromFEN(entry.fen)
		if err != nil {
			panic(err)
		}

		bucketIx := 0
		for pt := types.Pawn; pt < types.King; pt++ {
			bucketIx += b.Pieces[pt].Count() * eval.Phase[pt]
		}
		bucketIx = min(bucketIx , PhaseBucketCount-1)

		if r < keepProb[bucketIx] {
			fmt.Fprintf(outF, "%s; %.1f\n", entry.fen, wdlToEPD(entry.wdl))
			outBuckets[bucketIx]++
		}
	}
	fmt.Println(outBuckets)
}

type EPDEntry struct {
	fen string
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
	panic(fmt.Sprintf("unexpect wdl %d", n))
}
