package search

import (
	. "github.com/paulsonkoly/chess-3/chess"
)

type historyMove struct {
	piece Piece
	to    Square
	score Score
}

type historyStack struct {
	data [MaxPlies]historyMove
	sp   int
}

func newHistStack() *historyStack {
	return &historyStack{}
}

func (h *historyStack) reset() {
	h.sp = 0
}

func (h *historyStack) push(piece Piece, to Square, score Score) {
	h.data[h.sp] = historyMove{piece: piece, to: to, score: score}
	h.sp++
}

func (h *historyStack) pop() {
	h.sp--
}

func (h *historyStack) top(n int) historyMove {
	return h.data[h.sp-n-1]
}

func (h *historyStack) oldScore() Score {
	if h.sp >= 2 && h.data[h.sp-2].score != Inv {
		return h.data[h.sp-2].score
	} else if h.sp >= 4 {
		return h.data[h.sp-4].score
	}
	return Inv
}

func (h *historyStack) size() int {
	return h.sp
}
