package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/debug"
	"github.com/paulsonkoly/chess-3/eval"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"
	"github.com/paulsonkoly/chess-3/search"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

var debugFEN = flag.String("debugFEN", "", "Debug a given fen to a given depth using stockfish perft")
var debugDepth = flag.Int("debugDepth", 3, "Debug a given depth")
var cpuProf = flag.String("cpuProf", "", "cpu profile file name")

type UciEngine struct {
	board       *board.Board
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
	depth := 10 // Default depth if none is specified
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
			depth = parseInt(args[i+1])
		case "movetime":
			timeAllowed = parseMilliseconds(args[i+1]) - 30 // safety margin
		}
	}

	timeAllowed = e.TimeControl(timeAllowed)

	if timeAllowed > 0 {
		// Timeout handling with iterative deepening
		stop := make(chan struct{})
		var bestMove move.Move

		wg := sync.WaitGroup{}
		wg.Add(1)

		score := 0
		go func() {
			defer wg.Done()
			s, moves := search.Search(e.board, 100, stop)
			fmt.Printf("info awfail %d ableaf %d qdepth %d qdelta %d qsee %d\n", search.AWFail, search.ABLeaf, search.QDepth, search.QDelta, search.QSEE)
			search.AWFail = 0
			search.ABLeaf = 0
			search.QDelta = 0
			search.QDepth = 0
			search.QSEE = 0
			if len(moves) > 0 {
				bestMove = moves[0]
				score = s
			}
		}()

		time.Sleep(time.Duration(timeAllowed) * time.Millisecond)
		close(stop)
		wg.Wait()
		fmt.Printf("bestmove %s info score cp %d\n", bestMove, score)
	} else {
		// Fixed depth search
		start := time.Now()
		score, moves := search.Search(e.board, depth, nil)

		if len(moves) > 0 {
			bestMove := moves[0]
			elapsed := time.Since(start).Milliseconds()
			fmt.Printf("bestmove %s info score cp %d time %d\n", bestMove, score, elapsed)
		} else {
			fmt.Println("bestmove 0000") // No legal move
		}
	}
}

// 7800 that factors 39 * 200
var initialMatCount = 16*eval.PieceValues[Pawn] +
	4*eval.PieceValues[Knight] +
	4*eval.PieceValues[Bishop] +
	4*eval.PieceValues[Rook] +
	2*eval.PieceValues[Queen]

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

	matCount := e.board.Pieces[Queen].Count()*eval.PieceValues[Queen] +
		e.board.Pieces[Rook].Count()*eval.PieceValues[Rook] +
		e.board.Pieces[Bishop].Count()*eval.PieceValues[Bishop] +
		e.board.Pieces[Knight].Count()*eval.PieceValues[Knight] +
		e.board.Pieces[Pawn].Count()*eval.PieceValues[Pawn]

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
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		e.handleCommand(scanner.Text())
	}
}
