package transp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatch64(t *testing.T) {
	tests := []struct {
		name   string
		word   uint64
		key    partialKey
		wantIx int
		wantOk bool
	}{
		{"not found", 0xabcd432112345678, 0xf001, 0, false},
		{"ix1", 0xabcd432112345678, 0x5678, 0, true},
		{"ix2", 0xabcd432112345678, 0x1234, 1, true},
		{"ix3", 0xabcd432112345678, 0x4321, 2, true},
		{"ix4", 0xabcd432112345678, 0xabcd, 3, true},
		{"first index", 0x1234432112345678, 0x1234, 1, true},
		{"0", 0x0, 0x1234, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIx, gotOk := match64(tt.word, tt.key)
			assert.Equal(t, tt.wantIx, gotIx)
			assert.Equal(t, tt.wantOk, gotOk)
		})
	}
}
