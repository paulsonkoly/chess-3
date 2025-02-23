package heur

import (
	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

// MaxHistoryScore is the maximal score the history table can contain.
const MaxHistoryScore = Captures - QuietHistory - 1 // this needs to be poswer of 2 - 1

// History heuristics.
//
// Stores move weights for queit moves.
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
				h.data[color][sqFrom][sqTo] >>= 1
			}
		}
	}
}

// Add increments the history heuristics for the move by d*d.
func (h *History) Add(stm Color, from, to Square, d Depth) {
	hist := h.data[stm][from][to] + Score(d)*Score(d)
	if hist <= MaxHistoryScore {
		h.data[stm][from][to] = hist
	}
}

// Probe returns the history heuristics entry for the move.
func (h *History) Probe(stm Color, from, to Square) Score {
	return h.data[stm][from][to]
}
