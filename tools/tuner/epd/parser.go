package epd

import (
	"bytes"
	"errors"

	"github.com/paulsonkoly/chess-3/board"
)

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
