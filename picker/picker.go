// Package picker is a lazy move loop iterator. Generating moves or ranking
// them are delayed in the hopes of beta cut or pruning.
//
// Use AllMoves picker in case all pseudo legal moves are needed, this is the
// case from the main search.
//
// Use NoisyOrEvasions in case only noisy moves are needed or in check all
// evasions are needed. This is the case in a qsearch.
package picker

import (
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/heur"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"
	"github.com/paulsonkoly/chess-3/stack"
)

// AllMoves is the move iterator for a given position that iterates all pseudo
// legal moves.
type AllMoves struct {
	yielder
	board    *board.Board
	ranker   *heur.MoveRanker
	hstack   *stack.Stack[heur.StackMove]
	hashMove move.Move
	state    state
}

// NewAllMoves creates a new move iterator for the position represented by b.
// ms points to the move store. ranker points to heur.Ranker. hstack points to
// the history stack.
func NewAllMoves(
	b *board.Board,
	ms *move.Store,
	ranker *heur.MoveRanker,
	hashMove move.Move,
	hstack *stack.Stack[heur.StackMove],
) AllMoves {
	return AllMoves{yielder: yielder{ms: ms}, board: b, ranker: ranker, hashMove: hashMove, hstack: hstack}
}

// NoisyOrEvasions is the move iterator for a given position that iterates
// noisy moves or when in check all evasions.
type NoisyOrEvasions struct {
	yielder
	board    *board.Board
	ranker   *heur.MoveRanker
	checkers BitBoard
	state    state
}

// NewNoisyOrEvasions creates a new move iterator for the position represented by b.
// ms points to the move store. ranker points to heur.Ranker. checkers has the
// squares of pieces giving check.
func NewNoisyOrEvasions(b *board.Board, ms *move.Store, ranker *heur.MoveRanker, checkers BitBoard) NoisyOrEvasions {
	var state state
	if checkers == 0 {
		state = genNoisy
	} else {
		state = genNoisyEvasion
	}
	return NoisyOrEvasions{yielder: yielder{ms: ms}, board: b, ranker: ranker, state: state, checkers: checkers}
}

type state byte

const (
	pickHash state = iota
	genNoisy
	yieldGoodNoisy
	genQuiet
	genNoisyEvasion
	yieldNoisyEvasion
	genQuietEvasion
	yieldRest
)

func (am *AllMoves) Next() bool {
	switch am.state {

	case pickHash:
		am.state = genNoisy
		if am.board.IsPseudoLegal(am.hashMove) {
			// we put the hash move in the actual store move buffer, in case we need
			// to update histories on fail high
			m := am.ms.Alloc(am.hashMove)
			m.Weight = heur.HashMove
			am.ix++
			return true
		}
		fallthrough

	case genNoisy:
		am.state = yieldGoodNoisy
		movegen.Noisy(am.ms, am.board)
		moves := am.ms.Frame()

		for i := am.ix; i < len(moves); i++ {
			if am.hashMove == moves[i].Move {
				// hash move was already yielded
				moves[i].Weight = -heur.HashMove
			} else {
				moves[i].Weight = am.ranker.RankNoisy(moves[i].Move, am.board)
			}
		}

		fallthrough

	case yieldGoodNoisy:
		if am.yield(0) {
			return true
		}

		am.state = genQuiet
		fallthrough

	case genQuiet:
		am.state = yieldRest

		quietStart := len(am.ms.Frame())
		movegen.Quiet(am.ms, am.board)
		moves := am.ms.Frame()

		for i := quietStart; i < len(moves); i++ {
			if am.hashMove == moves[i].Move {
				// hash move was already yielded
				moves[i].Weight = -heur.HashMove
			} else {
				moves[i].Weight = am.ranker.RankQuiet(moves[i].Move, am.board, am.hstack)
			}
		}
		fallthrough

	case yieldRest:
		return am.yield(-heur.HashMove + 1)
	}

	return false
}

func (noe *NoisyOrEvasions) Next() bool {
	for {
		switch noe.state {

		case genNoisy:
			noe.state = yieldRest
			movegen.Noisy(noe.ms, noe.board)
			moves := noe.ms.Frame()

			for i := noe.ix; i < len(moves); i++ {
				moves[i].Weight = noe.ranker.RankNoisy(moves[i].Move, noe.board)
			}
			continue

		case genNoisyEvasion:
			noe.state = yieldNoisyEvasion
			movegen.NoisyEvasions(noe.ms, noe.board, noe.checkers)
			moves := noe.ms.Frame()

			for i := noe.ix; i < len(moves); i++ {
				moves[i].Weight = noe.ranker.RankNoisyEvasion(moves[i].Move, noe.board)
			}
			fallthrough

		case yieldNoisyEvasion:
			if noe.yield(-Inf) {
				return true
			}

			noe.state = genQuietEvasion
			fallthrough

		case genQuietEvasion:
			noe.state = yieldRest

			quietStart := len(noe.ms.Frame())
			movegen.QuietEvasions(noe.ms, noe.board, noe.checkers)
			moves := noe.ms.Frame()

			for i := quietStart; i < len(moves); i++ {
				moves[i].Weight = noe.ranker.RankQuietEvasion(moves[i].Move, noe.board)
			}
			fallthrough

		case yieldRest:
			return noe.yield(-Inf)
		}
	}
}

type yielder struct {
	ms *move.Store
	ix int
}

func (y *yielder) yield(threshold Score) bool {
	moves := y.ms.Frame()
	maxim := threshold
	best := -1
	for i := y.ix; i < len(moves); i++ {
		if maxim < moves[i].Weight {
			maxim = moves[i].Weight
			best = i
		}
	}

	if best != -1 {
		moves[y.ix], moves[best] = moves[best], moves[y.ix]
		y.ix++
		return true
	}

	return false
}

// Move is the currently yielded move. It's only valid if Next() is called
// first and if it returned true.
func (y *yielder) Move() *move.Weighted {
	return &y.ms.Frame()[y.ix-1]
}

// YieldedMoves returns a slice of yielded moves so far.
func (y *yielder) YieldedMoves() []move.Weighted {
	return y.ms.Frame()[:y.ix]
}
