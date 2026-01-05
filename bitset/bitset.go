// Package bitset provides a 256 capacity index map with Kernighan iteration.
//
// # example
//
//	// Declare a BitSet with null value.
//	s := BitSet{}
//
//	// Set some bits
//	s.Set(13)
//	s.Set(40)
//	s.Set(113)
//
//	// Loop over set bits.
//	for ix := s.Next(t); ix != -1; ix, ix = s.Next(t) {
//		fmt.Println(ix) // prints 13, 40, 113
//		// Remove ix otherwise Next yields it again.
//		t.Clear(ix)
//	}
package bitset

import (
	"math/bits"
	"strconv"
	"strings"
)

// BitSet can track numbers from 0 to 255.
type BitSet [4]uint64

// Set adds ix to b.
//
// Index is expect to be between 0 and 255.
func (b *BitSet) Set(ix int) {
	b[ix>>6] |= uint64(1) << (ix & 63)
}

// Clear clears ix in b.
//
// Index is expect to be between 0 and 255.
func (b *BitSet) Clear(ix int) {
	b[ix>>6] &= ^(uint64(1) << (ix & 63))
}

// AndNot performs a self modifying and with not o.
func (b *BitSet) AndNot(o *BitSet) {
	for i := range 4 {
		b[i] &= ^o[i]
	}
}

// Next returns the lowest significant set bit index in b.
func (b *BitSet) Next() int {
	for i := range 4 {
		if x := b[i]; x != 0 {
			return bits.TrailingZeros64(x) + i*64
		}
	}
	return -1
}

// String is a debug string represention of b.
func (b BitSet) String() string {
	vals := make([]string, 0)
	for i := b.Next(); i != -1; i = b.Next() {
		b.Clear(i)
		vals = append(vals, strconv.Itoa(i))
	}
	return "[" + strings.Join(vals, ", ") + "]"
}
