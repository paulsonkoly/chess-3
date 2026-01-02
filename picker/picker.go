package picker

import (
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/heur"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"
	"github.com/paulsonkoly/chess-3/stack"
)

type state byte

const (
	pickHash state = iota
	genMoves
	yieldMoves
)

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
		p.state = genMoves
		if p.board.IsPseudoLegal(p.hashMove) {
			(*p.ms.Alloc()).Move = p.hashMove
			p.moves = p.ms.Frame()
			p.ix++
			return true
		}
		fallthrough

	case genMoves:
		p.state = yieldMoves
		movegen.GenMoves(p.ms, p.board)
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
			m := moves[i]
			var val Score
			switch {
			case m.Promo() != NoPiece || p.board.SquaresToPiece[p.board.CaptureSq(m.Move)] != NoPiece:
				val = p.ranker.RankNoisy(m.Move, p.board, p.hstack)
			default:
				val = p.ranker.RankQuiet(m.Move, p.board, p.hstack)
			}

			moves[i].Weight = val
		}
		p.moves = moves

		fallthrough

	case yieldMoves:

		maxim := -Inf - 1
		best := -1
		for i := p.ix; i < len(p.moves); i++ {
			if maxim < p.moves[i].Weight {
				maxim = p.moves[i].Weight
				best = i
			}
		}

		if best == -1 {
			return false
		}

		p.moves[p.ix], p.moves[best] = p.moves[best], p.moves[p.ix]
		p.ix++

		return true
	}

	return false
}

func (p *Picker) Move() move.Move {
	return p.moves[p.ix-1].Move
}

func (p *Picker) FailHigh(d Depth) {
	p.ranker.FailHigh(d, p.board, p.moves[:p.ix], p.hstack)
}
