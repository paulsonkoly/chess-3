package search

import (
	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

const MaxPlies = 64

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

func (h *historyStack) size() int {
	return h.sp
}
