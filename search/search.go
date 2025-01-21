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
	"github.com/paulsonkoly/chess-3/mstore"
	"github.com/paulsonkoly/chess-3/transp"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

const WindowSize = 50 // half a pawn left and right around score
const MaxPlies = 64

var AWFail int

var ms = mstore.New()

type searchSt struct {
	transpT *transp.Table
}

func Search(b *board.Board, d Depth, stop <-chan struct{}) (score Score, moves []move.Move) {
	// otherwise a checkmate score would always fail high
	alpha := -Inf - 1
	beta := Inf + 1
	aborting = false
	searchSt := searchSt{transpT: transp.NewTable()}

	for d := range d + 1 { // +1 for 0 depth search (quiesence eval)
		awOk := false // aspiration window succeeded
		factor := Score(1)
		var (
			scoreSample Score
			movesSample []move.Move
		)

		for !awOk {
			scoreSample, movesSample = AlphaBeta(b, alpha, beta, d, stop, &searchSt)

			switch {

			case scoreSample <= alpha:
				AWFail++
				alpha -= factor * WindowSize
				factor *= 2

			case scoreSample >= beta:
				AWFail++
				beta += factor * WindowSize
				factor *= 2

			default:
				awOk = true
			}

			if abort(stop) {
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

var (
	ABLeaf    int
	ABBreadth int
	ABCnt     int
	TTHit     int
)

func AlphaBeta(b *board.Board, alpha, beta Score, d Depth, stop <-chan struct{}, sst *searchSt) (Score, []move.Move) {

	transpT := sst.transpT
	pv := []move.Move{}

	if transpE, ok := transpT.LookUp(b.Hashes[len(b.Hashes)-1]); ok {
		TTHit++
		if transpE.Depth >= d {
			switch transpE.Type {

			case transp.PVNode:
				if transpE.From|transpE.To != 0 {
					ms.Push()
					defer ms.Pop()

					movegen.GenMoves(ms, b, board.BitBoard(1<<transpE.To))

					for _, m := range ms.Frame() {
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
		}
		TTHit--
	}

	if d == 0 {
		ABLeaf++
		return Quiescence(b, alpha, beta, 0, stop), pv
	}

  ABCnt++

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

		value, _ := AlphaBeta(b, -beta, -beta+1, rd, stop, sst)
		value *= -1

		b.UndoNullMove(enP)

		if value >= beta {
			return value, pv
		}
	}

	ms.Push()
	defer ms.Pop()

	movegen.GenMoves(ms, b, board.Full)
	moves := ms.Frame()
	sortMoves(b, moves, d, sst)

	for ix, m := range moves {
		b.MakeMove(&m)

		king := b.Colors[b.STM.Flip()] & b.Pieces[King]
		if movegen.IsAttacked(b, b.STM, king) {
			b.UndoMove(&m)
			continue
		}

		hasLegal = true

		// late move reduction
		rd := lmr(d, ix)
		if rd < d-1 && !inCheck {
			value, _ := AlphaBeta(b, -alpha-1, -alpha, rd, stop, sst)
			value *= -1

			if value <= alpha {
				b.UndoMove(&m)
				continue
			}
		}

		value, curr := AlphaBeta(b, -beta, -alpha, d-1, stop, sst)
		value *= -1
		b.UndoMove(&m)

		if value > alpha {
			failLow = false
			alpha = value
			pv = append(curr, m)
		}

		if value >= beta {
			// store node as fail high (cut-node)
			transpT.Insert(b.Hashes[len(b.Hashes)-1], d, m.From, m.To, m.Promo, value, transp.CutNode)

      ABBreadth += ix

			return value, nil
		}

		if abort(stop) {
			return alpha, pv
		}
	}

  ABBreadth += len(moves)

	if !hasLegal {
		// checkmate score
		value := -Inf

		if b.FiftyCnt >= 100 || b.Threefold() {
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
		transpT.Insert(b.Hashes[len(b.Hashes)-1], d, 0, 0, NoPiece, alpha, transp.AllNode)
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

		transpT.Insert(b.Hashes[len(b.Hashes)-1], d, from, to, promo, alpha, transp.PVNode)
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

var (
	QDepth int
	QDelta int
	QSEE   int
)

func Quiescence(b *board.Board, alpha, beta Score, d int, stop <-chan struct{}) Score {
	if d > QDepth {
		QDepth = d
	}

	ms.Push()
	defer ms.Pop()

	movegen.GenMoves(ms, b, board.Full)

	standPat := eval.Eval(b, ms.Frame())

	if standPat >= beta {
		return beta
	}

	delta := standPat + 110 // we only have psqt atm, which doesn't have bigger values than 50
	alpha = max(alpha, standPat)

	moves := ms.Frame()
	sortMoves(b, moves, 0, nil)

	for _, m := range moves {
		captured := b.SquaresToPiece[m.To]
		if m.EPP == Pawn {
			captured = Pawn
		}
		see := heur.SEE(b, &m)

		b.MakeMove(&m)

		check := false
		king := b.Colors[b.STM] & b.Pieces[King]
		if movegen.IsAttacked(b, b.STM.Flip(), king) {
			check = true
		}

		// legality check
		king = b.Colors[b.STM.Flip()] & b.Pieces[King]
		if movegen.IsAttacked(b, b.STM, king) {
			b.UndoMove(&m)
			continue
		}

		if !check {
			if eval.PieceValues[captured]+delta < alpha {
				QDelta++
				b.UndoMove(&m)
				continue
			}

			if see < 0 {
				QSEE++
				b.UndoMove(&m)
				continue
			}
		}

		if !check && captured == NoPiece {
			b.UndoMove(&m)
			continue
		}

		curr := -Quiescence(b, -beta, -alpha, d+1, stop)
		b.UndoMove(&m)

		if curr >= beta {
			return curr
		}
		alpha = max(alpha, curr)

		if abort(stop) {
			return alpha
		}
	}

	return alpha
}

func sortMoves(b *board.Board, moves []move.Move, _ Depth, sst *searchSt) {
	var transPE *transp.Entry

	if sst != nil {
		transPE, _ = sst.transpT.LookUp(b.Hashes[len(b.Hashes)-1])
		if transPE != nil && transPE.Type == transp.AllNode {
			transPE = nil
		}
	}

	for ix, m := range moves {
		weight := Score(0)

		weight += heur.SEE(b, &m)

		if transPE != nil && m.From == transPE.From && m.To == transPE.To && m.Promo == transPE.Promo {
			weight += 5000
		}

		toSq := m.To
		fromSq := m.From
		if b.STM == White {
			toSq ^= 56
			fromSq ^= 56
		}
		weight += eval.PSqT[(m.Piece-1)*2][toSq] - eval.PSqT[(m.Piece-1)*2][fromSq]
		moves[ix].Weight = weight
	}
	slices.SortFunc(moves, func(a, b move.Move) int { return int(b.Weight - a.Weight) })
}
