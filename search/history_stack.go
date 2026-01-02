package search

import (
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/heur"
)

type historyStack struct {
	data [MaxPlies]heur.StackMove
	sp   int
}

func newHistStack() *historyStack {
	return &historyStack{}
}

func (h *historyStack) reset() {
	h.sp = 0
}

func (h *historyStack) push(piece Piece, to Square, score Score) {
	h.data[h.sp] = heur.StackMove{Piece: piece, To: to, Score: score}
	h.sp++
}

func (h *historyStack) pop() {
	h.sp--
}

func (h *historyStack) top(n int) []heur.StackMove {
	return h.data[max(0, h.sp-n-1):h.sp]
}

func (h *historyStack) oldScore() Score {
	if h.sp >= 2 && h.data[h.sp-2].Score != Inv {
		return h.data[h.sp-2].Score
	} else if h.sp >= 4 {
		return h.data[h.sp-4].Score
	}
	return Inv
}
