package search

import (
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

	for m := range movegen.Moves(b, board.Full) {
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
			}
			if score > beta {
				return
			}
			alpha = max(alpha, score)
		} else {
			value, curr := AlphaBeta(b, alpha, beta, depth-1)
			b.UndoMove(&m)
			if value < score {
				score = value
				moves = append(curr, m)
			}
			if score < alpha {
				return
			}
			beta = min(beta, score)
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

		alpha = min(alpha, score)
	} else {
		if score <= alpha {
			return
		}

		beta = max(beta, score)
	}

	for m := range movegen.Moves(b, b.Colors[b.STM.Flip()]) {
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
			alpha = min(alpha, curr)
		} else {
			if curr <= alpha {
				return curr
			}
			score = min(score, curr)
			beta = max(beta, curr)
		}
	}

	return
}
