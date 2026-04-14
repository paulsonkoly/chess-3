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
	key := hash(0)
	// loop unrolled on hot path. ~1-2% NPS
	key ^= board.PiecesRand[White][Pawn][b.Counts[White][Pawn]]
	key ^= board.PiecesRand[White][Knight][b.Counts[White][Knight]]
	key ^= board.PiecesRand[White][Bishop][b.Counts[White][Bishop]]
	key ^= board.PiecesRand[White][Rook][b.Counts[White][Rook]]
	key ^= board.PiecesRand[White][Queen][b.Counts[White][Queen]]
	key ^= board.PiecesRand[Black][Pawn][b.Counts[Black][Pawn]]
	key ^= board.PiecesRand[Black][Knight][b.Counts[Black][Knight]]
	key ^= board.PiecesRand[Black][Bishop][b.Counts[Black][Bishop]]
	key ^= board.PiecesRand[Black][Rook][b.Counts[Black][Rook]]
	key ^= board.PiecesRand[Black][Queen][b.Counts[Black][Queen]]

	entry := &e.materialCache[key%materialCacheSize]
	if entry.hash == key {
		return e.matFuncs[entry.evalID](e, b, c)
	}

	var evalID evalID

	switch {
	case insufficient(b):
		evalID = evalInsufficientID

	case knbvk(b):
		evalID = evalKNBvKID

	case knvkp(b):
		evalID = evalKNvKPID

	case kbvkp(b):
		evalID = evalKBvKPID

	default:
		evalID = evalPositionalID
	}

	entry.hash = key
	entry.evalID = evalID

	return e.matFuncs[evalID](e, b, c)
}

type evalID byte

const (
	evalInsufficientID = iota
	evalKNBvKID
	evalKNvKPID
	evalKBvKPID
	evalPositionalID
)

func evalInsufficient[T ScoreType](e *Eval[T], b *board.Board, c *CoeffSet[T]) T {
	return 0
}

func evalKNBvK[T ScoreType](e *Eval[T], b *board.Board, c *CoeffSet[T]) T {
	return e.knbvk(b, c)
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

func evalPositional[T ScoreType](e *Eval[T], b *board.Board, c *CoeffSet[T]) T {
	// drawishness
	fifty := int(100 - b.FiftyCnt)
	fifty *= fifty
	sf := T((fifty * MaxScaleFactor) / 10_000)
	e.scaleFactor[White] = sf
	e.scaleFactor[Black] = sf
	return e.positional(b, c)
}
