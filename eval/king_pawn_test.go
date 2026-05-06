package eval

import (
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/stretchr/testify/assert"
)

func TestShelter(t *testing.T) {
	tests := [...]struct {
		name        string
		fen         string
		color       Color
		wantShelter int
		wantKind    shelterKind
	}{
		{"white normal king side", "4k3/8/4P3/3P4/4K3/8/8/8 w - - 0 1", White, 0b_010_001, normalShelter},
		{"white normal queen side", "4k3/8/3P4/2P5/3K4/8/8/8 w - - 0 1", White, 0b_010_100, normalShelter},
		{"white opponents home rank", "1K2k3/8/8/8/8/8/8/8 w - - 0 1", White, 0, invalidShelter},
		{"white A file", "4k3/8/8/P7/1P6/K7/8/8 w - - 0 1", White, 0b_10_01, smallShelter},
		{"white H file", "4k3/8/8/8/8/7P/6P1/7K w - - 0 1", White, 0b_10_01, smallShelter},

		{"black normal king side", "8/4k3/3p4/4p3/8/8/8/4K3 w - - 0 1", Black, 0b_010_001, normalShelter},
		{"black normal queen side", "8/3k4/3p4/4p3/8/8/8/4K3 w - - 0 1", Black, 0b_001_010, normalShelter},
		{"black in opponents area", "8/8/8/8/8/8/1pk5/4K3 w - - 0 1", Black, 0, invalidShelter},
		{"black A file", "8/8/8/k7/p7/1p6/8/4K3 w - - 0 1", Black, 0b_01_10, smallShelter},
		{"black H file", "8/8/7k/6p1/7p/8/8/4K3 w - - 0 1", Black, 0b_10_01, smallShelter},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Must(board.FromFEN(tt.fen))
			gotShelter, gotKind := shelter(b, tt.color)
			assert.Equal(t, tt.wantKind, gotKind, "fen %s", tt.fen)
			assert.Equal(t, tt.wantShelter, gotShelter, "fen %s", tt.fen)
		})
	}
}
