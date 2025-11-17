// package transp is the transposition table cache.
package transp

import (
	"fmt"
	"unsafe"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"

	. "github.com/paulsonkoly/chess-3/types"
)

const (
	// MegaBytes is the count of bytes in a MegaByte. It is useful for code like hash.New(16 * MegaBytes).
	MegaBytes = 1024 * 1024
)

const (
	// entrySize is the transposition table entry size in bytes.
	entrySize = 8
	// bucketSize is the number of entries per bucket. A bucket should match a most common CPU cache line.
	bucketSize = 8
	// alignment is the byte alignment of buckets.
	alignment = bucketSize * entrySize
	// partialKeyBits is the number of bits of the Zobrist-hash stored per entry.
	partialKeyBits = 16
)

type Table struct {
	raw    []byte // reference to unaligned underlying data to keep it from GC
	data   []Entry
	ixMask board.Hash
}

type Type byte

const (
	// Exact is an exact score.
	Exact Type = iota
	// UpperBound is an upper bound for the score.
	UpperBound
	// LowerBound is a lower bound for the score.
	LowerBound
)

type partialKey = uint16

// Entry is a transposition table entry.
//
//go:packed
type Entry struct {
	move.SimpleMove            // SimpleMove is the simplified move data. (2 bytes)
	value           Score      // value is the score for the entry where one is present. (2 bytes)
	Depth           Depth      // Depth of the entry. (1 byte)
	Type            Type       // Type is the entry type. (1 byte)
	pKey            partialKey // pKey is the high bits of the Zobrist-hash. (2 bytes)
}

func init() {
	packSize := unsafe.Sizeof(Entry{})
	if packSize != entrySize {
		panic(fmt.Sprintf("tt entry isn't packed to %d bytes it is %d bytes, check alignments", entrySize, packSize))
	}
}

// Value is the score of the entry corrected for current ply in case of mate score.
func (e Entry) Value(ply Depth) Score {
	if e.value > Inf-MaxPlies {
		return e.value - Score(ply)
	}

	if e.value < -Inf+MaxPlies {
		return e.value + Score(ply)
	}

	return e.value
}

// New creates a new transposition table. size is the table size in bytes, and
// only power of 2 sizes are supported.
func New(size int) *Table {
	if size < 1 || size&(size-1) != 0 {
		panic(fmt.Sprintf("invalid transposition table size %d", size))
	}

	numBuckets := size / alignment
	numEntries := numBuckets * bucketSize

	raw := make([]byte, size+alignment-1)
	base := uintptr(unsafe.Pointer(&raw[0]))
	aligned := (base + alignment - 1) &^ (alignment - 1)

	entries := (*Entry)(unsafe.Pointer(aligned))
	data := unsafe.Slice(entries, numEntries)

	return &Table{raw: raw, data: data, ixMask: board.Hash(numBuckets - 1)}
}

// HashFull is the permill count of the hash usage.
func (t Table) HashFull() int {
	return 1000
}

// Clear clears the transposition table for the next search.Search().
func (t *Table) Clear() {
	for ix := range t.data {
		t.data[ix].Depth = 0
	}
}

// Insert inserts an entry to the transposition table.
func (t *Table) Insert(hash board.Hash, d, ply Depth, sm move.SimpleMove, value Score, typ Type) {
	ix := (hash & t.ixMask)*bucketSize


	minD := Depth(MaxPlies + 1)
	var entry *Entry
	for eix := ix; eix < ix+bucketSize; eix++ {
		if t.data[eix].Depth < minD {
			minD = t.data[eix].Depth
			entry = &t.data[eix]
		}
	}

	if value < -Inf+MaxPlies {
		value -= Score(ply)
	}

	if value > Inf-MaxPlies {
		value += Score(ply)
	}

	*entry = Entry{
		SimpleMove: sm,
		pKey:       partialKey(hash >> (64 - partialKeyBits)),
		value:      value,
		Type:       typ,
		Depth:      d,
	}
}

// Lookup looks up the transposition table entry, using hash as the key.
func (t *Table) LookUp(hash board.Hash) (*Entry, bool) {
	ix := (hash & t.ixMask) * bucketSize

	for eix := ix; eix < ix+bucketSize; eix++ {
		if t.data[eix].pKey == partialKey(hash>>(64-partialKeyBits)) {
			return &t.data[eix], true
		}
	}

	return nil, false
}
