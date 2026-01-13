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
	"github.com/paulsonkoly/chess-3/params"
	"github.com/paulsonkoly/chess-3/picker"
	"github.com/paulsonkoly/chess-3/transp"

	. "github.com/paulsonkoly/chess-3/chess"
)

// Go is the main entry point into the engine. It kicks off the search on board
// b, and returns the final score and best move.
func (s *Search) Go(b *board.Board, opts ...Option) (score Score, move move.Move) {
	s.refresh()
	defer func() {
		s.gen++
	}()

	options := options{depth: MaxPlies, nodes: -1, softNodes: -1, info: true}
	for _, opt := range opts {
		opt(&options)
	}

	if options.counters == nil {
		options.counters = &Counters{}
	}

	return s.iterativeDeepen(b, &options)
}

// iterativeDeepen performs an iterative-deepened alpha-beta with aspiration
// window. depth is iterated between 0 and d inclusive.
func (s *Search) iterativeDeepen(b *board.Board, opts *options) (score Score, move move.Move) {
	// otherwise a checkmate score would always fail high
	alpha := -Inf - 1
	beta := Inf + 1

	start := time.Now()

	for idD := range opts.depth + 1 {
		awOk := false // aspiration window succeeded
		factor := Score(1)
		var scoreSample Score

		for !awOk {
			scoreSample = s.alphaBeta(b, alpha, beta, idD, 0, PVNode, opts)

			switch {

			case scoreSample <= alpha:
				alpha -= factor * Score(params.WindowSize)
				factor *= 2

			case scoreSample >= beta:
				beta += factor * Score(params.WindowSize)
				factor *= 2

			default:
				awOk = true
			}

			if s.abort(opts) {
				// have a final node count for debugging purposes
				if opts.info {
					fmt.Printf("info depth %d nodes %d\n", idD, opts.counters.Nodes)
				}

				// we hit hard timeout/abort and we don't have a move. We try to return
				// the first legal move, regardless of its quality, if there is none we
				// return null move.
				if move == 0 {
					s.ms.Push()
					defer s.ms.Pop()

					movegen.GenNoisy(s.ms, b)
					movegen.GenNotNoisy(s.ms, b)
					moves := s.ms.Frame()

					for _, pseudo := range moves {
						r := b.MakeMove(pseudo.Move)
						if !b.InCheck(b.STM.Flip()) { // legal
							move = pseudo.Move
							b.UndoMove(pseudo.Move, r)
							break
						}
						b.UndoMove(pseudo.Move, r)
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
		if opts.info {
			fmt.Printf("info depth %d score %s nodes %d time %d hashfull %d pv %s\n",
				idD, score, cnts.Nodes, miliSec, s.tt.HashFull(s.gen), pvInfo(s.pv.active()))
		}

		if move != 0 && opts.softAbort(miliSec, opts.counters.Nodes) {
			return
		}

		alpha = score - Score(params.WindowSize)
		beta = score + Score(params.WindowSize)
	}
	return
}

func (s *Search) abort(opts *options) bool {
	if s.aborted {
		return true
	}
	if opts.stop != nil {
		select {
		case <-opts.stop:
			s.aborted = true
			return true
		default:
		}
	}
	return false
}

// incrementNodes increments node count in opts.counters, except if it would
// overrun the alloted nodes in which case it sets abort.
func (s *Search) incrementNodes(opts *options) {
	if opts.nodes == -1 || opts.counters.Nodes < opts.nodes {
		opts.counters.Nodes++
	} else {
		s.aborted = true
	}
}

func pvInfo(moves []move.Move) string {
	sb := strings.Builder{}
	space := ""
	for _, m := range moves {
		sb.WriteString(space)
		fmt.Fprint(&sb, m)
		space = " "
	}
	return sb.String()
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
	s.pv.setNull(ply)

	if d == 0 || ply >= MaxPlies-1 {
		return s.quiescence(b, alpha, beta, ply, opts)
	}

	s.incrementNodes(opts)
	opts.counters.ABNodes ++

	if s.abort(opts) {
		return Inv
	}

	tfCnt := b.Threefold()
	// this condition is trying to avoid returning 0 move on ply 0 if it's the second repetition
	if b.FiftyCnt >= 100 || tfCnt >= 3-min(ply, 1) {
		return 0
	}

	var hashMove move.Move
	if transpE, ok := s.tt.LookUp(b.Hash()); ok {
		hashMove = transpE.Move

		if nType != PVNode && transpE.Depth() >= d {
			tpVal := transpE.Value(ply)

			switch transpE.Type() {

			case transp.Exact:
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
		}
	}

	inCheck := b.InCheck(b.STM)
	improving := false
	staticEval := Inv

	if !inCheck {
		staticEval = eval.Eval(b, &eval.Coefficients)

		oldScore := Inv
		if old, ok := s.hstack.Top(1); ok && old.Score != Inv {
			oldScore = old.Score
		} else if old, ok := s.hstack.Top(3); ok {
			oldScore = old.Score
		}

		improving = oldScore < staticEval

		// RFP
		if d < Depth(params.RFPDepthLimit) &&
			staticEval >= beta+Score(d)*Score(params.RFPScoreFactor) &&
			beta > -Inf+MaxPlies {
			return staticEval
		}

		// null move pruning
		if d > Depth(params.NMPDepthLimit) &&
			staticEval >= beta &&
			b.Colors[b.STM] & ^(b.Pieces[Pawn]|b.Pieces[King]) != 0 {

			rev := b.MakeNullMove()

			red := Depth(params.NMPInit) + Depth(Clamp((staticEval-beta)/Score(params.NMPDiffFactor), 0, MaxPlies))

			value := -s.alphaBeta(b, -beta, -beta+1, max(d-red, 0), ply+1, CutNode, opts)

			b.UndoNullMove(rev)

			if value >= beta {
				if value >= Inf-MaxPlies {
					return beta
				}

				return value
			}
		}
	}

	pck := picker.New(b, hashMove, s.ms, &s.ranker, s.hstack)
	s.ms.Push()
	defer s.ms.Pop()

	var (
		bestMove move.Move
	)

	hasLegal := false
	failLow := true
	maxim := -Inf - 1
	moveCnt := 0
	quietCnt := 0

	for pck.Next() {
		w := pck.Move()
		m := w.Move

		moved := b.SquaresToPiece[m.From()]
		captured := b.SquaresToPiece[b.CaptureSq(m)]

		r := b.MakeMove(m)

		if b.InCheck(b.STM.Flip()) {
			b.UndoMove(m, r)
			continue
		}

		hasLegal = true
		moveCnt++

		s.hstack.Push(heur.StackMove{Piece: moved, To: m.To(), Score: staticEval})

		var value Score

		quiet := captured == NoPiece && m.Promo() == NoPiece
		if quiet {
			quietCnt++
		}

		next := nextNodeType(nType, moveCnt)

		// Late move reduction and null-window search. Skip it on the first legal
		// move, which is likely to be the hash move.
		fullSearched := false
		if d > 1 && quietCnt > params.LMRStart && !inCheck {
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

		b.UndoMove(m, r)
		s.hstack.Pop()

		if value > maxim {
			maxim = value
		}

		// it is important that we check abort *before* updating any of the
		// persistent states, for being able to replicate previous runs with go
		// nodes
		if s.abort(opts) {
			return Inv
		}

		if value > alpha {
			if value >= beta {
				// store node as fail high (cut-node)
				s.tt.Insert(b.Hash(), s.gen, d, ply, m, value, transp.LowerBound)
				s.ranker.FailHigh(d, b, pck.YieldedMoves(), s.hstack)
				opts.counters.Moves += moveCnt
				if moveCnt == 1 {
					opts.counters.FirstCut++
				}

				return value
			}

			// value > alpha
			w.Weight = value
			failLow = false
			alpha = value
			bestMove = m
			s.pv.insert(ply, m)
		} else {
			// upbound move, this will be useful in history penalties
			w.Weight = -Inf
		}

		// LMP
		quietLimit := int(d) * int(d)
		if !improving {
			quietLimit /= 2
		}
		if !inCheck && alpha+1 == beta && quietCnt > 1+quietLimit {
			break
		}
	}

	opts.counters.Moves += moveCnt

	if !hasLegal {
		maxim = Score(0)

		if inCheck {
			maxim = -Inf + Score(ply)
		}

		failLow = false
	}

	if failLow {
		// store node as fail low (All-node)
		s.tt.Insert(b.Hash(), s.gen, d, ply, 0, maxim, transp.UpperBound)
	} else {
		s.tt.Insert(b.Hash(), s.gen, d, ply, bestMove, maxim, transp.Exact)
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

// log is precomputed logarithmic scale.
//
// can be reproduced with:
//
//	(1..300).map {|i| (Math.log2(i) * 69).round }.each_slice(10) {|a| puts a.join(", ") }
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

// lmr is late move depth reduction.
//
// check values with:
//
//	x = (1..200).map {|i| (Math.log2(i) * 69).round }.unshift(0)
//	10.times.map {|d| 30.times.map {|m| (x[d] * x[m] )>>14}}
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
func (s *Search) quiescence(b *board.Board, alpha, beta Score, ply Depth, opts *options) Score {

	s.incrementNodes(opts)

	if s.abort(opts) {
		return Inv
	}

	if b.FiftyCnt >= 100 || b.Threefold() >= 3 {
		return 0
	}

	transpT := s.tt
	if transpE, ok := transpT.LookUp(b.Hash()); ok {
		tpVal := transpE.Value(ply)

		switch transpE.Type() {

		case transp.Exact:
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
	}

	inCheck := b.InCheck(b.STM)

	if inCheck {
		if b.IsCheckmate() {
			return -Inf + Score(ply)
		}
	} else {
		if b.IsStalemate() {
			return 0
		}
	}

	standPat := eval.Eval(b, &eval.Coefficients)

	if !inCheck && standPat >= beta {
		return standPat
	}

	s.ms.Push()
	defer s.ms.Pop()

	movegen.GenNoisy(s.ms, b)

	delta := standPat + Score(params.StandPatDelta)
	// fail soft upper bound
	maxim := standPat
	alpha = max(alpha, standPat)

	moves := s.ms.Frame()

	s.rankMovesQ(b, moves)

	for m, ix := getNextMove(moves, -1); m != nil; m, ix = getNextMove(moves, ix) {

		if m.Weight < 0 {
			break
		}

		captured := b.SquaresToPiece[b.CaptureSq(m.Move)]

		r := b.MakeMove(m.Move)

		if b.InCheck(b.STM.Flip()) {
			b.UndoMove(m.Move, r)
			continue
		}

		gain := heur.PieceValues[captured]

		if m.Promo() != NoPiece {
			gain += heur.PieceValues[m.Promo()] - heur.PieceValues[Pawn]
		}

		if gain+delta < alpha {
			b.UndoMove(m.Move, r)
			break
		}

		curr := -s.quiescence(b, -beta, -alpha, ply+1, opts)
		b.UndoMove(m.Move, r)

		if curr >= beta {
			transpT.Insert(b.Hash(), s.gen, 0, ply, m.Move, curr, transp.LowerBound)
			return curr
		}
		maxim = max(maxim, curr)
		alpha = max(alpha, curr)

		if s.abort(opts) {
			return Inv
		}
	}

	transpT.Insert(b.Hash(), s.gen, 0, ply, 0, maxim, transp.UpperBound)

	return maxim
}

func (s *Search) rankMovesQ(b *board.Board, moves []move.Weighted) {
	for ix, m := range moves {
		moves[ix].Weight = s.ranker.RankNoisy(m.Move, b, s.hstack)
	}
}

func getNextMove(moves []move.Weighted, ix int) (*move.Weighted, int) {
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
