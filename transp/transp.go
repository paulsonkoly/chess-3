// Package transp is a transposition table.
package transp

import (
	"fmt"
	"math/bits"
	"unsafe"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"

	. "github.com/paulsonkoly/chess-3/chess"
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
	UpperBound Type = iota // Entry score is upper bound only.
	LowerBound             // Entry score is lower bound only.
	Exact                  // Entry score is exact.
)

// packed depth and type into a single byte.
type packed byte

// Depth is the entry depth.
func (p packed) Depth() Depth { return Depth(p >> 2) }

// Type indicates the node type / whether the score is exact or bound.
func (p packed) Type() Type { return Type(p & 3) }

type entry struct {
	move.Move       // SimpleMove is the hash move. (2 bytes)
	value     Score // (2 bytes)
	packed          // packed depth and type. (1 byte)
	gen       Gen   // (1 byte)
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

func (e *entry) quality(curr Gen) int { return quality(curr, e.gen, e.Depth()) }

// partialKey is the bits of the Zobrist stored in the table.
type partialKey uint16

type bucket struct {
	pKeys   uint64                // pKeys are the partial keys for entries present in this bucket.
	entries [bucketEntryCnt]entry // entries set of entries that compete in replacement.

}

// Table is the transposition table.
type Table struct {
	raw  []byte   // raw is reference to unaligned underlying data to keep it from GC
	data []bucket // Cache entries minus the pKey.
}

func init() {
	trueSize := unsafe.Sizeof(bucket{})
	if trueSize != bucketSize {
		panic(fmt.Sprintf("unexpected tt bucket size %d expected %d. This is a tt bug.\n", trueSize, bucketSize))
	}
}

// New creates a new transposition table. size is the table size in bytes.
func New(size int) *Table {
	t := Table{}
	t.Resize(size)
	return &t
}

// Resize resizes the table to size bytes, potentially reallocating its
// resources. The table data is not guaranteed to be kept in-tact. It is
// recommended, but not a must to clear the table after a resize.
func (t *Table) Resize(size int) {
	validateSize(size)

	requiredBuckets := size / bucketSize

	if len(t.data) >= requiredBuckets {
		// underlying raw buffer doesn't need to be touched, we just resize the buckets slice
		t.data = t.data[:requiredBuckets]
	} else {
		// over allocate raw pool, so we can start at bucket alignment. We want
		// buckets to fall on a single 64 byte CPU cache line, contained in either
		// the low or the high 32 bytes.
		t.raw = make([]byte, size+bucketSize-1)
		base := uintptr(unsafe.Pointer(&t.raw[0]))
		aligned := (base + uintptr(bucketSize-1)) &^ uintptr(bucketSize-1)

		ptr := unsafe.Pointer(aligned)
		t.data = unsafe.Slice((*bucket)(ptr), requiredBuckets)
	}
}

func validateSize(size int) {
	if size < bucketSize || size%bucketSize != 0 {
		panic(fmt.Sprintf("invalid transposition table size %d", size))
	}
}

// HashFull is the permill use estimate of the tt.
func (t Table) HashFull(gen Gen) int {
	cnt := 0
	if len(t.data) < 1000 {
		panic("tt size is too small to measure permill HashFull")
	}
	for _, bucket := range t.data[:1000] {
		for i, entry := range bucket.entries {
			if (bucket.pKeys>>(i*partialKeyBits))&(1<<partialKeyBits-1) != 0 && entry.gen == gen {
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
	// Lemire's fast modulo trick. If the table grows beyond 4GB, only the
	// collision rate increases; correctness is unaffected. See:
	// https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
	h := uint32(hash)
	return int(uint64(h) * uint64(len(t.data)) >> 32)
}

const (
	rep16 = 0x0001_0001_0001_0001
	hi16  = 0x8000_8000_8000_8000
)

// match64 finds the index of matching 16 bit lane in the 64 bit word w,
// comparing each 16 bit lanes with key. If there are no matches it returns ok
// false.
func match64(w uint64, key partialKey) (ix int, ok bool) {
	r := uint64(key) * rep16
	x := w ^ r
	mask := (x - rep16) & ^x & hi16
	if mask == 0 {
		return
	}
	return bits.TrailingZeros64(mask) / 16, true
}

// LookUp looks up the entry for hash.
func (t *Table) LookUp(hash board.Hash) (*entry, bool) {
	bucket := &t.data[t.bucketIx(hash)]
	bucketKeys := bucket.pKeys
	hashKey := partialKey(hash >> (64 - partialKeyBits))

	if ix, ok := match64(bucketKeys, hashKey); ok {
		return &bucket.entries[ix], true
	}

	return nil, false
}

// Insert writes an entry into the transposition table.
func (t *Table) Insert(hash board.Hash, gen Gen, d, ply Depth, sm move.Move, value Score, typ Type) {
	bucket := &t.data[t.bucketIx(hash)]

	hashKey := partialKey(hash >> (64 - partialKeyBits))
	bucketKeys := bucket.pKeys

	// sufficiently large start value for minimum search
	minQ := 1 << 50
	var replace int
	var target *entry
	for i := range bucketEntryCnt {
		target = &bucket.entries[i]
		entryQ := target.quality(gen)

		if partialKey(bucketKeys) == hashKey {
			if typ != Exact && target.Depth() > d+2 && target.gen == gen {
				return
			}

			// if sm is null but we have a move in the entry keep it "stockfish" trick
			if sm == 0 {
				sm = target.Move
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
		Move:   sm,
		value:  value,
		packed: packed(d)<<2 | packed(typ),
		gen:    gen,
	}

	bucket.pKeys &= ^(((1 << partialKeyBits) - 1) << (replace * partialKeyBits))
	bucket.pKeys |= uint64(hashKey) << (replace * partialKeyBits)
}

func quality(curr, g Gen, d Depth) int {
	return int(d) + 2*(int(g)-int(curr))
}
