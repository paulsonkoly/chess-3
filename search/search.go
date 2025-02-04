package search

import (
	"fmt"
	"slices"
	"strings"

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
	MaxPlies   = 64
)

// State is a persistent state storage between searches.
type State struct {
	tt   *transp.Table
	hist *heur.History
	ms   *move.Store

	// Stop channel signals an immediate Stop requiest to the search. Current
	// depth will be abandoned.
	Stop chan struct{}

	AWFail int // AwFail is the count of times the score fell outside of the aspiration window.
	ABLeaf int // ABLeaf is the count of alpha-beta leafs.
	// ABBreadth is the total count of explored moves in alpha-beta. Thus
	// (ABBreadth / ABCnt) is the average alpha-beta branching factor.
	ABBreadth int
	ABCnt     int // ABCnt is the inner node count in alpha-beta.
	TTHit     int // TThit is the transposition table hit-count.
	QDepth    int // QDepth is the maximal quiesence search depth.
	QDelta    int // QDelta is the count of times a delta pruning happened in quiesence search.
	QSEE      int // QSEE is the count of times the static exchange evaluation fell under 0 in quiesence search.
}

// NewState creates a new search state. It's supposed to be called once, and
// re-used between Search() calls.
func NewState() *State {
	return &State{tt: transp.New(), ms: move.NewStore(), hist: heur.NewHistory()}
}

// Clear resets the counters, and various stores for the search, assuming a new
// position.
func (s *State) Clear() {
	s.tt.Clear()
	s.ms.Clear()
	s.AWFail = 0
	s.ABLeaf = 0
	s.ABBreadth = 0
	s.ABCnt = 0
	s.TTHit = 0
	s.QDepth = 0
	s.QDelta = 0
	s.QSEE = 0
}

// Search is the main entry point to the engine. It performs and
// iterative-deepened alpha-beta with aspiration window.
func Search(b *board.Board, d Depth, sst *State) (score Score, moves []move.Move) {
	// otherwise a checkmate score would always fail high
	alpha := -Inf - 1
	beta := Inf + 1
	aborting = false

	sst.Clear()

	for d := range d + 1 { // +1 for 0 depth search (quiesence eval)
		awOk := false // aspiration window succeeded
		factor := Score(1)
		var (
			scoreSample Score
			movesSample []move.Move
		)

		for !awOk {
			scoreSample, movesSample = AlphaBeta(b, alpha, beta, d, sst)

			switch {

			case scoreSample <= alpha:
				sst.AWFail++
				alpha -= factor * WindowSize
				factor *= 2

			case scoreSample >= beta:
				sst.AWFail++
				beta += factor * WindowSize
				factor *= 2

			default:
				awOk = true
			}

			if abort(sst.Stop) {
				return
			}
		}
		score, moves = scoreSample, movesSample
		slices.Reverse(moves)
		fmt.Printf("info depth %d score cp %d pv %s\n", d, score, pvInfo(moves))

		alpha = score - WindowSize
		beta = score + WindowSize
	}
	return
}

var aborting = false

func abort(stop <-chan struct{}) bool {
	if stop != nil {
		select {
		case <-stop:
			aborting = true
			return true
		default:
		}
	}
	return aborting
}

func pvInfo(moves []move.Move) string {
	sb := strings.Builder{}
	space := ""
	for _, m := range moves {
		sb.WriteString(space)
		sb.WriteString(fmt.Sprint(m))
		space = " "
	}
	return sb.String()
}

