package uci

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/debug"
	"github.com/paulsonkoly/chess-3/eval"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/params"
	"github.com/paulsonkoly/chess-3/search"
	"github.com/paulsonkoly/chess-3/transp"

	. "github.com/paulsonkoly/chess-3/chess"
)

const (
	defaultHash = 1
	minimalHash = 1
	maximalHash = 1024
)

var GitVersion = "dev"

// Driver is an UCI (universal chess interface) driver for the underlying chess
// engine logic. It is responsible to receive and interpret UCI protocol
// commands and invoke relevant engine functions - such as search in turn. It
// is also responsible for handling asynchronous searches and issuing stop to
// the engine code if needed.
type Driver struct {
	board      *board.Board
	search     Search
	input      *bufio.Scanner
	output     io.Writer
	err        io.Writer
	inputLines chan string
	stop       chan struct{}
	debug      bool
}

type Search interface {
	Go(*board.Board, ...search.Option) (Score, move.Move)
	Clear()
	ResizeTT(int)
}

type driverOpts struct {
	input  io.Reader
	output io.Writer
	err    io.Writer
	search Search
}

// WithInput replaces the default os.Stdin in the driver with the user specified io.Reader.
func WithInput(input io.Reader) DriverOpt { return func(o *driverOpts) { o.input = input } }

// WithOutput replaces the default os.Stdout in the driver with the user specified io.Writer.
func WithOutput(output io.Writer) DriverOpt { return func(o *driverOpts) { o.output = output } }

// WithError replaces the default os.Stderr in the driver with the user specified io.Writer.
func WithError(err io.Writer) DriverOpt { return func(o *driverOpts) { o.err = err } }

// WithSearch replaces the default Search with the user specified one.
func WithSearch(s Search) DriverOpt { return func(o *driverOpts) { o.search = s } }

// DriverOpt is an option for creating a new UCI driver.
type DriverOpt func(*driverOpts)

// NewDriver creates a new UCI driver based on opts.
func NewDriver(opts ...DriverOpt) *Driver {
	actual := driverOpts{
		input:  os.Stdin,
		output: os.Stdout,
		err:    os.Stderr,
	}

	for _, opt := range opts {
		opt(&actual)
	}

	if actual.search == nil {
		actual.search = search.New(1 * transp.MegaBytes)
	}

	return &Driver{
		board:  board.StartPos(),
		search: actual.search,
		input:  bufio.NewScanner(actual.input),
		output: actual.output,
		err:    actual.err,
	}
}

// Run executes an input loop reading from stdin and in parallel running and
// controlling the search. It supports search interrupts with time control or
// stop command.
func (d *Driver) Run() {
	wg := sync.WaitGroup{}

	d.inputLines = make(chan string)
	d.stop = make(chan struct{})

	wg.Add(2)

	go func() {
		d.readInput()
		close(d.inputLines)
		wg.Done()
	}()

	go func() {
		d.handleInput()
		wg.Done()
	}()

	wg.Wait()

	close(d.stop)
}

func (d *Driver) readInput() {
	for d.input.Scan() {
		line := d.input.Text()
		args := strings.Fields(line)

		if len(args) < 1 {
			continue
		}

		switch args[0] {

		case "stop":
			select {
			case d.stop <- struct{}{}:
			default:
				// no search is running. Ignore.
			}

		case "quit":
			select {
			case d.stop <- struct{}{}:
			default:
				// no search is running. Ignore.
			}
			return

		case "isready":
			fmt.Fprintln(d.output, "readyok")

		default:
			d.inputLines <- line
		}
	}
}

func (d *Driver) handleInput() {
	for line := range d.inputLines {
		d.handleCommand(line)
	}
}

func (d *Driver) handleCommand(command string) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case "uci":
		fmt.Fprintf(d.output, "id name chess-3 %s\n", GitVersion)
		fmt.Fprintln(d.output, "id author Paul Sonkoly")
		fmt.Fprintf(d.output, "option name Hash type spin default %d min %d max %d\n", defaultHash, minimalHash, maximalHash)
		// these are here to conform ob. we don't actually support these options.
		fmt.Fprintln(d.output, "option name Threads type spin default 1 min 1 max 1")
		// spsa options
		fmt.Fprint(d.output, params.UCIOptions())
		fmt.Fprintln(d.output, "uciok")

	case "ucinewgame":
		d.search.Clear()

	case "position":
		d.handlePosition(parts[1:])

	case "go":
		d.handleGo(parts[1:])

	case "fen":
		fmt.Fprintln(d.output, d.board.FEN())

	case "setoption":
		d.handleSetOption(parts[1:])

	case "eval":
		d.handleEval()

	case "perft":
		if len(parts) < 2 {
			fmt.Fprintln(d.err, "depth missing")
			break
		}

		depth, err := strconv.Atoi(parts[1])
		if err != nil {
			fmt.Fprintln(d.err, err)
			break
		}
		if depth < 0 || depth > 30 {
			fmt.Fprintln(d.err, "unsupported depth")
			break
		}

		fmt.Fprintln(d.output, debug.Perft(d.board, Depth(depth), true))

	case "debug":
		if len(parts) < 2 {
			fmt.Fprintln(d.err, "on/off missing")
			break
		}

		switch parts[1] {

		case "on":
			d.debug = true

		case "off":
			d.debug = false
		}

	case "spsa":
		fmt.Fprint(d.output, params.OpenbenchInfo())
	}
}

