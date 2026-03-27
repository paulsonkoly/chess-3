package picker

import (
	"testing"

	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/stretchr/testify/assert"
)

func TestPartialSort(t *testing.T) {
	tests := []struct {
		name      string
		moves     []move.Weighted
		threshold Score
		want      int
	}{
		{
			name:      "empty input",
			moves:     []move.Weighted{},
			threshold: 50,
			want:      0,
		},
		{
			name:      "no element less than or equal to threshold",
			moves:     []move.Weighted{{Weight: 100}, {Weight: 200}, {Weight: 300}},
			threshold: 50,
			want:      3,
		},
		{
			name:      "all elements less than or equal to threshold",
			moves:     []move.Weighted{{Weight: 10}, {Weight: 20}, {Weight: 30}},
			threshold: 50,
			want:      0,
		},
		{
			name:      "single element greater than threshold",
			moves:     []move.Weighted{{Weight: 100}},
			threshold: 50,
			want:      1,
		},
		{
			name:      "single element less than or equal to threshold",
			moves:     []move.Weighted{{Weight: 50}},
			threshold: 50,
			want:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx := partialSort(tt.moves, tt.threshold)

			assert.Equal(t, tt.want, idx)

			for _, m := range tt.moves[:idx] {
				assert.Greater(t, m.Weight, tt.threshold)
			}
			for _, m := range tt.moves[idx:] {
				assert.LessOrEqual(t, m.Weight, tt.threshold)
			}
		})
	}
}

func FuzzPartialSort(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte, thr int16) {
		if len(data) == 0 {
			return
		}

		moves := make([]move.Weighted, len(data))
		for i, d := range data {
			x := uint32(d) * 0x45d9f3b
			x ^= (x >> 16)
			w := Score(x) - (1 << 14)
			moves[i] = move.Weighted{Weight: w}
		}

		threshold := Score(thr)

		idx := partialSort(moves, threshold)

		for _, m := range moves[:idx] {
			assert.Greater(t, m.Weight, thr)
		}

		for _, m := range moves[idx:] {
			assert.LessOrEqual(t, m.Weight, thr)
		}
	})
}
