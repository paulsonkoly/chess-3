package eval_test

import (
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/eval"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/stretchr/testify/assert"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

func TestEval(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		b    *board.Board
		want Score
	}{
		{name: "checkmate",
			b:    board.FromFEN("8/1b3P2/2p5/1p6/8/P4k2/6q1/6K1 w - - 2 54"),
			want: -Inf,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := eval.Eval(tt.b, []move.Move{})
			assert.Equal(t, tt.want, got)
		})
	}
}