// AlphaBeta performs an alpha beta search to depth d, and then transitions
// into Quiesence() search.
func AlphaBeta(b *board.Board, alpha, beta Score, d Depth, sst *State) (Score, []move.Move) {

	transpT := sst.tt
	pv := []move.Move{}

	tfCnt := b.Threefold()
	if tfCnt >= 3 {
		return 0, pv
	}

	if transpE, ok := transpT.LookUp(b.Hashes[len(b.Hashes)-1]); ok && transpE.Depth >= d && transpE.TFCnt >= tfCnt {
		sst.TTHit++
		switch transpE.Type {

		case transp.PVNode:
			if transpE.From|transpE.To != 0 {
				sst.ms.Push()
				defer sst.ms.Pop()

				movegen.GenMoves(sst.ms, b, board.BitBoard(1<<transpE.To))

				for _, m := range sst.ms.Frame() {
					if m.From == transpE.From && m.Promo == transpE.Promo {
						pv = []move.Move{m}
					}
				}
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
		sst.TTHit--
	}

	if d == 0 {
		sst.ABLeaf++
		return Quiescence(b, alpha, beta, 0, sst), pv
	}

	sst.ABCnt++

	hasLegal := false
	failLow := true

	inCheck := false
	king := b.Colors[b.STM] & b.Pieces[King]
	if movegen.IsAttacked(b, b.STM.Flip(), king) {
		inCheck = true
	}

	// null move pruning
	if !inCheck && b.Colors[b.STM] & ^(b.Pieces[Pawn]|b.Pieces[King]) != 0 {

		enP := b.MakeNullMove()

		rd := max(0, d-3)

		value, _ := AlphaBeta(b, -beta, -beta+1, rd, sst)
		value *= -1

		b.UndoNullMove(enP)

		if value >= beta {
			return value, pv
		}
	}

	// deflate history
	if sst.ABCnt%10_000 == 0 {
		sst.hist.Deflate()
	}

	sst.ms.Push()
	defer sst.ms.Pop()

	movegen.GenMoves(sst.ms, b, board.Full)
	moves := sst.ms.Frame()

	rankMovesAB(b, moves, sst)

	var (
		m  *move.Move
		ix int
	)

	for m, ix = getNextMove(moves, -1); m != nil; m, ix = getNextMove(moves, ix) {
		b.MakeMove(m)

		king := b.Colors[b.STM.Flip()] & b.Pieces[King]
		if movegen.IsAttacked(b, b.STM, king) {
			b.UndoMove(m)
			continue
		}

		hasLegal = true

		// late move reduction
		rd := lmr(d, ix)
		if rd < d-1 && !inCheck {
			value, _ := AlphaBeta(b, -alpha-1, -alpha, rd, sst)
			value *= -1

			if value <= alpha {
				b.UndoMove(m)
				continue
			}
		}

		value, curr := AlphaBeta(b, -beta, -alpha, d-1, sst)
		value *= -1
		b.UndoMove(m)

		if value > alpha {
			failLow = false
			alpha = value
			pv = append(curr, *m)
		}

		if value >= beta {
			// store node as fail high (cut-node)
			transpT.Insert(b.Hashes[len(b.Hashes)-1], d, tfCnt, m.From, m.To, m.Promo, value, transp.CutNode)

			if m.Captured == NoPiece {
				sst.hist.Add(b.STM, m.From, m.To, d)
			}

			sst.ABBreadth += ix

			return value, nil
		}

		if abort(sst.Stop) {
			return alpha, pv
		}
	}

	sst.ABBreadth += ix

	if !hasLegal {
		// checkmate score
		value := -Inf

		if b.FiftyCnt >= 100 {
			value = 0
		} else {
			king := b.Colors[b.STM] & b.Pieces[King]
			if !movegen.IsAttacked(b, b.STM.Flip(), king) {
				// draw score
				value = 0
			}
		}

		if value > alpha {
			failLow = false
			alpha = value
		}
	}

	if failLow {
		// store node as fail low (All-node)
		transpT.Insert(b.Hashes[len(b.Hashes)-1], d, tfCnt, 0, 0, NoPiece, alpha, transp.AllNode)
	} else {
		// store node as exact (PV-node)
		// there might not be a move in case of !hasLegal
		var from, to Square
		var promo Piece
		if len(pv) > 0 {
			m := pv[len(pv)-1]
			from = m.From
			to = m.To
			promo = m.Promo
		}

		transpT.Insert(b.Hashes[len(b.Hashes)-1], d, tfCnt, from, to, promo, alpha, transp.PVNode)
	}

	return alpha, pv
}

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

func lmr(d Depth, mCount int) Depth {
	value := (log[int(d)] * log[mCount] / 19500)

	return max(0, d-Depth(value))
}

// Quiescence resolves the position to a quiet one, and then evaluates.
func Quiescence(b *board.Board, alpha, beta Score, d int, sst *State) Score {
	if d > sst.QDepth {
		sst.QDepth = d
	}

	if b.Threefold() >= 3 {
		return 0
	}

	sst.ms.Push()
	defer sst.ms.Pop()

	movegen.GenMoves(sst.ms, b, board.Full)

	standPat := eval.Eval(b, alpha, beta, sst.ms.Frame(), &eval.Coefficients)

	if standPat >= beta {
		return beta
	}

	delta := standPat + 110 // we only have psqt atm, which doesn't have bigger values than 50
	alpha = max(alpha, standPat)

	moves := sst.ms.Frame()

	rankMovesQ(b, moves)

	for m, ix := getNextMove(moves, -1); m != nil; m, ix = getNextMove(moves, ix) {
		captured := b.SquaresToPiece[m.To]
		if m.EPP == Pawn {
			captured = Pawn
		}

		b.MakeMove(m)

		// legality check
		king := b.Colors[b.STM.Flip()] & b.Pieces[King]
		if movegen.IsAttacked(b, b.STM, king) {
			b.UndoMove(m)
			continue
		}

		check := false
		king = b.Colors[b.STM] & b.Pieces[King]
		if movegen.IsAttacked(b, b.STM.Flip(), king) {
			check = true
		}

		if !check && captured == NoPiece {
			b.UndoMove(m)
			continue
		}

		if !check {
			if heur.PieceValues[captured]+delta < alpha {
				sst.QDelta++
				b.UndoMove(m)
				continue
			}

			if m.SEE < 0 {
				sst.QSEE++
				b.UndoMove(m)
				continue
			}
		}

		curr := -Quiescence(b, -beta, -alpha, d+1, sst)
		b.UndoMove(m)

		if curr >= beta {
			return curr
		}
		alpha = max(alpha, curr)

		if abort(sst.Stop) {
			return alpha
		}
	}

	return alpha
}

func rankMovesAB(b *board.Board, moves []move.Move, sst *State) {
	var transPE *transp.Entry

	transPE, _ = sst.tt.LookUp(b.Hashes[len(b.Hashes)-1])
	if transPE != nil && transPE.Type == transp.AllNode {
		transPE = nil
	}

	for ix, m := range moves {

		switch {
		case transPE != nil && m.From == transPE.From && m.To == transPE.To && m.Promo == transPE.Promo:
			moves[ix].Weight = 5000

		case b.SquaresToPiece[m.To] != NoPiece:
			see := heur.SEE(b, &m)
			if see < 0 {
				moves[ix].Weight = see - heur.MaxHistoryScore
			} else {
				moves[ix].Weight = see + heur.MaxHistoryScore
			}
			moves[ix].SEE = see

		default:
			moves[ix].Weight = sst.hist.Probe(b.STM, m.From, m.To)
		}
	}
}

func rankMovesQ(b *board.Board, moves []move.Move) {
	for ix, m := range moves {
		if b.SquaresToPiece[m.To] != NoPiece {
			see := heur.SEE(b, &m)
			moves[ix].SEE = see
			moves[ix].Weight = see
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
