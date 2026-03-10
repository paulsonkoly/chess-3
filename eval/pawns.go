package eval

import (
	"github.com/paulsonkoly/chess-3/attacks"
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

func (p *pawns) fromBoard(b *board.Board) {
	ps := [...]BitBoard{b.Pieces[Pawn] & b.Colors[White], b.Pieces[Pawn] & b.Colors[Black]}
	p.pawns = ps

	p.frontSpan = [...]BitBoard{frontFill(ps[White], White) << 8, frontFill(ps[Black], Black) >> 8}
	rearSpan := [...]BitBoard{frontFill(ps[White], Black) >> 8, frontFill(ps[Black], White) << 8}

	wFiles := ps[White] | p.frontSpan[White] | rearSpan[White]
	bFiles := ps[Black] | p.frontSpan[Black] | rearSpan[Black]
	p.neighbourF = [...]BitBoard{
		((wFiles & ^AFileBB) >> 1) | ((wFiles & ^HFileBB) << 1),
		((bFiles & ^HFileBB) << 1) | ((bFiles & ^AFileBB) >> 1),
	}

	p.frontLine = [...]BitBoard{^rearSpan[White] & ps[White], ^rearSpan[Black] & ps[Black]}

	p.cover = [...]BitBoard{
		((p.frontSpan[White] & ^AFileBB) >> 1) | ((p.frontSpan[White] & ^HFileBB) << 1),
		((p.frontSpan[Black] & ^HFileBB) << 1) | ((p.frontSpan[Black] & ^AFileBB) >> 1),
	}
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
	return (p.pawns[c]) &^ p.frontLine[c]
}

// isolatedPawns are pawns not having any friendly pawn on adjacent files.
func (p *pawns) isolatedPawns(c Color) BitBoard {
	return (p.pawns[c]) &^ p.neighbourF[c]
}

// backwardsPawns are pawns not supported by any friendly pawns, while their
// push is prevented by either enemy pawn capture or blockade.
func (p *pawns) backwardPawns(c Color) BitBoard {
	threatened := attacks.PawnCaptureMoves(p.pawns[c.Flip()], c.Flip()) & ^p.cover[c]
	stopped := attacks.PawnSinglePushMoves(p.pawns[c.Flip()]|threatened, c.Flip())
	return p.pawns[c] & ^p.cover[c] & stopped
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
