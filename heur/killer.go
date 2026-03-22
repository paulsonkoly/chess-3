package heur

import (
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/move"
)

// KillerStride is the number of killer moves per ply.
const KillerStride = 3

// Killer stores the killer moves per ply.
type Killer struct {
	data [MaxPlies][KillerStride]move.Move
}

// NewKiller creates a new killer table.
func NewKiller() *Killer {
	return &Killer{}
}

// Clear resets the killer table with 0 values.
func (k *Killer) Clear() {
	k.data = [MaxPlies][KillerStride]move.Move{}
}

// Add inserts m to ply.
func (k *Killer) Add(ply Depth, m move.Move) {
	ix := 0
	for i, v := range k.data[ply] {
		if v == m || v == 0 {
			ix = i
			break
		}
	}
	k.data[ply][0], k.data[ply][ix] = k.data[ply][ix], k.data[ply][0]
	k.data[ply][0] = m
}

// LookUp returns the killer move - potentially 0, from ply. sel
// selects the killer move priority, 0: most recent. sel has to be in
// 0..KillerStride.
func (k *Killer) LookUp(ply Depth, sel int) move.Move {
	return k.data[ply][sel]
}
