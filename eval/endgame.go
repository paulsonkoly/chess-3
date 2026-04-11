package eval

import (
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
)

const MaxScaleFactor = 128

func insufficient(b *board.Board) bool {
	if b.Pieces[Pawn]|b.Pieces[Queen]|b.Pieces[Rook] != 0 {
		return false
	}

	wN := (b.Colors[White] & b.Pieces[Knight]).Count()
	bN := (b.Colors[Black] & b.Pieces[Knight]).Count()
	wB := (b.Colors[White] & b.Pieces[Bishop]).Count()
	bB := (b.Colors[Black] & b.Pieces[Bishop]).Count()

	if wN+bN+wB+bB <= 3 { // draw cases
		wScr := wN + 3*wB
		bScr := bN + 3*bB

		if max(wScr-bScr, bScr-wScr) <= 3 {
			return true
		}
	}

	return false
}

func isKNBvK(b *board.Board) bool {
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

func (e *Eval[T]) knbvk(b *board.Board, c *CoeffSet[T]) {
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
}

func (e *Eval[T]) scaleOCB(b *board.Board, c *CoeffSet[T]) bool {
	wBishop := b.Colors[White] & b.Pieces[Bishop]
	bBishop := b.Colors[Black] & b.Pieces[Bishop]

	if !wBishop.IsPow2() || !bBishop.IsPow2() || wBishop.LowestSet().Parity() == bBishop.LowestSet().Parity() {
		return false
	}

	pawnDiff := Abs((b.Colors[White] & b.Pieces[Pawn]).Count() - (b.Colors[Black] & b.Pieces[Pawn]).Count())
	pawnDiff = Clamp(pawnDiff, 0, 3)

	knights, rooks, queens := b.Pieces[Knight], b.Pieces[Rook], b.Pieces[Queen]
	others := knights | rooks | queens
	if others == 0 {
		e.scaleFactor = c.OppositeColoredBishops[0][pawnDiff]
		return true
	}

	wN, bN := b.Colors[White]&knights, b.Colors[Black]&knights
	if wN.IsPow2() && bN.IsPow2() && rooks|queens == 0 {
		e.scaleFactor = c.OppositeColoredBishops[1][pawnDiff]
		return true
	}

	wR, bR := b.Colors[White]&rooks, b.Colors[Black]&rooks
	if wR.IsPow2() && bR.IsPow2() && knights|queens == 0 {
		e.scaleFactor = c.OppositeColoredBishops[2][pawnDiff]
		return true
	}

	wQ, bQ := b.Colors[White]&queens, b.Colors[Black]&queens
	if wQ.IsPow2() && bQ.IsPow2() && knights|rooks == 0 {
		e.scaleFactor = c.OppositeColoredBishops[3][pawnDiff]
		return true
	}

	return false
}

func (e *Eval[T]) scaleFifty(b *board.Board) bool {
	fifty := int(100 - b.FiftyCnt)
	fifty *= fifty * MaxScaleFactor
	e.scaleFactor = T(fifty / 10_000)
	return true
}
