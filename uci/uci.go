package uci

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/eval"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"
	"github.com/paulsonkoly/chess-3/search"
	"github.com/paulsonkoly/chess-3/transp"

	. "github.com/paulsonkoly/chess-3/types"
)

const (
	startPos    = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	defaultHash = 8
)

type Engine struct {
	Board      *board.Board
	Search     *search.Search
	debug      bool
	input      *bufio.Scanner
	inputLines chan string
	stop       chan struct{}
}

func NewEngine() *Engine {
	return &Engine{
		Board:  Must(board.FromFEN(startPos)),
		Search: search.New(defaultHash * transp.MegaBytes),
	}
}

// Run executes an input loop reading from stdin and in paralell running and
// controlling the search. It supports search interrupts with time control or
// stop command.
func (e *Engine) Run() {
	wg := sync.WaitGroup{}

	e.input = bufio.NewScanner(os.Stdin)
	e.inputLines = make(chan string)
	e.stop = make(chan struct{})

	wg.Add(2)

	go func() {
		e.readInput()
		close(e.inputLines)
		wg.Done()
	}()

	go func() {
		e.handleInput()
		wg.Done()
	}()

	wg.Wait()

	close(e.stop)
}

func (e *Engine) readInput() {
	for e.input.Scan() {
		line := e.input.Text()

		switch line {

		case "stop":
			select {
			case e.stop <- struct{}{}:
			default:
				// no search is running. Ignore.
			}

		case "quit":
			return

		case "isready":
			fmt.Println("readyok")

		default:
			e.inputLines <- line
		}
	}
}

func (e *Engine) handleInput() {
	for line := range e.inputLines {
		e.handleCommand(line)
	}
}

func (e *Engine) handleCommand(command string) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case "uci":
		fmt.Println("id name chess-3")
		fmt.Println("id author Paul Sonkoly")
		fmt.Printf("option name Hash type spin default %d min 1 max 128\n", defaultHash)
		// these are here to conform ob. we don't actually support these options.
		fmt.Println("option name Threads type spin default 1 min 1 max 1")
		fmt.Println("uciok")

	case "ucinewgame":
		e.Search.Clear()

	case "position":
		e.handlePosition(parts[1:])

	case "go":
		e.handleGo(parts[1:])

	case "fen":
		fmt.Println(e.Board.FEN())

	case "setoption":
		e.handleSetOption(parts[1:])

	case "eval":
		e.handleEval()

	case "quit":
		os.Exit(0)

	case "debug":
		switch parts[1] {

		case "on":
			e.debug = true

		case "off":
			e.debug = false
		}
	}
}

func (e *Engine) handleSetOption(args []string) {
	if len(args) != 4 || args[0] != "name" || args[2] != "value" {
		return
	}
	switch args[1] {

	case "Hash":
		val, err := strconv.Atoi(args[3])
		if err != nil || val < 1 || val&(val-1) != 0 {
			return
		}

		// TODO re-allocate or better yet increase  / reduce the tt only
		e.Search = search.New(val * transp.MegaBytes) // we need to re-allocate the hash table
	}
}

func (e *Engine) handlePosition(args []string) {
	if len(args) == 0 {
		return
	}

	switch args[0] {

	case "startpos":
		e.Board = Must(board.FromFEN(startPos))
		if len(args) > 2 && args[1] == "moves" {
			e.applyMoves(args[2:])
		}

	case "fen":
		if len(args) < 7 {
			fmt.Fprintf(os.Stderr, "not enough arguments %d\n", len(args))
			return
		}

		fen := strings.Join(args[1:7], " ")
		b, err := board.FromFEN(fen)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid fen %v\n", err)
			return
		}
		if b.InvalidPieceCount() {
			fmt.Fprintln(os.Stderr, "invalid piece counts")
			return
		}
		e.Board = b

		if len(args) >= 8 && args[7] == "moves" {
			e.applyMoves(args[8:])
		}
	}
}

