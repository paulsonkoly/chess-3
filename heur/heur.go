// Package heur provides move ordering heuristics.
//
// Move ordering stages
//
// - hash: [HashMove]
// - good captures: [Captures] ... [HashMove]
// - quiets: -6*[MaxHistory]..6*[MaxHistory]
//   (3 * cont[0] + 2 * cont[1] + hist) each <= [MaxHistory]
// - bad captures: -Inf..-[Captures]
//
package heur

import (
	. "github.com/paulsonkoly/chess-3/types"
)

// PieceValues approximates the value of each piece type for SEE and heuristic
// purposes.
var PieceValues = [...]Score{0, 100, 300, 300, 500, 900, Inf}

const (
	// HashMove is assigned to a move from Hash, either PV or fail-high.
	HashMove = Score(15000)
	// Captures is the minimal score for captures, actual score is this plus SEE.
	Captures = Score(8192)
	// MaxHistory is the maximal absolute value in either the history or the continuation stores. 
	MaxHistory = Score(1024)
)

func init() {
	// 1 * history + 2 * continuation[1] + 3 * continuation[0]
	if Captures < 6*MaxHistory {
		panic("gap is not big enough in move weight layout for history scores")
	}
}
