package hist

import (
	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

type Move struct {
	Piece Piece
	To    Square
	Score Score
}

type Stack struct {
	data [MaxPlies]Move
	sp   int
}

func NewStack() *Stack {
	return &Stack{}
}

func (h *Stack) Reset() {
	h.sp = 0
}

func (h *Stack) Push(piece Piece, to Square, score Score) {
	h.data[h.sp] = Move{Piece: piece, To: to, Score: score}
	h.sp++
}

func (h *Stack) Pop() {
	h.sp--
}

func (h *Stack) Top(n int) Move {
	return h.data[h.sp-n-1]
}

func (h *Stack) OldScore() Score {
	if h.sp >= 2 && h.data[h.sp-2].Score != Inv {
		return h.data[h.sp-2].Score
	} else if h.sp >= 4 {
		return h.data[h.sp-4].Score
	}
	return Inv
}

func (h *Stack) Size() int {
	return h.sp
}
