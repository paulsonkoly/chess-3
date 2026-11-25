package eval

import (
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/stretchr/testify/assert"
)

func TestKBNvKPositive(t *testing.T) {
	// random kbnvk stm is winning.
	fens := [...]string{
		"8/5k2/8/8/2K5/8/4B3/1N6 w - - 0 1",
		"8/2K2k2/8/1N3B2/8/8/8/8 w - - 0 1",
		"4n3/5k2/2K5/4b3/8/8/8/8 b - - 0 1",
		"2B5/8/K7/8/N4k2/8/8/8 w - - 0 1",
		"8/1b6/8/3k4/8/3K4/3n4/8 b - - 0 1",
		"8/8/2k5/8/3b2K1/3n4/8/8 b - - 0 1",
		"8/8/4K3/3n4/8/8/k7/6b1 b - - 0 1",
		"8/8/5K2/8/8/5b2/k7/1n6 b - - 0 1",
		"8/2k5/7n/2K5/8/8/7b/8 b - - 0 1",
		"8/1K1B4/8/8/5N2/8/8/k7 w - - 0 1",
		"8/n7/8/2b5/8/3k4/K7/8 b - - 0 1",
		"k2B4/8/8/K7/8/8/3N4/8 w - - 0 1",
		"8/1K6/8/8/3k4/3b4/4n3/8 b - - 0 1",
		"8/K3k3/8/8/8/8/2b5/4n3 b - - 0 1",
		"8/b3k3/8/8/7K/8/8/2n5 b - - 0 1",
		"6k1/8/8/8/K7/3B4/1N6/8 w - - 0 1",
		"2K5/8/6N1/5B2/2k5/8/8/8 w - - 0 1",
		"1b6/3K4/8/8/8/4n3/8/4k3 b - - 0 1",
		"8/k7/6N1/1K5B/8/8/8/8 w - - 0 1",
		"8/k3n1b1/8/8/8/8/6K1/8 b - - 0 1",
		"8/2n1K1k1/8/8/8/8/5b2/8 b - - 0 1",
		"2K5/8/8/8/6n1/8/2k5/6b1 b - - 0 1",
		"8/5b2/8/8/5n2/8/8/k6K b - - 0 1",
	}
	for _, fen := range fens {
		t.Run(fen, func(t *testing.T) {
			b := Must(board.FromFEN(fen))
			e := New[Score]()

			assert.Positive(t, e.kbnvk(b), "fen %s", fen)
		})
	}
}

func TestKBNvKMonotonicity(t *testing.T) {
	// winning side to move positions sampled from a stockfish self play
	// (5s/move) till mate, in order.
	positions := [...]string{
		"7k/8/8/8/8/8/8/KB5N w - - 0 1",
		"8/6k1/8/8/8/8/1K6/1B5N w - - 2 2",
		"8/8/5k2/8/8/2K5/8/1B5N w - - 4 3",
		"8/6k1/8/8/3K4/8/8/1B5N w - - 6 4",
		"7k/8/8/4K3/8/8/8/1B5N w - - 8 5",
		"8/6k1/8/4K3/8/8/B7/7N w - - 10 6",
		"8/8/7k/5K2/8/8/B7/7N w - - 12 7",
		"8/7k/5K2/8/8/8/B7/7N w - - 14 8",
		"7k/8/5K2/8/8/8/B4N2/8 w - - 16 9",
		"8/7k/5K2/8/8/7N/B7/8 w - - 18 10",
		"7k/8/5K2/6N1/8/8/B7/8 w - - 20 11",
		"6k1/8/5K2/6N1/8/8/8/1B6 w - - 22 12",
		"5k2/5N2/5K2/8/8/8/8/1B6 w - - 24 13",
		"4k3/5N1B/5K2/8/8/8/8/8 w - - 26 14",
		"3k4/7B/5K2/4N3/8/8/8/8 w - - 28 15",
		"4k3/8/5K2/4N3/4B3/8/8/8 w - - 30 16",
		"5k2/8/5K2/3BN3/8/8/8/8 w - - 32 17",
		"4k3/3N4/5K2/3B4/8/8/8/8 w - - 34 18",
		"3k4/3N4/4K3/3B4/8/8/8/8 w - - 36 19",
		"4k3/3N4/3K4/3B4/8/8/8/8 w - - 38 20",
		"3k4/3N4/3KB3/8/8/8/8/8 w - - 40 21",
		"2k5/3N1B2/3K4/8/8/8/8/8 w - - 42 22",
		"3k4/5B2/3K4/2N5/8/8/8/8 w - - 44 23",
		"2k5/1N3B2/3K4/8/8/8/8/8 w - - 46 24",
		"1k6/1N3B2/2K5/8/8/8/8/8 w - - 48 25",
		"8/k4B2/2KN4/8/8/8/8/8 w - - 50 26",
		"1k6/8/2KN4/8/2B5/8/8/8 w - - 52 27",
		"k7/8/1K1N4/8/2B5/8/8/8 w - - 54 28",
		"8/k1K5/3N4/8/2B5/8/8/8 w - - 56 29",
		"k1N5/2K5/8/8/2B5/8/8/8 w - - 58 30",
	}
	scores := make([]float64, 0, len(positions))
	e := New[float64]()
	for _, fen := range positions {
		b := Must(board.FromFEN(fen))
		scores = append(scores, e.kbnvk(b))
	}

	// the stockfish line does not completely align with our heuristic, so we
	// just assert an overall upward trend instead of per-sample monotonicity.
	var sumX, sumY, sumXY, sumXX float64
	for i, score := range scores {
		x, y := float64(i), score
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}
	n := float64(len(scores))
	slope := (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)

	assert.Greater(t, slope, float64(0), "path to mate has no positive trend: %v", scores)
}

func TestKPvKDrawing(t *testing.T) {
	fen := "3k4/3p4/8/8/8/3K4/8/8 b - - 0 1"
	b := Must(board.FromFEN(fen))
	e := New[Score]()
	assert.Zero(t, e.kpvk(b, &Coefficients), "fen %s should be drawing", fen)
}

func TestKPvKWinningScores(t *testing.T) {
	// position samples from a winning line where black queens, sample
	// taken after each pawn advancement.
	positions := [...]string{
		"3k4/3p4/8/8/8/8/5K2/8 b - - 0 1",
		"8/8/3p4/8/2k5/8/3K4/8 w - - 0 8",
		"8/8/8/3p4/2k5/4K3/8/8 w - - 0 9",
		"8/8/8/8/3p4/4k3/8/3K4 w - - 0 12",
		"8/8/8/8/8/3pk3/8/2K5 w - - 0 13",
		"8/8/8/8/8/4k3/3p4/3K4 w - - 0 14",
	}

	// normalize to the winning side's perspective
	normalize := func(stm Color, score Score) Score {
		t.Helper()
		if stm == Black {
			return score
		}
		return -score
	}

	queenValue := Coefficients.PieceValues[EG][Queen]

	e := New[Score]()
	b := Must(board.FromFEN(positions[0]))
	prev := normalize(b.STM, e.kpvk(b, &Coefficients))
	assert.Positive(t, prev, "fen %s", positions[0])
	assert.Less(t, prev, queenValue, "fen %s", positions[0])

	for i := 1; i < len(positions); i++ {
		b = Must(board.FromFEN(positions[i]))
		next := normalize(b.STM, e.kpvk(b, &Coefficients))
		assert.Greater(t, next, prev, "fen %s monotonicity", positions[i])
		assert.Less(t, next, queenValue, "fen %s", positions[i])

		prev = next
	}
}

type kbpvkCase struct {
	fen     string
	winning bool
}

