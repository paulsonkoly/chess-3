package bitset_test

import (
	"testing"

	"github.com/paulsonkoly/chess-3/bitset"
	"github.com/stretchr/testify/assert"
)

func TestBitMap(t *testing.T) {
	s := bitset.BitSet{}

	s.Set(13)
	s.Set(40)
	s.Set(113)

	ix := s.Next()
	assert.Equal(t, 13, ix)
	s.Clear(ix)

	ix = s.Next()
	assert.Equal(t, 40, ix)
	s.Clear(ix)

	ix = s.Next()
	assert.Equal(t, 113, ix)
	s.Clear(ix)

	assert.Equal(t, -1, s.Next())
}

func TestBitMapEmpty(t *testing.T) {
	b := bitset.BitSet{}
	assert.Equal(t, -1, b.Next())
}

func TestBitMapLargest(t *testing.T) {
	s := bitset.BitSet{}
	s.Set(255)
	assert.Equal(t, 255, s.Next())
}

func TestBitMapSetOutOfBounds(t *testing.T) {
	s := bitset.BitSet{}
	assert.Panics(t, func() { s.Set(-1) })
	assert.Panics(t, func() { s.Set(256) })
}

func TestBitMapString(t *testing.T) {
	s := bitset.BitSet{}

	s.Set(13)
	s.Set(40)
	s.Set(113)

	assert.Equal(t, "[13, 40, 113]", s.String())

}
