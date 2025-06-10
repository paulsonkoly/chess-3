package transp_test

import (
	"testing"

	"github.com/paulsonkoly/chess-3/transp"
	"github.com/stretchr/testify/assert"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

func TestUsableFor(t *testing.T) {
	var tests = []struct {
		name   string
		entryD Depth
		entryA transp.Age
		d      Depth
		age    transp.Age
		want   bool
	}{
		{
			name:   "same age / accept",
			entryD: 5,
			entryA: 13,
			d:      4,
			age:    13,
			want:   true,
		},
		{
			name:   "same age / too low depth",
			entryD: 5,
			entryA: 13,
			d:      6,
			age:    13,
			want:   false,
		},
		{
			name:   "older by 1 search / same depth",
			entryD: 5,
			entryA: 13,
			d:      5,
			age:    14,
			want:   false,
		},
		{
			name:   "older by 1 search / +1 in depth",
			entryD: 6,
			entryA: 13,
			d:      5,
			age:    14,
			want:   false,
		},
		{
			name:   "older by 1 search / +2 in depth",
			entryD: 7,
			entryA: 13,
			d:      5,
			age:    14,
			want:   true,
		},
		{
			name:   "overflow",
			entryD: 7,
			entryA: 255,
			d:      5,
			age:    0,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := transp.Entry{Depth: tt.entryD, Age: tt.entryA}

			assert.Equal(t, tt.want, entry.UsableFor(tt.d, tt.age))
		})
	}
}
