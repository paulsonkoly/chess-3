package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/olekukonko/tablewriter"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/debug"
	"github.com/paulsonkoly/chess-3/heur"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"
	"github.com/paulsonkoly/chess-3/search"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

var debugFEN = flag.String("debugFEN", "", "Debug a given fen to a given depth using stockfish perft")
var debugDepth = flag.Int("debugDepth", 3, "Debug a given depth")
var cpuProf = flag.String("cpuProf", "", "cpu profile file name")
var memProf = flag.String("memProf", "", "mem profile file name")
var bench = flag.Bool("bench", false, "run benchmark instead of UCI")

type UciEngine struct {
	board       *board.Board
	sst         *search.State
	timeControl struct {
		wtime int // White time in milliseconds
		btime int // Black time in milliseconds
		winc  int // White increment per move in milliseconds
		binc  int // Black increment per move in milliseconds
	}
}

func NewUciEngine() *UciEngine {
	return &UciEngine{
		board: board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"),
		sst:   search.NewState(),
	}
}

func (e *UciEngine) handleCommand(command string) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case "uci":
		fmt.Println("id name chess-3")
		fmt.Println("id author Paul Sonkoly")
		// these are here to conform ob. we don't actually support these options.
		fmt.Println("option name Hash type spin default 1 min 1 max 1")
		fmt.Println("option name Threads type spin default 1 min 1 max 1")
		fmt.Println("uciok")
	case "isready":
		fmt.Println("readyok")
	case "position":
		e.handlePosition(parts[1:])
	case "go":
		e.handleGo(parts[1:])
	case "trace":
		e.handleTrace(parts[1:])
	case "fen":
		fmt.Println(e.board.FEN())
	case "quit":
		os.Exit(0)
	case "debug":
		switch parts[1] {

		case "on":
			e.sst.Debug = true

		case "off":
			e.sst.Debug = false
		}
	}
}

func (e *UciEngine) handlePosition(args []string) {
	if len(args) == 0 {
		return
	}

	if args[0] == "startpos" {
		e.board = board.FromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
		if len(args) > 2 && args[1] == "moves" {
			e.applyMoves(args[2:])
		}
	} else if args[0] == "fen" {
		fen := strings.Join(args[1:], " ")
		spaceIndex := strings.Index(fen, " moves ")
		if spaceIndex != -1 {
			e.board = board.FromFEN(fen[:spaceIndex])
			e.applyMoves(strings.Fields(fen[spaceIndex+7:]))
		} else {
			e.board = board.FromFEN(fen)
		}
	}
}

func (e *UciEngine) applyMoves(moves []string) {
	b := e.board
	for _, ms := range moves {
		sm := parseUCIMove(ms)

		m := movegen.FromSimple(b, sm)

		b.MakeMove(&m)
	}
}

func (e *UciEngine) handleTrace(args []string) {
	e.sst.Trace = make([]search.Trace, 0, 16)
	depth := Depth(1)
	skip := false

	for i := 0; i < len(args); i++ {
		if !skip {
			switch args[i] {

			case "depth":
				depth = Depth(parseInt(args[i+1]))
				skip = true

			default:
				e.writeTraceBuf(args[i:])
				goto End
			}
		} else {
      skip = false
    }
	}
End:

	e.sst.Stop = make(chan struct{})
	e.Search(depth)
}

func (e *UciEngine) writeTraceBuf(args []string) {
	if len(args) == 0 {
		return
	}

	sm := parseUCIMove(args[0])
	m := movegen.FromSimple(e.board, sm)
	e.sst.Trace = append(e.sst.Trace, search.Trace{Move: sm, Hash: e.board.Hash()})

	e.board.MakeMove(&m)

	e.writeTraceBuf(args[1:])

	e.board.UndoMove(&m)
}

func (e *UciEngine) handleGo(args []string) {
	depth := Depth(1) // Default depth if none is specified
	timeAllowed := 0

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "wtime":
			e.timeControl.wtime = parseMilliseconds(args[i+1])
		case "btime":
			e.timeControl.btime = parseMilliseconds(args[i+1])
		case "winc":
			e.timeControl.winc = parseMilliseconds(args[i+1])
		case "binc":
			e.timeControl.binc = parseMilliseconds(args[i+1])
		case "depth":
			depth = Depth(parseInt(args[i+1]))
		case "movetime":
			timeAllowed = parseMilliseconds(args[i+1]) - 30 // safety margin
		}
	}

	timeAllowed = e.TimeControl(timeAllowed)

	// Timeout handling with iterative deepening. If we issue a time based go
	// that closes Stop, and then subsequently a depth based go the Stop should
	// still be re-initialised otherwise the depth based go would abort
	// immediately.
	e.sst.Stop = make(chan struct{})

	if timeAllowed > 0 {
		var bestMove move.SimpleMove

		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer wg.Done()

			_, moves := e.Search(search.MaxPlies)

			if len(moves) > 0 {
				bestMove = moves[0]
			}
		}()

		time.Sleep(time.Duration(timeAllowed) * time.Millisecond)
		close(e.sst.Stop)
		wg.Wait()
		fmt.Printf("bestmove %s\n", bestMove)
	} else {
		// Fixed depth search

		_, moves := e.Search(depth)

		if len(moves) > 0 {
			bestMove := moves[0]
			fmt.Printf("bestmove %s\n", bestMove)
		} else {
			fmt.Println("bestmove 0000") // No legal move
		}
	}
}

