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

const entrySize = 16 // EntrySize is the transposition table entry size in bytes.

type Table struct {
	data   []Entry
	numE   int
	ixMask board.Hash
	cnt    int
}

type NodeT byte

const (
	// PVNode type entry is an exact score entry for positions that are in the alpha-beta window.
	PVNode NodeT = iota
	// AllNode type entry is a fail-low node, score failed under alpha.
	AllNode
	// CutNode type entry is a fail-high node, score failed above beta.
	CutNode
)

// Entry is a transposition table entry.
//
// We use up 16 bytes
type Entry struct {
	move.SimpleMove            // SimpleMove is the simplified move data.
	Depth           Depth      // Depth of the entry.
	Value           Score      // Value is the entry score value. Not valid for nodes where the score is not established.
	TFCnt           Depth      // Three-fold repetation count of the entry.
	Type            NodeT      // Type is the entry type.
	Hash            board.Hash // Hash is the board Zobrist-hash.
}

// New creates a new transposition table.
func New(sizeInMb int) *Table {
	if sizeInMb < 1 || sizeInMb&(sizeInMb-1) != 0 {
		panic(fmt.Sprintf("invalid transposition table size %d", sizeInMb))
	}

	sizeInBytes := sizeInMb * 1024 * 1024
	numEntries := sizeInBytes / entrySize
	numEntriesL2 := bits.TrailingZeros(uint(numEntries))

	return &Table{
		data:   make([]Entry, numEntries),
		numE:   numEntries,
		ixMask: (1 << numEntriesL2) - 1,
	}
}

// HashFull is the permill count of the hash usage.
func (t Table) HashFull() int {
	return 1000 * t.cnt / t.numE
}

// Clear clears the transposition table for the next search.Search().
func (t *Table) Clear() {
	t.cnt = 0
	for ix := range t.data {
		t.data[ix].Depth = 0
	}
}

// Insert inserts an entry to the transposition table if the current hash in
// the table has a lower depth than d.
func (t *Table) Insert(hash board.Hash, d, tfCnt Depth, sm move.SimpleMove, value Score, typ NodeT) {
	ix := hash & t.ixMask

	if t.data[ix].Depth > d {
		return
	}

	if t.data[ix].Depth == 0 {
		t.cnt++
	}

	t.data[ix] = Entry{
		SimpleMove: sm,
		Hash:       hash,
		Value:      value,
		Type:       typ,
		Depth:      d,
		TFCnt:      tfCnt,
	}
}

// Lookup looks up the transposition table entry, using hash as the key.
func (t *Table) LookUp(hash board.Hash) (*Entry, bool) {
	ix := hash & t.ixMask

	if t.data[ix].Hash == hash {
		return &t.data[ix], true
	}

	return nil, false
}
