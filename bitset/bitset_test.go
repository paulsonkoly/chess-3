package bitset_test

import (
	"testing"

	"github.com/paulsonkoly/chess-3/bitset"
	"github.com/stretchr/testify/assert"
)

func TestBitMap(t *testing.T) {
	s := bitset.BitSet{}
	o := bitset.BitSet{}

	s.Set(13)
	s.Set(40)
	s.Set(113)

	ix := s.Next(o)
	assert.Equal(t, 13, ix)
	o.Set(ix)

	ix = s.Next(o)
	assert.Equal(t, 40, ix)
	o.Set(ix)

	ix = s.Next(o)
	assert.Equal(t, 113, ix)
	o.Set(ix)

	assert.Equal(t, -1, s.Next(o))
}

func TestBitMapEmpty(t *testing.T) {
	assert.Equal(t, -1, bitset.BitSet{}.Next(bitset.BitSet{}))
}

func TestBitMapLargest(t *testing.T) {
	s := bitset.BitSet{}
	s.Set(255)
	assert.Equal(t, 255, s.Next(bitset.BitSet{}))
}

func TestBitMapSetOutOfBounds(t *testing.T) {
	s := bitset.BitSet{}
	assert.Panics(t, func() { s.Set(-1) })
	assert.Panics(t, func() { s.Set(256) })
}
