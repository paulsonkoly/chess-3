package heur

import (
	. "github.com/paulsonkoly/chess-3/chess"
)

// History heuristics.
//
// Stores move weights for quiet moves.
type History struct {
	data [Colors][Squares][Squares]Score
}

// NewHistory creates a new history heuristics.
func NewHistory() *History {
	return &History{}
}

// Clear resets all entries to 0.
func (h *History) Clear() {
	for color := range Colors {
		for from := range Squares {
			for to := range Squares {
				h.data[color][from][to] = 0
			}
		}
	}
}

// Add increments the history heuristics for the move by bonus.
func (h *History) Add(stm Color, from, to Square, bonus Score) {
	clampedBonus := Clamp(bonus, -MaxHistory, MaxHistory)
	h.data[stm][from][to] += clampedBonus - Score(int(h.data[stm][from][to])*int(Abs(clampedBonus))/int(MaxHistory))
}

// LookUp returns the history heuristics entry for the move.
func (h *History) LookUp(stm Color, from, to Square) Score {
	return h.data[stm][from][to]
}
