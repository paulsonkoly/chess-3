// package transp is a transposition table.
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
	// bucketEntryCnt is the number of entries per bucket.
	bucketEntryCnt = 4
	// bucketSize is the bucket size in bytes.
	bucketSize = 32
	// partialKeyBits is the number of bits of the Zobrist-hash stored per entry.
	partialKeyBits = 16
)

// Type represents the stored bound type.
type Type byte

const (
	Exact Type = iota
	UpperBound
	LowerBound
)

type entry struct {
	move.SimpleMove       // SimpleMove is the hash move. (2 bytes)
	value           Score // (2 bytes).
	Depth           Depth // Depth is the entry depth. (1 byte)
	Type            Type  // Type indicates the node type / whether the score is exact or bound. (1 byte)
}

// Value is the score of the entry corrected for current ply in case of mate score.
func (e entry) Value(ply Depth) Score {
	if e.value > Inf-MaxPlies {
		return e.value - Score(ply)
	}

	if e.value < -Inf+MaxPlies {
		return e.value + Score(ply)
	}

	return e.value
}

// partialKey is the bits of the Zobrist stored in the table.
type partialKey uint16

type bucket struct {
	pKeys   uint64                // pKeys are the partial keys for entries present in this bucket.
	entries [bucketEntryCnt]entry // entries set of entries that compete in replecament.

}

// Table is the transposition table.
type Table struct {
	raw    []byte   // raw is reference to unaligned underlying data to keep it from GC
	data   []bucket // Cache entries minus the pKey.
	ixMask board.Hash
}

func init() {
	trueSize := unsafe.Sizeof(bucket{})
	if trueSize != bucketSize {
		panic(fmt.Sprintf("enexpected tt bucket size %d expected %d. This is a tt bug.\n", trueSize, bucketSize))
	}
}

// New creates a new transposition table. size is the table size in bytes, and
// only power of 2 sizes are supported.
func New(size int) *Table {
	if size < bucketSize || size&(size-1) != 0 {
		panic(fmt.Sprintf("invalid transposition table size %d", size))
	}

	numBuckets := size / bucketSize

	// overallocate raw pool, so we can start at bucket alignment. We want
	// buckets to fall on a single 64 byte CPU cache line.
	raw := make([]byte, size+bucketSize-1)
	base := uintptr(unsafe.Pointer(&raw[0]))
	aligned := (base + uintptr(bucketSize-1)) &^ uintptr(bucketSize-1)

	ptr := unsafe.Pointer(aligned)
	buckets := unsafe.Slice((*bucket)(ptr), numBuckets)

	return &Table{raw: raw, data: buckets, ixMask: board.Hash(numBuckets - 1)}
}

// HashFull returns a dummy usage percentage (implement as you need).
func (t Table) HashFull() int {
	cnt := 0
	for i := range min(1000, len(t.data)) {
		if t.data[i].entries[0].Depth != 0 {
			cnt++
		}
	}
	return cnt
}

// Clear zeros the depth field for all entries.
func (t *Table) Clear() {
	for i := range t.data {
		for j := range t.data[i].entries {
			// this is a soft clear, a cheap implementation of aging. As the hashes are
			// still in-tact these entries are reusable in the next search, they just
			// always lose on replacement.
			t.data[i].entries[j].Depth = 0
		}
	}
}

// bucketIx returns the index of the bucket for hash.
func (t *Table) bucketIx(hash board.Hash) int {
	return int(hash & t.ixMask)
}

// LookUp looks up the entry for hash.
func (t *Table) LookUp(hash board.Hash) (*entry, bool) {
	bucket := &t.data[t.bucketIx(hash)]
	bucketKeys := bucket.pKeys
	hashKey := partialKey(hash >> (64 - partialKeyBits))

	for i := range bucketEntryCnt {
		if hashKey == partialKey(bucketKeys) {
			return &bucket.entries[i], true
		}
		bucketKeys >>= partialKeyBits
	}

	return nil, false
}

// Insert writes an entry into the transposition table.
func (t *Table) Insert(hash board.Hash, d, ply Depth, sm move.SimpleMove, value Score, typ Type) {
	bucket := &t.data[t.bucketIx(hash)]

	var replace int
	for replace = range bucketEntryCnt {
		entry := &bucket.entries[replace]
		if entry.Depth < d {
			break
		}

		if entry.Depth == d && (entry.Type == UpperBound || typ != UpperBound) {
			break
		}
	}

	if value < -Inf+MaxPlies {
		value -= Score(ply)
	}
	if value > Inf-MaxPlies {
		value += Score(ply)
	}

	bucket.entries[replace] = entry{
		SimpleMove: sm,
		value:      value,
		Depth:      d,
		Type:       typ,
	}

	hashKey := hash >> (64 - partialKeyBits)
	bucket.pKeys &= ^(((1 << partialKeyBits) - 1) << (replace * partialKeyBits))
	bucket.pKeys |= uint64(hashKey) << (replace * partialKeyBits)
}
