package search

import (
	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

const MaxPlies = 64

type historyMove struct {
	piece Piece
	to    Square
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

func (h *historyStack) push(piece Piece, to Square) {
	h.data[h.sp] = historyMove{piece: piece, to: to}
	h.sp++
}

func (h *historyStack) pop() {
	h.sp--
}

func (h *historyStack) top(n int) (Piece, Square) {
	d := h.data[h.sp-n-1]
	return d.piece, d.to
}

func (h *historyStack) size() int {
	return h.sp
}
