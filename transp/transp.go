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
	// bucketSize is the number of entries per bucket.
	bucketSize = 4
	// entrySize is the number of bytes per entry.
	entrySize = 8
	// partialKeyBits is the number of bits of the Zobrist-hash stored per entry.
	partialKeyBits = 16
)

// Table is the transposition table.
type Table struct {
	raw    []byte   // raw is reference to unaligned underlying data to keep it from GC
	data   []Entry  // Cache entries minus the pKey.
	pKeys  []uint64 // 64-bit packed pKeys.
	ixMask board.Hash
}

func init() {
	// entry size has to be 8 bytes.
	packSize := unsafe.Sizeof(Entry{})
	if packSize != entrySize {
		panic(fmt.Sprintf("tt entry must be %d bytes but is %d bytes; adjust layout", entrySize, packSize))
	}

	// we need to be able to fit pKeys into a number of ui64s per bucket. We
	// could support larger buckets, as long as this is multiple of 64.
	if (bucketSize * partialKeyBits) != 64 {
		panic("tt partial key packing error")
	}
}

// Type represents the stored bound type.
type Type byte

const (
	Exact Type = iota
	UpperBound
	LowerBound
)

type partialKey = uint16

type Entry struct {
	move.SimpleMove        // SimpleMove is the hash move. (2 bytes)
	value           Score  // (2 bytes).
	Depth           Depth  // Depth is the entry depth. (1 byte)
	Type            Type   // Type indicates the node type / whether the score is exact or bound. (1 byte)
	_               uint16 // pad to 8 bytes (2 bytes)
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

	alignment := bucketSize * entrySize
	numBuckets := size / alignment
	numEntries := numBuckets * bucketSize

	// Overallocate by cacheLine-1 and align the returned pointer to cacheLine
	raw := make([]byte, size+alignment-1)
	base := uintptr(unsafe.Pointer(&raw[0]))
	aligned := (base + uintptr(alignment-1)) &^ uintptr(alignment-1)

	ptr := unsafe.Pointer(aligned)
	entries := unsafe.Slice((*Entry)(ptr), numEntries)

	// each entry has a 16 bit pKey, and we pack pKeys in 64 bits.
	numPKeys := numEntries * partialKeyBits / 64
	pKeys := make([]uint64, numPKeys)

	return &Table{
		raw:    raw,
		data:   entries,
		pKeys:  pKeys,
		ixMask: board.Hash(numBuckets - 1),
	}
}

// HashFull returns a dummy usage percentage (implement as you need).
func (t Table) HashFull() int {
	cnt := 0
	for i := range min(1000, len(t.data)/bucketSize) {
		if t.data[i*bucketSize].Depth != 0 {
			cnt++
		}
	}
	return cnt
}

// Clear zeros the depth field for all entries.
func (t *Table) Clear() {
	for i := range t.data {
		// this is a soft clear, a cheap implementation of aging. As the hashes are
		// still in-tact these entries are reusable in the next search, they just
		// always lose on replacement.
		t.data[i].Depth = 0
	}
}

// bucketIx returns the starting entry index of the bucket for hash.
func (t *Table) bucketIx(hash board.Hash) int {
	return int((hash & t.ixMask)) * bucketSize
}

// pkeyIx returns the index of the starting ui64 for hash .
func (t *Table) pkeyIx(hash board.Hash) int {
	// this logic is hard wired atm for 4 16 bit pKeys per 64 bit word.
	return int(hash & t.ixMask)
}

// LookUp looks up the entry for hash.
func (t *Table) LookUp(hash board.Hash) (*Entry, bool) {
	bucketKeys := t.pKeys[t.pkeyIx(hash)]
	hashKey := partialKey(hash >> (64 - partialKeyBits))

	for i := range bucketSize {
		if hashKey == partialKey(bucketKeys) {
			bIx := t.bucketIx(hash)
			eIx := bIx + i

			return &t.data[eIx], true
		}
		bucketKeys >>= partialKeyBits
	}

	return nil, false
}

// Insert writes an entry into the transposition table.
func (t *Table) Insert(hash board.Hash, d, ply Depth, sm move.SimpleMove, value Score, typ Type) {
	bIx := t.bucketIx(hash)

	var replace int
	// -1 to make sure we have a replacement. If no entries in the bucket are
	// replecable, replace the last one unconditionally.
	for replace = bIx; replace < bIx+bucketSize-1; replace++ {
		if t.data[replace].Depth < d {
			break
		}

		if t.data[replace].Depth == d && (t.data[replace].Type == UpperBound || typ != UpperBound) {
			break
		}
	}

	if value < -Inf+MaxPlies {
		value -= Score(ply)
	}
	if value > Inf-MaxPlies {
		value += Score(ply)
	}

	t.data[replace] = Entry{
		SimpleMove: sm,
		value:      value,
		Depth:      d,
		Type:       typ,
	}

	pkIx := t.pkeyIx(hash)
	hashKey := partialKey(hash >> (64 - partialKeyBits))
	bucketKeys := t.pKeys[pkIx]
	eIx := replace - bIx
	bucketKeys &= ^(((1<<partialKeyBits) - 1) << (eIx * partialKeyBits))
	bucketKeys |= uint64(hashKey) << (eIx * partialKeyBits)

	t.pKeys[pkIx] = bucketKeys
}
