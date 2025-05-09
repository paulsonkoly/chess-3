package heur

import (
	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

type Continuation struct {
	// color, pt, sq, pt2, sq2
	data [2][6][64][6][64]Score
}

func NewContinuation() *Continuation {
	return &Continuation{}
}

// Deflate divides every entry in the store by 2.
func (c *Continuation) Deflate() {
	for i := range c.data {
		for j := range c.data[i] {
			for k := range c.data[i][j] {
				for h := range c.data[i][j][k] {
					for l := range c.data[i][j][k][h] {
						c.data[i][j][k][h][l] /= 2
					}
				}
			}
		}
	}
}

// Add increments the continuation history heuristics for the move by d*d.
func (c *Continuation) Add(stm Color, ptHist Piece, toHist Square, pt Piece, to Square, bonus Score) {
	clampedBonus := Clamp(bonus, -MaxHistory, MaxHistory)
	old := c.data[stm][ptHist-1][toHist][pt-1][to]
	c.data[stm][ptHist-1][toHist][pt-1][to] += clampedBonus - Score(int(old)*int(Abs(clampedBonus))/MaxHistory)
}

// Probe returns the continuation history heuristics entry for the move.
func (c *Continuation) Probe(stm Color, ptHist Piece, toHist Square, pt Piece, to Square) Score {
	return c.data[stm][ptHist-1][toHist][pt-1][to]
}
