// package transp is the transposition table cache.
package transp

import (
	"github.com/paulsonkoly/chess-3/board"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

const (
	TableSize = 1 << 18 // 4Mb
)

type Table struct {
	data []Entry
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
	Hash  board.Hash // Hash is the board Zobrist-hash.
	Value Score      // Value is the entry score value. Not valid for nodes where the score is not established.
	Depth Depth      // Depth of the entry.
	TFCnt Depth      // Three-fold repetation count of the entry.
	From  Square     // From is the move from square. Only valid for entries where a move was chosen.
	To    Square     // To is the move to square. Only valid for entries where a move was chosen.
	Promo Piece      // Promo is the promotion piece of the move. Only valid where a move was chosen.
	Type  NodeT      // Type is the entry type.
}

// New creates a new transposition table.
func New() *Table {
	return &Table{
		data: make([]Entry, TableSize),
	}
}

// Clear clears the transposition table for the next search.Search().
func (t *Table) Clear() {
	for ix := range t.data {
		t.data[ix].Depth = 0
	}
}

// Insert inserts an entry to the transposition table if the current hash in
// the table has a lower depth than d.
func (t *Table) Insert(hash board.Hash, d, tfCnt Depth, from, to Square, promo Piece, value Score, typ NodeT) {
	ix := hash % TableSize

	if t.data[ix].Depth > d {
		return
	}

	t.data[ix] = Entry{
		Hash:  hash,
		From:  from,
		To:    to,
		Promo: promo,
		Value: value,
		Type:  typ,
		Depth: d,
		TFCnt: tfCnt,
	}
}

// Lookup looks up the transposition table entry, using hash as the key.
func (t *Table) LookUp(hash board.Hash) (*Entry, bool) {
	ix := hash % TableSize

	if t.data[ix].Hash == hash {
		return &t.data[ix], true
	}

	return nil, false
}
