package eval

import (
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/stretchr/testify/assert"

	. "github.com/paulsonkoly/chess-3/chess"
)

func TestCalcPawnStructure(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		want pieceWise
	}{
		{
			name: "Pawn endgame",
			fen:  "8/2pp1k1p/5p2/5p2/8/P7/3P2P1/5K2 w - - 0 1",
			want: pieceWise{
				holes:         [2]BitBoard{0x8494bffff, 0xffffa1a100000000},
				passers:       [2]BitBoard{BitBoardFromSquares(A3), 0},
				doubledPawns:  [2]BitBoard{0, BitBoardFromSquares(F6)},
				isolatedPawns: [2]BitBoard{BitBoardFromSquares(A3, D2, G2), BitBoardFromSquares(F6, F5, H7)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := board.FromFEN(tt.fen)
			if err != nil {
				t.Fatal(err)
			}

			actual := pieceWise{}
			actual.calcPawnStructure(b)

			assert.Equal(t, tt.want.holes, actual.holes, "holes")
			assert.Equal(t, tt.want.passers, actual.passers, "passers")
			assert.Equal(t, tt.want.doubledPawns, actual.doubledPawns, "doubledPawns")
			assert.Equal(t, tt.want.isolatedPawns, actual.isolatedPawns, "isolatedPawns")
		})
	}
}
