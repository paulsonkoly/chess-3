package search

import (
	"fmt"
	"strings"
	"time"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/eval"
	"github.com/paulsonkoly/chess-3/heur"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"
	"github.com/paulsonkoly/chess-3/transp"

	. "github.com/paulsonkoly/chess-3/types"
)

const (
	NMPDiffFactor = Score(51)
	NMPDepthLimit = Depth(1)
	NMPInit       = Depth(4)

	RFPDepthLimit = Depth(8)
)

const (
	WindowSize = 50 // half a pawn left and right around score
)

func (s *Search) WithOptions(b *board.Board, d Depth, opts ...Option) (score Score, move move.SimpleMove) {
	s.refresh()
	defer func() {
		s.gen++
	}()

	options := options{}
	for _, opt := range opts {
		opt(&options)
	}

	if options.counters == nil {
		options.counters = &Counters{}
	}

	return s.iterativeDeepen(b, d, &options)
}

// iterativeDeepen is the main entry point to the engine. It performs and
// iterative-deepened alpha-beta with aspiration window. depth is iterated
// between 0 and d inclusive.
func (s *Search) iterativeDeepen(b *board.Board, d Depth, opts *options) (score Score, move move.SimpleMove) {
	// otherwise a checkmate score would always fail high
	alpha := -Inf - 1
	beta := Inf + 1

	start := time.Now()

	for idD := range d + 1 {
		awOk := false // aspiration window succeeded
		factor := Score(1)
		var scoreSample Score

		for !awOk {
			scoreSample = s.alphaBeta(b, alpha, beta, idD, 0, PVNode, opts)

			switch {

			case scoreSample <= alpha:
				opts.counters.AWFail++
				alpha -= factor * WindowSize
				factor *= 2

			case scoreSample >= beta:
				opts.counters.AWFail++
				beta += factor * WindowSize
				factor *= 2

			default:
				awOk = true
			}

			if abort(opts) {
				// we hit hard timeout, and we don't have a move. We try to return the
				// PV move if we have it, regardless of its quality, if there is none
				// we return the first legal move we find. If there is none, we return
				// null move.
				if move == 0 {
					// there is no previous move, so we produce something from the
					// aborted search. The scoreSample is our best bet at this point.
					score = scoreSample
					if len(s.pv.active()) > 0 {
						move = s.pv.active()[0]
						return
					}

					s.ms.Push()
					defer s.ms.Pop()

					movegen.GenMoves(s.ms, b)
					moves := s.ms.Frame()

					for _, pseudo := range moves {
						b.MakeMove(&pseudo)
						if !movegen.InCheck(b, b.STM.Flip()) { // legal
							move = pseudo.SimpleMove
							b.UndoMove(&pseudo)
							break
						}
						b.UndoMove(&pseudo)
					}
				}
				return
			}
		}
		score = scoreSample
		if len(s.pv.active()) > 0 {
			move = s.pv.active()[0]
		}

		elapsed := time.Since(start)
		miliSec := elapsed.Milliseconds()
		cnts := opts.counters
		cnts.Time = miliSec
		fmt.Printf("info depth %d score %s nodes %d time %d hashfull %d pv %s\n",
			idD, scInfo(score), cnts.ABCnt+cnts.QCnt, miliSec, s.tt.HashFull(s.gen), pvInfo(s.pv.active()))

		if opts.debug {
			ABBF := float64(cnts.ABBreadth) / float64(cnts.ABCnt)

			fmt.Printf("info awfail %d ableaf %d abbf %.2f tthits %d qdepth %d\n",
				cnts.AWFail, cnts.ABLeaf, ABBF, cnts.TTHit, cnts.QDepth)
		}

		if move != 0 && (opts.softTime > 0 && miliSec > opts.softTime) {
			return
		}

		alpha = score - WindowSize
		beta = score + WindowSize
	}
	return
}

func abort(opts *options) bool {
	if opts.stop != nil {
		select {
		case <-opts.stop:
			opts.abort = true
			return true
		default:
		}
	}
	return opts.abort
}

func pvInfo(moves []move.SimpleMove) string {
	sb := strings.Builder{}
	space := ""
	for _, m := range moves {
		sb.WriteString(space)
		sb.WriteString(fmt.Sprint(m))
		space = " "
	}
	return sb.String()
}

