package chess_test

import (
	"fmt"
	"testing"

	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/stretchr/testify/assert"
)

func TestLowestSet(t *testing.T) {
	tests := []struct {
		bb   BitBoard
		want Square
	}{
		{0x0000000000000000, 64},
		{0x0000000000000001, 0},
		{0x0000000000000002, 1},
		{0x0000000000000003, 0},
		{0x8000000000000000, 63},
		{0x4000000000000000, 62},
		{0xc000000000000000, 62},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("0x%016x", tt.bb)
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.bb.LowestSet())
		})
	}
}

func TestBitBoardFromSquares(t *testing.T) {
	tests := []struct {
		sqrs []Square
		want BitBoard
	}{
		{[]Square{}, 0x0000000000000000},
		{[]Square{A1}, 0x0000000000000001},
		{[]Square{H1}, 0x0000000000000080},
		{[]Square{A8}, 0x0100000000000000},
		{[]Square{H8}, 0x8000000000000000},
		{[]Square{B2, D5, F2}, 0x800002200},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("0x%016x", tt.want)
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, BitBoardFromSquares(tt.sqrs...))
		})
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		bb   BitBoard
		want int
	}{
		{0x0000000000000000, 0},
		{0x0000000000000001, 1},
		{0x0000000000000080, 1},
		{0x0100000000000000, 1},
		{0x8000000000000000, 1},
		{0x800002200, 3},
		{0xffffffffffffffff, 64},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("0x%016x", tt.bb)
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.bb.Count())
		})
	}
}

func TestIsPow2(t *testing.T) {
	tests := []struct {
		bb   BitBoard
		want bool
	}{
		{0x0000000000000000, false},
		{0x0000000000000001, true},
		{0x0000000000000080, true},
		{0x0100000000000000, true},
		{0x8000000000000000, true},
		{0x800002200, false},
		{0xffffffffffffffff, false},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("0x%016x", tt.bb)
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.bb.IsPow2())
		})
	}
}
