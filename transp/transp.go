// package transp is the transposition table cache.
package transp

import (
	"fmt"
	"math/bits"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

const (
	bucketSize    = 32 // entrySize is the transposition table entry size in bytes.
	bucketEntries = 2  // bucket is the number of entries in a bucket.
)

type Table struct {
	data   []Bucket
	numE   int
	ixMask board.Hash
	cnt    int
}

// 32 bytes
type Bucket struct {
	data [bucketEntries]Entry // data is the pair of hash entries, one
}

// replIx is the replacement index in a bucket
func (b *Bucket) replIx(hash board.Hash, age Age) int {
	minD := int(255)<<10 + int(MaxPlies-1)<<2 + 3
	minIx := -1
	for ix := range bucketEntries {
		// if hash matches replace. This should guarantee different hashes in a bucket per entry.
		if b.data[ix].hash == hash {
			return ix
		}

		ageDiff := (int(b.data[ix].Age) - int(age) + 255) % 256

		importance := (ageDiff)<<10 + int(b.data[ix].Depth)<<2 + int(b.data[ix].Type)

		if importance < minD {
			minD = importance
			minIx = ix
		}
	}

	return minIx
}

type NodeT byte

const (
	// AllNode type entry is a fail-low node, score failed under alpha.
	AllNode NodeT = iota
	// CutNode type entry is a fail-high node, score failed above beta.
	CutNode
	// PVNode type entry is an exact score entry for positions that are in the alpha-beta window.
	PVNode
)

// Entry is a transposition table entry.
//
// 16 bytes
type Entry struct {
	move.SimpleMove            // SimpleMove is the simplified move data.
	Value           Score      // Value is the entry score value. Not valid for nodes where the score is not established.
	Depth           Depth      // Depth of the entry.
	TFCnt           Depth      // Three-fold repetation count of the entry.
	Age             Age        // Age is the rolling search counter.
	Type            NodeT      // Type is the entry type.
	hash            board.Hash // Hash is the board Zobrist-hash.
}

// New creates a new transposition table.
func New(sizeInMb int) *Table {
	if sizeInMb < 1 || sizeInMb&(sizeInMb-1) != 0 {
		panic(fmt.Sprintf("invalid transposition table size %d", sizeInMb))
	}

	sizeInBytes := sizeInMb * 1024 * 1024
	numEntries := sizeInBytes / bucketSize
	numEntriesL2 := bits.TrailingZeros(uint(numEntries))

	return &Table{
		data:   make([]Bucket, numEntries),
		numE:   numEntries,
		ixMask: (1 << numEntriesL2) - 1,
	}
}

// Clear clears the transposition table.
func (t *Table) Clear() {
	t.cnt = 0
	for ix := range t.data {
		for jx := range bucketEntries {
			t.data[ix].data[jx].hash = 0
			t.data[ix].data[jx].Age = 0
			t.data[ix].data[jx].Depth = 0
		}
	}
}

// Insert inserts an entry to the transposition table kicking out the worst entry from the bucket.
func (t *Table) Insert(hash board.Hash, d, tfCnt Depth, age Age, sm move.SimpleMove, value Score, typ NodeT) {
	ix := hash & t.ixMask

	wx := t.data[ix].replIx(hash, age)

	repl := &t.data[ix].data[wx]

	if repl.Age == age && repl.hash == hash && repl.Type == PVNode && typ != PVNode {
		return
	}

	if repl.Depth == 0 {
		t.cnt++
	}

	*repl = Entry{
		SimpleMove: sm,
		hash:       hash,
		Value:      value,
		TFCnt:      tfCnt,
		Type:       typ,
		Depth:      d,
		Age:        age,
	}
}

// Lookup looks up the transposition table entry, using hash as the key.
func (t *Table) Probe(hash board.Hash) (*Entry, bool) {
	ix := hash & t.ixMask

	for jx := range bucketEntries {
		if t.data[ix].data[jx].hash == hash {
			return &t.data[ix].data[jx], true
		}
	}

	return nil, false
}
