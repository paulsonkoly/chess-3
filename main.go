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
		fmt.Println("uciok")
	case "isready":
		fmt.Println("readyok")
	case "position":
		e.handlePosition(parts[1:])
	case "go":
		e.handleGo(parts[1:])
	case "fen":
		fmt.Println(e.board.FEN())
	case "quit":
		os.Exit(0)
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
		from, to, promo := parseUCIMove(ms)

		m := movegen.UCIMove(b, from, to, promo)

		b.MakeMove(&m)
	}
}

func (e *UciEngine) handleGo(args []string) {
	depth := Depth(10) // Default depth if none is specified
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

	if timeAllowed > 0 {
		// Timeout handling with iterative deepening
		e.sst.Stop = make(chan struct{})
		var bestMove move.SimpleMove

		wg := sync.WaitGroup{}
		wg.Add(1)

		score := Score(0)
		go func() {
			defer wg.Done()

			s, moves := e.Search(100)

			if len(moves) > 0 {
				bestMove = moves[0]
				score = s
			}
		}()

		time.Sleep(time.Duration(timeAllowed) * time.Millisecond)
		close(e.sst.Stop)
		wg.Wait()
		fmt.Printf("bestmove %s info score cp %d\n", bestMove, score)
	} else {
		// Fixed depth search
		start := time.Now()

		score, moves := e.Search(depth)

		if len(moves) > 0 {
			bestMove := moves[0]
			elapsed := time.Since(start).Milliseconds()
			fmt.Printf("bestmove %s info score cp %d time %d\n", bestMove, score, elapsed)
		} else {
			fmt.Println("bestmove 0000") // No legal move
		}
	}
}

func (e *UciEngine) Search(d Depth) (Score, []move.SimpleMove) {
	s, moves := search.Search(e.board, d, e.sst)

	ABBF := float64(e.sst.ABBreadth) / float64(e.sst.ABCnt)

	fmt.Printf("info awfail %d ableaf %d abbf %.2f tthits %d qdepth %d qdelta %d qsee %d\n",
		e.sst.AWFail, e.sst.ABLeaf, ABBF, e.sst.TTHit, e.sst.QDepth, e.sst.QDelta, e.sst.QSEE)

	return s, moves
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

func parseUCIMove(move string) (Square, Square, Piece) {
	from := Square((move[0] - 'a') + (move[1]-'1')*8)
	to := Square((move[2] - 'a') + (move[3]-'1')*8)
	var promo Piece
	if len(move) == 5 {
		switch move[4] {
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
	return from, to, promo
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

	if *bench {
		bratkoKopec := []string{
			"1k1r4/pp1b1R2/3q2pp/4p3/2B5/4Q3/PPP2B2/2K5 b - - 0 1",
			"3r1k2/4npp1/1ppr3p/p6P/P2PPPP1/1NR5/5K2/2R5 w - - 0 1",
			"2q1rr1k/3bbnnp/p2p1pp1/2pPp3/PpP1P1P1/1P2BNNP/2BQ1PRK/7R b - - 0 1",
			"rnbqkb1r/p3pppp/1p6/2ppP3/3N4/2P5/PPP1QPPP/R1B1KB1R w KQkq - 0 1",
			"r1b2rk1/2q1b1pp/p2ppn2/1p6/3QP3/1BN1B3/PPP3PP/R4RK1 w - - 0 1",
			"2r3k1/pppR1pp1/4p3/4P1P1/5P2/1P4K1/P1P5/8 w - - 0 1",
			"1nk1r1r1/pp2n1pp/4p3/q2pPp1N/b1pP1P2/B1P2R2/2P1B1PP/R2Q2K1 w - - 0 1",
			"4b3/p3kp2/6p1/3pP2p/2pP1P2/4K1P1/P3N2P/8 w - - 0 1",
			"2kr1bnr/pbpq4/2n1pp2/3p3p/3P1P1B/2N2N1Q/PPP3PP/2KR1B1R w - - 0 1",
			"3rr1k1/pp3pp1/1qn2np1/8/3p4/PP1R1P2/2P1NQPP/R1B3K1 b - - 0 1",
			"2r1nrk1/p2q1ppp/bp1p4/n1pPp3/P1P1P3/2PBB1N1/4QPPP/R4RK1 w - - 0 1",
			"r3r1k1/ppqb1ppp/8/4p1NQ/8/2P5/PP3PPP/R3R1K1 b - - 0 1",
			"r2q1rk1/4bppp/p2p4/2pP4/3pP3/3Q4/PP1B1PPP/R3R1K1 w - - 0 1",
			"rnb2r1k/pp2p2p/2pp2p1/q2P1p2/8/1Pb2NP1/PB2PPBP/R2Q1RK1 w - - 0 1",
			"2r3k1/1p2q1pp/2b1pr2/p1pp4/6Q1/1P1PP1R1/P1PN2PP/5RK1 w - - 0 1",
			"r1bqkb1r/4npp1/p1p4p/1p1pP1B1/8/1B6/PPPN1PPP/R2Q1RK1 w kq - 0 1",
			"r2q1rk1/1ppnbppp/p2p1nb1/3Pp3/2P1P1P1/2N2N1P/PPB1QP2/R1B2RK1 b - - 0 1",
			"r1bq1rk1/pp2ppbp/2np2p1/2n5/P3PP2/N1P2N2/1PB3PP/R1B1QRK1 b - - 0 1",
			"3rr3/2pq2pk/p2p1pnp/8/2QBPP2/1P6/P5PP/4RRK1 b - - 0 1",
			"r4k2/pb2bp1r/1p1qp2p/3pNp2/3P1P2/2N3P1/PPP1Q2P/2KRR3 w - - 0 1",
			"3rn2k/ppb2rpp/2ppqp2/5N2/2P1P3/1P5Q/PB3PPP/3RR1K1 w - - 0 1",
			"2r2rk1/1bqnbpp1/1p1ppn1p/pP6/N1P1P3/P2B1N1P/1B2QPP1/R2R2K1 b - - 0 1",
			"r1bqk2r/pp2bppp/2p5/3pP3/P2Q1P2/2N1B3/1PP3PP/R4RK1 b kq - 0 1",
			"r2qnrnk/p2b2b1/1p1p2pp/2pPpp2/1PP1P3/PRNBB3/3QNPPP/5RK1 w - - 0 1",
		}

		bf := 0.0
		start := time.Now()

		for _, fen := range bratkoKopec {
			e.board = board.FromFEN(fen)
			e.Search(9)

			bf += float64(e.sst.ABBreadth) / float64(e.sst.ABCnt)
		}
		elapsed := time.Since(start).Milliseconds()
		fmt.Printf("total time %dms\n", elapsed)
		fmt.Println("average branching factor ", bf/float64(len(bratkoKopec)))
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			e.handleCommand(scanner.Text())
		}
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