func (d *Driver) handleSetOption(args []string) {
	if len(args) < 4 {
		fmt.Fprintln(d.err, "argument missing")
		return
	}
	if args[0] != "name" || args[2] != "value" {
		return
	}
	switch args[1] {

	case "Hash":
		val, err := strconv.Atoi(args[3])
		if err != nil || val < minimalHash || val > maximalHash {
			return
		}

		d.search.ResizeTT(val * transp.MegaBytes)

	default:
		val, err := strconv.Atoi(args[3])
		if err != nil {
			return
		}
		if err := params.Set(args[1], val); err != nil {
			fmt.Fprintln(d.err, err)
		}
	}
}

func (d *Driver) handlePosition(args []string) {
	if len(args) == 0 {
		return
	}

	switch args[0] {

	case "startpos":
		d.board = board.StartPos()
		if len(args) > 2 && args[1] == "moves" {
			d.applyMoves(args[2:])
		}

	case "fen":
		if len(args) < 7 {
			fmt.Fprintf(d.err, "not enough arguments %d\n", len(args))
			return
		}

		fen := strings.Join(args[1:7], " ")
		b, err := board.FromFEN(fen)
		if err != nil {
			fmt.Fprintf(d.err, "invalid fen %v\n", err)
			return
		}
		if b.InvalidPieceCount() {
			fmt.Fprintln(d.err, "invalid piece counts")
			return
		}
		d.board = b

		if len(args) >= 8 && args[7] == "moves" {
			d.applyMoves(args[8:])
		}
	}
}

func (d *Driver) applyMoves(moves []string) {
	b := d.board
	for _, ms := range moves {
		m, err := parseUCIMove(b, ms)

		if err != nil {
			fmt.Fprintln(d.err, err)
			return
		}

		b.MakeMove(m)
	}
}

func parseUCIMove(b *board.Board, uciM string) (move.Move, error) {
	if len(uciM) != 4 && len(uciM) != 5 {
		return 0, errors.New("invalid uci move")
	}
	from := Square((uciM[0] - 'a') + (uciM[1]-'1')*8)
	to := Square((uciM[2] - 'a') + (uciM[3]-'1')*8)
	if from < A1 || from > H8 || to < A1 || to > H8 {
		return 0, errors.New("invalid uci move")
	}
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
		default:
			return 0, errors.New("invalid uci move")
		}
	}

	m := move.From(from) | move.To(to) | move.Promo(promo)
	if !b.IsPseudoLegal(m) {
		return 0, errors.New("uci move not pseudo-legal")
	}

	return m, nil
}

func (d *Driver) handleEval() {
	fmt.Fprintln(d.output, eval.Eval(d.board, &eval.Coefficients))
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
	PredictedMoves   = 30
	TimeInf          = int64(1 << 50)
)

func (tc timeControl) softLimit(stm Color) int64 {
	if tc.mtime > 0 {
		return tc.mtime
	}

	if stm == White && tc.wtime > 0 {
		return tc.wtime/PredictedMoves + tc.winc/2
	}

	if stm == Black && tc.btime > 0 {
		return tc.btime/PredictedMoves + tc.binc/2
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

func (d *Driver) handleGo(args []string) {
	opts := make([]search.Option, 0, 4)

	tc := timeControl{}

	for i := range args {
		if slices.Contains([]string{"wtime", "btime", "winc", "binc", "depth", "nodes", "movetime"}, args[i]) &&
			len(args) <= i+1 {
			fmt.Fprintln(d.err, "argument missing")
			return
		}

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
			depth := Depth(parseInt(args[i+1]))
			opts = append(opts, search.WithDepth(depth))
		case "nodes":
			nodes := parseInt(args[i+1])
			opts = append(opts, search.WithNodes(nodes))
		case "movetime":
			tc.mtime = parseInt64(args[i+1])
		}
	}

	// stop is always needed in order to support stop command, regardless of timeouts.
	stop := make(chan struct{})
	opts = append(opts, search.WithStop(stop))

	stm := d.board.STM
	if tc.timedMode(stm) {
		opts = append(opts, search.WithSoftTime(tc.softLimit(stm)))
	}

	if d.debug {
		opts = append(opts, search.WithDebug(true))
	}

	var bm move.Move

	searchFin := make(chan struct{})

	go func() {
		_, bm = d.search.Go(d.board, opts...)
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

		case <-d.stop:
			if !stopped {
				stopped = true
				close(stop)
			}
		}
	}

	if !stopped {
		close(stop)
	}

	fmt.Fprintf(d.output, "bestmove %s\n", bm)
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
