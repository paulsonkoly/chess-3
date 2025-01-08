package eval

import (
	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/movegen"
	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

const Inf = 100000

func Eval(b *board.Board) int {
	hasLegal := false
	for m := range movegen.Moves(b, board.Full) {
		b.MakeMove(&m)

		king := b.Colors[b.STM.Flip()] & b.Pieces[King]
		hasLegal = hasLegal || !movegen.IsAttacked(b, b.STM, king)
		b.UndoMove(&m)

		if hasLegal {
			break
		}
	}

	if !hasLegal {
		king := b.Colors[b.STM] & b.Pieces[King]
		if movegen.IsAttacked(b, b.STM.Flip(), king) {
			if b.STM == White {
				return -Inf
			}
			return Inf
		}
		return 0
	}

	return eval(b, White) - eval(b, Black)
}

func eval(b *board.Board, color Color) int {
	return (b.Colors[color]&b.Pieces[Pawn]).Count()*100 +
		(b.Colors[color]&b.Pieces[Knight]).Count()*300 +
		(b.Colors[color]&b.Pieces[Bishop]).Count()*300 +
		(b.Colors[color]&b.Pieces[Rook]).Count()*500 +
		(b.Colors[color]&b.Pieces[Queen]).Count()*900
}
