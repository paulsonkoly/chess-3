package eval

import (
	"github.com/paulsonkoly/chess-3/attacks"
	"github.com/paulsonkoly/chess-3/board"

	. "github.com/paulsonkoly/chess-3/chess"
)

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

// outposts are squares in our 5the, 6th or 7th rank that cannot be defended by
// any of the enemy's pawns and simultaneously attacked by one of our pawns.
func (p *pawns) outposts(c Color) BitBoard {
	var territory BitBoard
	if c == White {
		territory = FifthRankBB | SixthRankBB | SeventhRankBB
	} else {
		territory = SecondRankBB | ThirdRankBB | FourthRankBB
	}
	attacked := attacks.PawnCaptureMoves(p.pawns[c], c)
	return territory & attacked & ^p.cover[c.Flip()]
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