func scInfo(score Score) string {
	if Abs(score) >= Inf-MaxPlies {
		diff := Inf - score

		if score < 0 {
			diff = -score - Inf
		}

		return fmt.Sprintf("mate %d", (diff+1)/2)
	}

	return fmt.Sprintf("cp %d", score)
}

// Node is the predicted type of the node.
type Node = byte

const (
	// PVNode expects the score to be in the window.
	PVNode Node = iota
	// CutNode expects the node to fail high.
	CutNode
	// AllNode expects the node to fail low.
	AllNode
)

// AlphaBeta performs an alpha beta search to depth d, and then transitions
// into Quiesence() search.
func (s *Search) alphaBeta(b *board.Board, alpha, beta Score, d, ply Depth, nType Node, opts *options) Score {

	transpT := s.tt
	s.pv.setNull(ply)

	tfCnt := b.Threefold()
	// this condition is trying to avoid returning 0 move on ply 0 if it's the second repetation
	if b.FiftyCnt >= 100 || tfCnt >= 3-min(ply, 1) {
		return 0
	}

	if transpE, ok := transpT.LookUp(b.Hash()); ok && transpE.Depth() >= d {
		opts.counters.TTHit++

		tpVal := transpE.Value(ply)

		switch transpE.Type() {

		case transp.Exact:
			if transpE.SimpleMove != 0 {
				s.pv.setTip(ply, transpE.SimpleMove)
			}
			return tpVal

		case transp.LowerBound:
			if tpVal >= beta {
				return tpVal
			}

		case transp.UpperBound:
			if tpVal <= alpha {
				return tpVal
			}
		}
		opts.counters.TTHit--
	}

	if d == 0 {
		opts.counters.ABLeaf++
		return s.quiescence(b, alpha, beta, 0, ply, opts)
	}

	opts.counters.ABCnt++

	inCheck := movegen.InCheck(b, b.STM)
	improving := false
	staticEval := Inv

	if !inCheck {
		staticEval = eval.Eval(b, &eval.Coefficients)

		improving = s.hstack.oldScore() < staticEval

		// RFP
		if d < RFPDepthLimit && staticEval >= beta+Score(d)*105 && beta > -Inf+MaxPlies {
			return staticEval
		}

		// null move pruning
		if d > NMPDepthLimit && staticEval >= beta && b.Colors[b.STM] & ^(b.Pieces[Pawn]|b.Pieces[King]) != 0 {

			enP := b.MakeNullMove()

			r := Depth(NMPInit)

			if improving {
				r++
			}

			r += Depth(Clamp((staticEval-beta)/NMPDiffFactor, 0, MaxPlies))

			value := -s.alphaBeta(b, -beta, -beta+1, max(d-r, 0), ply+1, CutNode, opts)

			b.UndoNullMove(enP)

			// In case the null move search left a PV fragment this removes it,
			// normally, it shouldn't matter because it's expected to fail low higher
			// up anyway. But going up the window can widen, and it would be possible
			// that a nullmove line bubbles up.
			s.pv.setNull(ply)

			if value >= beta {
				if value >= Inf-MaxPlies {
					return beta
				}

				return value
			}
		}
	}

	s.ms.Push()
	defer s.ms.Pop()

	movegen.GenMoves(s.ms, b)
	moves := s.ms.Frame()

	s.rankMovesAB(b, moves)

	var (
		m        *move.Move
		ix       int
		bestMove move.SimpleMove
	)

	hasLegal := false
	failLow := true
	maxim := -Inf - 1
	moveCnt := 0
	quietCnt := 0

	for m, ix = getNextMove(moves, -1); m != nil; m, ix = getNextMove(moves, ix) {

		b.MakeMove(m)

		if movegen.InCheck(b, b.STM.Flip()) {
			b.UndoMove(m)
			continue
		}

		hasLegal = true
		moveCnt++

		s.hstack.push(m.Piece, m.To(), staticEval)

		var value Score

		quiet := m.Captured == NoPiece && m.Promo() == NoPiece
		if quiet {
			quietCnt++
		}

		next := nextNodeType(nType, moveCnt)

		// Late move reduction and null-window search. Skip it on the first legal
		// move, which is likely to be the hash move.
		fullSearched := false
		if d > 1 && quietCnt > 2 && !inCheck {
			rd := lmr(d, moveCnt-1, improving, nType)

			// reduced depth first, then re-try with full depth and null window.
			if rd < d-1 {
				value = -s.alphaBeta(b, -alpha-1, -alpha, rd, ply+1, next, opts)
			}

			if value <= alpha {
				goto Fin
			}

			value = -s.alphaBeta(b, -alpha-1, -alpha, d-1, ply+1, next, opts)

			if value <= alpha {
				goto Fin
			}

			// if null window is the full window
			fullSearched = beta == alpha+1
		}

		// null window search failed (meaning didn't fail low).
		if !fullSearched {
			value = -s.alphaBeta(b, -beta, -alpha, d-1, ply+1, next, opts)
		}

	Fin:

		b.UndoMove(m)
		s.hstack.pop()

		if value > maxim {
			maxim = value
		}

		if value > alpha {
			if value >= beta {
				// store node as fail high (cut-node)
				transpT.Insert(b.Hash(), s.gen, d, ply, m.SimpleMove, value, transp.LowerBound)

				hSize := s.hstack.size()
				bonus := -(Score(d)*20 - 15)

				for i, m := range moves {
					if i == ix {
						bonus = -bonus
					}

					if m.Captured == NoPiece && m.Promo() == NoPiece {
						s.hist.Add(b.STM, m.From(), m.To(), bonus)

						if hSize >= 1 {
							hist := s.hstack.top(0)
							s.cont[0].Add(b.STM, hist.piece, hist.to, m.Piece, m.To(), bonus)
						}

						if hSize >= 2 {
							hist := s.hstack.top(1)
							s.cont[1].Add(b.STM, hist.piece, hist.to, m.Piece, m.To(), bonus)
						}
					}

					if i == ix {
						break
					}
				}

				opts.counters.ABBreadth += moveCnt

				return value
			}

			// value > alpha
			failLow = false
			alpha = value
			bestMove = m.SimpleMove
			s.pv.insert(ply, m.SimpleMove)
		}

		// LMP
		quietLimit := int(d) * int(d)
		if !improving {
			quietLimit /= 2
		}
		if !inCheck && alpha+1 == beta && quietCnt > 1+quietLimit {
			break
		}

		if abort(opts) {
			return maxim
		}
	}

	opts.counters.ABBreadth += moveCnt

	if !hasLegal {
		maxim = Score(0)

		if inCheck {
			maxim = -Inf + Score(ply)
		}

		failLow = false
	}

	if failLow {
		// store node as fail low (All-node)
		transpT.Insert(b.Hash(), s.gen, d, ply, 0, maxim, transp.UpperBound)
	} else {
		transpT.Insert(b.Hash(), s.gen, d, ply, bestMove, maxim, transp.Exact)
	}

	return maxim
}

