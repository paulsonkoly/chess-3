package heur

import (
	. "github.com/paulsonkoly/chess-3/chess"
)

type ContinuationType byte

const (
	FollowUp = ContinuationType(iota)
	Counter

	Continuations
)

// Continuation is the heuristics table indexed with color, old move piece type
// & to square, new move piece type and to square.
type Continuation struct {
	data [Continuations][Colors][6][Squares][6][Squares]Score
}

// NewContinuation creates a continuation history table.
func NewContinuation() *Continuation {
	return &Continuation{}
}

// Clear clears the continuation history table.
func (c *Continuation) Clear() {
	c.data = [Continuations][Colors][6][Squares][6][Squares]Score{}
}

// Add increments the continuation history heuristics for the move by bonus.
func (c *Continuation) Add(t ContinuationType, stm Color, ptHist Piece, toHist Square, pt Piece, to Square, bonus Score) {
	entry := &c.data[t][stm][ptHist-1][toHist][pt-1][to]

	clampedBonus := Clamp(bonus, -MaxHistory, MaxHistory)
	*entry += clampedBonus - Score(int(*entry)*int(Abs(clampedBonus))/int(MaxHistory))
}

// LookUp returns the continuation history heuristics entry for the move.
func (c *Continuation) LookUp(t ContinuationType, stm Color, ptHist Piece, toHist Square, pt Piece, to Square) Score {
	return c.data[t][stm][ptHist-1][toHist][pt-1][to]
}
