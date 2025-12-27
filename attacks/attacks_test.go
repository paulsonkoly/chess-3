package attacks_test

import (
	"testing"

	"github.com/paulsonkoly/chess-3/attacks"
	"github.com/stretchr/testify/assert"

	. "github.com/paulsonkoly/chess-3/chess"
)

func TestPawnSinglePushMoves(t *testing.T) {
	tests := []struct {
		name  string
		from  BitBoard
		color Color
		to    BitBoard
	}{
		{
			name:  "pawn push white",
			from:  BitBoardFromSquares(B1, E5, D7, C8, G8),
			color: White,
			to:    BitBoardFromSquares(B2, E6, D8),
		},
		{
			name:  "pawn push black",
			from:  BitBoardFromSquares(B1, E5, D7, C8, G8),
			color: Black,
			to:    BitBoardFromSquares(E4, D6, C7, G7),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.to, attacks.PawnSinglePushMoves(tt.from, tt.color))
		})
	}

}

func TestPawnCaptureMoves(t *testing.T) {
	tests := []struct {
		name  string
		from  BitBoard
		color Color
		to    BitBoard
	}{
		{
			name:  "pawn capture white",
			from:  BitBoardFromSquares(B1, E5, D7, G8),
			color: White,
			to:    BitBoardFromSquares(A2, C2, D6, F6, C8, E8),
		},
		{
			name:  "pawn capture black",
			from:  BitBoardFromSquares(B1, E5, D7, G8),
			color: Black,
			to:    BitBoardFromSquares(D4, F4, C6, E6, F7, H7),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.to, attacks.PawnCaptureMoves(tt.from, tt.color))
		})
	}
}
