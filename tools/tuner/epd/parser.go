package epd

import (
	"errors"
	"strconv"
	"strings"

	"github.com/paulsonkoly/chess-3/board"
)

type Entry struct {
	Board  *board.Board
	Result float64
}

var ErrLineInvalid = errors.New("invalid epd line")

func Parse(line string) (Entry, error) {
	splits := strings.Split(line, "; ")
	if len(splits) != 2 {
		return Entry{}, ErrLineInvalid
	}

	b, err := board.FromFEN(splits[0])
	if err != nil {
		return Entry{}, ErrLineInvalid
	}

	r, err := strconv.ParseFloat(splits[1], 64)
	if err != nil {
		return Entry{}, ErrLineInvalid
	}

	return Entry{Board: b, Result: r}, nil
}
