package heur

import (
	. "github.com/paulsonkoly/chess-3/types"
)

// Continuation is the heuristics table indexed with color, old move piece type
// & to square, new move piece type and to square.
type Continuation struct {
	data [Colors][6][Squares][6][Squares]Score
}

// NewContinuation creates a continuation history table.
func NewContinuation() *Continuation {
	return &Continuation{}
}

// Clear clears the continuation history table.
func (c *Continuation) Clear() {
	c.data = [Colors][6][Squares][6][Squares]Score{}
}

// Add increments the continuation history heuristics for the move by bonus.
func (c *Continuation) Add(stm Color, ptHist Piece, toHist Square, pt Piece, to Square, bonus Score) {
	entry := &c.data[stm][ptHist-1][toHist][pt-1][to]

	clampedBonus := Clamp(bonus, -MaxHistory, MaxHistory)
	*entry += clampedBonus - Score(int(*entry)*int(Abs(clampedBonus))/int(MaxHistory))
}

// LookUp returns the continuation history heuristics entry for the move.
func (c *Continuation) LookUp(stm Color, ptHist Piece, toHist Square, pt Piece, to Square) Score {
	return c.data[stm][ptHist-1][toHist][pt-1][to]
}
