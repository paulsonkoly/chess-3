package eval

import (
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
)

// ScoreType defines the evaluation result type. The engine uses int16 for
// score type, as defined in types. The tuner uses float64.
type ScoreType interface{ Score | float64 }

type Phase byte

const (
	MG = Phase(iota)
	EG

	Phases
)

type Eval[T ScoreType] struct {
	sp            [Colors][Phases]T
	scaleFactor   T
	kingAttacks   [Colors]T
	attacks       [Colors][Pieces]BitBoard
	cover         [Colors]BitBoard
	pawns         [Colors]Pawns
	kings         [Colors]Kings
	matFuncs      [6]evalFunc[T]
	pawnCache     []PawnCache
	pawnKingCache []PawnKingCache
	materialCache []MaterialCache[T]
}

type Pawns struct {
	cover      BitBoard
	frontline  BitBoard
	backmost   BitBoard
	frontspan  BitBoard
	neighbourF BitBoard
}

type Kings struct {
	nb BitBoard
	sq Square
}

func New[T ScoreType]() *Eval[T] {
	return &Eval[T]{
		pawnCache:     make([]PawnCache, PawnCacheSize),
		pawnKingCache: make([]PawnKingCache, PawnKingCacheSize),
		materialCache: make([]MaterialCache[T], materialCacheSize),
		matFuncs: [6]evalFunc[T]{
			evalInsufficient[T],
			evalKNBvK[T],
			evalOCB[T],
			evalOCBKnights[T],
			evalOCBRooks[T],
			evalPositional[T],
		},
	}
}

func (e *Eval[T]) Clear() {
	for i := range e.pawnCache {
		e.pawnCache[i] = PawnCache{}
	}

	for i := range e.pawnKingCache {
		e.pawnKingCache[i] = PawnKingCache{}
	}

	for i := range e.materialCache {
		e.materialCache[i] = MaterialCache[T]{}
	}
}

func (e *Eval[T]) Score(b *board.Board, c *CoeffSet[T]) T {
	return e.material(b, c)
}
