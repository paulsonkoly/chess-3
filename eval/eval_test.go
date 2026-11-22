package eval

import (
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/stretchr/testify/assert"
)

func TestEvalCacheTransparency(t *testing.T) {
	tests := [...]struct {
		name    string
		base    string
		variant string
	}{
		// Same pawns, different kings/knight -> same pawn hash.
		{"pawn", "8/8/4k1p1/5p2/P7/1P1P4/2K5/8 w - - 0 1", "8/8/2k3p1/5p2/PK6/1P1P4/6N1/8 w - - 0 1"},
		// Same pawns and kings, different non-pawn piece -> same pawn+king hash.
		{"pawn_king", "8/8/4k1p1/5p2/P7/1P1P4/2K5/8 w - - 0 1", "8/8/4k1p1/5p2/P7/1P1P4/2K3N1/8 w - - 0 1"},
		// Same material counts, different placement -> same material key.
		{"material", "8/8/4k1p1/5p2/P7/1P1P4/2K5/8 w - - 0 1", "8/8/3k3p/6p1/1P6/P2P4/7K/8 w - - 0 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Must(board.FromFEN(tt.base))
			other := Must(board.FromFEN(tt.variant))
			e := New[Score]()

			baseline := e.Score(b, &Coefficients)

			_ = e.Score(other, &Coefficients)

			assert.Equal(t, baseline, e.Score(b, &Coefficients), "warmed cache changed the score")

			e.Clear()

			assert.Equal(t, baseline, e.Score(b, &Coefficients), "clear changed the score")
		})
	}
}
