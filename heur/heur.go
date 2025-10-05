package heur

import (
	. "github.com/paulsonkoly/chess-3/types"
)

// PieceValues approximates the value of each piece type for SEE and heuristic
// purposes.
var PieceValues = [...]Score{0, 100, 300, 300, 500, 900, Inf}

// Move ordering stages
//
// | hash | .. Good captures .. | .. Continuation .. | .. History .. | .. Quiet .. | .. unused .. | .. Bad captures .. |
//
//	 ^                         ^                                    ^       ^                      ^
//	HashMove            Captures                                Quiet       0                      -Captures
const (
	// HashMove is assigned to a move from Hash, either PV or fail-high.
	HashMove = Score(15000)
	// Captures is the minimal score for captures, actual score is this plus SEE.
	Captures = Score(8192)
)

const MaxHistory = 1024

func init() {
	// 1 * history + 2 * continuation[1] + 3 * continuation[0]
	if Captures < 6*MaxHistory {
		panic("gap is not big enough in move weight layout for history scores")
	}
}
