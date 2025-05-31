package eval

import (
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/stretchr/testify/assert"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

func TestEGScale(t *testing.T) {
	tests := []struct {
		name string
		b    *board.Board
		want Score
	}{
		{
			name: "not endgame",
			b:    Must(board.FromFEN("2n5/4k3/p4b1p/1p4p1/2p3P1/2P5/1B1P4/4K1R1 w - - 0 1")),
			want: 128,
		},
		{
			name: "opposite color bishop endgame",
			b:    Must(board.FromFEN("8/4k3/p4b1p/1p4p1/2p3P1/2P5/2BP4/4K3 w - - 0 1")),
			want: Coefficients.OppositeBishops,
		},
		{
			name: "matching color bishop endgame",
			b:    Must(board.FromFEN("8/4k3/p4b1p/1p4p1/2p3P1/2P5/1B1P4/4K3 w - - 0 1")),
			want: 128,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := egScale(tt.b, &Coefficients)

			assert.Equal(t, tt.want, actual)
		})
	}
}
