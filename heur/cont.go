package heur

import (
	. "github.com/paulsonkoly/chess-3/chess"
)

// Continuation is the heuristics table.
//
// Indexed by oldSTM, oldPiece, oldTo, currSTM, currPiece, currTo where old
// refers to a move that happened some plies ago, curr refers to the current
// move.
type Continuation struct {
	data [Colors][6][Squares][Colors][6][Squares]Score
}

// NewContinuation creates a continuation history table.
func NewContinuation() *Continuation {
	return &Continuation{}
}

// Clear clears the continuation history table.
func (c *Continuation) Clear() {
	c.data = [Colors][6][Squares][Colors][6][Squares]Score{}
}

// Add increments the continuation history heuristics for the move by bonus.
func (c *Continuation) Add(
	oldSTM Color,
	oldPiece Piece,
	oldTo Square,
	currSTM Color,
	currPiece Piece,
	currTo Square,
	bonus Score,
) {
	entry := &c.data[oldSTM][oldPiece-1][oldTo][currSTM][currPiece-1][currTo]

	clampedBonus := Clamp(bonus, -MaxHistory, MaxHistory)
	*entry += clampedBonus - Score(int(*entry)*int(Abs(clampedBonus))/int(MaxHistory))
}

// LookUp returns the continuation history heuristics entry for the move.
func (c *Continuation) LookUp(
	oldSTM Color,
	oldPiece Piece,
	oldTo Square,
	currSTM Color,
	currPiece Piece,
	currTo Square,
) Score {
	return c.data[oldSTM][oldPiece-1][oldTo][currSTM][currPiece-1][currTo]
}
