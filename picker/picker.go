package picker

import (
	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/heur"
	"github.com/paulsonkoly/chess-3/hist"
	"github.com/paulsonkoly/chess-3/move"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

type Picker struct {
	b      *board.Board          // b is board pointer
	moves  []move.Move           // moves is the slice of moves from which we pick
	yix    int                   // yix is the index of first move not yet picked
	swapIx int                   // swapIx points to the end of good captures
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
	moves []move.Move,
	hist *heur.History,
	cont [2]*heur.Continuation,
	hstack *hist.Stack,
) *Picker {
	phase := hashMove
	if hash == 0 {
		phase = weighCaptures
	}
	return &Picker{b: b, hash: hash, moves: moves, phase: phase, hist: hist, cont: cont, hstack: hstack}
}

func (p *Picker) Pick() *move.Move {
	var result *move.Move

	success := false
	for !success {
		switch p.phase {

		case hashMove:
			p.phase = weighCaptures

			for ix, m := range p.moves {
				if p.hash.Matches(&m) {
					m.Weight = heur.HashMove // this is not needed here, but helps understanding in debug
					result = p.yield(ix)
					success = true
					break
				}
			}

		case weighCaptures:
			p.phase = goodCaptures

			p.swapIx = p.yix
			for ix := p.yix; ix < len(p.moves); ix++ {
				m := &p.moves[ix]

				if p.b.SquaresToPiece[m.To()] != NoPiece {
					see := heur.SEE(p.b, m)
					if see < 0 {
						m.Weight = see - heur.Captures
					} else {
						m.Weight = see + heur.Captures
						// bring good captures forward, so the good capture loop can
						// terminate when the weight drops under the Capture threshold
						p.moves[p.swapIx], p.moves[ix] = p.moves[ix], p.moves[p.swapIx]
						p.swapIx++
					}
				}
			}

			if p.swapIx == p.yix {
				p.phase = weighNonCaptures
			}

		case goodCaptures:
			maxim := -Inf - 1
			best := -1

			for ix := p.yix; ix < p.swapIx; ix++ {
				if maxim < p.moves[ix].Weight {
					maxim = p.moves[ix].Weight
					best = ix
				}
			}
			if best != -1 && maxim >= heur.Captures {
				result = p.yield(best)
				success = true
			} else {
				p.phase = weighNonCaptures
			}

		case weighNonCaptures:
			p.phase = rest

			for ix := p.yix; ix < len(p.moves); ix++ {
				m := &p.moves[ix]

				if m.Weight == 0 {

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

			for ix := p.yix; ix < len(p.moves); ix++ {
				if maxim < p.moves[ix].Weight {
					maxim = p.moves[ix].Weight
					best = ix
				}
			}
			if best != -1 {
				result = p.yield(best)
			}
			success = true
		}
	}

	return result
}

func (p *Picker) yield(ix int) *move.Move {
	p.moves[p.yix], p.moves[ix] = p.moves[ix], p.moves[p.yix]
	result := &p.moves[p.yix]
	p.yix++
	return result
}