func nextNodeType(nType Node, cnt int) Node {
	switch nType {
	case PVNode:
		if cnt == 1 {
			return PVNode
		} else {
			return CutNode
		}

	case CutNode:
		if cnt == 1 {
			return AllNode
		} else {
			return CutNode
		}

	case AllNode:
		return CutNode
	}

	return CutNode
}

// (1..300).map {|i| (Math.log2(i) * 69).round }.each_slice(10) {|a| puts a.join(", ") }
var log = [...]int{
	0,
	0, 69, 109, 138, 160, 179, 194, 207, 219, 230,
	239, 248, 256, 263, 270, 277, 283, 289, 294, 299,
	304, 309, 313, 317, 321, 325, 329, 333, 336, 340,
	343, 346, 349, 352, 355, 358, 361, 363, 366, 368,
	371, 373, 376, 378, 380, 382, 385, 387, 389, 391,
	393, 395, 397, 398, 400, 402, 404, 406, 407, 409,
	411, 412, 414, 415, 417, 418, 420, 421, 423, 424,
	426, 427, 429, 430, 431, 433, 434, 435, 436, 438,
	439, 440, 441, 443, 444, 445, 446, 447, 448, 449,
	451, 452, 453, 454, 455, 456, 457, 458, 459, 460,
}

