package heur

import (
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/move"
)

// Killer stores the killer moves per ply.
type Killer struct {
	data [MaxPlies]move.Move
}

// NewKiller creates a new killer table.
func NewKiller() *Killer {
	return &Killer{}
}

// Clear resets the killer table with 0 values.
func (k *Killer) Clear() {
	k.data = [MaxPlies]move.Move{}
}

// Add inserts m to ply.
func (k *Killer) Add(ply Depth, m move.Move) {
	k.data[ply] = m
}

// LookUp returns the killer move - potentially 0, from ply.
func (k *Killer) LookUp(ply Depth) move.Move {
	return k.data[ply]
}
