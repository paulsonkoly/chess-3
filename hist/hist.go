package hist

import (
	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

const MaxHistoryScore = 1023

type Store struct {
	data [2][64][64]Score
}

func New() *Store {
	return &Store{}
}

func (h *Store) Deflate() {
	for color := White; color <= Black; color++ {
		for sqFrom := A1; sqFrom <= H8; sqFrom++ {
			for sqTo := A1; sqTo <= H8; sqTo++ {
				h.data[color][sqFrom][sqTo] >>= 1
			}
		}
	}
}

func (h *Store) Add(stm Color, from, to Square, d Depth) {
	hist := h.data[stm][from][to] + Score(d)*Score(d)
	if hist <= MaxHistoryScore {
		h.data[stm][from][to] = hist
	}
}

func (h *Store) Probe(stm Color, from, to Square) Score {
	return h.data[stm][from][to]
}
