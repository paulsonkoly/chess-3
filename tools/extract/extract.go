package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/movegen"
	"github.com/paulsonkoly/chess-3/search"
	"github.com/paulsonkoly/chess-3/types"
)

var (
	inputF           *os.File
	output           string
	outputF          *os.File
	inCheck          bool
	bestCapture      bool
	bestCaptureDepth int
)

func main() {
	var err error
	var input string
	var cpuProf string

	flag.StringVar(&input, "input", "-", "input epd file. `-'for stdin.")
	flag.StringVar(&output, "output", "-", "output epd file. `-' for stdout.")
	flag.BoolVar(&inCheck, "incheck", true, "filter in check positions.")
	flag.BoolVar(&bestCapture, "cap", true, "filter positions where best move is capture")
	flag.IntVar(&bestCaptureDepth, "capd", 10, "depth to determine if best move is capture")
	flag.StringVar(&cpuProf, "cpuprof", "", "cpu pprof file name. Empty for disable.")

	flag.Parse()

	if cpuProf != "" {
		f, err := os.Create(cpuProf)
		if err != nil {
			log.Fatalf("pprof file creation error %v\n", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatalf("pprof start error %v\n", err)
		}
		defer pprof.StopCPUProfile()
	}

	if input == "-" {
		inputF = os.Stdin
	} else {
		if inputF, err = os.Open(input); err != nil {
			log.Fatalf("%s: %v\n", input, err)
		}
		defer inputF.Close()
	}

	if output == "-" {
		outputF = os.Stdout
	} else {
		if outputF, err = os.Create(output); err != nil {
			log.Fatalf("%s: %v\n", output, err)
		}
		defer outputF.Close()
	}

	filters := make([]Filter, 0)

	if inCheck {
		filters = append(filters, InCheck{})
	}

	if bestCapture {
		filters = append(filters, BestCapture{search: search.New(1), depth: types.Depth(bestCaptureDepth)})
	}

	run(filters)
}

func run(filters []Filter) {
	// TODO use the epd functionality from the tuner.
	b := board.Board{}
	nl := []byte{'\n'}
	var wdl float64

	scn := bufio.NewScanner(inputF)
	for scn.Scan() {
		buf := scn.Bytes()

		if err := Parse(buf, &b, &wdl); err != nil {
			fmt.Fprintf(os.Stderr, "parse failed %s %v\n", string(buf), err)
			continue
		}

		// filter
		for _, filter := range filters {
			if filter.Filter(&b, wdl) {
				continue
			}
		}
		// sample

		if _, err := outputF.Write(buf); err != nil {
			log.Fatalf("%s: %v", output, err)
		}

		if _, err := outputF.Write(nl); err != nil {
			log.Fatalf("%s: %v", output, err)
		}
	}
}

// TODO this shouldn't be duplicated here
// ErrLineInvalid indicates parse error of the epd line.
var ErrLineInvalid = errors.New("invalid epd line")

// Parse helper function provides an allocation free epd line parser.
func Parse(line []byte, b *board.Board, res *float64) error {
	if len(line) < 5 {
		return ErrLineInvalid
	}
	splitIx := len(line) - 5 // index of ';'

	if err := board.ParseFEN(b, line[:splitIx]); err != nil {
		return ErrLineInvalid
	}

	switch {

	case bytes.Equal(line[splitIx:], []byte("; 1.0")):
		*res = 1.0

	case bytes.Equal(line[splitIx:], []byte("; 0.5")):
		*res = 0.5

	case bytes.Equal(line[splitIx:], []byte("; 0.0")):
		*res = 0.0

	default:
		return ErrLineInvalid

	}

	return nil
}

type Filter interface {
	Filter(b *board.Board, wdl float64) bool
}

type InCheck struct{}

func (i InCheck) Filter(b *board.Board, _ float64) bool {
	return movegen.InCheck(b, b.STM)
}

type BestCapture struct {
	search *search.Search
	depth  types.Depth
}

func (bc BestCapture) Filter(b *board.Board, _ float64) bool {
	b.ResetHash()
	_, move := bc.search.WithOptions(b, bc.depth, search.WithInfo(false))
	return b.SquaresToPiece[move.To()] != types.NoPiece
}
