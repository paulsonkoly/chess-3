package eval

import (
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/stretchr/testify/assert"
)

func TestInsufficient(t *testing.T) {
	tests := [...]struct {
		name string
		fen  string
		want bool
	}{
		{
			"empty board",
			"4k3/8/8/8/8/8/8/4K3 w - - 0 1",
			true,
		},
		{
			"full board",
			"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			false,
		},
		{
			"bishop pair",
			"1k6/8/8/8/8/8/8/1K2BB2 w - - 0 1",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Must(board.FromFEN(tt.fen))
			assert.Equal(t, tt.want, insufficient(b))
		})
	}
}

func TestKBNvK(t *testing.T) {
	tests := [...]struct {
		name string
		fen  string
		want bool
	}{
		{
			"empty board",
			"4k3/8/8/8/8/8/8/4K3 w - - 0 1",
			false,
		},
		{
			"full board",
			"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			false,
		},
		{
			"knight/bishop",
			"1k6/8/8/8/8/8/8/1K2BN2 w - - 0 1",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Must(board.FromFEN(tt.fen))
			assert.Equal(t, tt.want, knbvk(b))
		})
	}
}
