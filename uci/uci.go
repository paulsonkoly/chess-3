package uci

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/heur"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"
	"github.com/paulsonkoly/chess-3/search"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

const (
	startPos    = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	defaultHash = 8
)

type Engine struct {
	Board      *board.Board
	SST        *search.State
	input      *bufio.Scanner
	inputLines chan string
	stop       chan struct{}
}

func NewEngine() *Engine {
	return &Engine{
		Board: board.FromFEN(startPos),
		SST:   search.NewState(defaultHash),
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

	case "isready":
		fmt.Println("readyok")

	case "position":
		e.handlePosition(parts[1:])

	case "go":
		e.handleGo(parts[1:])

	case "fen":
		fmt.Println(e.Board.FEN())

	case "setoption":
		e.handleSetOption(parts[1:])

	case "quit":
		os.Exit(0)

	case "debug":
		switch parts[1] {

		case "on":
			e.SST.Debug = true

		case "off":
			e.SST.Debug = false
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

		e.SST = search.NewState(val) // we need to re-allocate the hash table
	}
}

func (e *Engine) handlePosition(args []string) {
	if len(args) == 0 {
		return
	}

	if args[0] == "startpos" {
		e.Board = board.FromFEN(startPos)
		if len(args) > 2 && args[1] == "moves" {
			e.applyMoves(args[2:])
		}
	} else if args[0] == "fen" {
		fen := strings.Join(args[1:], " ")
		spaceIndex := strings.Index(fen, " moves ")
		if spaceIndex != -1 {
			e.Board = board.FromFEN(fen[:spaceIndex])
			e.applyMoves(strings.Fields(fen[spaceIndex+7:]))
		} else {
			e.Board = board.FromFEN(fen)
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

type timeControl struct {
	wtime int64 // White time in milliseconds
	btime int64 // Black time in milliseconds
	winc  int64 // White increment per move in milliseconds
	binc  int64 // Black increment per move in milliseconds
	mtime int64 // move time
}

// 7800 that factors 39 * 200
var initialMatCount = int(16*heur.PieceValues[Pawn] +
	4*heur.PieceValues[Knight] +
	4*heur.PieceValues[Bishop] +
	4*heur.PieceValues[Rook] +
	2*heur.PieceValues[Queen])

const MinTime = 30

func (tc timeControl) allocate(b *board.Board) int64 {
	if tc.mtime >= MinTime {
		return tc.mtime // safety margin
	}
	if tc.mtime != 0 {
		return 0
	}

	gameTime := int64(0)
	if b.STM == White {
		if tc.wtime == 0 {
			return 0
		}
		gameTime = tc.wtime
	}

	if b.STM == Black {
		if tc.btime == 0 {
			return 0
		}
		gameTime = tc.btime
	}

	// TODO use the same functionality from eval

	matCount := b.Pieces[Queen].Count()*int(heur.PieceValues[Queen]) +
		b.Pieces[Rook].Count()*int(heur.PieceValues[Rook]) +
		b.Pieces[Bishop].Count()*int(heur.PieceValues[Bishop]) +
		b.Pieces[Knight].Count()*int(heur.PieceValues[Knight]) +
		b.Pieces[Pawn].Count()*int(heur.PieceValues[Pawn])

	matCount = min(matCount, initialMatCount)

	// linear interpolate initialMatCount -> 44 .. 0 -> 5 moves left
	movesLeft := (matCount / 200) + 5

	complexity := float64(matCount) / float64(initialMatCount)
	complexity = 1 - complexity // 1-(1-x)**2 tapers off around 1 (d = 0) and steep around 0
	complexity *= complexity
	complexity = 1 - complexity
	complexity *= 3.0 // scale up
	complexity += 0.2 // safety margin

	return int64(math.Floor((complexity * float64(gameTime)) / float64(movesLeft)))
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

	allocTime := tc.allocate(e.Board)

	e.SST.Stop = make(chan struct{})
	e.SST.SoftTime = 0

	if allocTime > 0 {
		depth = search.MaxPlies
		e.SST.SoftTime = allocTime * 3 / 4
	} else {
		allocTime = 1 << 50 // not timed mode, essentially disable timeout
	}

	var move move.SimpleMove

	searchFin := make(chan struct{})

	go func() {
		_, move = e.Search(depth)
		close(searchFin)
	}()

	stopped := false
	for finished := false; !finished; {
		select {

		case <-searchFin:
			finished = true

		case <-time.After(time.Duration(allocTime) * time.Millisecond):
			stopped = true
			close(e.SST.Stop)

		case <-e.stop:
			stopped = true
			close(e.SST.Stop)
		}
	}

	if !stopped {
		close(e.SST.Stop)
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

func (e *Engine) Search(d Depth) (Score, move.SimpleMove) {
	return search.Search(e.Board, d, e.SST)
}
