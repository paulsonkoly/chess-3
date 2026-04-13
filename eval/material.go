package eval

import (
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
)

const MaterialCacheSize = 2 * 1024

type (
	Hash                  = board.Hash
	EvalFunc[T ScoreType] = func(*Eval[T], *board.Board, *CoeffSet[T]) T
)

type MaterialCache[T ScoreType] struct {
	hash   Hash
	evalID evalID
}

// material count dispatcher and cache.
func (e *Eval[T]) material(b *board.Board, c *CoeffSet[T]) T {
	hash := Hash(0)
	shift := 0
	for color := range Colors {
		for pType := Pawn; pType <= Queen; pType++ {
			count := (b.Colors[color] & b.Pieces[pType]).Count()
			hash |= Hash(count) << shift
			// maximal number of a piece of a single colour is 10, 2 + 8 promotions
			// 4 bits enough, 6 fits in Hash
			shift += 6
		}
	}
	// the empty position (just kings) hash would be 0, which coincidentally the
	// zero value of the hash key in the hash slot, causing a false hit.
	hash = murmur(hash) + 1

	entry := &e.materialCache[hash%MaterialCacheSize]
	if entry.hash == hash {
		return e.matFuncs[entry.evalID](e, b, c)
	}

	var evalID evalID

	switch {
	case insufficient(b):
		evalID = evalInsufficientID

	case knbvk(b):
		evalID = evalKNBvKID

	default:
		evalID = evalPositionalID
	}

	entry.hash = hash
	entry.evalID = evalID

	return e.matFuncs[evalID](e, b, c)
}

func murmur(key Hash) Hash {
	key ^= key >> 33
	key *= 0xff51afd7ed558ccd
	key ^= key >> 33
	key *= 0xc4ceb9fe1a85ec53
	key ^= key >> 33
	return key
}

type evalID byte

const (
	evalInsufficientID = iota
	evalKNBvKID
	evalPositionalID
)

func evalInsufficient[T ScoreType](e *Eval[T], b *board.Board, c *CoeffSet[T]) T {
	return 0
}

func evalKNBvK[T ScoreType](e *Eval[T], b *board.Board, c *CoeffSet[T]) T {
	return e.knbvk(b, c)
}

func evalPositional[T ScoreType](e *Eval[T], b *board.Board, c *CoeffSet[T]) T {
	// drawishness
	fifty := int(100 - b.FiftyCnt)
	fifty *= fifty
	e.scaleFactor = T((fifty * MaxScaleFactor) / 10_000)
	return e.positional(b, c)
}
