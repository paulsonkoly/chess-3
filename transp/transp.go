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

// Gen is the search counter for aging.
type Gen byte

const (
	Exact Type = iota
	UpperBound
	LowerBound
)

// packed depth and type into a single byte
type packed byte

// Depth is the entry depth.
func (p packed) Depth() Depth { return Depth(p >> 2) }

// Type indicates the node type / whether the score is exact or bound.
func (p packed) Type() Type { return Type(p & 3) }

type entry struct {
	move.SimpleMove       // SimpleMove is the hash move. (2 bytes)
	value           Score // (2 bytes).
	packed                // packed depth and type (1 byte)
	gen             Gen
}

// Value is the score of the entry corrected for current ply in case of mate score.
func (e *entry) Value(ply Depth) Score {
	if e.value > Inf-MaxPlies {
		return e.value - Score(ply)
	}

	if e.value < -Inf+MaxPlies {
		return e.value + Score(ply)
	}

	return e.value
}

func (e *entry) quality(curr Gen) int { return quality(curr, e.gen, e.Depth(), e.Type()) }

// partialKey is the bits of the Zobrist stored in the table.
type partialKey uint16

type bucket struct {
	pKeys   uint64                // pKeys are the partial keys for entries present in this bucket.
	entries [bucketEntryCnt]entry // entries set of entries that compete in replacement.

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
		panic(fmt.Sprintf("unexpected tt bucket size %d expected %d. This is a tt bug.\n", trueSize, bucketSize))
	}
}

// New creates a new transposition table. size is the table size in bytes, and
// only power of 2 sizes are supported.
func New(size int) *Table {
	if size < bucketSize || size&(size-1) != 0 {
		panic(fmt.Sprintf("invalid transposition table size %d", size))
	}

	numBuckets := size / bucketSize

	// over allocate raw pool, so we can start at bucket alignment. We want
	// buckets to fall on a single 64 byte CPU cache line.
	raw := make([]byte, size+bucketSize-1)
	base := uintptr(unsafe.Pointer(&raw[0]))
	aligned := (base + uintptr(bucketSize-1)) &^ uintptr(bucketSize-1)

	ptr := unsafe.Pointer(aligned)
	buckets := unsafe.Slice((*bucket)(ptr), numBuckets)

	return &Table{raw: raw, data: buckets, ixMask: board.Hash(numBuckets - 1)}
}

// HashFull is the permill use estimate of the tt.
func (t Table) HashFull(gen Gen) int {
	cnt := 0
	if len(t.data) < 1000 {
		panic("tt size is too small to measure permill HashFull")
	}
	for _, bucket := range t.data[:1000] {
		for _, entry := range bucket.entries {
			if entry.Depth() > 0 && entry.gen == gen {
				cnt++
			}
		}
	}
	return cnt / 4 // 4 entries per bucket
}

// Clear empties the tt.
func (t *Table) Clear() {
	for i, bucket := range t.data {
		t.data[i].pKeys = 0

		for j := range bucket.entries {
			t.data[i].entries[j] = entry{}
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
func (t *Table) Insert(hash board.Hash, gen Gen, d, ply Depth, sm move.SimpleMove, value Score, typ Type) {
	bucket := &t.data[t.bucketIx(hash)]

	hashKey := partialKey(hash >> (64 - partialKeyBits))
	bucketKeys := bucket.pKeys

	currQ := quality(gen, gen, d, typ)
	minQ := 1000
	var replace int
	for i := range bucketEntryCnt {
		entry := &bucket.entries[i]
		entryQ := entry.quality(gen)

		if partialKey(bucketKeys) == hashKey {
			if entryQ > currQ {
				return
			}
			replace = i
			break
		}

		if entryQ < minQ {
			minQ = entryQ
			replace = i
		}

		bucketKeys >>= partialKeyBits
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
		packed:     packed(d)<<2 | packed(typ),
		gen:        gen,
	}

	bucket.pKeys &= ^(((1 << partialKeyBits) - 1) << (replace * partialKeyBits))
	bucket.pKeys |= uint64(hashKey) << (replace * partialKeyBits)
}

func quality(curr, g Gen, d Depth, typ Type) int {
	typQ := 1
	if typ == UpperBound {
		typQ = 0
	}
	return int(d) + typQ + int(g)-int(curr)
}
