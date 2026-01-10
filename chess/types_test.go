package chess_test

import (
	"strconv"
	"testing"

	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/stretchr/testify/assert"
)

func TestScoreString(t *testing.T) {
	tests := [...]struct {
		name  string
		score Score
		want  string
	}{
		{"0 score", 0, "cp 0"},
		{"positive score", 73, "cp 73"},
		{"negative score", -50, "cp -50"},
		// 2 impossible edge cases:
		//  - mate 0 means we were asked to run on a (legal) mate position
		//  - mate -0 means the opponent is mated while it's our turn to move (illegal)
		{"score for mating opponent", Inf, "mate 0"},
		{"score for being mated ", -Inf, "mate -0"},
		{"score for mating opponent in 1 move", Inf - 1, "mate 1"},
		{"score for being mated in 1 move", -Inf + 2, "mate -1"},
		{"score for mating opponent in 2 moves", Inf - 3, "mate 2"},
		{"score for being mated in 2 moves", -Inf + 4, "mate -2"},
		{"score under mate boundary", Inf - MaxPlies - 1, "cp " + strconv.Itoa(int(Inf-MaxPlies-1))},
		{"score in mate boundary", Inf - MaxPlies, "mate " + strconv.Itoa(MaxPlies/2)},
		{"score for being mated in 2 moves", -Inf + 4, "mate -2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.score.String())
		})
	}
}
