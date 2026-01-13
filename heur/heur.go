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
	"github.com/paulsonkoly/chess-3/params"
	"github.com/paulsonkoly/chess-3/stack"
)

// PieceValues approximates the value of each piece type for SEE and heuristic
// purposes.
var PieceValues = [...]Score{0, 100, 300, 300, 500, 900, Inf}

const (
	k = Score(1024)
	// HashMove is assigned to a move from Hash, either PV or fail-high.
	HashMove = Score(15000)
	// Captures is the minimal score for captures, actual score is this plus SEE.
	Captures     = 7 * k
	CaptureRange = 8 * k
	// MaxHistory is the maximal absolute value in either the history or the continuation stores.
	MaxHistory = k
	// MaxCaptHistory is the maximal score in the capthist table.
	MaxCaptHistory = Score(128)
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
	captHist      *CaptHist
	continuations [2]*Continuation
}

// NewMoveRanker creates a new move ranker.
func NewMoveRanker() MoveRanker {
	return MoveRanker{
		history:       NewHistory(),
		captHist:      NewCaptHist(),
		continuations: [2]*Continuation{NewContinuation(), NewContinuation()},
	}
}

// Clear clears the stores in mr.
func (mr *MoveRanker) Clear() {
	mr.history.Clear()
	mr.captHist.Clear()
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
	var score Score

	promo := m.Promo()
	attacker := b.SquaresToPiece[m.From()]
	victim := b.SquaresToPiece[b.CaptureSq(m)]

	if promo != NoPiece {
		promo -= Pawn // Knight, Bishop, Rook, Queen => buckets: 0: NoPiece, 1: Knight, ... etc.
	}

	bucket := int(promo)*6 + int(victim) // bucket in range of 0 .. 30

	// MVV/LVA the bucket index is determined by promotion / victim; within the
	// bucket the score is a blend of inverted attacker and captHist.
	var adjCaptHist Score
	if victim != NoPiece {
		adjCaptHist = mr.captHist.LookUp(attacker, victim, m.To()) + MaxCaptHistory // translate -max..+max range to 0..2*max
	}
	invAttacker := Score(King - attacker)                  // attacker reversing order
	invAttacker = (invAttacker - 2) * (MaxCaptHistory / 8) // aligning attacker value with captHist Range

	score = (2*MaxCaptHistory)*Score(bucket) + Clamp(adjCaptHist+invAttacker, 0, 2*MaxCaptHistory)

	if SEE(b, m, 0) {
		// good capture
		return Captures + score
	} else {
		// bad capture
		return -Captures - CaptureRange + score
	}
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
		captured := b.SquaresToPiece[b.CaptureSq(m)]

		if captured == NoPiece && m.Promo() == NoPiece {
			mr.history.Add(b.STM, m.From(), m.To(), bonus)

			if hist, ok := stack.Top(0); ok {
				mr.continuations[0].Add(b.STM, hist.Piece, hist.To, moved, m.To(), bonus)
			}

			if hist, ok := stack.Top(1); ok {
				mr.continuations[1].Add(b.STM, hist.Piece, hist.To, moved, m.To(), bonus)
			}
		} else if captured != NoPiece {
			mr.captHist.Add(moved, captured, m.To(), Score(d))
		}
	}

	bonus := Score(d)*Score(params.HistBonusMul) - Score(params.HistBonusLin)
	penalty := -bonus

	rng := Score(1) << params.HistAdjRange
	red := Score(1) << params.HistAdjReduction

	if len(moves) >= 2 {
		for _, m := range moves[:len(moves)-1] {
			// m.Weight was set to score by the search, or -Inf for upbounds.
			adj := Score(rng+Clamp(m.Weight, -rng, rng)) / red

			adjustScores(m.Move, penalty+adj)
		}
	}
	adjustScores(moves[len(moves)-1].Move, bonus)
}
