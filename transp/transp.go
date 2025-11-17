// package transp is a cache-optimized transposition table implementation for Go.
// This version packs the partial key as the first field, aligns buckets to 64
// byte cache-lines, avoids subslices in hot loops and minimizes bounds/index
// arithmetic in the lookup/insert hot paths.

package transp

import (
	"fmt"
	"unsafe"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"

	. "github.com/paulsonkoly/chess-3/types"
)

const (
	// MegaBytes is the count of bytes in a MegaByte.
	MegaBytes = 1024 * 1024
)

const (
	// cacheLine is the assumed CPU cache line size we target.
	cacheLine = 64
	// partialKeyBits is the number of bits of the Zobrist-hash stored per entry.
	partialKeyBits = 16
)

// Table is the transposition table.
type Table struct {
	raw        []byte  // reference to unaligned underlying data to keep it from GC
	data       []Entry // entries in flat layout: numBuckets * bucketSize
	ixMask     board.Hash
	bucketSize int // derived from cacheLine / entrySize
}

// Type represents the stored bound type.
type Type byte

const (
	Exact Type = iota
	UpperBound
	LowerBound
)

type partialKey = uint16

// Entry layout: put pKey first so the common miss path loads the partial key at
// offset 0 (fast aligned 16-bit loads). Remaining fields follow. The order is
// chosen so the natural struct size is 8 bytes on common architectures.
//
// Field sizes assumed by this file:
//
//	pKey (2) | SimpleMove (2) | value (2) | Depth (1) | Type (1)  => 8 bytes
//
// Keep this layout — changes will trigger the init size check.
type Entry struct {
	pKey partialKey
	move.SimpleMove
	value Score
	Depth Depth
	Type  Type
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

func init() {
	// Safety check: ensure the entry is exactly 8 bytes.
	// If this panics on your target, inspect the field types' sizes and
	// re-order or pack manually. We prefer natural alignment (no //go:packed)
	// because unaligned accesses can be slower on some architectures.
	packSize := unsafe.Sizeof(Entry{})
	if packSize != 8 {
		panic(fmt.Sprintf("tt entry must be 8 bytes but is %d bytes; adjust layout", packSize))
	}
}

// New creates a new transposition table. size is the table size in bytes, and
// only power of 2 sizes are supported. The size should be a multiple of
// cacheLine (64) for best performance.
func New(size int) *Table {
	if size < 1 || size&(size-1) != 0 {
		panic(fmt.Sprintf("invalid transposition table size %d", size))
	}

	entrySize := int(unsafe.Sizeof(Entry{}))
	if cacheLine%entrySize != 0 {
		panic(fmt.Sprintf("cacheLine (%d) must be multiple of entry size (%d)", cacheLine, entrySize))
	}

	bucketSize := cacheLine / entrySize
	numBuckets := size / cacheLine
	if numBuckets == 0 {
		panic("table size too small; must be at least one cacheLine")
	}
	numEntries := numBuckets * bucketSize

	// Overallocate by cacheLine-1 and align the returned pointer to cacheLine
	raw := make([]byte, size+cacheLine-1)
	base := uintptr(unsafe.Pointer(&raw[0]))
	aligned := (base + uintptr(cacheLine-1)) &^ uintptr(cacheLine-1)

	ptr := unsafe.Pointer(aligned)
	entries := unsafe.Slice((*Entry)(ptr), numEntries)

	return &Table{
		raw:        raw,
		data:       entries,
		ixMask:     board.Hash(numBuckets - 1),
		bucketSize: bucketSize,
	}
}

// HashFull returns a dummy usage percentage (implement as you need).
func (t Table) HashFull() int { return 1000 }

// Clear zeros the depth field for all entries; cheap and avoids memzeroing entire table.
func (t *Table) Clear() {
	for i := range t.data {
		t.data[i].Depth = 0
	}
}

// bucketStart returns the starting entry index for the bucket of hash.
func (t *Table) bucketStart(hash board.Hash) int {
	return int((hash & t.ixMask)) * t.bucketSize
}

// LookUp looks up the entry for hash. This function is micro-optimized for the
// hot path: local copy of slice header, no subslices, single loop with simple
// index arithmetic and pKey first access.
func (t *Table) LookUp(hash board.Hash) (*Entry, bool) {
	data := t.data
	start := t.bucketStart(hash)
	end := start + t.bucketSize
	key := partialKey(hash >> (64 - partialKeyBits))

	// Warm the cache line by doing a single read; this is a lightweight hint.
	_ = data[start]

	for i := start; i < end; i++ {
		if data[i].pKey == key {
			return &data[i], true
		}
	}
	return nil, false
}

// Insert writes an entry into the bucket, using a simple "replace lowest depth"
// policy. It's written to avoid extra pointer arithmetic inside the loop.
func (t *Table) Insert(hash board.Hash, d, ply Depth, sm move.SimpleMove, value Score, typ Type) {
	data := t.data
	start := t.bucketStart(hash)
	end := start + t.bucketSize
	pkey := partialKey(hash >> (64 - partialKeyBits))

	minD := Depth(MaxPlies + 1)
	minIdx := start
	for i := start; i < end; i++ {
		if data[i].Depth < minD {
			minD = data[i].Depth
			minIdx = i
		}
	}

	if value < -Inf+MaxPlies {
		value -= Score(ply)
	}
	if value > Inf-MaxPlies {
		value += Score(ply)
	}

	// One assignment to the slot; keeps writes compact.
	data[minIdx] = Entry{
		pKey:       pkey,
		SimpleMove: sm,
		value:      value,
		Depth:      d,
		Type:       typ,
	}
}