// x = (1..200).map {|i| (Math.log2(i) * 69).round }.unshift(0)
// 10.times.map {|d| 30.times.map {|m| (x[d] * x[m] )>>14}}
func lmr(d Depth, mCount int, improving bool, nType Node) Depth {
	value := (log[d] * log[min(mCount, len(log)-1)]) >> 14

	// if !quiet {
	// 	value /= 2
	// }
	//

	if nType != PVNode {
		value++
	}

	if !improving {
		value++
	}

	return Clamp(d-1-Depth(value), 0, d-1)
}

// Quiescence resolves the position to a quiet one, and then evaluates.
func (s *Search) quiescence(b *board.Board, alpha, beta Score, d, ply Depth, opts *options) Score {
	if d > opts.counters.QDepth {
		opts.counters.QDepth = d
	}

	opts.counters.QCnt++

	if b.FiftyCnt >= 100 || b.Threefold() >= 3 {
		return 0
	}

	inCheck := movegen.InCheck(b, b.STM)

	if inCheck {
		if movegen.IsCheckmate(b) {
			return -Inf + Score(ply)
		}
	} else {
		if movegen.IsStalemate(b) {
			return 0
		}
	}

	standPat := eval.Eval(b, &eval.Coefficients)

	if !inCheck && standPat >= beta {
		return standPat
	}

	s.ms.Push()
	defer s.ms.Pop()

	movegen.GenForcing(s.ms, b)

	delta := standPat + 110
	// fail soft upper bound
	maxim := standPat
	alpha = max(alpha, standPat)

	moves := s.ms.Frame()

	rankMovesQ(b, moves)

	for m, ix := getNextMove(moves, -1); m != nil; m, ix = getNextMove(moves, ix) {

		if m.Weight < 0 {
			break
		}

		b.MakeMove(m)

		if movegen.InCheck(b, b.STM.Flip()) {
			b.UndoMove(m)
			continue
		}

		gain := heur.PieceValues[m.Captured]

		if m.Promo() != NoPiece {
			gain += heur.PieceValues[m.Promo()] - heur.PieceValues[Pawn]
		}

		if gain+delta < alpha {
			b.UndoMove(m)
			break
		}

		curr := -s.quiescence(b, -beta, -alpha, d+1, ply+1, opts)
		b.UndoMove(m)

		if curr >= beta {
			return curr
		}
		maxim = max(maxim, curr)
		alpha = max(alpha, curr)

		if abort(opts) {
			return maxim
		}
	}

	return maxim
}

func (s *Search) rankMovesAB(b *board.Board, moves []move.Move) {
	transPE, _ := s.tt.LookUp(b.Hash())

	for ix, m := range moves {

		switch {
		case transPE != nil && transPE.Matches(&m):
			moves[ix].Weight = heur.HashMove

		case b.SquaresToPiece[m.To()] != NoPiece || m.Promo() != NoPiece:
			see := heur.SEE(b, &m)
			if see < 0 {
				moves[ix].Weight = see - heur.Captures
			} else {
				moves[ix].Weight = see + heur.Captures
			}

		default:
			score := s.hist.LookUp(b.STM, m.From(), m.To())

			if s.hstack.size() >= 1 {
				hist := s.hstack.top(0)
				score += 3 * s.cont[0].LookUp(b.STM, hist.piece, hist.to, m.Piece, m.To())
			}

			if s.hstack.size() >= 2 {
				hist := s.hstack.top(1)
				score += 2 * s.cont[1].LookUp(b.STM, hist.piece, hist.to, m.Piece, m.To())
			}

			moves[ix].Weight = score
		}
	}
}

func rankMovesQ(b *board.Board, moves []move.Move) {
	for ix, m := range moves {
		switch {

		case b.SquaresToPiece[m.To()] != NoPiece:
			see := heur.SEE(b, &m)
			moves[ix].Weight = see

		case m.Promo() != NoPiece:
			moves[ix].Weight = 0

		default:
			moves[ix].Weight = -1
		}
	}
}

func getNextMove(moves []move.Move, ix int) (*move.Move, int) {
	maxim := -Inf - 1
	best := -1
	for jx := ix + 1; jx < len(moves); jx++ {
		if maxim < moves[jx].Weight {
			maxim = moves[jx].Weight
			best = jx
		}
	}

	if best == -1 {
		return nil, ix
	}
	ix++

	moves[ix], moves[best] = moves[best], moves[ix]

	return &moves[ix], ix
}
