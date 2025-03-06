package heur

import (
	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

type Continuation struct {
	// color, pt, sq, pt2, sq2
	data [2 * 6 * 64 * 6 * 64]Score
}

func NewContinuation() *Continuation {
	return &Continuation{}
}

// Deflate divides every entry in the store by 2.
func (c *Continuation) Deflate() {
	for i := range c.data {
		c.data[i] >>= 1
	}
}

func ix(stm Color, ptHist Piece, toHist Square, pt Piece, to Square) int {
	return int(to) + 64*int(pt-1) + 6*64*int(toHist) + 64*6*64*int(ptHist-1) + 6*64*6*64*int(stm)
}

// Add increments the continuation history heuristics for the move by d*d.
func (c *Continuation) Add(stm Color, ptHist Piece, toHist Square, pt Piece, to Square, s Score) {
	ix := ix(stm, ptHist, toHist, pt, to)
	bonus := c.data[ix] + s
	bonus = min(bonus, MaxHistory)
	bonus = max(bonus, -MaxHistory)

	c.data[ix] = bonus
}

// Probe returns the continuation history heuristics entry for the move.
func (c *Continuation) Probe(stm Color, ptHist Piece, toHist Square, pt Piece, to Square) Score {
	return c.data[ix(stm, ptHist, toHist, pt, to)]
}
