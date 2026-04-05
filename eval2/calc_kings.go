package eval2

import (
	"github.com/paulsonkoly/chess-3/attacks"
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
)

func (k *Kings) calc(b *board.Board, color Color) {
	king := b.Colors[color] & b.Pieces[King]
	k.sq = king.LowestSet()
	k.nb = attacks.KingMoves(k.sq) | king
}
