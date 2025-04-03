package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/olekukonko/tablewriter"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/debug"
	"github.com/paulsonkoly/chess-3/uci"
	"slices"
)

var debugFEN = flag.String("debugFEN", "", "Debug a given fen to a given depth using stockfish perft")
var debugDepth = flag.Int("debugDepth", 3, "Debug a given depth")
var cpuProf = flag.String("cpuProf", "", "cpu profile file name")
var memProf = flag.String("memProf", "", "mem profile file name")
var bench = flag.Bool("bench", false, "run benchmark instead of UCI")

func main() {

	flag.Parse()

	if *cpuProf != "" {
		cpu, err := os.Create(*cpuProf)
		if err != nil {
			panic(err)
		}
		err = pprof.StartCPUProfile(cpu)
		if err != nil {
			panic(err)
		}
		defer pprof.StopCPUProfile()
	}

	if *debugFEN != "" {
		b := board.FromFEN(*debugFEN)

		debug.MatchPerft(b, *debugDepth)
		return
	}

	e := uci.NewEngine()

	// our normal benchmark table
	if *bench {
		runBench(e)
		return
	}

	// openbench compatibility bench
	if slices.Contains(os.Args, "bench") {
		runOBBench(e)
		return
	}

	e.Run()

	if *memProf != "" {
		f, err := os.Create(*memProf)
		if err != nil {
			panic(err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			panic(err)
		}
	}
}

// Stats contains search statistics. See search.State.
type Stats struct {
	Ok     bool
	AWFail int
	ABCnt  int
	ABBF   float32
	TTHit  int
	QCnt   int
	QDepth int
	QDelta int
	QSEE   int
	Time   int64
	KNps   int
}

func runBench(e *uci.Engine) {
	bratkoKopec := []struct {
		fen string
		bm  string
	}{
		{"1k1r4/pp1b1R2/3q2pp/4p3/2B5/4Q3/PPP2B2/2K5 b - -0 1", "d6d1"},
		{"3r1k2/4npp1/1ppr3p/p6P/P2PPPP1/1NR5/5K2/2R5 w - -0 1", "d4d5"},
		{"2q1rr1k/3bbnnp/p2p1pp1/2pPp3/PpP1P1P1/1P2BNNP/2BQ1PRK/7R b - -0 1", "f6f5"},
		{"rnbqkb1r/p3pppp/1p6/2ppP3/3N4/2P5/PPP1QPPP/R1B1KB1R w KQkq -0 1", "e5e6"},
		{"r1b2rk1/2q1b1pp/p2ppn2/1p6/3QP3/1BN1B3/PPP3PP/R4RK1 w - -0 1", "c3d5"}, // a2a4
		{"2r3k1/pppR1pp1/4p3/4P1P1/5P2/1P4K1/P1P5/8 w - -0 1", "g5g6"},
		{"1nk1r1r1/pp2n1pp/4p3/q2pPp1N/b1pP1P2/B1P2R2/2P1B1PP/R2Q2K1 w - -0 1", "h5f6"},
		{"4b3/p3kp2/6p1/3pP2p/2pP1P2/4K1P1/P3N2P/8 w - -0 1", "f4f5"},
		{"2kr1bnr/pbpq4/2n1pp2/3p3p/3P1P1B/2N2N1Q/PPP3PP/2KR1B1R w - -0 1", "f4f5"},
		{"3rr1k1/pp3pp1/1qn2np1/8/3p4/PP1R1P2/2P1NQPP/R1B3K1 b - -0 1", "f3e5"},
		{"2r1nrk1/p2q1ppp/bp1p4/n1pPp3/P1P1P3/2PBB1N1/4QPPP/R4RK1 w - -0 1", "f2f4"},
		{"r3r1k1/ppqb1ppp/8/4p1NQ/8/2P5/PP3PPP/R3R1K1 b - -0 1", "d7f5"},
		{"r2q1rk1/4bppp/p2p4/2pP4/3pP3/3Q4/PP1B1PPP/R3R1K1 w - -0 1", "b2b4"},
		{"rnb2r1k/pp2p2p/2pp2p1/q2P1p2/8/1Pb2NP1/PB2PPBP/R2Q1RK1 w - -0 1", "d1d2 d1e1"},
		{"2r3k1/1p2q1pp/2b1pr2/p1pp4/6Q1/1P1PP1R1/P1PN2PP/5RK1 w - -0 1", "g4g7"},
		{"r1bqkb1r/4npp1/p1p4p/1p1pP1B1/8/1B6/PPPN1PPP/R2Q1RK1 w kq -0 1", "d2e4"},
		{"r2q1rk1/1ppnbppp/p2p1nb1/3Pp3/2P1P1P1/2N2N1P/PPB1QP2/R1B2RK1 b - -0 1", "h7h5"},
		{"r1bq1rk1/pp2ppbp/2np2p1/2n5/P3PP2/N1P2N2/1PB3PP/R1B1QRK1 b - -0 1", "c5b3"},
		{"3rr3/2pq2pk/p2p1pnp/8/2QBPP2/1P6/P5PP/4RRK1 b - -0 1", "e8e4"},
		{"r4k2/pb2bp1r/1p1qp2p/3pNp2/3P1P2/2N3P1/PPP1Q2P/2KRR3 w - -0 1", "g3g4"},
		{"3rn2k/ppb2rpp/2ppqp2/5N2/2P1P3/1P5Q/PB3PPP/3RR1K1 w - -0 1", "f5h6"},
		{"2r2rk1/1bqnbpp1/1p1ppn1p/pP6/N1P1P3/P2B1N1P/1B2QPP1/R2R2K1 b - -0 1", "b7e4"},
		{"r1bqk2r/pp2bppp/2p5/3pP3/P2Q1P2/2N1B3/1PP3PP/R4RK1 b kq -0 1", "f7f6"},
		{"r2qnrnk/p2b2b1/1p1p2pp/2pPpp2/1PP1P3/PRNBB3/3QNPPP/5RK1 w - -0 1", "f5f4"},
	}

	stats := []Stats{}

	for _, bk := range bratkoKopec {
		e.Board = board.FromFEN(bk.fen)
		_, ms := e.Search(9)

		ok := ms[0].String() == bk.bm

		stats = append(stats, Stats{
			ok,
			e.SST.AWFail,
			e.SST.ABCnt + e.SST.ABLeaf,
			float32(e.SST.ABBreadth) / float32(e.SST.ABCnt),
			e.SST.TTHit,
			e.SST.QCnt,
			e.SST.QDepth,
			e.SST.QDelta,
			e.SST.QSEE,
			e.SST.Time,
			(e.SST.ABCnt + e.SST.ABLeaf + e.SST.QCnt) / int(e.SST.Time),
		})
	}

	avg := Stats{}

	for _, stat := range stats {
		avg.AWFail += stat.AWFail
		avg.ABCnt += stat.ABCnt
		avg.ABBF += stat.ABBF
		avg.TTHit += stat.TTHit
		avg.QCnt += stat.QCnt
		avg.QDepth += stat.QDepth
		avg.QDelta += stat.QDelta
		avg.QSEE += stat.QSEE
		avg.Time += stat.Time
		avg.KNps += stat.KNps
	}

	table := tablewriter.NewWriter(os.Stdout)

	okCnt := 0

	for ix, stat := range stats {
		var ok string
		if stat.Ok {
			ok = "✓"
			okCnt++
		} else {
			ok = "❌"
		}
		aWFail := fmt.Sprintf("%d", stat.AWFail)
		aBCnt := fmt.Sprintf("%d", stat.ABCnt)
		abBF := fmt.Sprintf("%.4f", stat.ABBF)
		tTHit := fmt.Sprintf("%d", stat.TTHit)
		qCnt := fmt.Sprintf("%d", stat.QCnt)
		qDepth := fmt.Sprintf("%d", stat.QDepth)
		qDelta := fmt.Sprintf("%d", stat.QDelta)
		qSEE := fmt.Sprintf("%d", stat.QSEE)
		timeMs := fmt.Sprintf("%d", stat.Time)
		kNps := fmt.Sprintf("%d", stat.KNps)

		table.Append([]string{fmt.Sprintf("BK %d", ix+1), ok, aWFail, aBCnt, abBF, tTHit, qCnt, qDepth, qDelta, qSEE, timeMs, kNps})
	}

	table.Append([]string{"average",
		fmt.Sprintf("%d / %d", okCnt, len(bratkoKopec)),
		fmt.Sprintf("%.2f", float32(avg.AWFail)/float32(len(bratkoKopec))),
		fmt.Sprintf("%.2f", float64(avg.ABCnt)/float64(len(bratkoKopec))),
		fmt.Sprintf("%.4f", avg.ABBF/float32(len(bratkoKopec))),
		fmt.Sprintf("%.2f", float64(avg.TTHit)/float64(len(bratkoKopec))),
		fmt.Sprintf("%.2f", float64(avg.QCnt)/float64(len(bratkoKopec))),
		fmt.Sprintf("%.2f", float32(avg.QDepth)/float32(len(bratkoKopec))),
		fmt.Sprintf("%.2f", float64(avg.QDelta)/float64(len(bratkoKopec))),
		fmt.Sprintf("%.2f", float64(avg.QSEE)/float64(len(bratkoKopec))),
		fmt.Sprintf("%.2f", float64(avg.Time)/float64(len(bratkoKopec))),
		fmt.Sprintf("%.2f", float64(avg.KNps)/float64(len(bratkoKopec))),
	})

	table.SetHeader([]string{"Test", "BM", "AWFail", "ABCnt", "ABBF", "TTHit", "QCnt", "QDepth", "QDelta", "QSEE", "Time (ms)", "Speed (Kn/s)"})
	table.SetAutoWrapText(false)
	table.Render()
}

func runOBBench(e *uci.Engine) {
	fen := "2q1rr1k/3bbnnp/p2p1pp1/2pPp3/PpP1P1P1/1P2BNNP/2BQ1PRK/7R b - - 0 1"
	e.Board = board.FromFEN(fen)
	e.Search(15)

	nodes := e.SST.ABCnt + e.SST.ABLeaf + e.SST.QCnt
	time := e.SST.Time

	fmt.Printf("%d nodes %d nps\n", nodes, 1000*nodes/int(time))
}
