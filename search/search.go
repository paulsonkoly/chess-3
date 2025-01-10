package search

import (
	"slices"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/eval"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

func AlphaBeta(b *board.Board, alpha, beta int, depth int) (score int, moves []move.Move) {
	if depth == 0 {
		return Quiescence(b, alpha, beta, 0), []move.Move{}
		// return eval.Eval(b), []move.Move{}
	}

	score = -eval.Inf

	hasLegal := false

	moveList := sortedMoves(b)

	for _, m := range moveList {
		b.MakeMove(&m)

		king := b.Colors[b.STM.Flip()] & b.Pieces[King]
		if movegen.IsAttacked(b, b.STM, king) {
			b.UndoMove(&m)
			continue
		}

		hasLegal = true

		value, curr := AlphaBeta(b, -beta, -alpha, depth-1)
		value *= -1
		b.UndoMove(&m)
		if value > score {
			score = value
			moves = append(curr, m)
			alpha = max(alpha, score)
		}
		if score >= beta {
			return
		}
	}

	if !hasLegal {
		king := b.Colors[b.STM] & b.Pieces[King]
		if !movegen.IsAttacked(b, b.STM, king) {
			score = 0
		}
	}

	return
}

var (
	QDepth  int
	QDelta  int
	QWeight int
)

func Quiescence(b *board.Board, alpha, beta int, d int) int {
	if d > QDepth {
		QDepth = d
	}
	standPat := eval.Eval(b)

	if standPat >= beta {
		return beta
	}

	delta := standPat + 110 // we only have psqt atm, which doesn't have bigger values than 50
	alpha = max(alpha, standPat)

	moveList := sortedMoves(b)

	for _, m := range moveList {
		captured := b.SquaresToPiece[m.To]
		if m.EPP == Pawn {
			captured = Pawn
		}

		if eval.PieceValues[captured]+delta < alpha {
			QDelta++
			continue
		}
		//
		// if m.Weight < 0 {
		//   QWeight++ // this should be SSE
		//   continue
		// }
		//
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

		if !check && captured == NoPiece {
			b.UndoMove(&m)
			continue
		}

		curr := -Quiescence(b, -beta, -alpha, d+1)
		b.UndoMove(&m)

		if curr >= beta {
			return curr
		}
		alpha = max(alpha, curr)
	}

	return alpha
}

func sortedMoves(b *board.Board) []move.Move {
	result := make([]move.Move, 0, 30)
	for m := range movegen.Moves(b, board.Full) {
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

		m.Weight = eval.Psqt[m.Piece-1][sqTo] - eval.Psqt[m.Piece-1][sqFrom]

		if b.SquaresToPiece[m.To] != NoPiece {
			m.Weight += eval.PieceValues[b.SquaresToPiece[m.To]] - eval.PieceValues[m.Piece]
		}
		result = append(result, m)
	}
	slices.SortFunc(result, func(a, b move.Move) int { return b.Weight - a.Weight })
	return result
}
