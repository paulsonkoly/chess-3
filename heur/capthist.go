package heur

import (
	. "github.com/paulsonkoly/chess-3/chess"
)

// CaptHist is capture history heuristics store.
type CaptHist struct {
	data [6][5][Squares]Score
}

// NewCaptHist creates a new history heuristics.
func NewCaptHist() *CaptHist {
	return &CaptHist{}
}

// Clear resets all entries to 0.
func (c *CaptHist) Clear() {
	c.data = [6][5][Squares]Score{}
}

// Add increments the capture history heuristics for the move by bonus.
// moved should be at least a Pawn and captured should be between a Pawn and a
// Queen, otherwise this function can panic.
func (c *CaptHist) Add(moved, captured Piece, sq Square, bonus Score) {
	clampedBonus := Clamp(bonus, -MaxHistory, MaxHistory)
	moved -= Pawn // offset range ignoring NoPiece
	captured -= Pawn
	c.data[moved][captured][sq] +=
		clampedBonus - Score(int(c.data[moved][captured][sq])*int(Abs(clampedBonus))/int(MaxHistory))
}

// LookUp returns the history heuristics entry for the move.
// moved should be at least a Pawn and captured should be between a Pawn and a
// Queen, otherwise this function can panic.
func (c *CaptHist) LookUp(moved, captured Piece, sq Square) Score {
	moved -= Pawn // offset range ignoring NoPiece
	captured -= Pawn
	return c.data[moved][captured][sq]
}
