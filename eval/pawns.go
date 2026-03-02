package eval

import (
	"github.com/paulsonkoly/chess-3/board"

	. "github.com/paulsonkoly/chess-3/chess"
)

// the player's side of the board with the extra 2 central squares included at
// enemy side.
var sideOfBoard = [2]BitBoard{0x00000018_ffffffff, 0xffffffff_18000000}

type pawns struct {
	pawns      [2]BitBoard
	cover      [2]BitBoard
	frontLine  [2]BitBoard
	frontSpan  [2]BitBoard
	neighbourF [2]BitBoard // neighbourF is files adjacent to files with pawns
}

func calcPawns(b *board.Board) *pawns {
	pawns := pawns{}

	ps := [...]BitBoard{b.Pieces[Pawn] & b.Colors[White], b.Pieces[Pawn] & b.Colors[Black]}
	pawns.pawns = ps

	pawns.frontSpan = [...]BitBoard{frontFill(ps[White], White) << 8, frontFill(ps[Black], Black) >> 8}
	rearSpan := [...]BitBoard{frontFill(ps[White], Black) >> 8, frontFill(ps[Black], White) << 8}

	wFiles := ps[White] | pawns.frontSpan[White] | rearSpan[White]
	bFiles := ps[Black] | pawns.frontSpan[Black] | rearSpan[Black]
	pawns.neighbourF = [...]BitBoard{
		((wFiles & ^AFileBB) >> 1) | ((wFiles & ^HFileBB) << 1),
		((bFiles & ^HFileBB) << 1) | ((bFiles & ^AFileBB) >> 1),
	}

	pawns.frontLine = [...]BitBoard{^rearSpan[White] & ps[White], ^rearSpan[Black] & ps[Black]}

	pawns.cover = [...]BitBoard{
		((pawns.frontSpan[White] & ^AFileBB) >> 1) | ((pawns.frontSpan[White] & ^HFileBB) << 1),
		((pawns.frontSpan[Black] & ^HFileBB) << 1) | ((pawns.frontSpan[Black] & ^AFileBB) >> 1),
	}

	return &pawns
}

// holes are squares that cannot be protected by one of our pawns on our side of the board.
func (p *pawns) holes(c Color) BitBoard {
	return sideOfBoard[c] &^ p.cover[c]
}

// passers are pawns not stoppable by enemy pawns without them changing file.
func (p *pawns) passers(c Color) BitBoard {
	return p.frontLine[c] & ^(p.frontSpan[c.Flip()] | (p.cover[c.Flip()]))
}

// doubledPawns are pawns that have a friendly further advanced pawn on the same file.
func (p *pawns) doubledPawns(c Color) BitBoard {
	return p.pawns[c] &^ p.frontLine[c]
}

// isolatedPawns are pawns not having any friendly pawn on adjacent files.
func (p *pawns) isolatedPawns(c Color) BitBoard {
	return p.pawns[c] &^ p.neighbourF[c]
}

// blockadedPawns are pawns that have an enemy pawn in front of them but not necessarily adjacent.
func (p *pawns) blockaded(c Color) BitBoard {
	return p.pawns[c] & p.frontSpan[c.Flip()]
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
