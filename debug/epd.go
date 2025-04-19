package debug

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/paulsonkoly/chess-3/board"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

type EPDReader struct {
	inp    *bufio.Scanner
	fen    string
	depths []string
}

type EPDEntry struct {
	D     Depth
	Cnt   int
	Fen   string
	Board *board.Board
}

func NewEPDReader(fn string) (*EPDReader, error) {
	inp, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	return &EPDReader{inp: bufio.NewScanner(inp), depths: make([]string, 0)}, nil
}

func (e *EPDReader) Scan() bool {
	if len(e.depths) > 1 {
		e.depths = e.depths[1:]
		return true
	}

	for e.inp.Scan() {
		line := e.inp.Text()
		parts := strings.Split(line, " ;")
		if len(parts) < 2 {
			continue
		}
		e.fen = parts[0]
		e.depths = parts[1:]
		return true
	}

	return false
}

func (e *EPDReader) Entry() EPDEntry {
	board, err := board.FromFEN(e.fen)
	if err != nil {
		panic(err)
	}

	dstr := e.depths[0]
	if dstr[0] != 'D' {
		panic("D expected in depth ")
	}

	parts := strings.Split(dstr[1:], " ")
	if len(parts) != 2 {
		panic("malformed perft depth info")
	}
	d, err := strconv.Atoi(parts[0])
	if err != nil {
		panic(err)
	}
	exp, err := strconv.Atoi(parts[1])
	if err != nil {
		panic(err)
	}
	return EPDEntry{Board: board, Fen: e.fen, D: Depth(d), Cnt: exp}
}
