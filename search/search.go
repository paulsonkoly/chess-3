package search

import (
	"fmt"
	"slices"
	"time"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/eval"
	"github.com/paulsonkoly/chess-3/heur"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"
	"github.com/paulsonkoly/chess-3/transp"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

const (
	WindowSize = 50 // half a pawn left and right around score
)

// Search is a persistent state storage between searches.
type Search struct {
	tt     *transp.Table
	hist   *heur.History
	cont   [2]*heur.Continuation
	ms     *move.Store
	hstack *historyStack

	Debug bool // Debug determines if additional debug info output is enabled.

	// Stop channel signals an immediate Stop request to the search. Current
	// depth will be abandoned.
	Stop chan struct{}

	aborted bool

	AWFail int // AwFail is the count of times the score fell outside of the aspiration window.
	ABLeaf int // ABLeaf is the count of alpha-beta leafs.
	// ABBreadth is the total count of explored moves in alpha-beta. Thus
	// (ABBreadth / ABCnt) is the average alpha-beta branching factor.
	ABBreadth int
	ABCnt     int   // ABCnt is the inner node count in alpha-beta.
	TTHit     int   // TThit is the transposition table hit-count.
	QCnt      int   // Quiesence node count
	QDepth    int   // QDepth is the maximal quiesence search depth.
	QDelta    int   // QDelta is the count of times a delta pruning happened in quiesence search.
	QSEE      int   // QSEE is the count of times the static exchange evaluation fell under 0 in quiesence search.
	Time      int64 // Time is the search time in milliseconds.
}

// New creates a new search state. It's supposed to be called once, and
// re-used between Go() calls.
func New() *Search {
	return &Search{
		tt:     transp.New(),
		ms:     move.NewStore(),
		hist:   heur.NewHistory(),
		cont:   [2]*heur.Continuation{heur.NewContinuation(), heur.NewContinuation()},
		hstack: newHistStack(),
	}
}

// Clear resets the counters, and various stores for the search, assuming a new
// position.
func (s *Search) Clear() {
	s.aborted = false
	s.tt.Clear()
	s.ms.Clear()
	s.hstack.reset()
	s.AWFail = 0
	s.ABLeaf = 0
	s.ABBreadth = 0
	s.ABCnt = 0
	s.TTHit = 0
	s.QCnt = 0
	s.QDepth = 0
	s.QDelta = 0
	s.QSEE = 0
}

// Go is the main entry point to the engine. It performs and
// iterative-deepened alpha-beta with aspiration window.
func (s *Search) Go(b *board.Board, d Depth) (score Score, moves []move.SimpleMove) {
	// otherwise a checkmate score would always fail high
	alpha := -Inf - 1
	beta := Inf + 1

	start := time.Now()

	s.Clear()

	for d := range d + 1 { // +1 for 0 depth search (quiesence eval)
		awOk := false // aspiration window succeeded
		factor := Score(1)
		var (
			scoreSample Score
			movesSample []move.SimpleMove
		)

		for !awOk {
			scoreSample, movesSample = s.AlphaBeta(b, alpha, beta, d)

			switch {

			case scoreSample <= alpha:
				s.AWFail++
				alpha -= factor * WindowSize
				factor *= 2

			case scoreSample >= beta:
				s.AWFail++
				beta += factor * WindowSize
				factor *= 2

			default:
				awOk = true
			}

			if s.abort() {
				return
			}
		}
		score, moves = scoreSample, movesSample
		slices.Reverse(moves)

		elapsed := time.Since(start)
		miliSec := elapsed.Milliseconds()
		s.Time = miliSec
		fmt.Printf("info depth %d score cp %d nodes %d time %d pv %s\n",
			d, score, s.ABCnt+s.ABLeaf+s.QCnt, miliSec, pvInfo(moves))

		if s.Debug {
			ABBF := float64(s.ABBreadth) / float64(s.ABCnt)

			fmt.Printf("info awfail %d ableaf %d abbf %.2f tthits %d qdepth %d qdelta %d qsee %d\n",
				s.AWFail, s.ABLeaf, ABBF, s.TTHit, s.QDepth, s.QDelta, s.QSEE)
		}

		alpha = score - WindowSize
		beta = score + WindowSize
	}
	return
}

// AlphaBeta performs an alpha beta search to depth d, and then transitions
// into Quiesence() search.
func (s *Search) AlphaBeta(b *board.Board, alpha, beta Score, d Depth) (Score, []move.SimpleMove) {

	transpT := s.tt
	pv := []move.SimpleMove{}

	tfCnt := b.Threefold()
	if b.FiftyCnt >= 100 || tfCnt >= 3 {
		return 0, pv
	}

	if transpE, ok := transpT.LookUp(b.Hash()); ok && transpE.Depth >= d && transpE.TFCnt >= tfCnt {
		s.TTHit++
		switch transpE.Type {

		case transp.PVNode:
			if transpE.From|transpE.To != 0 {
				pv = []move.SimpleMove{transpE.SimpleMove}
			}
			return transpE.Value, pv

		case transp.CutNode:
			if transpE.Value >= beta {
				return transpE.Value, pv
			}

		case transp.AllNode:
			if transpE.Value <= alpha {
				return transpE.Value, pv
			}
		}
		s.TTHit--
	}

	if d == 0 {
		s.ABLeaf++
		return s.Quiescence(b, alpha, beta, 0), pv
	}

	s.ABCnt++

	hasLegal := false
	failLow := true

	inCheck := movegen.InCheck(b, b.STM)

	// RFP
	if !inCheck {
		staticEval := eval.Eval(b, alpha, beta, &eval.Coefficients)

		if staticEval >= beta+Score(d)*105 {
			return staticEval, pv
		}
	}

	// null move pruning
	if !inCheck && b.Colors[b.STM] & ^(b.Pieces[Pawn]|b.Pieces[King]) != 0 {

		enP := b.MakeNullMove()

		rd := max(0, d-3)

		value, _ := s.AlphaBeta(b, -beta, -beta+1, rd)
		value *= -1

		b.UndoNullMove(enP)

		if value >= beta {
			return value, pv
		}
	}

	// deflate history
	if s.ABCnt%10_000 == 0 {
		s.hist.Deflate()
		s.cont[0].Deflate()
		s.cont[1].Deflate()
	}

	s.ms.Push()
	defer s.ms.Pop()

	movegen.GenMoves(s.ms, b)
	moves := s.ms.Frame()

	s.rankMovesAB(b, moves)

	var (
		m  *move.Move
		ix int
	)

	for m, ix = getNextMove(moves, -1); m != nil; m, ix = getNextMove(moves, ix) {

		b.MakeMove(m)

		if movegen.InCheck(b, b.STM.Flip()) {
			b.UndoMove(m)
			continue
		}

		hasLegal = true

		s.hstack.push(m.Piece, m.To)

		// late move reduction
		rd := lmr(d, ix)
		if rd < d-1 && !inCheck {
			value, _ := s.AlphaBeta(b, -alpha-1, -alpha, rd)
			value *= -1

			if value <= alpha {
				b.UndoMove(m)
				s.hstack.pop()
				continue
			}
		}

		value, curr := s.AlphaBeta(b, -beta, -alpha, d-1)
		value *= -1
		b.UndoMove(m)
		s.hstack.pop()

		if value > alpha {
			failLow = false
			alpha = value
			pv = append(curr, m.SimpleMove)
		}

		if value >= beta {
			// store node as fail high (cut-node)
			transpT.Insert(b.Hash(), d, tfCnt, m.SimpleMove, value, transp.CutNode)

			hSize := s.hstack.size()
			bonus := -Score(d * d)

			for i, m := range moves {
				if i == ix {
					bonus = -bonus
				}

				if m.Captured == NoPiece && m.Promo == NoPiece {
					s.hist.Add(b.STM, m.From, m.To, bonus)

					if hSize >= 1 {
						pt, to := s.hstack.top(0)
						s.cont[0].Add(b.STM, pt, to, m.Piece, m.To, bonus)
					}

					if hSize >= 2 {
						pt, to := s.hstack.top(1)
						s.cont[1].Add(b.STM, pt, to, m.Piece, m.To, bonus)
					}
				}

				if i == ix {
					break
				}
			}

			s.ABBreadth += ix + 1

			return value, nil
		}

		if s.abort() {
			return alpha, pv
		}
	}

	s.ABBreadth += ix + 1

	if !hasLegal {
		value := Score(0)

		if inCheck {
			value = -Inf
		}

		if value > alpha {
			failLow = false
			alpha = value
		}
	}

	if failLow {
		// store node as fail low (All-node)
		transpT.Insert(b.Hash(), d, tfCnt, move.SimpleMove{}, alpha, transp.AllNode)
	} else {
		// store node as exact (PV-node)
		// there might not be a move in case of !hasLegal
		var sm move.SimpleMove
		if len(pv) > 0 {
			sm = pv[len(pv)-1]
		}

		transpT.Insert(b.Hash(), d, tfCnt, sm, alpha, transp.PVNode)
	}

	return alpha, pv
}

// Quiescence resolves the position to a quiet one, and then evaluates.
func (s *Search) Quiescence(b *board.Board, alpha, beta Score, d int) Score {
	if d > s.QDepth {
		s.QDepth = d
	}

	s.QCnt++

	if b.FiftyCnt >= 100 || b.Threefold() >= 3 {
		return 0
	}

	if movegen.IsCheckmate(b) {
		return -Inf
	}

	if movegen.IsStalemate(b) {
		return 0
	}

	s.ms.Push()
	defer s.ms.Pop()

	movegen.GenForcing(s.ms, b)

	standPat := eval.Eval(b, alpha, beta, &eval.Coefficients)

	if standPat >= beta {
		return beta
	}

	delta := standPat + 110
	alpha = max(alpha, standPat)

	moves := s.ms.Frame()

	rankMovesQ(b, moves)

	for m, ix := getNextMove(moves, -1); m != nil; m, ix = getNextMove(moves, ix) {

		b.MakeMove(m)

		if movegen.InCheck(b, b.STM.Flip()) {
			b.UndoMove(m)
			continue
		}

		check := movegen.InCheck(b, b.STM)

		if !check {
			if m.Captured == NoPiece && m.Promo == NoPiece {
				b.UndoMove(m)
				continue
			}

			gain := heur.PieceValues[m.Captured]

			if m.Promo != NoPiece {
				gain += heur.PieceValues[m.Promo] - heur.PieceValues[Pawn]
			}

			if gain+delta < alpha {
				s.QDelta++
				b.UndoMove(m)
				continue
			}

			if m.Weight < 0 {
				s.QSEE++
				b.UndoMove(m)
				continue
			}
		}

		curr := -s.Quiescence(b, -beta, -alpha, d+1)
		b.UndoMove(m)

		if curr >= beta {
			return curr
		}
		alpha = max(alpha, curr)

		if s.abort() {
			return alpha
		}
	}

	return alpha
}
