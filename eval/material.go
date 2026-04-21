package eval

import (
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
)

const materialCacheSize = 2 * 1024

type (
	hash = board.Hash

	// This crashes the go compiler with go 1.26 being type alias.
	// https://github.com/golang/go/issues/78343
	// Replace with type alias once fixed in 1.27.
	evalFunc[T ScoreType] func(*Eval[T], *board.Board, *CoeffSet[T]) T
)

type MaterialCache[T ScoreType] struct {
	hash   hash
	evalID evalID
}

// material count dispatcher and cache.
func (e *Eval[T]) material(b *board.Board, c *CoeffSet[T]) T {

	wP := b.Counts[White][Pawn]
	wN := b.Counts[White][Knight]
	wB := b.Counts[White][Bishop]
	wR := b.Counts[White][Rook]
	wQ := b.Counts[White][Queen]
	bP := b.Counts[Black][Pawn]
	bN := b.Counts[Black][Knight]
	bB := b.Counts[Black][Bishop]
	bR := b.Counts[Black][Rook]
	bQ := b.Counts[Black][Queen]

	key := hash(0)

	wBishop := b.Colors[White] & b.Pieces[Bishop]
	bBishop := b.Colors[Black] & b.Pieces[Bishop]
	ocb := wB == 1 && bB == 1 && wBishop.LowestSet().Parity() != bBishop.LowestSet().Parity()
	if ocb {
		// index 63 is guaranteed to be unused as there can't be 63 bishops
		key ^= board.PiecesRand[White][Bishop][63]
	}

	// loop unrolled on hot path. ~1-2% NPS
	key ^= board.PiecesRand[White][Pawn][wP]
	key ^= board.PiecesRand[White][Knight][wN]
	key ^= board.PiecesRand[White][Bishop][wB]
	key ^= board.PiecesRand[White][Rook][wR]
	key ^= board.PiecesRand[White][Queen][wQ]
	key ^= board.PiecesRand[Black][Pawn][bP]
	key ^= board.PiecesRand[Black][Knight][bN]
	key ^= board.PiecesRand[Black][Bishop][bB]
	key ^= board.PiecesRand[Black][Rook][bR]
	key ^= board.PiecesRand[Black][Queen][bQ]

	entry := &e.materialCache[key%materialCacheSize]
	if entry.hash == key {
		return e.matFuncs[entry.evalID](e, b, c)
	}

	var evalID evalID
	switch {

	case wP == 0 && bP == 0 && wR == 0 && bR == 0 && wQ == 0 && bQ == 0 &&
		wN+bN+wB+bB <= 3 && max((wN+3*wB)-(bN+3*bB), (bN+3*bB)-(wN+3*wB)) <= 3:
		evalID = evalInsufficientID

	case wP == 0 && bP == 0 && wR == 0 && bR == 0 && wQ == 0 && bQ == 0 &&
		((wN == 1 && wB == 1 && bN == 0 && bB == 0) || (wN == 0 && wB == 0 && bN == 1 && bB == 1)):
		evalID = evalKNBvKID

	case ocb && wN == 0 && bN == 0 && wR == 0 && bR == 0 && wQ == 0 && bQ == 0 && pawnDiff(b) <= 3:
		evalID = evalOCBID

	case ocb && wN == 1 && bN == 1 && wR == 0 && bR == 0 && wQ == 0 && bQ == 0 && pawnDiff(b) <= 3:
		evalID = evalOCBKnightsID

	case ocb && wN == 0 && bN == 0 && wR == 1 && bR == 1 && wQ == 0 && bQ == 0 && pawnDiff(b) <= 3:
		evalID = evalOCBRooksID

	case wB == 0 && bB == 0 && wR == 0 && bR == 0 && wQ == 0 && bQ == 0 &&
		((wN == 1 && wP == 0 && bN == 0 && bP < 3) || (bN == 1 && bP == 0 && wN == 0 && wP < 3)):
		evalID = evalKNvKPID

	case wN == 0 && bN == 0 && wR == 0 && bR == 0 && wQ == 0 && bQ == 0 &&
		((wB == 1 && wP == 0 && bB == 0 && bP < 3) || (bB == 1 && bP == 0 && wB == 0 && wP < 3)):
		evalID = evalKBvKPID

	case wP == 0 && bP == 0 && wB == 0 && bB == 0 && wR == 1 && bR == 1 && wQ == 0 && bQ == 0 && wN+bN == 1:
		evalID = evalKRNvKRID

	case wP == 0 && bP == 0 && wN == 0 && bN == 0 && wR == 1 && bR == 1 && wQ == 0 && bQ == 0 && wB+bB == 1:
		evalID = evalKRBvKRID

	default:
		evalID = evalPositionalID
	}

	entry.hash = key
	entry.evalID = evalID

	return e.matFuncs[evalID](e, b, c)
}

