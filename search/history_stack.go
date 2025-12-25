package search

import (
	"github.com/paulsonkoly/chess-3/heur"
	"github.com/paulsonkoly/chess-3/stack"
	. "github.com/paulsonkoly/chess-3/types"
)

type historyStack struct {
	stack.Stack[heur.StackMove]
}

func newHistoryStack() *historyStack {
	return &historyStack{stack.New[heur.StackMove]()}
}

func (h historyStack) oldScore() Score {
	top := h.Top(4)
	if len(top) >= 2 && top[len(top)-2].Score != Inv {
		return top[len(top)-2].Score
	} else if len(top) >= 4 {
		return top[len(top)-4].Score
	}
	return Inv
}
