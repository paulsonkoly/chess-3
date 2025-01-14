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

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

const WindowSize = 50 // half a pawn left and right around score

var AWFail int

var ms = mstore.New()

func Search(b *board.Board, depth int, stop <-chan struct{}) (score int, moves []move.Move) {
	alpha := -eval.Inf
	beta := eval.Inf
	aborting = false

	for d := range depth + 1 { // +1 for 0 depth search (quiesence eval)
		awOk := false // aspiration window succeeded
		factor := 1
		var (
			scoreSample int
			movesSample []move.Move
		)

		for !awOk {
			scoreSample, movesSample = AlphaBeta(b, alpha, beta, d, stop)

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
	ABLeaf int
)

func AlphaBeta(b *board.Board, alpha, beta int, depth int, stop <-chan struct{}) (score int, pv []move.Move) {
	if depth == 0 {
		ABLeaf++
		return Quiescence(b, alpha, beta, 0, stop), []move.Move{}
	}

	score = -eval.Inf

	hasLegal := false

	ms.Push()
	defer ms.Pop()

	movegen.GenMoves(ms, b, board.Full)
	moves := ms.Frame()
	sortMoves(b, moves)

	for _, m := range moves {
		b.MakeMove(&m)

		king := b.Colors[b.STM.Flip()] & b.Pieces[King]
		if movegen.IsAttacked(b, b.STM, king) {
			b.UndoMove(&m)
			continue
		}

		value, curr := AlphaBeta(b, -beta, -alpha, depth-1, stop)
		value *= -1
		b.UndoMove(&m)
		if value > score || value == score && !hasLegal {
			score = value
			pv = append(curr, m)
			alpha = max(alpha, score)
		}

		hasLegal = true

		if score >= beta {
			return
		}

		if abort(stop) {
			return
		}
	}

	if !hasLegal {
		king := b.Colors[b.STM] & b.Pieces[King]
		if !movegen.IsAttacked(b, b.STM.Flip(), king) {
			score = 0
		}
	}

	return
}

var (
	QDepth int
	QDelta int
	QSEE   int
)

func Quiescence(b *board.Board, alpha, beta int, d int, stop <-chan struct{}) int {
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
	sortMoves(b, moves)

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

func sortMoves(b *board.Board, moves []move.Move) {
	for ix, m := range moves {
		sqFrom := m.From
		sqTo := m.To
		if b.STM == White {
			file := sqFrom % 8
			rank := sqFrom / 8
			sqFrom = file + (7-rank)*8

			file = sqTo % 8
			rank = sqTo / 8
			sqTo = file + (7-rank)*8
		}

		weight := eval.Psqt[m.Piece-1][sqTo] - eval.Psqt[m.Piece-1][sqFrom]

		if b.SquaresToPiece[m.To] != NoPiece {
			weight += heur.SEE(b, &m)
		}
		moves[ix].Weight = weight
	}
	slices.SortFunc(moves, func(a, b move.Move) int { return b.Weight - a.Weight })
}
