package heur

import (
	. "github.com/paulsonkoly/chess-3/chess"
)

const contStoreSize = int(Colors) * 6 * int(Squares) * int(Colors) * 6 * int(Squares)

// Continuation is the continuation history heuristics table.
//
// Indexed by oldSTM, oldPiece, oldTo, currSTM, currPiece, currTo where old
// refers to a move that happened some plies ago, curr refers to the current
// move.
type Continuation struct {
	data [contStoreSize]Score
}

// NewContinuation creates a continuation history table.
func NewContinuation() *Continuation {
	return &Continuation{}
}

// Clear clears the continuation history table.
func (c *Continuation) Clear() {
	c.data = [contStoreSize]Score{}
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
	entry := &c.data[ix(oldSTM, oldPiece, oldTo, currSTM, currPiece, currTo)]

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
	return c.data[ix(oldSTM, oldPiece, oldTo, currSTM, currPiece, currTo)]
}

func ix(oldSTM Color, oldPiece Piece, oldTo Square, currSTM Color, currPiece Piece, currTo Square) int {
	const (
		X0 = 6 * int(Colors) * int(Squares) * int(Colors) * int(Squares)
		X1 = int(Colors) * int(Squares) * int(Colors) * int(Squares)
		X2 = int(Squares) * int(Colors) * int(Squares)
		X3 = int(Colors) * int(Squares)
		X4 = int(Squares)
	)
	return int(oldPiece-1)*X0 + int(currPiece-1)*X1 + int(oldSTM)*X2 + int(oldTo)*X3 + int(currSTM)*X4 + int(currTo)
}
