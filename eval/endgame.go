package eval

import (
	"github.com/paulsonkoly/chess-3/attacks"
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/eval/kpvk"
)

// KBCorners are knight-bishop checkmate corners based on parity of square.
var KBCorners = [2][2]Square{{A1, H8}, {H1, A8}}

func (e *Eval[T]) kbnvk(b *board.Board) T {
	e.sp = [Colors][Phases]T{}

	const (
		kingDistFactor   = 5
		knightDistFactor = 3
		cornerDistFactor = 60
		// base score to ensure we transition into this unless the position was more winning.
		// (don't give up a queen or rook just to simplify into this mate).
		base = 750
	)

	bishopSq := b.Pieces[Bishop].LowestSet()
	knightSq := b.Pieces[Knight].LowestSet()

	weakSide := White
	if b.Pieces[Bishop]&b.Colors[White] != 0 {
		weakSide = Black
	}
	strongSide := weakSide.Flip()
	weakKSq := (b.Pieces[King] & b.Colors[weakSide]).LowestSet()
	strongKSq := (b.Pieces[King] & b.Colors[strongSide]).LowestSet()

	// strong king far from weak king
	e.sp[weakSide][EG] += T(Chebyshev(weakKSq, strongKSq)) * kingDistFactor

	// knight far from weak king
	e.sp[weakSide][EG] += T(Chebyshev(weakKSq, knightSq)) * knightDistFactor

	parity := bishopSq.Parity()
	cornerDist := min(Chebyshev(weakKSq, KBCorners[parity][0]), Chebyshev(weakKSq, KBCorners[parity][1]))

	cornerDist = 7 - cornerDist
	cornerDist *= cornerDist

	e.sp[strongSide][EG] += base + T(cornerDist)*cornerDistFactor

	return e.endgameScore(b)
}

func (e *Eval[T]) kpvk(b *board.Board, c *CoeffSet[T]) T {
	if !kpvk.Winning(b) {
		return 0
	}
	pawn := b.Pieces[Pawn]
	strongSide := White
	if b.Colors[Black]&pawn != 0 {
		strongSide = Black
	}
	weakSide := strongSide.Flip()

	e.sp = [Colors][2]T{}

	pawnSq := pawn.LowestSet()

	// we incrementally increase the score based on pawn rank, encouraging pawn
	// advancement, the score should be kept below queen value so we queen at the
	// end, but large enough so we simplify into this if it's a winning endgame.
	//
	// we purposefully leave king distances out, because if we are here we know
	// the king can support the pawn. (this might need reconsidering in the
	// future).
	pawnRank := pawnSq.Rank().FromPerspectiveOf(weakSide)
	e.sp[strongSide][EG] = c.PieceValues[EG][Queen] - T(pawnRank)*15

	return e.endgameScore(b)
}

// kbpvk is the bishop+pawn endgame where a rook pawn draws if the opponent's
// king can reach the queening square and the bishop is on the opposite color
// complex.
// There is no way to give exact answer without table bases, so this code is a
// mere estimation.
func (e *Eval[T]) kbpvk(b *board.Board, c *CoeffSet[T]) T {
	// if not a Rook-pawn then normal positional eval
	e.scaleFactor = [2]T{MaxScaleFactor, MaxScaleFactor}
	var flank Coord
	switch {

	case b.Pieces[Pawn]&^AFileBB == 0:
		flank = 0

	case b.Pieces[Pawn]&^HFileBB == 0:
		flank = 1

	default:
		return e.positional(b, c)
	}

	strongSide := Black
	if b.Counts[White][Bishop] == 1 {
		strongSide = White
	}
	weakSide := strongSide.Flip()

	queeningSq := KBCorners[flank^Coord(weakSide)][weakSide]
	bishopSq := b.Pieces[Bishop].LowestSet()

	// if bishop controls queening square then normal positional eval
	if bishopSq.Parity() == queeningSq.Parity() {
		return e.positional(b, c)
	}

	pawnSq := b.Pieces[Pawn].LowestSet()
	weakK := b.Colors[weakSide] & b.Pieces[King]
	weakKSq := weakK.LowestSet()
	strongK := b.Colors[strongSide] & b.Pieces[King]
	strongKSq := strongK.LowestSet()

	if b.STM == weakSide &&
		attacks.KingMoves(weakKSq)&b.Pieces[Pawn] != 0 && // pawn is attacked
		(attacks.KingMoves(strongKSq)|attacks.BishopMoves(bishopSq, strongK))&b.Pieces[Pawn] == 0 { // not defended
		return 0
	}

	weakDist := Chebyshev(queeningSq, weakKSq)
	if weakDist <= 1 {
		return 0
	}
	if b.STM == strongSide {
		weakDist++
	}

	pawnDist := Chebyshev(pawnSq, queeningSq)
	// double pushes
	pawnDist = min(pawnDist, 5)

	// if the pawn can be pushed to the queening square without any support
	if weakDist-pawnDist >= 2 {
		return e.positional(b, c)
	}

	strongDist := Chebyshev(queeningSq, strongKSq)
	if strongDist <= 1 {
		return e.positional(b, c)
	}

	margin := weakDist - strongDist
	var scale T
	switch {
	case margin <= 0:
		scale = 0
	case margin == 1:
		scale = MaxScaleFactor * 7 / 10
	case margin == 2:
		scale = MaxScaleFactor * 85 / 100
	default:
		scale = MaxScaleFactor
	}
	e.scaleFactor[strongSide] = scale
	e.scaleFactor[weakSide] = 0 // weak side cannot win

	return e.positional(b, c)
}
