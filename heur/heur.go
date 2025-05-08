package heur

import (
	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

// PieceValues approximates the value of each piece type for SEE and heuristic
// purposes.
var PieceValues = [...]Score{0, 100, 300, 300, 500, 900, Inf}

// | Score Region     | Contents        | Details                                                      |
// | ---------------- | --------------- | ------------------------------------------------------------ |
// | **< -6k**        | *Bad captures*  | Bucketed by captured piece (P..Q), each in \[-1k, +1k] range |
// | **≈ -6k to +6k** | *Quiets*        | `hist + 2·cont[1] + 3·cont[0]`, each term clamped to ±1k     |
// | **> +6k**        | *Good captures* | Same 6 piece buckets as above, based on captured piece type  |
// | **Top priority** | *HashMove*      | Always on top (or near top)                                  |

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