type evalID byte

const (
	evalInsufficientID = evalID(iota)
	evalKNBvKID
	evalOCBID
	evalOCBKnightsID
	evalOCBRooksID
	evalKNvKPID
	evalKBvKPID
	evalKRNvKRID
	evalKRBvKRID
	evalPositionalID

	evalIDs
)

func evalInsufficient[T ScoreType](e *Eval[T], b *board.Board, c *CoeffSet[T]) T {
	return 0
}

func evalKNBvK[T ScoreType](e *Eval[T], b *board.Board, c *CoeffSet[T]) T {
	return e.knbvk(b, c)
}

func evalOCB[T ScoreType](e *Eval[T], b *board.Board, c *CoeffSet[T]) T {
	sf := c.OppositeColoredBishops[0][pawnDiff(b)]
	e.scaleFactor = [Colors]T{sf, sf}
	return e.positional(b, c)
}

func evalOCBKnights[T ScoreType](e *Eval[T], b *board.Board, c *CoeffSet[T]) T {
	sf := c.OppositeColoredBishops[1][pawnDiff(b)]
	e.scaleFactor = [Colors]T{sf, sf}
	return e.positional(b, c)
}

func evalOCBRooks[T ScoreType](e *Eval[T], b *board.Board, c *CoeffSet[T]) T {
	sf := c.OppositeColoredBishops[2][pawnDiff(b)]
	e.scaleFactor = [Colors]T{sf, sf}
	return e.positional(b, c)
}

func pawnDiff(b *board.Board) int {
	return int(Abs(b.Counts[White][Pawn] - b.Counts[Black][Pawn]))
}

func evalKNvKP[T ScoreType](e *Eval[T], b *board.Board, c *CoeffSet[T]) T {
	strongSide := Black
	if b.Counts[White][Knight] == 1 {
		strongSide = White
	}
	weakSide := strongSide.Flip()

	e.scaleFactor[strongSide] = c.InsufficientKnight
	e.scaleFactor[weakSide] = MaxScaleFactor

	return e.positional(b, c)
}

func evalKBvKP[T ScoreType](e *Eval[T], b *board.Board, c *CoeffSet[T]) T {
	strongSide := Black
	if b.Counts[White][Bishop] == 1 {
		strongSide = White
	}
	weakSide := strongSide.Flip()

	e.scaleFactor[strongSide] = c.InsufficientBishop
	e.scaleFactor[weakSide] = MaxScaleFactor

	return e.positional(b, c)
}

func evalKRNvKR[T ScoreType](e *Eval[T], b *board.Board, c *CoeffSet[T]) T {
	strongSide := Black
	if b.Counts[White][Knight] == 1 {
		strongSide = White
	}
	weakSide := strongSide.Flip()

	e.scaleFactor[strongSide] = c.KRNvKR
	e.scaleFactor[weakSide] = MaxScaleFactor

	return e.positional(b, c)
}

func evalKRBvKR[T ScoreType](e *Eval[T], b *board.Board, c *CoeffSet[T]) T {
	strongSide := Black
	if b.Counts[White][Bishop] == 1 {
		strongSide = White
	}
	weakSide := strongSide.Flip()

	e.scaleFactor[strongSide] = c.KRBvKR
	e.scaleFactor[weakSide] = MaxScaleFactor

	return e.positional(b, c)
}

func evalPositional[T ScoreType](e *Eval[T], b *board.Board, c *CoeffSet[T]) T {
	// drawishness
	fifty := int(100 - b.FiftyCnt)
	fifty *= fifty
	sf := T((fifty * MaxScaleFactor) / 10_000)
	e.scaleFactor = [Colors]T{sf, sf}
	return e.positional(b, c)
}
