package search

import (
	"github.com/paulsonkoly/chess-3/move"

	. "github.com/paulsonkoly/chess-3/types"
)

// pv is the principal variation buffer.
//
// It holds an array of moves for each ply. When a line is accepted at ply n
// the moves from ply n+1 are copied into the array of ply n, offseted by n.
//
// For each ply the length of the required array is one less than the previous
// ply's array. Ply 0 assumes the length to be MaxPlies. Thus the total buffer
// size is MaxPlies + (MaxPlies-1) + ... + 1 == MaxPlies * (MaxPlies + 1) / 2.
type pv struct {
	// moves is the double buffered PV
	moves [MaxPlies * (MaxPlies + 1) / 2]move.SimpleMove

	depth [MaxPlies]Depth
}

// newPV creates a new PV buffer.
func newPV() *pv {
	return &pv{}
}

// insert inserts the move m at depth ply saving the tail of the PV.
func (pv *pv) insert(ply Depth, m move.SimpleMove) {
	i := bufIx(ply)
	j := bufIx(ply + 1)
	l := pv.depth[ply+1]

	pv.moves[i] = m
	copy(pv.moves[i+1:i+1+int(l)], pv.moves[j:j+int(l)])
	pv.depth[ply] = l + 1
}

// setTip inserts the move m at depth ply setting it to be the end of the PV.
func (pv *pv) setTip(ply Depth, m move.SimpleMove) {
	pv.moves[bufIx(ply)] = m
	pv.depth[ply] = 1
}

// setNull sets the current pv length to 0 at depth ply.
func (pv * pv)setNull(ply Depth) {
	pv.depth[ply] = 0
}

func bufIx(ply Depth) int {
	return int(ply)*MaxPlies - int(ply)*int(ply-1)/2
}

// active is the principal variation held in the PV buffer.
func (pv *pv) active() []move.SimpleMove {
	return pv.moves[0:pv.depth[0]]
}
