package eval2

import (
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
)

func (e *Eval[T]) insufficient(b *board.Board) bool {
	if b.Pieces[Pawn]|b.Pieces[Queen]|b.Pieces[Rook] != 0 {
		return false
	}

	wN := e.pieceCounts[White][Knight]
	bN := e.pieceCounts[Black][Knight]
	wB := e.pieceCounts[White][Bishop]
	bB := e.pieceCounts[Black][Bishop]

	if wN+bN+wB+bB <= 3 { // draw cases
		wScr := wN + 3*wB
		bScr := bN + 3*bB

		if max(wScr-bScr, bScr-wScr) <= 3 {
			return true
		}
	}

	return false
}
