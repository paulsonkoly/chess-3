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
		return Quiescence(b, alpha, beta), []move.Move{}
	}

	if b.STM == White {
		score = -eval.Inf
	} else {
		score = eval.Inf
	}

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

		if b.STM.Flip() == White {
			value, curr := AlphaBeta(b, alpha, beta, depth-1)
			b.UndoMove(&m)
			if value > score {
				score = value
				moves = append(curr, m)
				alpha = max(alpha, score)
			}
			if score >= beta {
				return
			}
		} else {
			value, curr := AlphaBeta(b, alpha, beta, depth-1)
			b.UndoMove(&m)
			if value < score {
				score = value
				moves = append(curr, m)
				beta = min(beta, score)
			}

			if score <= alpha {
				return
			}
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

func Quiescence(b *board.Board, alpha, beta int) (score int) {
	score = eval.Eval(b)

	if b.STM == White {
		if score >= beta {
			return
		}

		alpha = max(alpha, score)
	} else {
		if score <= alpha {
			return
		}

		beta = min(beta, score)
	}

	moveList := sortedMoves(b)

	for _, m := range moveList {
		if m.Weight < 0 {
			return
		}
		if (1<<m.To)&b.Colors[b.STM.Flip()] == 0 {
			continue
		}

		b.MakeMove(&m)

		king := b.Colors[b.STM.Flip()] & b.Pieces[King]
		if movegen.IsAttacked(b, b.STM, king) {
			b.UndoMove(&m)
			continue
		}

		curr := Quiescence(b, alpha, beta)
		b.UndoMove(&m)

		if b.STM == White {
			if curr >= beta {
				return curr
			}
			score = max(score, curr)
			alpha = max(alpha, curr)
		} else {
			if curr <= alpha {
				return curr
			}
			score = min(score, curr)
			beta = min(beta, curr)
		}
	}

	return
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
