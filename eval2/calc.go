package eval2

import (
	"github.com/paulsonkoly/chess-3/attacks"
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
)

func (e *Eval[T]) calcCounts(b *board.Board) {
	for color := range Colors {
		for pType := Pawn; pType <= Queen; pType++ {
			e.pieceCounts[color][pType] = (b.Colors[color] & b.Pieces[pType]).Count()
		}
	}
}

func (e *Eval[T]) calcAttacks(b *board.Board) {
	e.attacks[White][Pawn] = attacks.PawnCaptureMoves(b.Pieces[Pawn]&b.Colors[White], White)
	e.attacks[Black][Pawn] = attacks.PawnCaptureMoves(b.Pieces[Pawn]&b.Colors[Black], Black)
}