func (e *Engine) applyMoves(moves []string) {
	b := e.Board
	for _, ms := range moves {
		sm := parseUCIMove(ms)

		m := movegen.FromSimple(b, sm)

		b.MakeMove(&m)
	}
}

func parseUCIMove(uciM string) move.SimpleMove {
	var m move.SimpleMove

	m.SetFrom(Square((uciM[0] - 'a') + (uciM[1]-'1')*8))
	m.SetTo(Square((uciM[2] - 'a') + (uciM[3]-'1')*8))
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
	m.SetPromo(promo)
	return m
}

func (e *Engine) handleEval() {
	fmt.Println(eval.Eval(e.Board, &eval.Coefficients))
}

type timeControl struct {
	wtime int64 // White time in milliseconds
	btime int64 // Black time in milliseconds
	winc  int64 // White increment per move in milliseconds
	binc  int64 // Black increment per move in milliseconds
	mtime int64 // move time
}

func (tc timeControl) timedMode(stm Color) bool {
	return (stm == White && tc.wtime > 0) || (stm == Black && tc.btime > 0) || tc.mtime > 0
}

const (
	TimeSafetyMargin = 30
	TimeInf          = int64(1 << 50)
)

func (tc timeControl) softLimit(stm Color) int64 {
	if tc.mtime > 0 {
		return tc.mtime
	}

	if stm == White && tc.wtime > 0 {
		return tc.wtime/20 + tc.winc/2
	}

	if stm == Black && tc.btime > 0 {
		return tc.btime/20 + tc.binc/2
	}

	return TimeInf
}

func (tc timeControl) hardLimit(stm Color) int64 {
	if tc.mtime > 0 {
		return tc.mtime
	}

	timeLeft := TimeInf

	if stm == White && tc.wtime > 0 {
		timeLeft = tc.wtime
	}

	if stm == Black && tc.btime > 0 {
		timeLeft = tc.btime
	}

	if timeLeft <= TimeSafetyMargin {
		// we are losing on time anyway, but at least allocate time
		return timeLeft
	}

	return Clamp(4*tc.softLimit(stm), TimeSafetyMargin, timeLeft-TimeSafetyMargin)
}

func (e *Engine) handleGo(args []string) {
	depth := Depth(1) // Default depth if none is specified

	tc := timeControl{}

	for i := range len(args) {
		switch args[i] {
		case "wtime":
			tc.wtime = parseInt64(args[i+1])
		case "btime":
			tc.btime = parseInt64(args[i+1])
		case "winc":
			tc.winc = parseInt64(args[i+1])
		case "binc":
			tc.binc = parseInt64(args[i+1])
		case "depth":
			depth = Depth(parseInt(args[i+1]))
		case "movetime":
			tc.mtime = parseInt64(args[i+1])
		}
	}

	stm := e.Board.STM

	if tc.timedMode(stm) {
		depth = MaxPlies
	}

	stop := make(chan struct{})
	softTime := tc.softLimit(stm)

	var move move.SimpleMove

	searchFin := make(chan struct{})

	go func() {
		_, move = e.Search.WithOptions(
			e.Board,
			depth,
			search.WithStop(stop),
			search.WithSoftTime(softTime),
			search.WithDebug(e.debug),
		)
		close(searchFin)
	}()

	stopped := false
	for finished := false; !finished; {
		select {

		case <-searchFin:
			finished = true

		case <-time.After(time.Duration(tc.hardLimit(stm)) * time.Millisecond):
			if !stopped {
				stopped = true
				close(stop)
			}

		case <-e.stop:
			if !stopped {
				stopped = true
				close(stop)
			}
		}
	}

	if !stopped {
		close(stop)
	}

	fmt.Printf("bestmove %s\n", move)
}

func parseInt(value string) int {
	result, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return result
}

func parseInt64(value string) int64 {
	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return result
}
