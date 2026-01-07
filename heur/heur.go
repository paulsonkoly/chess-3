// Package heur provides move ordering heuristics.
//
// # Move ordering stages
//
//   - hash: [HashMove]
//   - good captures: [Captures] ... [HashMove]
//   - quiets: -6*[MaxHistory]..6*[MaxHistory]
//     (3 * cont[0] + 2 * cont[1] + hist) each <= [MaxHistory]
//   - bad captures: -Inf..-[Captures]
package heur

import (
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/stack"
)

// PieceValues approximates the value of each piece type for SEE and heuristic
// purposes.
var PieceValues = [...]Score{0, 100, 300, 300, 500, 900, Inf}

const (
	// HashMove is assigned to a move from Hash, either PV or fail-high.
	HashMove = Score(15000)
	// Captures is the minimal score for captures, actual score is this plus SEE.
	Captures = Score(8192)
	// MaxHistory is the maximal absolute value in either the history or the continuation stores.
	MaxHistory = Score(1024)
)

func init() {
	// 1 * history + 2 * continuation[1] + 3 * continuation[0]
	if Captures < 6*MaxHistory {
		panic("gap is not big enough in move weight layout for history scores")
	}
}

// MoveRanker is a composition of heuristic stores that can rank a move.
type MoveRanker struct {
	history       *History
	continuations [2]*Continuation
}

// NewMoveRanker creates a new move ranker.
func NewMoveRanker() MoveRanker {
	return MoveRanker{history: NewHistory(), continuations: [2]*Continuation{NewContinuation(), NewContinuation()}}
}

// Clear clears the stores in mr.
func (mr *MoveRanker) Clear() {
	mr.history.Clear()
	mr.continuations[0].Clear()
	mr.continuations[1].Clear()
}

// StackMove represents an already played move, identified by moving piece type
// and to squares. It is coupled with static evaluation of the position.
type StackMove struct {
	Piece Piece  // Piece is the moved piece type.
	To    Square // To is the destination square of the move.
	Score Score  // Score is the static evaluation of the position.
}

// RankNoisy returns the heuristic rank of a noisy move.
func (mr *MoveRanker) RankNoisy(m move.Move, b *board.Board, _ *stack.Stack[StackMove]) Score {
	return MVVLVA(b, m, SEE(b, m, 0))
}

// RankQuiet returns the heuristic rank of a quiet move.
func (mr *MoveRanker) RankQuiet(m move.Move, b *board.Board, stack *stack.Stack[StackMove]) Score {
	score := mr.history.LookUp(b.STM, m.From(), m.To())
	moved := b.SquaresToPiece[m.From()]

	if hist, ok := stack.Top(0); ok {
		score += 3 * mr.continuations[0].LookUp(b.STM, hist.Piece, hist.To, moved, m.To())
	}

	if hist, ok := stack.Top(1); ok {
		score += 2 * mr.continuations[1].LookUp(b.STM, hist.Piece, hist.To, moved, m.To())
	}

	return score
}

// FailHigh updates the history / continuation stores based on the move buffer
// moves. We assume all moves preceding the last are bad, and the last one is
// good. Naturally this would be true in a move loop.
//
// This function panics if the moves buffer is empty.
func (mr *MoveRanker) FailHigh(d Depth, b *board.Board, moves []move.Weighted, stack *stack.Stack[StackMove]) {
	adjustScores := func(m move.Move, bonus Score) {
		moved := b.SquaresToPiece[m.From()]
		// TODO en-passant
		captured := b.SquaresToPiece[m.To()]

		if captured == NoPiece && m.Promo() == NoPiece {
			mr.history.Add(b.STM, m.From(), m.To(), bonus)

			if hist, ok := stack.Top(0); ok {
				mr.continuations[0].Add(b.STM, hist.Piece, hist.To, moved, m.To(), bonus)
			}

			if hist, ok := stack.Top(1); ok {
				mr.continuations[1].Add(b.STM, hist.Piece, hist.To, moved, m.To(), bonus)
			}
		}
	}

	bonus := Score(d)*20 - 15
	penalty := -bonus

	if len(moves) >= 2 {
		for _, m := range moves[:len(moves)-1] {
			var adj Score
			// adjustment 0..8
			if m.Weight != 0 {
				adj = Score(256 + Clamp(int(m.Weight), -256, 256)) / 64
			}

			adjustScores(m.Move, penalty+adj)
		}
	}
	adjustScores(moves[len(moves)-1].Move, bonus)
}
