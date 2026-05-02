package picker_test

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/heur"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"
	"github.com/paulsonkoly/chess-3/picker"
	"github.com/paulsonkoly/chess-3/stack"
	"github.com/stretchr/testify/assert"
)

func TestAllMoves(t *testing.T) {
	tests := []struct {
		fen string
	}{
		{"r3k2r/2pb1ppp/2pp1q2/p7/1nP1B3/1P2P3/P2N1PPP/R2QK2R w KQkq a6 0 14"},
		{"4rrk1/2p1b1p1/p1p3q1/4p3/2P2n1p/1P1NR2P/PB3PP1/3R1QK1 b - - 2 24"},
		{"r3qbrk/6p1/2b2pPp/p3pP1Q/PpPpP2P/3P1B2/2PB3K/R5R1 w - - 16 42"},
		{"6k1/1R3p2/6p1/2Bp3p/3P2q1/P7/1P2rQ1K/5R2 b - - 4 44"},
		{"8/8/1p2k1p1/3p3p/1p1P1P1P/1P2PK2/8/8 w - - 3 54"},
		{"7r/2p3k1/1p1p1qp1/1P1Bp3/p1P2r1P/P7/4R3/Q4RK1 w - - 0 36"},
		{"r1bq1rk1/pp2b1pp/n1pp1n2/3P1p2/2P1p3/2N1P2N/PP2BPPP/R1BQ1RK1 b - - 2 10"},
		{"3r3k/2r4p/1p1b3q/p4P2/P2Pp3/1B2P3/3BQ1RP/6K1 w - - 3 87"},
		{"2r4r/1p4k1/1Pnp4/3Qb1pq/8/4BpPp/5P2/2RR1BK1 w - - 0 42"},
		{"4q1bk/6b1/7p/p1p4p/PNPpP2P/KN4P1/3Q4/4R3 b - - 0 37"},
		{"2q3r1/1r2pk2/pp3pp1/2pP3p/P1Pb1BbP/1P4Q1/R3NPP1/4R1K1 w - - 2 34"},
		{"1r2r2k/1b4q1/pp5p/2pPp1p1/P3Pn2/1P1B1Q1P/2R3P1/4BR1K b - - 1 37"},
		{"r3kbbr/pp1n1p1P/3ppnp1/q5N1/1P1pP3/P1N1B3/2P1QP2/R3KB1R b KQkq b3 0 17"},
		{"8/6pk/2b1Rp2/3r4/1R1B2PP/P5K1/8/2r5 b - - 16 42"},
		{"1r4k1/4ppb1/2n1b1qp/pB4p1/1n1BP1P1/7P/2PNQPK1/3RN3 w - - 8 29"},
		{"8/p2B4/PkP5/4p1pK/4Pb1p/5P2/8/8 w - - 29 68"},
		{"3r4/ppq1ppkp/4bnp1/2pN4/2P1P3/1P4P1/PQ3PBP/R4K2 b - - 2 20"},
		{"5rr1/4n2k/4q2P/P1P2n2/3B1p2/4pP2/2N1P3/1RR1K2Q w - - 1 49"},
		{"1r5k/2pq2p1/3p3p/p1pP4/4QP2/PP1R3P/6PK/8 w - - 1 51"},
		{"q5k1/5ppp/1r3bn1/1B6/P1N2P2/BQ2P1P1/5K1P/8 b - - 2 34"},
		{"r1b2k1r/5n2/p4q2/1ppn1Pp1/3pp1p1/NP2P3/P1PPBK2/1RQN2R1 w - - 0 22"},
		{"r1bqk2r/pppp1ppp/5n2/4b3/4P3/P1N5/1PP2PPP/R1BQKB1R w KQkq - 0 5"},
		{"r1bqr1k1/pp1p1ppp/2p5/8/3N1Q2/P2BB3/1PP2PPP/R3K2n b Q - 1 12"},
		{"r1bq2k1/p4r1p/1pp2pp1/3p4/1P1B3Q/P2B1N2/2P3PP/4R1K1 b - - 2 19"},
		{"r4qk1/6r1/1p4p1/2ppBbN1/1p5Q/P7/2P3PP/5RK1 w - - 2 25"},
		{"r7/6k1/1p6/2pp1p2/7Q/8/p1P2K1P/8 w - - 0 32"},
		{"r3k2r/ppp1pp1p/2nqb1pn/3p4/4P3/2PP4/PP1NBPPP/R2QK1NR w KQkq - 1 5"},
		{"3r1rk1/1pp1pn1p/p1n1q1p1/3p4/Q3P3/2P5/PP1NBPPP/4RRK1 w - - 0 12"},
		{"5rk1/1pp1pn1p/p3Brp1/8/1n6/5N2/PP3PPP/2R2RK1 w - - 2 20"},
		{"8/1p2pk1p/p1p1r1p1/3n4/8/5R2/PP3PPP/4R1K1 b - - 3 27"},
		{"8/4pk2/1p1r2p1/p1p4p/Pn5P/3R4/1P3PP1/4RK2 w - - 1 33"},
		{"8/5k2/1pnrp1p1/p1p4p/P6P/4R1PK/1P3P2/4R3 b - - 1 38"},
		{"8/8/1p1kp1p1/p1pr1n1p/P6P/1R4P1/1P3PK1/1R6 b - - 15 45"},
		{"8/8/1p1k2p1/p1prp2p/P2n3P/6P1/1P1R1PK1/4R3 b - - 5 49"},
		{"8/8/1p4p1/p1p2k1p/P2npP1P/4K1P1/1P6/3R4 w - - 6 54"},
		{"8/8/1p4p1/p1p2k1p/P2n1P1P/4K1P1/1P6/6R1 b - - 6 59"},
		{"8/5k2/1p4p1/p1pK3p/P2n1P1P/6P1/1P6/4R3 b - - 14 63"},
		{"8/1R6/1p1K1kp1/p6p/P1p2P1P/6P1/1Pn5/8 w - - 0 67"},
		{"1rb1rn1k/p3q1bp/2p3p1/2p1p3/2P1P2N/PP1RQNP1/1B3P2/4R1K1 b - - 4 23"},
		{"4rrk1/pp1n1pp1/q5p1/P1pP4/2n3P1/7P/1P3PB1/R1BQ1RK1 w - - 3 22"},
		{"r2qr1k1/pb1nbppp/1pn1p3/2ppP3/3P4/2PB1NN1/PP3PPP/R1BQR1K1 w - - 4 12"},
		{"2r2k2/8/4P1R1/1p6/8/P4K1N/7b/2B5 b - - 0 55"},
		{"6k1/5pp1/8/2bKP2P/2P5/p4PNb/B7/8 b - - 1 44"},
		{"2rqr1k1/1p3p1p/p2p2p1/P1nPb3/2B1P3/5P2/1PQ2NPP/R1R4K w - - 3 25"},
		{"r1b2rk1/p1q1ppbp/6p1/2Q5/8/4BP2/PPP3PP/2KR1B1R b - - 2 14"},
		{"6r1/5k2/p1b1r2p/1pB1p1p1/1Pp3PP/2P1R1K1/2P2P2/3R4 w - - 1 36"},
		{"rnbqkb1r/pppppppp/5n2/8/2PP4/8/PP2PPPP/RNBQKBNR b KQkq c3 0 2"},
		{"2rr2k1/1p4bp/p1q1p1p1/4Pp1n/2PB4/1PN3P1/P3Q2P/2RR2K1 w - f6 0 20"},
		{"3br1k1/p1pn3p/1p3n2/5pNq/2P1p3/1PN3PP/P2Q1PB1/4R1K1 w - - 0 23"},
		{"2r2b2/5p2/5k2/p1r1pP2/P2pB3/1P3P2/K1P3R1/7R w - - 23 93"},
	}

	rng := rand.New(rand.NewPCG(832473287, 23292478578))

	for tix, tt := range tests {
		ms := move.NewStore()
		hStack := stack.New[heur.StackMove]()

		t.Run(fmt.Sprintf("picker test %d", tix), func(t *testing.T) {
			b := Must(board.FromFEN(tt.fen))

			ranker := heur.NewMoveRanker()
			hStack.Reset()
			ms.Clear()
			ms.Push()
			movegen.Noisy(ms, b)
			movegen.Quiet(ms, b)
			allMoves := slices.Clone(ms.Frame())
			numMoves := len(allMoves)

			rand.Shuffle(len(allMoves), func(i, j int) {
				allMoves[i], allMoves[j] = allMoves[j], allMoves[i]
			})

			failHighIx := rng.IntN(numMoves)
			ranker.FailHigh(3, b, allMoves[:failHighIx], hStack)

			hashMoveIx := rng.IntN(numMoves)
			hashMove := ms.Frame()[hashMoveIx].Move

			ms.Clear()

			pck := picker.NewAllMoves(b, ms, &ranker, hashMove, hStack)

			state := verifyHash

			yielded := make([]move.Weighted, 0)

			for pck.Next() {
				m := pck.Move().Move

				assert.NotContains(t, yielded, *pck.Move(), "fen %s hashMove %s double yield %s", tt.fen, hashMove, m)
				yielded = append(yielded, *pck.Move())

				switch state {

				case verifyHash:
					state = verifyGoodNoisies
					assert.Equal(t, hashMove, m, "fen %s, hashmove %s yielded %s", tt.fen, hashMove, pck.Move())

				case verifyGoodNoisies:
					if (m.Promo() != NoPiece || b.SquaresToPiece[b.CaptureSq(m)] != NoPiece) && heur.SEE(b, m, 0) {
						assert.GreaterOrEqual(t, pck.Move().Weight, heur.Captures, "fen %s capture weight too low %s", tt.fen, m)
						continue
					}

					state = verifyQuiets
					fallthrough

				case verifyQuiets:
					if m.Promo() == NoPiece && b.SquaresToPiece[b.CaptureSq(m)] == NoPiece {
						assert.Greater(t, pck.Move().Weight, -heur.Captures, "fen %s quiet weight too low %s", tt.fen, m)
						assert.Less(t, pck.Move().Weight, heur.Captures, "fen %s quiet weight too high %s", tt.fen, m)
						continue
					}
					state = verifyBadNoisies
					fallthrough

				case verifyBadNoisies:
					assert.Less(t, pck.Move().Weight, -heur.Captures, "fen %s capture weight too high %s", tt.fen, m)
				}
			}

			ym := pck.YieldedMoves()
			assert.Equal(t, ym, yielded, "fen %s", tt.fen)

			assertNonIncreasing(t, ym, "weights increasing - fen %s", tt.fen)
			assertMovesMatch(t, yielded, allMoves, "fen %s", tt.fen)
		})
	}
}

