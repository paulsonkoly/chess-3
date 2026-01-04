// Package picker is a lazy move loop iterator. Generating moves or ranking
// them are delayed in the hopes of beta cut or pruning.
package picker

import (
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/heur"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"
	"github.com/paulsonkoly/chess-3/stack"
)

// Picker is the move iterator for a given position.
type Picker struct {
	board    *board.Board
	ms       *move.Store
	ranker   *heur.MoveRanker
	hstack   *stack.Stack[heur.StackMove]
	moves    []move.Weighted
	ix       int
	hashMove move.Move
	state    state
}

type state byte

const (
	pickHash state = iota
	genNoisy
	yieldGoodNoisy
	genQuiet
	yieldRest
)

// NewPicker creates a new move iterator for the position represented by b.
// hashMove will be yielded first. ms points to the move store. ranker points
// to heur.Ranker. hstack points to the history stack.
func NewPicker(
	b *board.Board,
	hashMove move.Move,
	ms *move.Store,
	ranker *heur.MoveRanker,
	hstack *stack.Stack[heur.StackMove],
) Picker {
	return Picker{board: b, hashMove: hashMove, ms: ms, hstack: hstack, ranker: ranker}
}

func (p *Picker) Next() bool {
	switch p.state {

	case pickHash:
		p.state = genNoisy
		if p.board.IsPseudoLegal(p.hashMove) {
			(*p.ms.Alloc()).Move = p.hashMove
			p.moves = p.ms.Frame()
			p.ix++
			return true
		}
		fallthrough

	case genNoisy:
		p.state = yieldGoodNoisy
		movegen.GenNoisy(p.ms, p.board)
		moves := p.ms.Frame()

		// remove duplicate hashmove
		if p.ix > 0 {
			for i := p.ix; i < len(moves); i++ {
				if p.hashMove == moves[i].Move {
					moves[len(moves)-1], moves[i] = moves[i], moves[len(moves)-1]
					moves = moves[:len(moves)-1]
					break
				}
			}
		}

		for i := p.ix; i < len(moves); i++ {
			moves[i].Weight = p.ranker.RankNoisy(moves[i].Move, p.board, p.hstack)
		}
		p.moves = moves

		fallthrough

	case yieldGoodNoisy:

		maxim := Score(0) // start at 0 to filter out bad noisy
		best := -1
		for i := p.ix; i < len(p.moves); i++ {
			if maxim < p.moves[i].Weight {
				maxim = p.moves[i].Weight
				best = i
			}
		}

		if best != -1 {
			p.moves[p.ix], p.moves[best] = p.moves[best], p.moves[p.ix]
			p.ix++
			return true
		}

		p.state = genQuiet
		fallthrough

	case genQuiet:
		p.state = yieldRest

		quietStart := len(p.ms.Frame())
		movegen.GenNotNoisy(p.ms, p.board)
		moves := p.ms.Frame()

		// remove duplicate hashmove
		for i := quietStart; i < len(moves); i++ {
			if p.hashMove == moves[i].Move {
				moves[len(moves)-1], moves[i] = moves[i], moves[len(moves)-1]
				moves = moves[:len(moves)-1]
				break
			}
		}

		for i := quietStart; i < len(moves); i++ {
			moves[i].Weight = p.ranker.RankQuiet(moves[i].Move, p.board, p.hstack)
		}
		p.moves = moves

		fallthrough

	case yieldRest:

		maxim := -Inf - 1
		best := -1
		for i := p.ix; i < len(p.moves); i++ {
			if maxim < p.moves[i].Weight {
				maxim = p.moves[i].Weight
				best = i
			}
		}

		if best != -1 {
			p.moves[p.ix], p.moves[best] = p.moves[best], p.moves[p.ix]
			p.ix++
			return true
		}
	}

	return false
}

// Move is the currently yielded move. It's only valid if Next() is called
// first and if it returned true.
func (p *Picker) Move() move.Move {
	return p.moves[p.ix-1].Move
}

// YieldedMoves returns a slice of yielded moves so far.
func (p *Picker) YieldedMoves() []move.Weighted {
	return p.moves[:p.ix]
}
