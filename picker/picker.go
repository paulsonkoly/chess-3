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
	ix       int
	split    int
	hashMove move.Move
	state    state
}

type state byte

const (
	pickHash state = iota
	genNoisy
	yieldGoodNoisy
	genQuiet
	yieldGtSplit
	yieldLeSplit
)

// New creates a new move iterator for the position represented by b.
// hashMove will be yielded first. ms points to the move store. ranker points
// to heur.Ranker. hstack points to the history stack.
func New(
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
			// we put the hash move in the actual store move buffer, in case we need
			// to update histories on fail high
			m := p.ms.Alloc(p.hashMove)
			m.Weight = heur.HashMove
			p.ix++
			return true
		}
		fallthrough

	case genNoisy:
		p.state = yieldGoodNoisy
		movegen.GenNoisy(p.ms, p.board)
		moves := p.ms.Frame()

		for i := p.ix; i < len(moves); i++ {
			if p.hashMove == moves[i].Move {
				// hash move was already yielded
				moves[i].Weight = -heur.HashMove
			} else {
				moves[i].Weight = p.ranker.RankNoisy(moves[i].Move, p.board, p.hstack)
			}
		}

		fallthrough

	case yieldGoodNoisy:
		moves := p.ms.Frame()

		maxim := Score(0) // start at 0 to filter out bad noisy
		best := -1
		for i := p.ix; i < len(moves); i++ {
			if maxim < moves[i].Weight {
				maxim = moves[i].Weight
				best = i
			}
		}

		if best != -1 {
			moves[p.ix], moves[best] = moves[best], moves[p.ix]
			p.ix++
			return true
		}

		p.state = genQuiet
		fallthrough

	case genQuiet:
		p.state = yieldGtSplit

		quietStart := len(p.ms.Frame())
		movegen.GenNotNoisy(p.ms, p.board)
		moves := p.ms.Frame()

		for i := quietStart; i < len(moves); i++ {
			if p.hashMove == moves[i].Move {
				// hash move was already yielded
				moves[i].Weight = -heur.HashMove
			} else {
				moves[i].Weight = p.ranker.RankQuiet(moves[i].Move, p.board, p.hstack)
			}
		}

		p.split = partialSort(moves[p.ix:], -100) + p.ix

		fallthrough

	case yieldGtSplit:
		moves := p.ms.Frame()

		if best, ok := findMax(moves[p.ix:p.split], -heur.HashMove); ok {
			moves[p.ix], moves[best+p.ix] = moves[best+p.ix], moves[p.ix]
			p.ix++
			return true
		}

		p.state = yieldLeSplit
		fallthrough

	case yieldLeSplit:
		moves := p.ms.Frame()

		if best, ok := findMax(moves[p.ix:], -heur.HashMove); ok {
			moves[p.ix], moves[best+p.ix] = moves[best+p.ix], moves[p.ix]
			p.ix++
			return true
		}

		p.state = yieldLeSplit
	}

	return false
}

// Move is the currently yielded move. It's only valid if Next() is called
// first and if it returned true.
func (p *Picker) Move() *move.Weighted {
	return &p.ms.Frame()[p.ix-1]
}

// YieldedMoves returns a slice of yielded moves so far.
func (p *Picker) YieldedMoves() []move.Weighted {
	return p.ms.Frame()[:p.ix]
}

func findMax(moves []move.Weighted, threshold Score) (ix int, ok bool) {
	maxim := threshold
	for i := range moves {
		if maxim < moves[i].Weight {
			maxim = moves[i].Weight
			ix = i
			ok = true
		}
	}
	return
}

func partialSort(moves []move.Weighted, threshold Score) int {
	i := 0
	j := len(moves) - 1
	for i <= j {
		for i < len(moves) && moves[i].Weight > threshold {
			i++
		}
		for j >= 0 && moves[j].Weight <= threshold {
			j--
		}
		if i >= j {
			return i
		}
		moves[i], moves[j] = moves[j], moves[i]
	}
	return i
}
