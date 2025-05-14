package picker

import (
	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/heur"
	"github.com/paulsonkoly/chess-3/hist"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

type Picker struct {
	b      *board.Board          // b is board pointer
	ms     *move.Store           // ms is the move store
	yix    int                   // yix is the index of first move not yet picked
	hist   *heur.History         // hist is the pointer to history heuristics
	cont   [2]*heur.Continuation // cont is the pointers to continuation heuristics
	hstack *hist.Stack           // hstack is the search stack
	hash   move.SimpleMove       // hash Move
	phase  phase                 // phase is the picker FSM state
}

type phase byte

const (
	hashMove = phase(iota)
	weighCaptures
	goodCaptures
	weighNonCaptures
	rest
)

func NewPicker(
	b *board.Board,
	hash move.SimpleMove,
	ms *move.Store,
	hist *heur.History,
	cont [2]*heur.Continuation,
	hstack *hist.Stack,
) *Picker {
	phase := hashMove
	if hash == 0 {
		phase = weighCaptures
	}
	return &Picker{b: b, hash: hash, ms: ms, phase: phase, hist: hist, cont: cont, hstack: hstack}
}

func (p *Picker) Pick() *move.Move {
	var result *move.Move

	moves := p.ms.Frame()

	success := false
	for !success {
		switch p.phase {

		case hashMove:
			p.phase = weighCaptures

			// delay generating moves, construct the hash move from SimpleMove instead
			result = p.ms.Alloc()
			// TODO: the hash move is not necessarily valid in position
			*result = movegen.FromSimple(p.b, p.hash)
			p.yix++
			success = true

		case weighCaptures:
			p.phase = goodCaptures

			// Generate forcing moves first
			movegen.GenForcing(p.ms, p.b)
			// re-obtain the frame because we generated new moves
			moves = p.ms.Frame()

			goodCnt := 0

			for ix := p.yix; ix < len(moves); ix++ {
				m := &moves[ix]

				switch {

				case p.hash.Matches(m):
					// the hash move was already yielded, swap it forward.
					moves[p.yix], moves[ix] = moves[ix], moves[p.yix]
					p.yix++

				case p.b.SquaresToPiece[m.To()] != NoPiece:
					see := heur.SEE(p.b, m)
					if see < 0 {
						m.Weight = see - heur.Captures
					} else {
						m.Weight = see + heur.Captures
						goodCnt++
					}

				case m.Promo() != NoPiece:
					m.Weight = heur.Captures + heur.PieceValues[m.Promo()]
				}
			}

			if goodCnt == 0 {
				p.phase = weighNonCaptures
			}

		case goodCaptures:
			maxim := -Inf - 1
			best := -1

			for ix := p.yix; ix < len(moves); ix++ {
				if maxim < moves[ix].Weight {
					maxim = moves[ix].Weight
					best = ix
				}
			}
			if best != -1 && maxim >= heur.Captures {
				result = p.yield(moves, best)
				success = true
			} else {
				p.phase = weighNonCaptures
			}

		case weighNonCaptures:
			p.phase = rest

			// generate quiet moves
			movegen.GenQuiets(p.ms, p.b)
			// re-obtain the frame
			moves = p.ms.Frame()

			for ix := p.yix; ix < len(moves); ix++ {
				m := &moves[ix]

				switch {

				case p.hash.Matches(m):
					// the hash move was already yielded, swap it forward.
					moves[p.yix], moves[ix] = moves[ix], moves[p.yix]
					p.yix++
					continue

				case m.Weight == 0:
					score := p.hist.Probe(p.b.STM, m.From(), m.To())

					if p.hstack.Size() >= 1 {
						hist := p.hstack.Top(0)
						score += 3 * p.cont[0].Probe(p.b.STM, hist.Piece, hist.To, m.Piece, m.To())
					}

					if p.hstack.Size() >= 2 {
						hist := p.hstack.Top(1)
						score += 2 * p.cont[1].Probe(p.b.STM, hist.Piece, hist.To, m.Piece, m.To())
					}

					m.Weight = score
				}
			}

		case rest:
			maxim := -Inf - 1
			best := -1

			for ix := p.yix; ix < len(moves); ix++ {
				if maxim < moves[ix].Weight {
					maxim = moves[ix].Weight
					best = ix
				}
			}
			if best != -1 {
				result = p.yield(moves, best)
			}
			success = true
		}
	}

	return result
}

func (p *Picker) yield(moves []move.Move, ix int) *move.Move {
	moves[p.yix], moves[ix] = moves[ix], moves[p.yix]
	result := &moves[p.yix]
	p.yix++
	return result
}
