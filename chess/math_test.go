package chess_test

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/stretchr/testify/assert"
)

func TestAbs(t *testing.T) {
	tests := []struct {
		num  int
		want int
	}{
		{-5, 5},
		{0, 0},
		{999, 999},
	}
	for _, tt := range tests {
		name := strconv.Itoa(tt.num)
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, Abs(tt.num))
		})
	}
}

func TestSignum(t *testing.T) {
	tests := []struct {
		num  int
		want int
	}{
		{-5, -1},
		{0, 0},
		{999, 1},
	}
	for _, tt := range tests {
		name := strconv.Itoa(tt.num)
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, Signum(tt.num))
		})
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		num  int
		lo   int
		hi   int
		want int
	}{
		{0, -5, 3, 0},
		{-6, -5, 3, -5},
		{8, -5, 3, 3},
		{-5, -5, 3, -5},
		{3, -5, 3, 3},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("%d %d %d", tt.num, tt.lo, tt.hi)
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, Clamp(tt.num, tt.lo, tt.hi))
		})
	}
}

func TestMust(t *testing.T) {
	tests := []struct {
		name string
		v    int
		err  error
	}{
		{"no error", 5, nil},
		{"error", 3, errors.New("oops")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err != nil {
				assert.PanicsWithValue(t, tt.err, func() {
					Must(tt.v, tt.err)
				})
			} else {
				assert.Equal(t, tt.v, Must(tt.v, tt.err))
			}
		})
	}
}
