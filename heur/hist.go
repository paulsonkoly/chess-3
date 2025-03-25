package heur

import (
	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

// History heuristics.
//
// Stores move weights for quiet moves.
type History struct {
	data [2][64][64]Score
}

// NewHistory creates a new history heuristics.
func NewHistory() *History {
	return &History{}
}

// Deflate divides every entry in the store by 2.
func (h *History) Deflate() {
	for color := White; color <= Black; color++ {
		for sqFrom := A1; sqFrom <= H8; sqFrom++ {
			for sqTo := A1; sqTo <= H8; sqTo++ {
				h.data[color][sqFrom][sqTo] /= 2
			}
		}
	}
}

// Add increments the history heuristics for the move by bonus.
func (h *History) Add(stm Color, from, to Square, bonus Score) {
	clampedBonus := Clamp(bonus, -MaxHistory, MaxHistory)
	h.data[stm][from][to] += clampedBonus - Score(int(h.data[stm][from][to])*int(Abs(clampedBonus))/MaxHistory)
}

// Probe returns the history heuristics entry for the move.
func (h *History) Probe(stm Color, from, to Square) Score {
	return h.data[stm][from][to]
}
