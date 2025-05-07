package heur

import (
	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

type CaptHist struct {
	// capturing piece, to square, captured piece
	data [6][64][6]Score
}

func NewCaptHist() *CaptHist {
	return &CaptHist{}
}

// Deflate divides every entry in the store by 2.
func (c *CaptHist) Deflate() {
	for p1 := range c.data {
		for t := range c.data {
			for p2 := range c.data {
				c.data[p1][t][p2] /= 2
			}
		}
	}
}

// Add increments the continuation history heuristics for the move by d*d.
func (c *CaptHist) Add(piece Piece, to Square, captured Piece, bonus Score) {
	ref := &c.data[piece-1][to][captured-1]

	clampedBonus := Clamp(bonus, -MaxCaptures, MaxCaptures)
	*ref += clampedBonus - Score(int(*ref)*int(Abs(clampedBonus))/MaxCaptures)
}

// Probe returns the continuation history heuristics entry for the move.
func (c *CaptHist) Probe(piece Piece, to Square, captured Piece) Score {
	return c.data[piece-1][to][captured-1]
}
