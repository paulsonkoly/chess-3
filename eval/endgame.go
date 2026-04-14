package eval

import (
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
)

func insufficient(b *board.Board) bool {
	if b.Pieces[Pawn]|b.Pieces[Queen]|b.Pieces[Rook] != 0 {
		return false
	}

	wN := b.Counts[White][Knight]
	bN := b.Counts[Black][Knight]
	wB := b.Counts[White][Bishop]
	bB := b.Counts[Black][Bishop]

	if wN+bN+wB+bB <= 3 { // draw cases
		wScr := wN + 3*wB
		bScr := bN + 3*bB

		if max(wScr-bScr, bScr-wScr) <= 3 {
			return true
		}
	}

	return false
}

var traditionalPieceValues = [...]int{0, 1, 3, 3, 5, 9, 0}

func knvkp(b *board.Board) bool {
	if b.Pieces[Bishop]|b.Pieces[Rook]|b.Pieces[Queen] != 0 {
		return false
	}

	var mat [Colors]int

	for color := range Colors {
		for pType := Pawn; pType <= Knight; pType++ {
			mat[color] += traditionalPieceValues[pType] * int(b.Counts[color][pType])
		}
	}

	switch {
	case mat[White] < mat[Black]:
		return b.Counts[Black][Knight] == 1 && b.Counts[Black][Pawn] == 0
	case mat[Black] < mat[White]:
		return b.Counts[White][Knight] == 1 && b.Counts[White][Pawn] == 0
	default:
		return false
	}
}

func knbvk(b *board.Board) bool {
	whiteN := b.Pieces[Knight] & b.Colors[White]
	blackN := b.Pieces[Knight] & b.Colors[Black]
	whiteB := b.Pieces[Bishop] & b.Colors[White]
	blackB := b.Pieces[Bishop] & b.Colors[Black]

	return b.Pieces[Pawn]|b.Pieces[Rook]|b.Pieces[Queen] == 0 &&
		((whiteN.IsPow2() && whiteB.IsPow2() && (blackN|blackB) == 0) ||
			(blackN.IsPow2() && blackB.IsPow2() && (whiteN|whiteB) == 0))
}

// KBCorners are knight-bishop checkmate corners based on parity of square.
var KBCorners = [2][2]Square{{A1, H8}, {H1, A8}}

func (e *Eval[T]) knbvk(b *board.Board, c *CoeffSet[T]) T {
	bishopSq := b.Pieces[Bishop].LowestSet()
	knightSq := b.Pieces[Knight].LowestSet()

	victim := White
	if b.Pieces[Bishop]&b.Colors[White] != 0 {
		victim = Black
	}
	victimKSq := (b.Pieces[King] & b.Colors[victim]).LowestSet()
	attackKSq := (b.Pieces[King] & b.Colors[victim.Flip()]).LowestSet()

	e.addPSqT(victim, King, victimKSq, c)
	e.addPSqT(victim.Flip(), King, attackKSq, c)
	e.addPSqT(victim.Flip(), Knight, knightSq, c)
	e.addPSqT(victim.Flip(), Bishop, bishopSq, c)
	e.sp[victim.Flip()][EG] += c.PieceValues[EG][Knight]
	e.sp[victim.Flip()][EG] += c.PieceValues[EG][Bishop]

	parity := (bishopSq.File() + bishopSq.Rank()) & 1

	cornerDist := min(Chebishev(victimKSq, KBCorners[parity][0]), Chebishev(victimKSq, KBCorners[parity][1]))
	cornerDist = 7 - cornerDist
	cornerDist *= cornerDist

	e.sp[victim.Flip()][EG] += T(cornerDist) * 30

	return e.endgameScore(b)
}
