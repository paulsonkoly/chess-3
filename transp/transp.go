package transp

import (
	"github.com/paulsonkoly/chess-3/board"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

const (
	TableSize      = 1 << 18 // 4Mb
	ReplacementAge = 1_000_000
	ProbeLength    = 20
)

type Table struct {
	data []Entry
}

type NodeT byte

const (
	PVNode NodeT = iota
	AllNode
	CutNode
)

// Entry is a transposition table entry.
//
// We use up 16 bytes, and have a spare byte atm.
type Entry struct {
	Hash  board.Hash
	Value Score
	Depth Depth
	From  Square
	To    Square
	Promo Piece
	Type  NodeT
}

func NewTable() *Table {
	return &Table{
		data: make([]Entry, TableSize),
	}
}

func (t *Table) Insert(hash board.Hash, d Depth, from, to Square, promo Piece, value Score, typ NodeT) {
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
	}
}

func (t *Table) LookUp(hash board.Hash) (*Entry, bool) {
	ix := hash % TableSize

	if t.data[ix].Hash == hash {
		return &t.data[ix], true
	}

	return nil, false
}
