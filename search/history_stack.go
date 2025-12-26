package search

import (
	"github.com/paulsonkoly/chess-3/heur"
	"github.com/paulsonkoly/chess-3/stack"
	. "github.com/paulsonkoly/chess-3/types"
)

// historyStack stores the current move piece and to square per ply along with
// its static eval. Useful for continuation history or improving heuristics.
type historyStack struct {
	stack.Stack[heur.StackMove]
}

// newHistoryStack Allocates a new history stack.
func newHistoryStack() *historyStack {
	return &historyStack{stack.New[heur.StackMove]()}
}

// oldScore is the static eval from previous plies relevant to improving.
func (h historyStack) oldScore() Score {
	top := h.Top(4)
	if len(top) >= 2 && top[len(top)-2].Score != Inv {
		return top[len(top)-2].Score
	} else if len(top) >= 4 {
		return top[len(top)-4].Score
	}
	return Inv
}
