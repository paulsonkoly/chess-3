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

type picker struct {
	board  *board.Board
	ms     *move.Store
	ranker *heur.MoveRanker
	ix     int
	state  state
}

// Main is the move iterator for a given position in the main search.
type Main struct {
	picker
	hstack   *stack.Stack[heur.StackMove]
	hashMove move.Move
}

// NewMain creates a new move iterator for the position represented by b.
// ms points to the move store. ranker points to heur.Ranker. hstack points to
// the history stack.
func NewMain(
	b *board.Board,
	ms *move.Store,
	ranker *heur.MoveRanker,
	hashMove move.Move,
	hstack *stack.Stack[heur.StackMove],
) Main {
	return Main{picker: picker{board: b, ms: ms, ranker: ranker}, hashMove: hashMove, hstack: hstack}
}

// QSearch is the move iterator for a given position in the quiessence search.
type QSearch struct {
	picker
	checkers BitBoard
}

// NewQSearch creates a new move iterator for the position represented by b.
// ms points to the move store. ranker points to heur.Ranker. hstack points to
// the history stack.
func NewQSearch(b *board.Board, ms *move.Store, ranker *heur.MoveRanker, checkers BitBoard) QSearch {
	return QSearch{picker: picker{board: b, ms: ms, ranker: ranker}, checkers: checkers}
}

type state byte

const (
	start state = iota
	genNoisy
	yieldGoodNoisy
	genQuiet
	genNoisyEvasion
	yieldNoisyEvasion
	genQuietEvasion
	yieldRest
)

func (p *Main) Next() bool {
	switch p.state {

	case start:
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
		movegen.Noisy(p.ms, p.board)
		moves := p.ms.Frame()

		for i := p.ix; i < len(moves); i++ {
			if p.hashMove == moves[i].Move {
				// hash move was already yielded
				moves[i].Weight = -heur.HashMove
			} else {
				moves[i].Weight = p.ranker.RankNoisy(moves[i].Move, p.board)
			}
		}

		fallthrough

	case yieldGoodNoisy:
		if p.yield(0) {
			return true
		}

		p.state = genQuiet
		fallthrough

	case genQuiet:
		p.state = yieldRest

		quietStart := len(p.ms.Frame())
		movegen.Quiet(p.ms, p.board)
		moves := p.ms.Frame()

		for i := quietStart; i < len(moves); i++ {
			if p.hashMove == moves[i].Move {
				// hash move was already yielded
				moves[i].Weight = -heur.HashMove
			} else {
				moves[i].Weight = p.ranker.RankQuiet(moves[i].Move, p.board, p.hstack)
			}
		}
		fallthrough

	case yieldRest:
		return p.yield(-heur.HashMove + 1)
	}

	return false
}

func (qs *QSearch) QSNext() bool {
	for {
		switch qs.state {

		case start:
			if qs.checkers != 0 {
				qs.state = genNoisyEvasion
				continue
			} else {
				qs.state = genNoisy
			}
			fallthrough

		case genNoisy:
			qs.state = yieldRest
			movegen.Noisy(qs.ms, qs.board)
			moves := qs.ms.Frame()

			for i := qs.ix; i < len(moves); i++ {
				moves[i].Weight = qs.ranker.RankNoisy(moves[i].Move, qs.board)
			}
			continue

		case genNoisyEvasion:
			qs.state = yieldNoisyEvasion
			movegen.NoisyEvasions(qs.ms, qs.board, qs.checkers)
			moves := qs.ms.Frame()

			for i := qs.ix; i < len(moves); i++ {
				moves[i].Weight = qs.ranker.RankNoisyEvasion(moves[i].Move, qs.board)
			}
			fallthrough

		case yieldNoisyEvasion:
			if qs.yield(-Inf) {
				return true
			}

			qs.state = genQuietEvasion
			fallthrough

		case genQuietEvasion:
			qs.state = yieldRest

			quietStart := len(qs.ms.Frame())
			movegen.QuietEvasions(qs.ms, qs.board, qs.checkers)
			moves := qs.ms.Frame()

			for i := quietStart; i < len(moves); i++ {
				moves[i].Weight = qs.ranker.RankQuietEvasion(moves[i].Move, qs.board)
			}
			fallthrough

		case yieldRest:
			return qs.yield(-Inf)
		}
	}
}

func (p *picker) yield(threshold Score) bool {
	moves := p.ms.Frame()
	maxim := threshold
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

	return false
}

// Move is the currently yielded move. It's only valid if Next() is called
// first and if it returned true.
func (p *picker) Move() *move.Weighted {
	return &p.ms.Frame()[p.ix-1]
}

// YieldedMoves returns a slice of yielded moves so far.
func (p *picker) YieldedMoves() []move.Weighted {
	return p.ms.Frame()[:p.ix]
}
