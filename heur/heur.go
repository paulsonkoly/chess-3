package heur

import (
	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

// PieceValues approximates the value of each piece type for SEE and heuristic
// purposes.
var PieceValues = [...]Score{0, 100, 300, 300, 500, 900, Inf}

const (
	// Captures is the minimal score for captures.
	Captures = 6 * MaxHistory
	// HashMove is assigned to a move from Hash, either PV or fail-high.
	HashMove = Captures + 12 * MaxCaptHist + 100
)

const (
	MaxHistory  = Score(1024)
	MaxCaptHist = Score(1024)
)
