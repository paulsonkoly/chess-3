package chess_test

import (
	"fmt"
	"testing"

	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/stretchr/testify/assert"
)

func TestSquare(t *testing.T) {
	squares := [...]Square{A1, C2, B7, A8, G1, H1, H8}

	for _, sq := range squares {
		assert.Equal(t, sq, SquareAt(sq.File(), sq.Rank()), "%s", sq)
	}
}

func TestSquareString(t *testing.T) {
	tests := [...]struct {
		sq   Square
		want string
	}{
		{E3, "e3"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.sq.String())
		})
	}
}

func TestFromPerspectiveOf(t *testing.T) {
	tests := [...]struct {
		rank Coord
		side Color
		want Coord
	}{
		{SecondRank, White, SecondRank},
		{SecondRank, Black, SeventhRank},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("rank %d side %d", tt.rank+1, tt.side), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.rank.FromPerspectiveOf(tt.side))
		})
	}
}