func TestKBPvK(t *testing.T) {
	// 500 random wrong coloured rook pawn positions, labeled with tb outcome.
	// this algorithm is heuristical and it cannot give an absolute answer.
	// Currently we are failing 8 out of 500. we assert that it's not worse than
	// that.
	tests := [...]kbpvkCase{
		{fen: "1k6/1B6/7P/8/8/2K5/8/8 w - - 0 1", winning: true},
		{fen: "5K2/8/8/P7/4k3/8/8/6B1 w - - 0 1", winning: true},
		{fen: "3K4/8/8/8/P2B3k/8/8/8 b - - 0 1", winning: true},
		{fen: "8/8/k3K3/7B/8/7P/8/8 b - - 0 1", winning: true},
		{fen: "8/5k2/8/8/P7/8/K7/6B1 b - - 0 1", winning: false},
		{fen: "8/4k3/B7/8/K6P/8/8/8 b - - 0 1", winning: false},
		{fen: "8/8/8/8/6B1/k6K/7P/8 b - - 0 1", winning: true},
		{fen: "8/5B2/8/8/K7/8/7P/2k5 b - - 0 1", winning: true},
		{fen: "8/8/8/7P/K5B1/3k4/8/8 b - - 0 1", winning: true},
		{fen: "2B5/8/k7/8/8/8/7P/2K5 b - - 0 1", winning: true},
		{fen: "8/K7/8/4B3/P7/5k2/8/8 b - - 0 1", winning: true},
		{fen: "8/3K4/8/1k6/8/8/P7/4B3 b - - 0 1", winning: false},
		{fen: "1k6/8/8/8/P7/8/1B2K3/8 w - - 0 1", winning: false},
		{fen: "8/8/8/2B5/1K6/3k4/P7/8 b - - 0 1", winning: true},
		{fen: "8/8/8/8/P7/8/1K4k1/6B1 w - - 0 1", winning: true},
		{fen: "8/8/5k2/8/2B5/8/2K4P/8 w - - 0 1", winning: false},
		{fen: "8/1B6/2K4P/8/8/3k4/8/8 b - - 0 1", winning: true},
		{fen: "2B5/8/8/8/7P/8/4k3/2K5 w - - 0 1", winning: true},
		{fen: "8/2B5/8/8/P4K2/8/7k/8 b - - 0 1", winning: true},
		{fen: "8/1k6/3B4/K7/8/8/P7/8 b - - 0 1", winning: false},
		{fen: "8/7P/8/3B4/8/4k1K1/8/8 b - - 0 1", winning: true},
		{fen: "8/6B1/P7/8/3K4/6k1/8/8 b - - 0 1", winning: true},
		{fen: "8/8/6K1/8/8/P3B3/8/2k5 b - - 0 1", winning: true},
		{fen: "8/P5B1/8/8/8/8/8/1K3k2 b - - 0 1", winning: true},
		{fen: "8/8/8/2B5/8/P1K5/6k1/8 b - - 0 1", winning: true},
		{fen: "8/P4kB1/8/8/7K/8/8/8 w - - 0 1", winning: true},
		{fen: "8/8/5B2/P7/1K6/8/8/7k w - - 0 1", winning: true},
		{fen: "2K5/7P/6B1/8/7k/8/8/8 w - - 0 1", winning: true},
		{fen: "8/1k5P/8/7B/8/8/8/6K1 w - - 0 1", winning: true},
		{fen: "3k2K1/8/8/8/8/P1B5/8/8 b - - 0 1", winning: false},
		{fen: "8/8/7k/8/4B3/7P/6K1/8 b - - 0 1", winning: false},
		{fen: "8/5B1P/3k4/8/8/1K6/8/8 w - - 0 1", winning: true},
		{fen: "8/6K1/8/8/8/P7/2k5/4B3 w - - 0 1", winning: true},
		{fen: "8/8/8/P7/7K/8/8/3kB3 w - - 0 1", winning: true},
		{fen: "k7/7B/8/8/4K2P/8/8/8 w - - 0 1", winning: true},
		{fen: "8/8/2B4P/8/1k6/8/8/4K3 w - - 0 1", winning: true},
		{fen: "8/1kB5/8/4K3/8/8/P7/8 b - - 0 1", winning: false},
		{fen: "8/8/B3K2P/8/8/8/8/1k6 b - - 0 1", winning: true},
		{fen: "8/8/8/7P/8/1B6/5K1k/8 w - - 0 1", winning: true},
		{fen: "4k3/K7/8/5B2/8/7P/8/8 w - - 0 1", winning: false},
		{fen: "7B/P7/K7/8/4k3/8/8/8 b - - 0 1", winning: true},
		{fen: "2B5/8/8/7P/8/8/8/1K3k2 b - - 0 1", winning: true},
		{fen: "8/7k/P7/2B5/8/8/4K3/8 w - - 0 1", winning: true},
		{fen: "8/7P/8/6K1/8/8/B2k4/8 w - - 0 1", winning: true},
		{fen: "k7/8/8/7P/6B1/8/3K4/8 b - - 0 1", winning: true},
		{fen: "8/P6k/5B2/8/8/1K6/8/8 w - - 0 1", winning: true},
		{fen: "2B5/7K/8/8/3k4/8/7P/8 b - - 0 1", winning: true},
		{fen: "8/8/3K4/2B5/P6k/8/8/8 b - - 0 1", winning: true},
		{fen: "8/8/7K/8/7P/2k5/8/5B2 b - - 0 1", winning: true},
		{fen: "8/8/4B2P/3k4/8/1K6/8/8 b - - 0 1", winning: true},
		{fen: "6k1/3K4/8/7P/8/8/4B3/8 b - - 0 1", winning: false},
		{fen: "8/8/8/8/8/P1B5/6k1/K7 b - - 0 1", winning: true},
		{fen: "8/4k3/6K1/7P/8/5B2/8/8 b - - 0 1", winning: true},
		{fen: "7k/2K5/8/7P/8/8/8/7B b - - 0 1", winning: false},
		{fen: "7K/7P/8/6k1/8/8/6B1/8 b - - 0 1", winning: true},
		{fen: "8/k5K1/2B5/8/7P/8/8/8 w - - 0 1", winning: true},
		{fen: "7k/8/6B1/2K5/8/7P/8/8 b - - 0 1", winning: false},
		{fen: "8/8/8/8/6BP/4k3/8/2K5 w - - 0 1", winning: true},
		{fen: "1k6/8/8/8/6K1/8/P2B4/8 b - - 0 1", winning: false},
		{fen: "8/8/8/6k1/7P/1B3K2/8/8 b - - 0 1", winning: false},
		{fen: "5k2/3B4/8/8/8/7P/3K4/8 w - - 0 1", winning: false},
		{fen: "8/8/8/6K1/8/8/7P/3B3k w - - 0 1", winning: true},
		{fen: "B7/8/8/8/K1k5/7P/8/8 w - - 0 1", winning: false},
		{fen: "8/P4k2/8/8/8/8/2K5/B7 w - - 0 1", winning: true},
		{fen: "4K3/7P/8/5k2/8/8/8/7B b - - 0 1", winning: true},
		{fen: "8/4B2k/8/P7/8/8/4K3/8 w - - 0 1", winning: true},
		{fen: "8/8/3K4/8/8/7P/5k2/3B4 w - - 0 1", winning: true},
		{fen: "8/8/P2k4/8/7B/8/2K5/8 w - - 0 1", winning: true},
		{fen: "8/3k4/8/6KP/8/8/8/3B4 w - - 0 1", winning: true},
		{fen: "8/P7/3B4/8/8/8/8/1k3K2 w - - 0 1", winning: true},
		{fen: "5K1B/8/8/P1k5/8/8/8/8 w - - 0 1", winning: false},
		{fen: "6K1/8/8/P7/8/B7/8/3k4 w - - 0 1", winning: true},
		{fen: "8/8/8/5K2/1Bk5/P7/8/8 b - - 0 1", winning: false},
		{fen: "8/8/5k2/8/8/3KB3/P7/8 b - - 0 1", winning: false},
		{fen: "8/8/8/5k1P/8/3K4/B7/8 w - - 0 1", winning: false},
		{fen: "8/8/8/8/5K2/P1B5/8/3k4 w - - 0 1", winning: true},
		{fen: "4k3/8/8/8/P7/B5K1/8/8 b - - 0 1", winning: false},
		{fen: "K7/4k3/7P/8/8/5B2/8/8 b - - 0 1", winning: false},
		{fen: "4k3/8/7P/8/8/K7/8/1B6 w - - 0 1", winning: true},
		{fen: "2k5/8/8/8/8/P7/K7/6B1 b - - 0 1", winning: false},
		{fen: "8/8/1k5K/8/8/8/P4B2/8 b - - 0 1", winning: false},
		{fen: "8/5k1P/8/8/K7/8/B7/8 b - - 0 1", winning: false},
		{fen: "8/8/3B4/8/7K/3k4/P7/8 w - - 0 1", winning: false},
		{fen: "7K/8/8/P7/1k6/8/8/6B1 w - - 0 1", winning: true},
		{fen: "8/8/8/7P/8/4kB2/7K/8 b - - 0 1", winning: true},
		{fen: "8/8/8/K7/7P/1k6/8/7B w - - 0 1", winning: true},
		{fen: "2K5/7P/8/8/8/8/3k4/5B2 w - - 0 1", winning: true},
		{fen: "2K5/8/6B1/8/8/5k1P/8/8 b - - 0 1", winning: false},
		{fen: "8/8/2B5/8/8/7P/2K1k3/8 b - - 0 1", winning: false},
		{fen: "8/8/8/7P/8/5B2/6K1/4k3 w - - 0 1", winning: true},
		{fen: "8/8/8/8/4B3/3K4/7P/6k1 b - - 0 1", winning: false},
		{fen: "8/8/7P/1B6/5K2/8/2k5/8 w - - 0 1", winning: true},
		{fen: "8/7P/8/8/5k2/8/1K6/1B6 w - - 0 1", winning: true},
		{fen: "5k2/2K5/8/8/P6B/8/8/8 b - - 0 1", winning: true},
		{fen: "8/7P/8/8/8/1B5k/8/7K b - - 0 1", winning: true},
		{fen: "8/8/8/8/2B5/3K3P/8/2k5 b - - 0 1", winning: true},
		{fen: "8/2B4k/8/8/P4K2/8/8/8 w - - 0 1", winning: true},
		{fen: "8/8/8/P1k5/8/8/2K5/4B3 w - - 0 1", winning: false},
		{fen: "8/8/3K4/8/P7/8/1B3k2/8 w - - 0 1", winning: true},
		{fen: "4B3/8/8/8/6k1/8/7P/1K6 w - - 0 1", winning: false},
		{fen: "1B6/8/P7/8/8/2k5/8/1K6 b - - 0 1", winning: true},
		{fen: "3k4/8/8/8/7P/5B2/8/6K1 b - - 0 1", winning: false},
		{fen: "1B6/8/8/8/P7/K7/2k5/8 w - - 0 1", winning: true},
		{fen: "4K3/7P/8/2k5/2B5/8/8/8 b - - 0 1", winning: true},
		{fen: "8/6k1/8/8/7P/8/K5B1/8 w - - 0 1", winning: false},
		{fen: "7K/8/B7/8/4k3/8/7P/8 b - - 0 1", winning: true},
		{fen: "k5B1/8/7P/8/8/1K6/8/8 b - - 0 1", winning: true},
		{fen: "8/P5k1/8/2K3B1/8/8/8/8 b - - 0 1", winning: true},
		{fen: "1B6/8/8/4K3/P7/2k5/8/8 w - - 0 1", winning: true},
		{fen: "8/8/8/7P/1K4B1/8/8/6k1 b - - 0 1", winning: true},
		{fen: "8/8/8/7P/4K3/8/6k1/3B4 b - - 0 1", winning: true},
		{fen: "8/8/8/2B5/8/P7/4k1K1/8 w - - 0 1", winning: true},
		{fen: "8/8/k6P/7K/8/8/8/1B6 b - - 0 1", winning: true},
		{fen: "6B1/8/8/8/7P/8/8/k2K4 w - - 0 1", winning: true},
		{fen: "8/8/P2B4/8/5K2/8/8/5k2 w - - 0 1", winning: true},
		{fen: "5K2/8/8/6B1/P7/2k5/8/8 b - - 0 1", winning: false},
		{fen: "8/8/4B3/4K2P/8/8/8/7k w - - 0 1", winning: true},
		{fen: "6K1/8/7B/P7/8/8/k7/8 w - - 0 1", winning: true},
		{fen: "B7/2K5/8/k7/8/8/7P/8 b - - 0 1", winning: true},
		{fen: "8/8/B6P/8/5k2/8/8/3K4 b - - 0 1", winning: true},
		{fen: "6B1/K7/2k4P/8/8/8/8/8 w - - 0 1", winning: true},
		{fen: "8/P7/1B4K1/8/8/8/8/3k4 b - - 0 1", winning: true},
		{fen: "8/k7/1B6/P7/8/8/8/1K6 b - - 0 1", winning: false},
		{fen: "7B/8/8/8/8/8/Pk6/6K1 b - - 0 1", winning: false},
		{fen: "8/P7/4k3/4B3/8/8/8/3K4 w - - 0 1", winning: true},
		{fen: "5K2/8/8/8/7P/6k1/2B5/8 b - - 0 1", winning: false},
		{fen: "2k5/B7/8/8/8/8/P7/K7 w - - 0 1", winning: false},
		{fen: "8/8/8/K3B3/8/P4k2/8/8 b - - 0 1", winning: true},
		{fen: "4k3/8/8/8/6K1/8/4B2P/8 b - - 0 1", winning: false},
		{fen: "8/7k/8/5KB1/P7/8/8/8 w - - 0 1", winning: true},
		{fen: "8/4k3/8/8/8/P7/8/3K2B1 b - - 0 1", winning: false},
		{fen: "5k2/4B3/8/P7/4K3/8/8/8 b - - 0 1", winning: true},
		{fen: "8/K7/8/4k3/B7/8/7P/8 b - - 0 1", winning: false},
		{fen: "6K1/8/7B/P7/2k5/8/8/8 w - - 0 1", winning: true},
		{fen: "5k2/8/K7/8/P7/B7/8/8 b - - 0 1", winning: true},
		{fen: "8/8/8/6k1/8/5K2/7P/5B2 b - - 0 1", winning: false},
		{fen: "6k1/8/8/4B3/8/7K/P7/8 w - - 0 1", winning: true},
		{fen: "8/P7/8/2K5/k7/8/5B2/8 b - - 0 1", winning: true},
		{fen: "8/5K1P/8/5B2/k7/8/8/8 b - - 0 1", winning: true},
		{fen: "5K2/8/k7/8/7P/5B2/8/8 b - - 0 1", winning: true},
		{fen: "8/8/2k5/8/8/4K3/PB6/8 w - - 0 1", winning: false},
		{fen: "2k5/8/8/1K6/7P/1B6/8/8 w - - 0 1", winning: true},
		{fen: "8/8/3k4/8/B7/3K3P/8/8 w - - 0 1", winning: false},
		{fen: "8/P4K2/2k5/8/8/8/1B6/8 b - - 0 1", winning: false},
		{fen: "8/7P/8/8/1K6/5k2/8/1B6 b - - 0 1", winning: true},
		{fen: "6k1/8/8/3K4/P7/8/5B2/8 w - - 0 1", winning: true},
		{fen: "8/8/8/K7/7P/8/2B1k3/8 b - - 0 1", winning: true},
		{fen: "8/P7/8/4B3/8/K7/8/4k3 w - - 0 1", winning: true},
		{fen: "8/8/7P/8/6K1/1B6/3k4/8 w - - 0 1", winning: true},
		{fen: "8/8/7k/2K5/8/8/P4B2/8 b - - 0 1", winning: true},
		{fen: "8/8/8/8/1k5P/3B4/8/2K5 b - - 0 1", winning: true},
		{fen: "6K1/1B6/8/7P/8/8/8/7k b - - 0 1", winning: true},
		{fen: "8/7K/P6B/2k5/8/8/8/8 w - - 0 1", winning: true},
		{fen: "8/8/4K3/7B/2k5/7P/8/8 w - - 0 1", winning: true},
		{fen: "8/8/P3K3/8/8/k7/7B/8 b - - 0 1", winning: true},
		{fen: "8/1B6/8/8/8/7P/4K1k1/8 b - - 0 1", winning: false},
		{fen: "8/8/5k2/8/6BP/4K3/8/8 b - - 0 1", winning: false},
		{fen: "8/7P/2K5/6k1/8/5B2/8/8 w - - 0 1", winning: true},
		{fen: "8/4B3/8/P7/8/8/8/4k1K1 b - - 0 1", winning: true},
		{fen: "7B/4K3/8/P7/8/8/4k3/8 w - - 0 1", winning: true},
		{fen: "4B3/1k6/8/8/7P/7K/8/8 w - - 0 1", winning: true},
		{fen: "k7/8/8/8/6K1/P7/7B/8 w - - 0 1", winning: false},
		{fen: "8/8/8/8/1k1B4/8/P1K5/8 b - - 0 1", winning: false},
		{fen: "8/2k5/8/8/6B1/7P/1K6/8 w - - 0 1", winning: false},
		{fen: "4K3/8/5B2/8/P7/k7/8/8 w - - 0 1", winning: true},
		{fen: "8/B7/P6K/8/1k6/8/8/8 b - - 0 1", winning: false},
		{fen: "7K/8/8/4k3/7P/5B2/8/8 w - - 0 1", winning: true},
		{fen: "8/8/4K3/8/P7/8/8/4k1B1 b - - 0 1", winning: true},
		{fen: "8/8/8/8/4K3/6k1/7P/5B2 b - - 0 1", winning: false},
		{fen: "8/8/5k2/2K5/PB6/8/8/8 b - - 0 1", winning: true},
		{fen: "8/8/7k/8/2B5/8/5K1P/8 b - - 0 1", winning: false},
		{fen: "5B2/8/P7/1K6/8/8/2k5/8 b - - 0 1", winning: true},
		{fen: "8/8/B3K3/7P/3k4/8/8/8 b - - 0 1", winning: true},
		{fen: "3k4/6K1/8/8/P6B/8/8/8 b - - 0 1", winning: false},
		{fen: "5k2/8/6K1/B7/8/8/P7/8 b - - 0 1", winning: false},
		{fen: "8/P7/1k6/6K1/8/8/8/B7 b - - 0 1", winning: false},
		{fen: "8/8/8/8/8/4B3/P7/K2k4 b - - 0 1", winning: true},
		{fen: "8/P7/8/8/7K/3k4/8/2B5 w - - 0 1", winning: true},
		{fen: "8/P5B1/8/8/7k/8/8/3K4 w - - 0 1", winning: true},
		{fen: "8/P7/8/8/1B6/2k2K2/8/8 b - - 0 1", winning: true},
		{fen: "3k4/8/5B2/8/8/8/P7/6K1 b - - 0 1", winning: false},
		{fen: "8/7P/5k2/3B4/8/8/8/2K5 b - - 0 1", winning: false},
		{fen: "5B1K/8/8/P6k/8/8/8/8 b - - 0 1", winning: true},
		{fen: "8/P7/8/8/3B4/4k3/6K1/8 b - - 0 1", winning: true},
		{fen: "8/1B5P/7K/8/7k/8/8/8 w - - 0 1", winning: true},
		{fen: "8/8/8/7P/2k5/1B6/1K6/8 b - - 0 1", winning: true},
		{fen: "1k1B4/8/8/8/3K4/P7/8/8 b - - 0 1", winning: false},
		{fen: "8/8/8/8/4kB2/P7/8/6K1 w - - 0 1", winning: false},
		{fen: "1B6/8/8/8/8/8/PK4k1/8 w - - 0 1", winning: true},
		{fen: "8/8/8/8/2Bk3P/8/8/2K5 b - - 0 1", winning: false},
		{fen: "5B2/8/2K5/P4k2/8/8/8/8 b - - 0 1", winning: true},
		{fen: "8/5B2/K7/8/8/2k5/7P/8 b - - 0 1", winning: false},
		{fen: "6B1/3K4/7P/8/8/7k/8/8 b - - 0 1", winning: true},
		{fen: "8/8/3K4/8/7P/8/8/3k1B2 w - - 0 1", winning: true},
		{fen: "8/8/6k1/3K4/8/8/B6P/8 b - - 0 1", winning: false},
		{fen: "8/7k/8/7P/6B1/8/3K4/8 w - - 0 1", winning: false},
		{fen: "8/8/7B/P7/8/8/1K2k3/8 w - - 0 1", winning: true},
		{fen: "2k5/8/2B5/8/8/8/7P/7K w - - 0 1", winning: false},
		{fen: "8/8/B7/8/K7/6kP/8/8 w - - 0 1", winning: false},
		{fen: "7B/8/8/8/8/P2K1k2/8/8 w - - 0 1", winning: true},
		{fen: "1K6/8/6B1/8/3k4/8/7P/8 b - - 0 1", winning: false},
		{fen: "8/8/8/6B1/P7/5k2/8/7K w - - 0 1", winning: true},
		{fen: "K7/8/8/8/8/5k2/P7/6B1 b - - 0 1", winning: true},
		{fen: "8/8/7P/8/8/1K6/8/3k1B2 w - - 0 1", winning: true},
		{fen: "8/8/K7/8/1B6/P7/5k2/8 w - - 0 1", winning: true},
		{fen: "2k2B2/8/2K5/8/8/P7/8/8 w - - 0 1", winning: true},
		{fen: "8/1k1B4/8/8/6K1/8/7P/8 b - - 0 1", winning: true},
		{fen: "3k4/8/8/P7/7B/1K6/8/8 b - - 0 1", winning: false},
		{fen: "8/8/8/1K6/8/8/P3k2B/8 w - - 0 1", winning: true},
		{fen: "8/8/8/2K5/6B1/7P/8/1k6 w - - 0 1", winning: true},
		{fen: "2k5/8/2K5/P7/1B6/8/8/8 b - - 0 1", winning: false},
		{fen: "8/2B5/P7/6k1/8/8/8/5K2 w - - 0 1", winning: true},
		{fen: "8/8/8/P5B1/8/6K1/8/6k1 w - - 0 1", winning: true},
		{fen: "8/8/7P/8/5k1K/5B2/8/8 b - - 0 1", winning: true},
		{fen: "8/k7/7B/8/8/P7/1K6/8 w - - 0 1", winning: false},
		{fen: "3B4/k7/8/8/7K/P7/8/8 w - - 0 1", winning: false},
		{fen: "8/k7/8/7B/7P/8/1K6/8 b - - 0 1", winning: true},
		{fen: "3B4/8/7K/8/8/P4k2/8/8 w - - 0 1", winning: true},
		{fen: "8/8/6K1/8/P2k4/6B1/8/8 b - - 0 1", winning: false},
		{fen: "8/P5K1/8/8/8/8/8/2B2k2 b - - 0 1", winning: true},
		{fen: "7K/B7/8/8/P7/3k4/8/8 w - - 0 1", winning: false},
		{fen: "2k5/8/8/8/7P/K7/8/5B2 b - - 0 1", winning: false},
		{fen: "8/8/7P/1B3K2/8/1k6/8/8 w - - 0 1", winning: true},
		{fen: "8/4B3/8/P7/4K3/8/8/6k1 w - - 0 1", winning: true},
		{fen: "8/8/2K4k/8/7P/8/8/5B2 w - - 0 1", winning: false},
		{fen: "8/8/8/1B6/8/7P/8/4k1K1 w - - 0 1", winning: true},
		{fen: "8/8/6K1/3k3P/8/8/8/3B4 b - - 0 1", winning: true},
		{fen: "8/8/8/1B1k4/8/8/1K5P/8 w - - 0 1", winning: false},
		{fen: "4K2B/1k6/8/8/8/P7/8/8 b - - 0 1", winning: false},
		{fen: "8/6B1/8/P7/8/8/3k4/1K6 w - - 0 1", winning: true},
		{fen: "8/8/7P/8/7K/4k3/8/7B b - - 0 1", winning: true},
		{fen: "1B6/8/P7/8/8/8/k4K2/8 w - - 0 1", winning: true},
		{fen: "8/8/3K3P/3B4/8/4k3/8/8 b - - 0 1", winning: true},
		{fen: "2K5/8/7P/1k6/8/3B4/8/8 b - - 0 1", winning: true},
		{fen: "8/8/k7/8/8/P5BK/8/8 w - - 0 1", winning: false},
		{fen: "8/8/2B5/4k3/2K5/8/7P/8 b - - 0 1", winning: false},
		{fen: "8/8/5BK1/P1k5/8/8/8/8 b - - 0 1", winning: false},
		{fen: "K7/7k/7P/8/8/5B2/8/8 b - - 0 1", winning: false},
		{fen: "3B4/P7/8/1K6/8/8/3k4/8 b - - 0 1", winning: true},
		{fen: "5K2/8/8/2B5/8/8/P7/3k4 b - - 0 1", winning: true},
		{fen: "8/8/7P/7k/8/1B6/8/6K1 b - - 0 1", winning: false},
		{fen: "8/8/7P/5k2/4B3/5K2/8/8 b - - 0 1", winning: true},
		{fen: "3k4/8/8/8/8/5K2/B6P/8 b - - 0 1", winning: false},
		{fen: "k7/8/K7/7P/8/8/B7/8 b - - 0 1", winning: true},
		{fen: "8/B2K4/8/P1k5/8/8/8/8 b - - 0 1", winning: true},
		{fen: "8/3B4/8/k5K1/8/8/7P/8 w - - 0 1", winning: true},
		{fen: "8/8/8/8/7K/5k2/P7/B7 b - - 0 1", winning: false},
		{fen: "1B1K4/8/8/P2k4/8/8/8/8 w - - 0 1", winning: true},
		{fen: "8/7k/8/2B2K2/8/P7/8/8 b - - 0 1", winning: true},
		{fen: "8/8/8/P7/8/8/k7/B2K4 b - - 0 1", winning: true},
		{fen: "8/8/8/P7/1k3B2/8/8/1K6 w - - 0 1", winning: true},
		{fen: "8/5B2/7P/5K2/k7/8/8/8 b - - 0 1", winning: true},
		{fen: "8/8/7K/2k5/8/7P/6B1/8 b - - 0 1", winning: true},
		{fen: "8/1B6/8/8/8/1K2k2P/8/8 b - - 0 1", winning: false},
		{fen: "8/5KB1/8/8/8/P7/3k4/8 b - - 0 1", winning: true},
		{fen: "8/8/K6P/8/8/8/k7/3B4 w - - 0 1", winning: true},
		{fen: "8/8/1B6/8/P7/8/2k5/4K3 b - - 0 1", winning: true},
		{fen: "1K6/3B4/8/8/8/4k3/7P/8 b - - 0 1", winning: false},
		{fen: "2k5/7K/P6B/8/8/8/8/8 b - - 0 1", winning: false},
		{fen: "8/8/1K1B4/P7/k7/8/8/8 b - - 0 1", winning: true},
		{fen: "8/8/8/7P/2k5/8/2B5/K7 b - - 0 1", winning: true},
		{fen: "6K1/7P/8/6k1/2B5/8/8/8 b - - 0 1", winning: true},
		{fen: "8/8/2K5/P7/8/B7/8/1k6 w - - 0 1", winning: true},
		{fen: "8/2k5/5B2/8/P6K/8/8/8 w - - 0 1", winning: false},
		{fen: "8/B7/P7/8/8/2K5/8/3k4 b - - 0 1", winning: true},
		{fen: "5B2/1k6/P7/8/1K6/8/8/8 b - - 0 1", winning: false},
		{fen: "1B6/8/P3k3/8/8/8/6K1/8 w - - 0 1", winning: true},
		{fen: "8/PK2B3/8/5k2/8/8/8/8 b - - 0 1", winning: true},
		{fen: "8/2B5/8/8/8/P7/k7/7K b - - 0 1", winning: false},
		{fen: "8/8/5k2/5B2/8/8/7P/7K b - - 0 1", winning: false},
		{fen: "8/8/7P/5B2/8/5k2/1K6/8 b - - 0 1", winning: true},
		{fen: "7k/8/8/8/8/8/PB6/6K1 b - - 0 1", winning: true},
		{fen: "k7/8/8/P7/1B6/8/2K5/8 w - - 0 1", winning: false},
		{fen: "8/7P/8/8/5k2/K7/2B5/8 w - - 0 1", winning: true},
		{fen: "8/8/8/1k6/8/2B5/P7/7K b - - 0 1", winning: false},
		{fen: "8/3K1k2/5B2/8/8/P7/8/8 w - - 0 1", winning: true},
		{fen: "8/2B5/P1K5/6k1/8/8/8/8 w - - 0 1", winning: true},
		{fen: "8/P7/4K1k1/8/8/8/3B4/8 b - - 0 1", winning: true},
		{fen: "8/7P/8/8/8/k4B2/8/K7 w - - 0 1", winning: true},
		{fen: "8/8/B3k3/8/8/7P/7K/8 w - - 0 1", winning: false},
		{fen: "8/P3B3/8/8/8/3k4/5K2/8 w - - 0 1", winning: true},
		{fen: "2K5/8/8/P7/3B4/3k4/8/8 w - - 0 1", winning: true},
		{fen: "3K4/8/7P/6k1/8/8/8/5B2 w - - 0 1", winning: true},
		{fen: "6K1/8/6B1/2k4P/8/8/8/8 b - - 0 1", winning: true},
		{fen: "6K1/8/8/B6k/8/P7/8/8 w - - 0 1", winning: true},
		{fen: "2K5/8/8/8/P7/8/8/4k1B1 w - - 0 1", winning: true},
		{fen: "1k6/8/8/8/2K5/8/P2B4/8 b - - 0 1", winning: false},
		{fen: "8/6B1/5K2/8/6k1/P7/8/8 b - - 0 1", winning: true},
		{fen: "8/6K1/P4B2/8/8/6k1/8/8 w - - 0 1", winning: true},
		{fen: "2B5/8/7P/3K4/7k/8/8/8 b - - 0 1", winning: true},
		{fen: "6K1/8/3B4/8/8/8/P7/1k6 b - - 0 1", winning: false},
		{fen: "8/8/K7/8/7k/7P/8/3B4 b - - 0 1", winning: false},
		{fen: "8/8/8/8/P6k/8/4K2B/8 b - - 0 1", winning: true},
		{fen: "8/1B6/1k6/7P/8/8/8/5K2 b - - 0 1", winning: true},
		{fen: "8/P6k/8/8/8/BK6/8/8 w - - 0 1", winning: true},
		{fen: "8/K7/P4B2/8/8/8/8/5k2 b - - 0 1", winning: true},
		{fen: "8/P6k/8/B7/8/8/8/4K3 b - - 0 1", winning: true},
		{fen: "8/8/K1k5/8/8/7P/2B5/8 w - - 0 1", winning: false},
		{fen: "6k1/8/7P/8/8/2K5/6B1/8 b - - 0 1", winning: false},
		{fen: "8/8/8/8/1B4K1/3k4/P7/8 w - - 0 1", winning: false},
		{fen: "B7/8/8/8/1K5P/4k3/8/8 w - - 0 1", winning: true},
		{fen: "1K6/8/3B4/8/8/P7/8/6k1 b - - 0 1", winning: true},
		{fen: "B7/8/4k2P/8/8/6K1/8/8 b - - 0 1", winning: false},
		{fen: "B1K5/8/8/8/8/4k3/7P/8 w - - 0 1", winning: true},
		{fen: "8/7P/8/K7/8/8/3k2B1/8 b - - 0 1", winning: true},
		{fen: "7k/4K3/8/7P/8/8/6B1/8 w - - 0 1", winning: false},
		{fen: "8/8/8/4K3/7P/5B2/8/2k5 w - - 0 1", winning: true},
		{fen: "8/8/8/1k6/P7/5K2/8/2B5 b - - 0 1", winning: false},
		{fen: "6B1/8/1k6/8/8/8/7P/2K5 w - - 0 1", winning: true},
		{fen: "7K/7P/8/8/B7/8/8/6k1 w - - 0 1", winning: true},
		{fen: "8/8/3B4/8/8/8/P4k1K/8 b - - 0 1", winning: false},
		{fen: "8/8/7P/K7/8/8/2B2k2/8 b - - 0 1", winning: true},
		{fen: "8/8/6k1/1K6/6BP/8/8/8 w - - 0 1", winning: false},
		{fen: "B6k/8/8/6K1/7P/8/8/8 w - - 0 1", winning: false},
		{fen: "8/1B1k4/8/7P/4K3/8/8/8 b - - 0 1", winning: false},
		{fen: "8/8/8/2K5/8/6B1/P7/4k3 b - - 0 1", winning: true},
		{fen: "8/4B3/8/Pk2K3/8/8/8/8 w - - 0 1", winning: false},
		{fen: "8/5k2/8/7P/8/7K/8/5B2 w - - 0 1", winning: false},
		{fen: "8/8/7P/8/6K1/7B/7k/8 w - - 0 1", winning: true},
		{fen: "8/7K/8/7B/7P/3k4/8/8 w - - 0 1", winning: true},
		{fen: "1K6/8/8/2k4P/8/8/4B3/8 w - - 0 1", winning: true},
		{fen: "8/8/5B2/P2k4/8/8/8/1K6 w - - 0 1", winning: true},
		{fen: "8/7P/8/8/8/6kB/8/5K2 w - - 0 1", winning: true},
		{fen: "8/8/8/8/8/8/P6K/2B1k3 w - - 0 1", winning: true},
		{fen: "3B4/8/3K4/P4k2/8/8/8/8 b - - 0 1", winning: true},
		{fen: "8/4k3/8/8/8/8/1K2B2P/8 w - - 0 1", winning: false},
		{fen: "8/P1k5/8/4K3/3B4/8/8/8 b - - 0 1", winning: false},
		{fen: "8/8/3k4/P7/8/1KB5/8/8 w - - 0 1", winning: false},
		{fen: "4K3/8/8/8/7k/7P/8/1B6 w - - 0 1", winning: true},
		{fen: "3k4/8/2K5/8/7P/8/B7/8 b - - 0 1", winning: false},
		{fen: "8/8/K7/7P/8/3B4/8/k7 b - - 0 1", winning: true},
		{fen: "8/8/P7/8/2k5/7K/1B6/8 b - - 0 1", winning: true},
		{fen: "7B/8/8/7K/8/8/P3k3/8 b - - 0 1", winning: false},
		{fen: "8/8/8/5B1P/8/1K6/8/k7 w - - 0 1", winning: true},
		{fen: "8/8/1k6/8/8/P1B5/5K2/8 w - - 0 1", winning: false},
		{fen: "4K3/1B6/8/8/8/8/7P/7k b - - 0 1", winning: false},
		{fen: "8/P7/8/6k1/1B6/8/2K5/8 w - - 0 1", winning: true},
		{fen: "8/8/8/2k5/8/5K1P/2B5/8 b - - 0 1", winning: false},
		{fen: "8/8/4k2P/8/8/7B/5K2/8 b - - 0 1", winning: false},
		{fen: "7B/8/8/P7/4k1K1/8/8/8 b - - 0 1", winning: false},
		{fen: "B4K2/8/8/7P/5k2/8/8/8 w - - 0 1", winning: true},
		{fen: "8/8/8/1k6/8/8/2B4P/2K5 w - - 0 1", winning: true},
		{fen: "5k2/8/8/8/1B2K3/P7/8/8 b - - 0 1", winning: true},
		{fen: "2B5/8/8/7P/5K2/8/8/6k1 w - - 0 1", winning: true},
		{fen: "8/2K4P/8/8/8/8/7k/1B6 b - - 0 1", winning: true},
		{fen: "2B5/8/5k2/K7/8/7P/8/8 w - - 0 1", winning: false},
		{fen: "2K5/8/8/P7/8/4B3/8/5k2 w - - 0 1", winning: true},
		{fen: "8/8/3k3P/8/2B5/8/8/1K6 w - - 0 1", winning: true},
		{fen: "4K3/8/8/6k1/8/3B4/7P/8 w - - 0 1", winning: true},
		{fen: "8/8/8/8/8/P7/3K2k1/6B1 w - - 0 1", winning: true},
		{fen: "2K5/7P/8/8/B7/1k6/8/8 b - - 0 1", winning: true},
		{fen: "1B6/k7/P7/K7/8/8/8/8 b - - 0 1", winning: false},
		{fen: "8/4k3/8/8/3K2B1/7P/8/8 b - - 0 1", winning: false},
		{fen: "8/5K2/8/7B/7P/8/8/k7 w - - 0 1", winning: true},
		{fen: "8/8/5k2/3K4/3B4/P7/8/8 b - - 0 1", winning: true},
		{fen: "k7/8/5K2/8/8/8/7P/3B4 w - - 0 1", winning: true},
		{fen: "8/3K4/8/4k3/8/7B/7P/8 b - - 0 1", winning: false},
		{fen: "5B1K/8/P7/8/8/8/8/6k1 w - - 0 1", winning: true},
		{fen: "8/8/5B2/P7/8/k7/6K1/8 b - - 0 1", winning: true},
		{fen: "5B2/8/4k3/8/8/P7/4K3/8 b - - 0 1", winning: false},
		{fen: "8/8/8/8/P2B4/1k6/8/4K3 w - - 0 1", winning: true},
		{fen: "8/5B2/7P/8/7K/8/4k3/8 w - - 0 1", winning: true},
		{fen: "B1K5/8/8/4k2P/8/8/8/8 w - - 0 1", winning: false},
		{fen: "k7/8/8/8/8/8/7P/5BK1 b - - 0 1", winning: true},
		{fen: "8/8/P1K5/5k2/8/8/3B4/8 b - - 0 1", winning: true},
		{fen: "8/5B2/K7/8/8/7P/6k1/8 b - - 0 1", winning: false},
		{fen: "8/8/8/8/4K3/P1B5/8/k7 b - - 0 1", winning: true},
		{fen: "8/8/4B3/8/7k/7P/K7/8 w - - 0 1", winning: false},
		{fen: "3k4/7P/8/8/B7/8/8/7K w - - 0 1", winning: true},
		{fen: "8/5B2/6kP/1K6/8/8/8/8 b - - 0 1", winning: false},
		{fen: "3K4/6k1/7P/8/B7/8/8/8 b - - 0 1", winning: false},
		{fen: "8/8/1K3B2/8/k7/P7/8/8 b - - 0 1", winning: false},
		{fen: "8/8/8/5B1P/8/k6K/8/8 b - - 0 1", winning: true},
		{fen: "8/PK6/8/8/6kB/8/8/8 w - - 0 1", winning: true},
		{fen: "3B4/8/PK2k3/8/8/8/8/8 b - - 0 1", winning: true},
		{fen: "8/8/8/7B/8/3K4/7P/1k6 b - - 0 1", winning: true},
		{fen: "4B3/7P/8/8/8/7k/1K6/8 w - - 0 1", winning: true},
		{fen: "k7/4B3/8/P7/2K5/8/8/8 b - - 0 1", winning: false},
		{fen: "8/P7/k7/8/1B6/8/2K5/8 w - - 0 1", winning: true},
		{fen: "2k5/P7/8/8/1B6/8/2K5/8 w - - 0 1", winning: true},
		{fen: "8/3k4/7P/4K3/8/8/8/1B6 b - - 0 1", winning: true},
		{fen: "8/3K4/Pk6/8/5B2/8/8/8 w - - 0 1", winning: false},
		{fen: "8/8/8/8/P6B/3K4/8/5k2 b - - 0 1", winning: true},
		{fen: "8/7P/8/8/8/1B2K3/8/1k6 b - - 0 1", winning: true},
		{fen: "6B1/7P/8/8/8/4K3/1k6/8 w - - 0 1", winning: true},
		{fen: "8/8/8/P7/2K1k3/8/8/6B1 w - - 0 1", winning: true},
		{fen: "8/6k1/8/P7/5B2/8/7K/8 b - - 0 1", winning: true},
		{fen: "8/6k1/B7/3K4/8/8/7P/8 b - - 0 1", winning: false},
		{fen: "7k/P7/8/8/8/8/3B4/3K4 w - - 0 1", winning: true},
		{fen: "8/8/8/P7/3K3k/8/8/2B5 w - - 0 1", winning: true},
		{fen: "8/8/6B1/8/5k1P/1K6/8/8 w - - 0 1", winning: false},
		{fen: "8/8/P7/3kB1K1/8/8/8/8 b - - 0 1", winning: false},
		{fen: "8/7P/8/7B/8/8/1K6/5k2 w - - 0 1", winning: true},
		{fen: "8/P7/8/8/8/8/5Bk1/3K4 w - - 0 1", winning: true},
		{fen: "8/2k5/7P/7K/8/7B/8/8 b - - 0 1", winning: true},
		{fen: "8/8/2k5/8/3K4/P7/5B2/8 b - - 0 1", winning: false},
		{fen: "3k4/8/8/8/K2B4/P7/8/8 b - - 0 1", winning: false},
		{fen: "K7/8/8/8/P7/4k3/8/4B3 w - - 0 1", winning: true},
		{fen: "8/P7/8/K2k4/1B6/8/8/8 w - - 0 1", winning: true},
		{fen: "8/6k1/3K3B/8/8/8/P7/8 b - - 0 1", winning: true},
		{fen: "8/7P/8/8/5k2/8/1K6/7B w - - 0 1", winning: true},
		{fen: "8/P5k1/8/8/8/4K3/8/2B5 w - - 0 1", winning: true},
		{fen: "8/8/8/1K6/7P/1B6/7k/8 b - - 0 1", winning: true},
		{fen: "8/3K4/6k1/8/7P/8/4B3/8 b - - 0 1", winning: false},
		{fen: "8/8/8/4k1KP/8/1B6/8/8 w - - 0 1", winning: true},
		{fen: "B6K/8/7P/8/k7/8/8/8 b - - 0 1", winning: true},
		{fen: "8/B7/8/5K2/8/P4k2/8/8 w - - 0 1", winning: true},
		{fen: "8/3k4/8/3K4/P7/8/8/4B3 b - - 0 1", winning: false},
		{fen: "8/8/3k4/8/P7/6K1/3B4/8 b - - 0 1", winning: false},
		{fen: "8/8/8/P7/7B/2k5/7K/8 b - - 0 1", winning: true},
		{fen: "8/4k3/8/3B4/8/8/3K3P/8 b - - 0 1", winning: false},
		{fen: "6K1/8/7k/8/P7/8/8/6B1 b - - 0 1", winning: true},
		{fen: "8/8/7P/8/8/8/2k1K1B1/8 w - - 0 1", winning: true},
		{fen: "7K/k7/8/7P/6B1/8/8/8 b - - 0 1", winning: true},
		{fen: "5B2/8/8/P4K2/8/8/8/1k6 b - - 0 1", winning: true},
		{fen: "8/8/8/8/P3K3/8/1k6/2B5 b - - 0 1", winning: true},
		{fen: "8/P7/7B/5k2/K7/8/8/8 b - - 0 1", winning: true},
		{fen: "8/8/8/1K6/8/8/PB1k4/8 w - - 0 1", winning: true},
		{fen: "7B/5K2/k7/8/8/P7/8/8 w - - 0 1", winning: false},
		{fen: "8/8/2k3BP/8/5K2/8/8/8 w - - 0 1", winning: true},
		{fen: "8/6B1/8/8/P7/k7/7K/8 w - - 0 1", winning: true},
		{fen: "8/k7/7P/8/3K4/8/8/1B6 b - - 0 1", winning: true},
		{fen: "8/6K1/P6B/8/8/8/8/k7 b - - 0 1", winning: true},
		{fen: "2B5/8/7P/8/8/K2k4/8/8 b - - 0 1", winning: true},
		{fen: "8/2K5/7P/2k5/8/8/6B1/8 w - - 0 1", winning: true},
		{fen: "8/8/8/P7/8/k3B3/2K5/8 w - - 0 1", winning: true},
		{fen: "8/P7/4K3/2k5/8/8/8/4B3 b - - 0 1", winning: true},
		{fen: "7B/P7/8/8/3k4/K7/8/8 b - - 0 1", winning: true},
		{fen: "8/8/6K1/8/4k3/P7/8/2B5 b - - 0 1", winning: false},
		{fen: "8/4K3/8/8/1k6/8/7P/5B2 w - - 0 1", winning: true},
		{fen: "6B1/7P/8/8/8/8/1K3k2/8 w - - 0 1", winning: true},
		{fen: "4K3/8/P7/8/8/B7/4k3/8 b - - 0 1", winning: true},
		{fen: "3B4/8/k7/8/P7/8/8/6K1 w - - 0 1", winning: false},
		{fen: "5k2/8/8/7P/8/2K5/4B3/8 b - - 0 1", winning: false},
		{fen: "5B2/P3K3/8/8/k7/8/8/8 b - - 0 1", winning: true},
		{fen: "8/8/3KB3/7P/8/8/3k4/8 b - - 0 1", winning: true},
		{fen: "8/1K1B4/7P/8/8/2k5/8/8 w - - 0 1", winning: true},
		{fen: "8/P5B1/K7/8/8/1k6/8/8 w - - 0 1", winning: true},
		{fen: "8/8/8/8/1k6/8/7P/1K3B2 w - - 0 1", winning: true},
		{fen: "8/4k3/8/1B5P/8/8/8/K7 b - - 0 1", winning: false},
		{fen: "8/6K1/B7/7P/8/4k3/8/8 b - - 0 1", winning: true},
		{fen: "8/8/8/8/4k2K/8/P7/6B1 b - - 0 1", winning: false},
		{fen: "7k/8/8/8/7P/8/4B2K/8 b - - 0 1", winning: false},
		{fen: "2K5/8/8/8/7P/8/4B3/3k4 b - - 0 1", winning: true},
		{fen: "8/6B1/8/P7/8/8/3K4/5k2 b - - 0 1", winning: true},
		{fen: "8/8/4B3/6k1/7P/8/8/1K6 b - - 0 1", winning: false},
		{fen: "6k1/8/8/8/8/B7/P7/3K4 b - - 0 1", winning: false},
		{fen: "8/3K4/8/4B3/8/8/P6k/8 b - - 0 1", winning: true},
		{fen: "K7/4B3/8/8/7k/8/P7/8 b - - 0 1", winning: true},
		{fen: "8/8/8/8/P7/4B2K/1k6/8 b - - 0 1", winning: true},
		{fen: "1B6/8/8/6k1/P7/8/8/7K w - - 0 1", winning: true},
		{fen: "8/2K5/P7/6B1/8/8/8/7k w - - 0 1", winning: true},
		{fen: "8/7k/7P/8/K7/8/8/7B b - - 0 1", winning: false},
		{fen: "8/3B4/8/7P/8/2K5/6k1/8 w - - 0 1", winning: true},
		{fen: "8/8/P7/B7/K7/7k/8/8 w - - 0 1", winning: true},
		{fen: "2k5/7K/8/8/8/4B3/P7/8 b - - 0 1", winning: false},
		{fen: "8/3K4/P7/8/8/6B1/8/6k1 w - - 0 1", winning: true},
		{fen: "8/8/P7/K7/8/B7/8/4k3 b - - 0 1", winning: true},
		{fen: "8/3k4/8/8/8/1B5P/4K3/8 w - - 0 1", winning: false},
		{fen: "2K5/8/6kP/8/8/8/8/5B2 w - - 0 1", winning: false},
		{fen: "8/8/8/1B6/8/3K4/7P/1k6 b - - 0 1", winning: true},
		{fen: "1K6/7k/P7/8/8/8/1B6/8 w - - 0 1", winning: true},
		{fen: "8/4K3/8/8/8/3B4/7P/7k b - - 0 1", winning: false},
		{fen: "1B6/4k3/8/P3K3/8/8/8/8 b - - 0 1", winning: false},
		{fen: "8/1K5k/8/8/8/8/PB6/8 b - - 0 1", winning: true},
		{fen: "8/1K6/8/k7/4B3/8/7P/8 b - - 0 1", winning: true},
		{fen: "8/8/6k1/8/P5K1/8/1B6/8 w - - 0 1", winning: true},
		{fen: "8/8/6k1/6B1/8/8/P6K/8 w - - 0 1", winning: true},
		{fen: "k7/2B4K/8/P7/8/8/8/8 b - - 0 1", winning: false},
		{fen: "8/8/7k/7B/7P/8/K7/8 b - - 0 1", winning: false},
		{fen: "4K3/8/1k6/8/8/7B/7P/8 w - - 0 1", winning: true},
		{fen: "8/6KP/8/1B6/8/5k2/8/8 w - - 0 1", winning: true},
		{fen: "8/8/8/8/2k2B2/6K1/P7/8 b - - 0 1", winning: false},
		{fen: "4k3/8/P1K5/B7/8/8/8/8 w - - 0 1", winning: true},
		{fen: "8/4B3/1k6/8/P3K3/8/8/8 b - - 0 1", winning: false},
		{fen: "8/8/K7/2k4P/8/7B/8/8 w - - 0 1", winning: true},
		{fen: "8/B7/8/8/8/P7/K1k5/8 b - - 0 1", winning: false},
		{fen: "3k4/8/8/4K3/P6B/8/8/8 b - - 0 1", winning: false},
		{fen: "8/5K2/P7/8/1k6/8/5B2/8 b - - 0 1", winning: true},
		{fen: "8/7P/6k1/8/8/1K6/8/7B b - - 0 1", winning: false},
		{fen: "8/8/8/3K4/3B4/P7/7k/8 w - - 0 1", winning: true},
		{fen: "4B2k/8/8/8/8/3K4/7P/8 b - - 0 1", winning: false},
		{fen: "8/8/5K2/8/7B/P7/8/3k4 b - - 0 1", winning: true},
		{fen: "8/8/7P/2kB4/K7/8/8/8 b - - 0 1", winning: true},
		{fen: "8/P3k3/8/8/K7/2B5/8/8 w - - 0 1", winning: true},
		{fen: "8/4B3/8/2k2K2/8/P7/8/8 b - - 0 1", winning: false},
		{fen: "6K1/P7/8/k7/8/B7/8/8 b - - 0 1", winning: true},
		{fen: "1B6/8/PK6/8/1k6/8/8/8 w - - 0 1", winning: true},
		{fen: "8/8/7P/2K5/8/1k6/8/1B6 w - - 0 1", winning: true},
		{fen: "1kB2K2/8/8/8/7P/8/8/8 w - - 0 1", winning: true},
		{fen: "8/8/8/KB1k3P/8/8/8/8 w - - 0 1", winning: true},
		{fen: "6K1/8/7P/8/8/4k3/B7/8 b - - 0 1", winning: true},
		{fen: "5k2/B7/6K1/8/8/P7/8/8 w - - 0 1", winning: true},
		{fen: "2k5/8/8/K7/P7/8/8/B7 b - - 0 1", winning: false},
		{fen: "8/8/7P/3k2K1/8/5B2/8/8 b - - 0 1", winning: true},
		{fen: "8/8/PB6/6K1/8/8/6k1/8 b - - 0 1", winning: true},
		{fen: "6K1/8/8/8/B4k2/8/7P/8 b - - 0 1", winning: true},
		{fen: "8/8/6B1/3k4/8/7P/8/7K b - - 0 1", winning: false},
		{fen: "8/8/8/8/5k2/8/4B2P/K7 w - - 0 1", winning: false},
		{fen: "8/8/4k1K1/P7/5B2/8/8/8 w - - 0 1", winning: true},
	}

	e := New[Score]()
	failures := make([]kbpvkCase, 0)

	for _, tt := range tests {
		b := Must(board.FromFEN(tt.fen))
		got := e.kbpvk(b, &Coefficients)
		if b.STM == Black {
			got = -got
		}

		assert.GreaterOrEqual(t, got, Score(0), "fen %s", tt.fen)
		if tt.winning && got <= Score(300) {
			failures = append(failures, tt)
		} else if !tt.winning && got >= Score(100) {
			failures = append(failures, tt)
		}
	}

	if len(failures) > 8 {
		for _, tt := range failures {
			t.Run(tt.fen, func(t *testing.T) {

				b := Must(board.FromFEN(tt.fen))
				got := e.kbpvk(b, &Coefficients)
				if b.STM == Black {
					got = -got
				}
				if tt.winning {
					assert.Greater(t, got, Score(300), "fen %s winning", tt.fen)
				} else if !tt.winning {
					assert.Less(t, got, Score(100), "fen %s drawing", tt.fen)
				}
			})
		}
	}
}
