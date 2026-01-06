// Package picker is a lazy move loop iterator. Generating moves or ranking
// them are delayed in the hopes of beta cut or pruning.
package picker

import (
	"github.com/paulsonkoly/chess-3/bitset"
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/heur"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"
	"github.com/paulsonkoly/chess-3/stack"
)

// Picker is the move iterator for a given position.
type Picker struct {
	board       *board.Board
	ms          *move.Store
	ranker      *heur.MoveRanker
	hstack      *stack.Stack[heur.StackMove]
	goodNoisies bitset.BitSet
	badNoisies  bitset.BitSet
	goodQuiets  bitset.BitSet
	badQuiets   bitset.BitSet
	yieldedHash bool
	yielded     bitset.BitSet
	hashMove    move.Move
	state       state
}

type state byte

const (
	pickHash state = iota
	genNoisy
	yieldGoodNoisy
	genQuiet
	yieldGoodQuiet
	yieldBadQuiet
	yieldBadNoisy
)

// badQuietThreshold controls when we switch from heuristic order to generation order.
const badQuietThreshold = -heur.MaxHistory / 4

// New creates a new move iterator for the position represented by b. hashMove
// will be yielded first. ms points to the move store. ranker points to
// heur.Ranker. hstack points to the history stack.
func New(
	b *board.Board,
	hashMove move.Move,
	ms *move.Store,
	ranker *heur.MoveRanker,
	hstack *stack.Stack[heur.StackMove],
) Picker {
	return Picker{board: b, hashMove: hashMove, ms: ms, hstack: hstack, ranker: ranker}
}

func (p *Picker) Next() (move.Move, bool) {
	switch p.state {

	case pickHash:
		p.state = genNoisy
		if p.board.IsPseudoLegal(p.hashMove) {
			p.yieldedHash = true
			return p.hashMove, true
		}
		fallthrough

	case genNoisy:
		p.state = yieldGoodNoisy
		movegen.GenNoisy(p.ms, p.board)
		moves := p.ms.Frame()

		for i, m := range moves {
			if p.hashMove == m.Move {
				p.yielded.Set(i)
			} else {
				weight := p.ranker.RankNoisy(m.Move, p.board, p.hstack)
				if weight >= 0 {
					p.goodNoisies.Set(i)
				} else {
					p.badNoisies.Set(i)
				}
				moves[i].Weight = weight
			}
		}

		fallthrough

	case yieldGoodNoisy:
		maxim := -Inf
		best := -1
		iter := p.goodNoisies
		iter.AndNot(&p.yielded)
		moves := p.ms.Frame()
		for ix := iter.Next(); ix != -1; ix = iter.Next() {
			iter.Clear(ix)
			if maxim < moves[ix].Weight {
				maxim = moves[ix].Weight
				best = ix
			}
		}

		if best != -1 {
			p.yielded.Set(best)
			return moves[best].Move, true
		}

		p.state = genQuiet
		fallthrough

	case genQuiet:
		p.state = yieldGoodQuiet

		quietStart := len(p.ms.Frame())
		movegen.GenNotNoisy(p.ms, p.board)
		moves := p.ms.Frame()[quietStart:]

		for i, m := range moves {
			if p.hashMove == m.Move {
				p.yielded.Set(quietStart + i)
			} else {
				weight := p.ranker.RankQuiet(m.Move, p.board, p.hstack)

				if weight < badQuietThreshold {
					p.badQuiets.Set(quietStart + i)
				} else {
					p.goodQuiets.Set(quietStart + i)
				}
				moves[i].Weight = weight
			}
		}

		fallthrough

	case yieldGoodQuiet:
		maxim := -Inf
		best := -1
		iter := p.goodQuiets
		iter.AndNot(&p.yielded)
		moves := p.ms.Frame()
		for ix := iter.Next(); ix != -1; ix = iter.Next() {
			iter.Clear(ix)
			if maxim < moves[ix].Weight {
				maxim = moves[ix].Weight
				best = ix
			}
		}

		if best != -1 {
			p.yielded.Set(best)
			return moves[best].Move, true
		}

		p.state = yieldBadQuiet
		fallthrough

	case yieldBadQuiet:
		// return badQuiets in generation order, this is up to debate. Whether
		// we want bucket system, or pure generation order. We are most likely in
		// an unfortunate AllNode.

		// this is destructive to p.badQuiets
		if ix := p.badQuiets.Next(); ix != -1 {
			p.badQuiets.Clear(ix)
			p.yielded.Set(ix)
			return p.ms.Frame()[ix].Move, true
		}
		p.state = yieldBadNoisy
		fallthrough

	case yieldBadNoisy:
		// return badNoisies in heuristic order, this is up to debate
		maxim := -Inf
		best := -1
		iter := p.badNoisies
		iter.AndNot(&p.yielded)
		moves := p.ms.Frame()
		for ix := iter.Next(); ix != -1; ix = iter.Next() {
			iter.Clear(ix)
			if maxim < moves[ix].Weight {
				maxim = moves[ix].Weight
				best = ix
			}
		}

		if best != -1 {
			p.yielded.Set(best)
			return moves[best].Move, true
		}
	}

	return 0, false
}

func (p *Picker) FailHigh(m move.Move, d Depth, failedSoft bool, nType Node) {
	bonus := Score(d) * 20
	if failedSoft {
		bonus++
	}
	switch nType {
	case AllNode:
		bonus += 2
	case CutNode:
		bonus += 1
	case PVNode:
	}
	malus := -bonus / 2

	if p.yieldedHash {
		adjustment := bonus
		if p.hashMove != m {
			adjustment = malus
		}
		moved := p.board.SquaresToPiece[p.hashMove.From()]
		// TODO en-passant
		captured := p.board.SquaresToPiece[p.hashMove.To()]

		if captured == NoPiece && p.hashMove.Promo() == NoPiece {
			p.ranker.Adjust(p.board.STM, p.hashMove, moved, p.hstack, adjustment)
		}
	}

	// we are destroying yielded at this point
	for ix := p.yielded.Next(); ix != -1; ix = p.yielded.Next() {
		p.yielded.Clear(ix)
		curr := p.ms.Frame()[ix].Move

		moved := p.board.SquaresToPiece[curr.From()]
		// TODO en-passant
		captured := p.board.SquaresToPiece[curr.To()]

		if captured != NoPiece || curr.Promo() != NoPiece {
			continue
		}

		switch curr {

		case p.hashMove:
			continue

		case m:
			p.ranker.Adjust(p.board.STM, curr, moved, p.hstack, bonus)

		default:
			p.ranker.Adjust(p.board.STM, curr, moved, p.hstack, malus)
		}
	}
}
