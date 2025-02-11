// package eval gives position evaluation measuerd in centipawns.
package eval

import (
	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/movegen"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

// Some of this code is derived from
// https://www.chessprogramming.org/PeSTO%27s_Evaluation_Function
// which is used in many simple engines. This gives a base evaluation with
// piece values and PSQT with tapered evaluation support between middle game
// and end game.
//
// Other aspects of the evaluation are additions on the PESTO values.

// ScoreType defines the evaluation result type. The engine uses int16 for
// score type, as defined in types. The tuner uses float64.
type ScoreType interface{ Score | float64 }

type CoeffSet[T ScoreType] struct {

	// PSqT is tapered piece square tables. (PESTO)
	PSqT [12][64]T

	// TPieceValues is tapered piece values between middle game and end game.
	// (PESTO)
	TPieceValues [2][7]T

	// TempoBonus is the advantage of the side to move.
	TempoBonus [2]T

	// KingAttackPieces is the bonus per piece type if piece is attacking a square in the enemy king's neighborhood.
	KingAttackPieces [2][4]T
	// KingAttackSquares is the bonus per attacked squares count in the enemy neighborhood.
	KingAttackCount [2][7]T

	// Mobility* is per piece mobility bonus.
	MobilityKnight [2][9]T
	MobilityBishop [2][14]T
	MobilityRook   [2][11]T

	// ProtectedPasser is the bonus for each protected passed pawn.
	ProtectedPasser [2]T
	// PasserKingDist is the bonus for our king being close / enemy king being far from passed pawn.
	PasserKingDist [2]T
	// PasserRank is the bonus for the passed pawn being on a specific rank.
	PasserRank [2][6]T

	// LazyMargin determines the early return margins at various points in the
	// evaluation. It's not tunable. Every time a new term is added to evaluation
	// one has to recompute the lazy margins. Modify the evaluation so that
	// instead of returning early, it records the partial score for the given
	// return point, then once the complete score is calculated, take the
	// difference between the final score and the partial score and across many
	// test positions store the maximal difference. This plus some safety margin
	// would be the lazy margin.
	LazyMargin [7]T
}

var Coefficients = CoeffSet[Score]{
	PSqT: [12][64]Score{
		// pawn middle game
		{
			0, 0, 0, 0, 0, 0, 0, 0,
			98, 134, 61, 95, 68, 126, 34, -11,
			-6, 7, 26, 31, 65, 56, 25, -20,
			-14, 13, 6, 21, 23, 12, 17, -23,
			-27, -2, -5, 12, 17, 6, 10, -25,
			-26, -4, -4, -10, 3, 3, 33, -12,
			-35, -1, -20, -23, -15, 24, 38, -22,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		// pawn end game
		{
			0, 0, 0, 0, 0, 0, 0, 0,
			178, 173, 158, 134, 147, 132, 165, 187,
			94, 100, 85, 67, 56, 53, 82, 84,
			32, 24, 13, 5, -2, 4, 17, 17,
			13, 9, -3, -7, -7, -8, 3, -1,
			4, 7, -6, 1, 0, -5, -1, -8,
			13, 8, 8, 10, 13, 0, 2, -7,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		// knight middle game
		{
			-167, -89, -34, -49, 61, -97, -15, -107,
			-73, -41, 72, 36, 23, 62, 7, -17,
			-47, 60, 37, 65, 84, 129, 73, 44,
			-9, 17, 19, 53, 37, 69, 18, 22,
			-13, 4, 16, 13, 28, 19, 21, -8,
			-23, -9, 12, 10, 19, 17, 25, -16,
			-29, -53, -12, -3, -1, 18, -14, -19,
			-105, -21, -58, -33, -17, -28, -19, -23,
		},
		// knight end game
		{
			-58, -38, -13, -28, -31, -27, -63, -99,
			-25, -8, -25, -2, -9, -25, -24, -52,
			-24, -20, 10, 9, -1, -9, -19, -41,
			-17, 3, 22, 22, 22, 11, 8, -18,
			-18, -6, 16, 25, 16, 17, 4, -18,
			-23, -3, -1, 15, 10, -3, -20, -22,
			-42, -20, -10, -5, -2, -20, -23, -44,
			-29, -51, -23, -15, -22, -18, -50, -64,
		},
		// bishop middle game
		{
			-29, 4, -82, -37, -25, -42, 7, -8,
			-26, 16, -18, -13, 30, 59, 18, -47,
			-16, 37, 43, 40, 35, 50, 37, -2,
			-4, 5, 19, 50, 37, 37, 7, -2,
			-6, 13, 13, 26, 34, 12, 10, 4,
			0, 15, 15, 15, 14, 27, 18, 10,
			4, 15, 16, 0, 7, 21, 33, 1,
			-33, -3, -14, -21, -13, -12, -39, -21,
		},
		// bishop end game
		{
			-14, -21, -11, -8, -7, -9, -17, -24,
			-8, -4, 7, -12, -3, -13, -4, -14,
			2, -8, 0, -1, -2, 6, 0, 4,
			-3, 9, 12, 9, 14, 10, 3, 2,
			-6, 3, 13, 19, 7, 10, -3, -9,
			-12, -3, 8, 10, 13, 3, -7, -15,
			-14, -18, -7, -1, 4, -9, -15, -27,
			-23, -9, -23, -5, -9, -16, -5, -17,
		},
		// rook middle game
		{
			32, 42, 32, 51, 63, 9, 31, 43,
			27, 32, 58, 62, 80, 67, 26, 44,
			-5, 19, 26, 36, 17, 45, 61, 16,
			-24, -11, 7, 26, 24, 35, -8, -20,
			-36, -26, -12, -1, 9, -7, 6, -23,
			-45, -25, -16, -17, 3, 0, -5, -33,
			-44, -16, -20, -9, -1, 11, -6, -71,
			-19, -13, 1, 17, 16, 7, -37, -26,
		},
		// rook end game
		{
			13, 10, 18, 15, 12, 12, 8, 5,
			11, 13, 13, 11, -3, 3, 8, 3,
			7, 7, 7, 5, 4, -3, -5, -3,
			4, 3, 13, 1, 2, 1, -1, 2,
			3, 5, 8, 4, -5, -6, -8, -11,
			-4, 0, -5, -1, -7, -12, -8, -16,
			-6, -6, 0, 2, -9, -9, -11, -3,
			-9, 2, 3, -1, -5, -13, 4, -20,
		},
		// queen middle game
		{
			-28, 0, 29, 12, 59, 44, 43, 45,
			-24, -39, -5, 1, -16, 57, 28, 54,
			-13, -17, 7, 8, 29, 56, 47, 57,
			-27, -27, -16, -16, -1, 17, -2, 1,
			-9, -26, -9, -10, -2, -4, 3, -3,
			-14, 2, -11, -2, -5, 2, 14, 5,
			-35, -8, 11, 2, 8, 15, -3, 1,
			-1, -18, -9, 10, -15, -25, -31, -50,
		},
		// gueen end game
		{
			-9, 22, 22, 27, 27, 19, 10, 20,
			-17, 20, 32, 41, 58, 25, 30, 0,
			-20, 6, 9, 49, 47, 35, 19, 9,
			3, 22, 24, 45, 57, 40, 57, 36,
			-18, 28, 19, 47, 31, 34, 39, 23,
			-16, -27, 15, 6, 9, 17, 10, 5,
			-22, -23, -30, -16, -16, -23, -36, -32,
			-33, -28, -22, -43, -5, -32, -20, -41,
		},
		// king middle game
		{
			-65, 23, 16, -15, -56, -34, 2, 13,
			29, -1, -20, -7, -8, -4, -38, -29,
			-9, 24, 2, -16, -20, 6, 22, -22,
			-17, -20, -12, -27, -30, -25, -14, -36,
			-49, -1, -27, -39, -46, -44, -33, -51,
			-14, -14, -22, -46, -44, -30, -15, -27,
			1, 7, -8, -64, -43, -16, 9, 8,
			-15, 36, 12, -54, 8, -28, 24, 14,
		},
		// king end game
		{
			-74, -35, -18, -18, -11, 15, 4, -17,
			-12, 17, 14, 17, 17, 38, 23, 11,
			10, 17, 23, 15, 20, 45, 44, 13,
			-8, 22, 24, 27, 26, 33, 26, 3,
			-18, -4, 21, 24, 27, 23, 9, -11,
			-19, -3, 11, 21, 23, 16, 7, -9,
			-27, -11, 4, 13, 14, 4, -5, -17,
			-53, -34, -21, -11, -28, -14, -24, -43,
		},
	},
	TPieceValues: [2][7]Score{
		{0, 82, 337, 365, 477, 1025, Inf},
		{0, 94, 281, 297, 512, 936, Inf},
	},
	TempoBonus: [2]Score{29, 27},
	KingAttackPieces: [2][4]Score{
		{3, 2, 17, 15},
		{-1, -1, -5, 24},
	},
	KingAttackCount: [2][7]Score{
		{0, 5, 10, 20, 26, 68, 45},
		{0, 8, 15, 21, 28, -11, 15},
	},
	MobilityKnight: [2][9]Score{
		{23, 43, 48, 50, 53, 56, 58, 60, 72},
		{-40, -22, 2, 4, -1, 1, 1, 0, -16},
	},
	MobilityBishop: [2][14]Score{
		{49, 50, 55, 55, 60, 64, 62, 60, 62, 59, 62, 66, 80, 64},
		{-46, -24, -35, -22, -16, -5, -2, 8, 7, 9, 3, 1, 1, -3},
	},
	MobilityRook: [2][11]Score{
		{50, 48, 49, 51, 53, 61, 58, 61, 63, 72, 86},
		{-3, 1, 6, 4, 6, 8, 15, 18, 25, 26, 16},
	},
	ProtectedPasser: [2]Score{36, 13},
	PasserKingDist:  [2]Score{-4, 12},
	PasserRank: [2][6]Score{
		{22, -4, -11, -2, -7, -35},
		{-29, -15, 11, 32, 46, 35},
	},
	LazyMargin: [...]Score{918, 704, 711, 711, 770, 842, 880},
}

// Phase is game phase.
var Phase = [...]int{0, 0, 1, 1, 2, 4, 0}

func Eval[T ScoreType](b *board.Board, _, beta T, c *CoeffSet[T]) T {
	if insuffientMat(b) {
		return 0
	}

	mg := [2]T{}
	eg := [2]T{}

	mg[b.STM] += c.TempoBonus[0]
	eg[b.STM] += c.TempoBonus[1]

	phase := 0

	for pType := Pawn; pType <= Queen; pType++ {
		for color := White; color <= Black; color++ {
			cnt := (b.Pieces[pType] & b.Colors[color]).Count()

			phase += cnt * Phase[pType]

			mg[color] += T(cnt) * c.TPieceValues[0][pType]
			eg[color] += T(cnt) * c.TPieceValues[1][pType]
		}
	}

  // see comment on LazyMargin
	// scoreHist := [7]T{}

	score := TaperedScore(b, phase, mg[:], eg[:])
	// scoreHist[0] = score
	if score > beta+c.LazyMargin[0] {
		return beta
	}

	pWise := newPieceWise(b, c)

	// This loop is going down in piece value for lazy return.
	for pType := King; pType >= Pawn; pType-- {
		for color := White; color <= Black; color++ {

			pieces := b.Pieces[pType] & b.Colors[color]
			for piece := board.BitBoard(0); pieces != 0; pieces ^= piece {
				piece = pieces & -pieces
				sq := piece.LowestSet()
				sqIx := sq

				if color == White {
					sqIx ^= 56 // upside down
				}

				ix := pType - 1

				mg[color] += c.PSqT[2*ix][sqIx]
				eg[color] += c.PSqT[2*ix+1][sqIx]

				pWise.Eval(pType, color, sq, mg[:], eg[:])
			}
		}

    score = TaperedScore(b, phase, mg[:], eg[:])
		// scoreHist[pType] = score

		if score > beta+c.LazyMargin[pType] {
			return beta
		}

		if pType == Knight {
			pWise.Passers()
		}
	}

	for color := White; color <= Black; color++ {
		kingACnt := min(len(c.KingAttackCount[0])-1, pWise.kingACount[color])
		mg[color] += pWise.kingAScore[0][color] * c.KingAttackCount[0][kingACnt] / 16
		eg[color] += pWise.kingAScore[1][color] * c.KingAttackCount[1][kingACnt] / 16
	}

	score = TaperedScore(b, phase, mg[:], eg[:])

	// for i, v := range scoreHist {
	// 	c.LazyMargin[i] = max(c.LazyMargin[i], score-v)
	// }
	//

	return score
}

func TaperedScore[T ScoreType](b *board.Board, phase int, mg, eg []T) T {
	mgScore := mg[b.STM] - mg[b.STM.Flip()]
	egScore := eg[b.STM] - eg[b.STM.Flip()]

	mgPhase := phase
	if mgPhase > 24 {
		mgPhase = 24 // in case of early promotion
	}
	egPhase := 24 - mgPhase

	if _, ok := (any(mgScore)).(Score); ok {
		return T((int(mgScore)*mgPhase + int(egScore)*egPhase) / 24)
	}

	return T((mgScore*T(mgPhase) + egScore*T(egPhase)) / 24)
}

type pieceWise[T ScoreType] struct {
	c          *CoeffSet[T]
	b          *board.Board
	occ        board.BitBoard
	passers    board.BitBoard
	kingNb     [2]board.BitBoard
	pawnCover  [2]board.BitBoard
	kingACount [2]int
	kingAScore [2][2]T
	kingSq     [2]Square
}

func newPieceWise[T ScoreType](b *board.Board, c *CoeffSet[T]) pieceWise[T] {
	result := pieceWise[T]{b: b, c: c}
	result.occ = b.Colors[White] | b.Colors[Black]

	for color := White; color <= Black; color++ {
		king := b.Colors[color] & b.Pieces[King]
		kingSq := king.LowestSet()
		kingA := movegen.KingMoves(kingSq)

		result.kingSq[color] = kingSq

		var kingNb board.BitBoard
		switch color {
		case White:
			kingNb = king | kingA | (kingA << 8)
		case Black:
			kingNb = king | kingA | (kingA >> 8)
		}

		result.kingNb[color] = kingNb
	}

	wP := b.Pieces[Pawn] & b.Colors[White]
	result.pawnCover[White] = ((wP & ^board.AFile) << 7) | ((wP & ^board.HFile) << 9)
	bP := b.Pieces[Pawn] & b.Colors[Black]
	result.pawnCover[Black] = ((bP & ^board.HFile) >> 7) | ((bP & ^board.AFile) >> 9)

	return result
}

func (p *pieceWise[T]) Passers() {
	b := p.b

	for color := White; color <= Black; color++ {
		myPawns := b.Pieces[Pawn] & b.Colors[color]
		theirPawns := b.Pieces[Pawn] & b.Colors[color.Flip()]

		myRearFill := frontFill(myPawns, color.Flip())
		theirFrontFill := frontFill(theirPawns, color.Flip())

		var (
			myRearSpan     board.BitBoard
			theirFrontSpan board.BitBoard
		)
		switch color {
		case White:
			myRearSpan = myRearFill >> 8
			theirFrontSpan = theirFrontFill >> 8

		case Black:
			myRearSpan = myRearFill << 8
			theirFrontSpan = theirFrontFill << 8
		}

		frontLine := ^myRearSpan & myPawns

		enemyCover := theirFrontSpan | ((theirFrontSpan & ^board.AFile) >> 1) | ((theirFrontSpan & ^board.HFile) << 1)

		p.passers |= frontLine & ^enemyCover
	}
}

func frontFill(b board.BitBoard, color Color) board.BitBoard {
	switch color {
	case White:
		b |= b << 8
		b |= b << 16
		b |= b << 32

	case Black:
		b |= b >> 8
		b |= b >> 16
		b |= b >> 32
	}

	return b
}

func (p *pieceWise[T]) Eval(pType Piece, color Color, sq Square, mg, eg []T) {

	occ := p.occ

	var attack board.BitBoard

	switch pType {

	case Queen:
		attack = movegen.BishopMoves(sq, occ) | movegen.RookMoves(sq, occ)

	case Rook:
		attack = movegen.RookMoves(sq, occ)

		rank := board.BitBoard(0xff) << (sq & 56)
		hmob := (attack & rank & ^p.b.Colors[color]).Count()
		vmob := (attack & ^rank & ^p.b.Colors[color]).Count()

		// count vertical mobility 2x compared to horizontal mobility
		mobCnt := (2*vmob + hmob) / 2

		mg[color] += p.c.MobilityRook[0][mobCnt]
		eg[color] += p.c.MobilityRook[1][mobCnt]

	case Bishop:
		attack = movegen.BishopMoves(sq, occ)

		mobCnt := (attack & ^p.b.Colors[color]).Count()
		mg[color] += p.c.MobilityBishop[0][mobCnt]
		eg[color] += p.c.MobilityBishop[1][mobCnt]

	case Knight:
		attack = movegen.KnightMoves(sq)

		mobCnt := (attack & ^p.b.Colors[color] & ^p.pawnCover[color.Flip()]).Count()
		mg[color] += p.c.MobilityKnight[0][mobCnt]
		eg[color] += p.c.MobilityKnight[1][mobCnt]

	case Pawn:
		pawn := board.BitBoard(1) << sq
		if p.passers&pawn != 0 {
			// if protected passers add protection bonus
			var support board.BitBoard
			rank := sq / 8

			switch color {
			case White:
				support = ((pawn & ^board.HFile) >> 7) | ((pawn & ^board.AFile) >> 9)
			case Black:
				support = ((pawn & ^board.AFile) << 7) | ((pawn & ^board.HFile) << 9)
				rank ^= 7
			}

			if support&p.b.Pieces[Pawn]&p.b.Colors[color] != 0 {
				mg[color] += p.c.ProtectedPasser[0]
				eg[color] += p.c.ProtectedPasser[1]
			}

			mg[color] += p.c.PasserRank[0][rank-1]
			eg[color] += p.c.PasserRank[1][rank-1]

			kingDist := Manhattan(sq, p.kingSq[color.Flip()]) - Manhattan(sq, p.kingSq[color])

			mg[color] += p.c.PasserKingDist[0] * T(kingDist)
			eg[color] += p.c.PasserKingDist[1] * T(kingDist)

		}
		return

	default:
		return
	}

	kingA := (p.kingNb[color.Flip()] & attack).Count()

	if kingA != 0 {
		p.kingAScore[0][color] += p.c.KingAttackPieces[0][pType-Knight] * T(kingA)
		p.kingAScore[1][color] += p.c.KingAttackPieces[1][pType-Knight] * T(kingA)
		p.kingACount[color]++
	}

}

func Manhattan(a, b Square) int {
	ax, ay, bx, by := int(a%8), int(a/8), int(b%8), int(b/8)
	return max(abs(ax-bx), abs(ay-by))
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func insuffientMat(b *board.Board) bool {
	if b.Pieces[Pawn]|b.Pieces[Queen]|b.Pieces[Rook] != 0 {
		return false
	}

	wN := (b.Colors[White] & b.Pieces[Knight]).Count()
	bN := (b.Colors[Black] & b.Pieces[Knight]).Count()
	wB := (b.Colors[White] & b.Pieces[Bishop]).Count()
	bB := (b.Colors[Black] & b.Pieces[Bishop]).Count()

	if wN+bN+wB+bB <= 3 { // draw cases
		wScr := wN + 3*wB
		bScr := bN + 3*bB

		if max(wScr-bScr, bScr-wScr) <= 3 {
			return true
		}
	}

	return false
}
