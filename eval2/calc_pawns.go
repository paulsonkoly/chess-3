package eval2

import (
	"github.com/paulsonkoly/chess-3/attacks"
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
)

func (p *Pawns) calc(b *board.Board, color Color) {
	pawns := b.Colors[color] & b.Pieces[Pawn]

	p.frontspan = attacks.PawnSinglePushMoves(frontFill(pawns, color), color)
	rearSpan := attacks.PawnSinglePushMoves(frontFill(pawns, color.Flip()), color.Flip())

	files := pawns | p.frontspan | rearSpan
	p.neighbourF = ((files & ^AFileBB) >> 1) | ((files & ^HFileBB) << 1)

	p.frontline = ^rearSpan & pawns
	p.backmost = ^p.frontspan & pawns
	p.cover = attacks.PawnCaptureMoves(p.frontspan, color)
}

func frontFill(b BitBoard, color Color) BitBoard {
	switch color {
	case White:
		b |= b << 8
		b |= b << 16
		b |= b << 32

	case Black:
		b |= b >> 8
		b |= b >> 16
		b |= b >> 32
	}

	return b
}