func (e *UciEngine) Search(d Depth) (Score, []move.SimpleMove) {
	return search.Search(e.board, d, e.sst)
}

// 7800 that factors 39 * 200
var initialMatCount = int(16*heur.PieceValues[Pawn] +
	4*heur.PieceValues[Knight] +
	4*heur.PieceValues[Bishop] +
	4*heur.PieceValues[Rook] +
	2*heur.PieceValues[Queen])

func (e *UciEngine) TimeControl(timeAllowed int) int {
	if timeAllowed != 0 {
		return timeAllowed
	}

	if e.board.STM == White {
		if e.timeControl.wtime == 0 {
			return timeAllowed
		}
		timeAllowed = e.timeControl.wtime
	}

	if e.board.STM == Black {
		if e.timeControl.btime == 0 {
			return timeAllowed
		}
		timeAllowed = e.timeControl.btime
	}

	// TODO use the same functionality from eval

	matCount := e.board.Pieces[Queen].Count()*int(heur.PieceValues[Queen]) +
		e.board.Pieces[Rook].Count()*int(heur.PieceValues[Rook]) +
		e.board.Pieces[Bishop].Count()*int(heur.PieceValues[Bishop]) +
		e.board.Pieces[Knight].Count()*int(heur.PieceValues[Knight]) +
		e.board.Pieces[Pawn].Count()*int(heur.PieceValues[Pawn])

	matCount = min(matCount, initialMatCount)

	// linear interpolate initialMatCount -> 44 .. 0 -> 5 moves left
	movesLeft := (matCount / 200) + 5

	complexity := float64(matCount) / float64(initialMatCount)
	complexity = 1 - complexity // 1-(1-x)**2 tapers off around 1 (d = 0) and steep around 0
	complexity *= complexity
	complexity = 1 - complexity
	complexity *= 3.0 // scale up
	complexity += 0.2 // safety margin

	return int(math.Floor((complexity * float64(timeAllowed)) / float64(movesLeft)))
}

func parseUCIMove(uciM string) move.SimpleMove {
	from := Square((uciM[0] - 'a') + (uciM[1]-'1')*8)
	to := Square((uciM[2] - 'a') + (uciM[3]-'1')*8)
	var promo Piece
	if len(uciM) == 5 {
		switch uciM[4] {
		case 'q':
			promo = Queen
		case 'r':
			promo = Rook
		case 'b':
			promo = Bishop
		case 'n':
			promo = Knight
		}
	}
	return move.SimpleMove{From: from, To: to, Promo: promo}
}

func parseMilliseconds(value string) int {
	result, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return result
}

func parseInt(value string) int {
	result, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return result
}

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

	e := NewUciEngine()

	// our normal benchmark table
	if *bench {
		e.bench()
		return
	}

	// openbench compatibility bench
	if len(os.Args) > 1 && os.Args[1] == "bench" {
		fen := "2q1rr1k/3bbnnp/p2p1pp1/2pPp3/PpP1P1P1/1P2BNNP/2BQ1PRK/7R b - - 0 1"
		e.board = board.FromFEN(fen)
		e.Search(9)

		nodes := e.sst.ABCnt + e.sst.ABLeaf + e.sst.QCnt
		time := e.sst.Time

		fmt.Printf("%d nodes %d nps\n", nodes, 1000*nodes/int(time))
		return
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		e.handleCommand(scanner.Text())
	}

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

func (e *UciEngine) bench() {
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
		e.board = board.FromFEN(bk.fen)
		_, ms := e.Search(9)

		ok := ms[0].String() == bk.bm

		stats = append(stats, Stats{
			ok,
			e.sst.AWFail,
			e.sst.ABCnt + e.sst.ABLeaf,
			float32(e.sst.ABBreadth) / float32(e.sst.ABCnt),
			e.sst.TTHit,
			e.sst.QCnt,
			e.sst.QDepth,
			e.sst.QDelta,
			e.sst.QSEE,
			e.sst.Time,
			(e.sst.ABCnt + e.sst.ABLeaf + e.sst.QCnt) / int(e.sst.Time),
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
