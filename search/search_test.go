package search

import (
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"
	"github.com/paulsonkoly/chess-3/transp"
	. "github.com/paulsonkoly/chess-3/types"
)

func BenchmarkRankMovesAB(b *testing.B) {
	s := New(1 * transp.MegaBytes)

	board := board.StartPos() // or a realistic midgame FEN

	moves := make([]move.Move, 0, 256)
	movegen.GenMoves(&moves, board)

	weights := make([]Score, 0, 256)

	b.ResetTimer()
	for range b.N {
		s.rankMovesAB(board, moves, &weights)
	}
}
