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
	shift := 0
	for color := range Colors {
		for pType := Pawn; pType <= Queen; pType++ {
			count := b.Counts[color][pType]
			key |= hash(count) << shift
			// maximal number of a piece of a single colour is 10, 2 + 8 promotions
			// 4 bits enough, 6 fits in Hash
			shift += 6
		}
	}
	// the empty position (just kings) hash would be 0, which coincidentally the
	// zero value of the hash key in the hash slot, causing a false hit.
	key = murmur(key) + 1

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

	default:
		evalID = evalPositionalID
	}

	entry.hash = key
	entry.evalID = evalID

	return e.matFuncs[evalID](e, b, c)
}

func murmur(key hash) hash {
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