func TestNoisyOrEvasions(t *testing.T) {
	tests := []struct {
		fen string
	}{
		{"r3k2r/2pb1ppp/2pp1q2/p7/1nP1B3/1P2P3/P2N1PPP/R2QK2R w KQkq a6 0 14"},
		{"4rrk1/2p1b1p1/p1p3q1/4p3/2P2n1p/1P1NR2P/PB3PP1/3R1QK1 b - - 2 24"},
		{"r3qbrk/6p1/2b2pPp/p3pP1Q/PpPpP2P/3P1B2/2PB3K/R5R1 w - - 16 42"},
		{"6k1/1R3p2/6p1/2Bp3p/3P2q1/P7/1P2rQ1K/5R2 b - - 4 44"},
		{"8/8/1p2k1p1/3p3p/1p1P1P1P/1P2PK2/8/8 w - - 3 54"},
		{"7r/2p3k1/1p1p1qp1/1P1Bp3/p1P2r1P/P7/4R3/Q4RK1 w - - 0 36"},
		{"r1bq1rk1/pp2b1pp/n1pp1n2/3P1p2/2P1p3/2N1P2N/PP2BPPP/R1BQ1RK1 b - - 2 10"},
		{"3r3k/2r4p/1p1b3q/p4P2/P2Pp3/1B2P3/3BQ1RP/6K1 w - - 3 87"},
		{"2r4r/1p4k1/1Pnp4/3Qb1pq/8/4BpPp/5P2/2RR1BK1 w - - 0 42"},
		{"4q1bk/6b1/7p/p1p4p/PNPpP2P/KN4P1/3Q4/4R3 b - - 0 37"},
		{"2q3r1/1r2pk2/pp3pp1/2pP3p/P1Pb1BbP/1P4Q1/R3NPP1/4R1K1 w - - 2 34"},
		{"1r2r2k/1b4q1/pp5p/2pPp1p1/P3Pn2/1P1B1Q1P/2R3P1/4BR1K b - - 1 37"},
		{"r3kbbr/pp1n1p1P/3ppnp1/q5N1/1P1pP3/P1N1B3/2P1QP2/R3KB1R b KQkq b3 0 17"},
		{"8/6pk/2b1Rp2/3r4/1R1B2PP/P5K1/8/2r5 b - - 16 42"},
		{"1r4k1/4ppb1/2n1b1qp/pB4p1/1n1BP1P1/7P/2PNQPK1/3RN3 w - - 8 29"},
		{"8/p2B4/PkP5/4p1pK/4Pb1p/5P2/8/8 w - - 29 68"},
		{"3r4/ppq1ppkp/4bnp1/2pN4/2P1P3/1P4P1/PQ3PBP/R4K2 b - - 2 20"},
		{"5rr1/4n2k/4q2P/P1P2n2/3B1p2/4pP2/2N1P3/1RR1K2Q w - - 1 49"},
		{"1r5k/2pq2p1/3p3p/p1pP4/4QP2/PP1R3P/6PK/8 w - - 1 51"},
		{"q5k1/5ppp/1r3bn1/1B6/P1N2P2/BQ2P1P1/5K1P/8 b - - 2 34"},
		{"r1b2k1r/5n2/p4q2/1ppn1Pp1/3pp1p1/NP2P3/P1PPBK2/1RQN2R1 w - - 0 22"},
		{"r1bqk2r/pppp1ppp/5n2/4b3/4P3/P1N5/1PP2PPP/R1BQKB1R w KQkq - 0 5"},
		{"r1bqr1k1/pp1p1ppp/2p5/8/3N1Q2/P2BB3/1PP2PPP/R3K2n b Q - 1 12"},
		{"r1bq2k1/p4r1p/1pp2pp1/3p4/1P1B3Q/P2B1N2/2P3PP/4R1K1 b - - 2 19"},
		{"r4qk1/6r1/1p4p1/2ppBbN1/1p5Q/P7/2P3PP/5RK1 w - - 2 25"},
		{"r7/6k1/1p6/2pp1p2/7Q/8/p1P2K1P/8 w - - 0 32"},
		{"r3k2r/ppp1pp1p/2nqb1pn/3p4/4P3/2PP4/PP1NBPPP/R2QK1NR w KQkq - 1 5"},
		{"3r1rk1/1pp1pn1p/p1n1q1p1/3p4/Q3P3/2P5/PP1NBPPP/4RRK1 w - - 0 12"},
		{"5rk1/1pp1pn1p/p3Brp1/8/1n6/5N2/PP3PPP/2R2RK1 w - - 2 20"},
		{"8/1p2pk1p/p1p1r1p1/3n4/8/5R2/PP3PPP/4R1K1 b - - 3 27"},
		{"8/4pk2/1p1r2p1/p1p4p/Pn5P/3R4/1P3PP1/4RK2 w - - 1 33"},
		{"8/5k2/1pnrp1p1/p1p4p/P6P/4R1PK/1P3P2/4R3 b - - 1 38"},
		{"8/8/1p1kp1p1/p1pr1n1p/P6P/1R4P1/1P3PK1/1R6 b - - 15 45"},
		{"8/8/1p1k2p1/p1prp2p/P2n3P/6P1/1P1R1PK1/4R3 b - - 5 49"},
		{"8/8/1p4p1/p1p2k1p/P2npP1P/4K1P1/1P6/3R4 w - - 6 54"},
		{"8/8/1p4p1/p1p2k1p/P2n1P1P/4K1P1/1P6/6R1 b - - 6 59"},
		{"8/5k2/1p4p1/p1pK3p/P2n1P1P/6P1/1P6/4R3 b - - 14 63"},
		{"8/1R6/1p1K1kp1/p6p/P1p2P1P/6P1/1Pn5/8 w - - 0 67"},
		{"1rb1rn1k/p3q1bp/2p3p1/2p1p3/2P1P2N/PP1RQNP1/1B3P2/4R1K1 b - - 4 23"},
		{"4rrk1/pp1n1pp1/q5p1/P1pP4/2n3P1/7P/1P3PB1/R1BQ1RK1 w - - 3 22"},
		{"r2qr1k1/pb1nbppp/1pn1p3/2ppP3/3P4/2PB1NN1/PP3PPP/R1BQR1K1 w - - 4 12"},
		{"2r2k2/8/4P1R1/1p6/8/P4K1N/7b/2B5 b - - 0 55"},
		{"6k1/5pp1/8/2bKP2P/2P5/p4PNb/B7/8 b - - 1 44"},
		{"2rqr1k1/1p3p1p/p2p2p1/P1nPb3/2B1P3/5P2/1PQ2NPP/R1R4K w - - 3 25"},
		{"r1b2rk1/p1q1ppbp/6p1/2Q5/8/4BP2/PPP3PP/2KR1B1R b - - 2 14"},
		{"6r1/5k2/p1b1r2p/1pB1p1p1/1Pp3PP/2P1R1K1/2P2P2/3R4 w - - 1 36"},
		{"rnbqkb1r/pppppppp/5n2/8/2PP4/8/PP2PPPP/RNBQKBNR b KQkq c3 0 2"},
		{"2rr2k1/1p4bp/p1q1p1p1/4Pp1n/2PB4/1PN3P1/P3Q2P/2RR2K1 w - f6 0 20"},
		{"3br1k1/p1pn3p/1p3n2/5pNq/2P1p3/1PN3PP/P2Q1PB1/4R1K1 w - - 0 23"},
		{"2r2b2/5p2/5k2/p1r1pP2/P2pB3/1P3P2/K1P3R1/7R w - - 23 93"},
		{"r3k2r/pp5p/3b1p2/1B3P2/3p2n1/8/PPP3PP/1RB1K2R b Kkq - 1 19"},
		{"3r3r/pp2k2p/5p2/1B3P2/3p1bn1/1PP5/P2K2P1/1RB4R w - - 1 23"},
		{"r2qk2r/3bbppp/p1nN4/1p1Q4/3P4/5N1P/PP3PP1/R3KB1R b KQkq - 3 16"},
		{"2rQk2r/5ppp/p1n1b3/1p6/3P4/3B1N1P/bP3PP1/2KRR3 b k - 1 21"},
		{"3rk2r/5ppp/p1n5/1p6/3P4/3B1N1P/bP3PP1/2KRR3 b k - 1 21"},
		{"3r1k1r/5pp1/p7/1p5p/1n1PB2P/4RN2/bP3PP1/1K1R4 w - - 1 26"},
		{"r2qkb1r/1p2pppp/p2p1n2/2p2b1P/3P4/2NnPN2/PPP2PPR/R1BQK3 w Qkq - 0 9"},
		{"2rqkb1r/5pp1/3pb2p/8/Qp1BP3/2N5/PP3PPN/1K1R4 b k - 1 22"},
		{"r4rk1/pp4p1/2ppb2p/q4n2/N1P1BB1b/3QP2P/PP3P2/3RK1R1 w - - 4 18"},
		{"3r1rk1/pp4p1/4bb2/3p3q/5B2/1PNQPn1P/P4P2/2R1K1RB w - - 2 27"},
		{"r1b1qrk1/ppp2p1p/3p2p1/2b5/1nP1P1n1/1Q2PN2/PPK2PPP/R1BN1B1R w - - 3 14"},
		{"r3qrk1/ppp2p1p/3p2p1/8/b1P1P1P1/1Pb1PN2/P1nK1PP1/R1BN1B1R w - - 0 19"},
		{"rnb1k1nr/p1pp1ppp/3b4/1P2p3/4P2q/5P1P/PP1P2P1/RNBQKBNR w KQkq - 1 6"},
		{"r1b2rk1/3p1ppq/5n1p/1P1pp2P/3nP1P1/3PBP2/P3K3/R3QBNR w - - 1 18"},
		{"r1b2rk1/5pp1/5n1p/1P1pP2P/3p2P1/5P2/P1q2K2/R3QBNR w - - 1 22"},
		{"2b3k1/2q2pp1/7p/3pr2P/3Q1NP1/3n1P2/5K2/7R w - - 0 33"},
		{"6k1/qb3pp1/4r2p/7P/1Q4P1/3N1P1R/5K2/8 w - - 7 38"},
		{"6k1/1Q3pp1/2r4p/2q4P/6P1/5P1R/5K2/8 w - - 0 40"},
		{"6k1/1Q3pp1/3q3p/7P/6P1/5P2/2r3K1/7R w - - 4 42"},
		{"Q5k1/5pp1/4q2p/7P/6P1/5P1K/2r5/7R b - - 7 43"},
		{"6k1/5p2/2q1rnp1/PN2b3/2P5/1p1B2PP/3P1PK1/3QR3 w - - 3 34"},
		{"1b6/5k2/2q1pnp1/PN6/2PQ1P2/3B2PP/1p1P2K1/8 w - - 6 41"},
		{"1N6/Q4k2/4p1p1/8/1qP2P2/3n2PP/3P3K/8 b - - 2 48"},
		{"8/6k1/4p1p1/4N3/2P2P2/6PP/5K2/3n4 w - - 5 57"},
		{"2r3Rk/3n1p2/1p1pb2B/pP2b3/P1PpP2P/q2P4/3Q3N/1KRB4 b - - 0 32"},
		{"6rk/5p2/1p1pb2B/pPn1b3/P1PpP2P/3q4/R2Q3N/1K1B4 w - - 0 35"},
		{"6r1/8/1p4k1/pP2ppR1/P1bp3P/8/8/1K1B4 b - - 0 44"},
		{"4r1k1/5p2/2Q5/R7/3bq3/6P1/PP4P1/1K1N4 w - - 0 28"},
		{"6k1/5p2/8/R7/8/2K1r1P1/PP4P1/8 w - - 1 32"},
		{"6k1/5p2/8/R7/1P6/8/P2K2r1/8 w - - 0 34"},
	}

	for _, tt := range tests {
		t.Run(tt.fen, func(t *testing.T) {
			b := Must(board.FromFEN(tt.fen))
			checkers := b.Checkers()
			ms := move.NewStore()
			ranker := heur.NewMoveRanker()

			ms.Clear()
			ms.Push()
			if checkers == 0 {
				movegen.Noisy(ms, b)
			} else {
				movegen.NoisyEvasions(ms, b, checkers)
				movegen.QuietEvasions(ms, b, checkers)
			}
			expected := slices.Clone(ms.Frame())
			ms.Pop()

			ms.Clear()
			pck := picker.NewNoisyOrEvasions(b, ms, &ranker, checkers)
			var yielded []move.Weighted
			for pck.Next() {
				yielded = append(yielded, *pck.Move())
			}

			assertMovesMatch(t, expected, yielded, "fen %s", tt.fen)
			assertNonIncreasing(t, yielded, "weights increasing - fen %s", tt.fen)
		})
	}
}

func assertNonIncreasing(t *testing.T, moves []move.Weighted, msgAndArgs ...any) {
	for i := 1; i < len(moves); i++ {
		assert.LessOrEqual(t, moves[i].Weight, moves[i-1].Weight, msgAndArgs...)
	}
}

func assertMovesMatch(t *testing.T, a, b []move.Weighted, msgAndArgs ...any) {
	assert.ElementsMatch(t, stripWeights(a), stripWeights(b), msgAndArgs...)
}

func stripWeights(l []move.Weighted) []move.Move {
	mod := make([]move.Move, 0, len(l))
	for _, m := range l {
		mod = append(mod, m.Move)
	}
	return mod
}

type state byte

const (
	verifyHash = state(iota)
	verifyGoodNoisies
	verifyQuiets
	verifyBadNoisies
)
