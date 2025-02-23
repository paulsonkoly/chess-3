package heur

import (
	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

// PieceValues approximates the value of each piece type for SEE and heuristic
// purposes.
var PieceValues = [...]Score{0, 100, 300, 300, 500, 900, Inf}

//  Move ordering stages
//
//  | hash | .. Good captures .. | History | Quiet | .. unused .. | Bad captures |
//     ^                         ^         ^     ^ ^              ^
//    HashMove            Captures  QuietHistory | |          -Captures
//                                               0 |
//                                             -QuietHistory
const (
  // HashMove is assigned to a move from Hash, either PV or fail-high
  HashMove = Score(15000)
  // Captures is the minimal score for captures, actual score is thisplus SEE.
  Captures = Score(2048)
  // QuietHistory is the minimal score for moves hitting the history heuristic (fail-high quiet).
  QuietHistory = Score(1024)
)
