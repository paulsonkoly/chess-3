// package eval gives position evaluation measuerd in centipawns.
package eval

import (
	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/movegen"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

// ScoreType defines the evaluation result type. The engine uses int16 for
// score type, as defined in types. The tuner uses float64.
type ScoreType interface{ Score | float64 }

type CoeffSet[T ScoreType] struct {

	// PSqT is tapered piece square tables.
	PSqT [12][64]T

	// PieceValues is tapered piece values between middle game and end game.
	PieceValues [2][7]T

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

	// KnightOutpost is a per square bonus for a knight being on an outpost, only
	// counting the 5 ranks covering sideOfBoard.
	KnightOutpost [2][40]T

	// ConnectedRooks is a bonus if rooks are connected.
	ConnectedRooks [2]T

	// BishopPair is the bonus for bishop pairs.
	BishopPair [2]T

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
		{
			0, 0, 0, 0, 0, 0, 0, 0,
			26, 58, 29, 56, 87, 60, 0, -54,
			9, 38, 63, 59, 72, 114, 41, 19,
			4, 26, 19, 45, 57, 34, 45, 10,
			-10, 18, 15, 40, 40, 37, 41, 9,
			-8, 12, 15, 16, 35, 16, 69, 24,
			-8, 25, 2, 15, 23, 57, 85, 21,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			0, 0, 0, 0, 0, 0, 0, 0,
			110, 106, 86, 52, 43, 56, 102, 101,
			23, 10, -9, -39, -46, -41, -11, -2,
			6, -1, -17, -39, -40, -27, -14, -12,
			-8, -10, -27, -34, -31, -29, -20, -22,
			-12, -12, -24, -24, -24, -20, -28, -28,
			-7, -14, -12, -14, -13, -21, -29, -28,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			-242, -30, -40, -101, 2, -181, -92, -128,
			-49, -22, 95, 12, 88, 62, 22, -11,
			24, 34, 21, 52, 92, 107, 60, 19,
			17, 11, -6, 37, 3, 22, -9, 26,
			3, 1, 17, 12, 13, 28, 4, -5,
			-11, 3, 14, 18, 33, 14, 31, 1,
			-5, -21, 0, 22, 23, 24, 15, 16,
			-41, -14, -28, 1, 0, 1, -13, -43,
		},
		{
			29, -40, -13, 4, -24, -7, -16, -73,
			0, 7, -28, -3, -24, -24, -20, -17,
			-26, -19, 15, 9, -13, -8, -25, -14,
			-1, 19, 27, 26, 24, 20, 7, -1,
			-3, 9, 25, 21, 29, 12, 3, 11,
			-15, 0, -5, 18, 15, -2, -9, -11,
			-13, -3, -4, -6, -7, -5, -5, -8,
			-21, -14, 6, 6, -8, -6, -22, -1,
		},
		{
			-37, -43, -52, -89, -121, -97, -74, -72,
			-34, 13, -18, -27, 6, 50, 0, -1,
			-18, 4, 60, 19, 63, 30, 37, 16,
			-23, -6, 5, 15, 11, -10, -4, -23,
			-6, 4, -10, 23, 18, -6, -1, -32,
			13, 10, 17, -7, 7, 18, 9, 3,
			-4, 28, 5, 8, 9, 41, 40, 21,
			10, -8, 1, -7, 3, -13, -17, 1,
		},
		{
			1, 3, -1, 16, 9, 7, 1, 9,
			-4, -4, 5, -4, 1, -19, 3, -32,
			6, 2, -11, 4, -11, 8, -5, -7,
			9, 10, 6, 11, 13, 2, -1, -7,
			-3, -5, 8, 4, 4, 8, -6, 1,
			-10, 1, 0, 6, 10, -3, -4, -5,
			5, -19, -12, -4, -5, -12, -12, -19,
			-14, 1, -7, 4, 1, 3, -11, -12,
		},
		{
			34, 44, 36, 34, 62, 73, 22, 42,
			0, -19, 26, 50, 8, 34, 45, 17,
			-20, -8, 0, 9, 34, 30, 75, 19,
			-23, -17, 6, 17, 14, 5, 4, 28,
			-32, -24, -9, -9, 5, -14, -9, -8,
			-27, -18, -5, 10, 9, -12, 18, 23,
			-25, -20, -9, 8, 1, 5, 18, -41,
			-8, -3, 10, 18, 27, 13, -12, 2,
		},
		{
			11, 4, 5, 5, -2, -6, 9, 6,
			8, 23, 11, -3, 6, 8, 2, 1,
			14, 10, 9, 6, -4, 1, -6, -6,
			8, 8, 4, 1, -3, 5, 3, -8,
			9, 6, 5, 8, 2, 2, 0, -4,
			-1, 1, -2, -4, -4, -1, -24, -18,
			3, 3, 8, 1, -6, -1, -9, -2,
			5, 1, -1, -3, -14, -5, -4, -27,
		},
		{
			-51, 12, -3, -16, -4, -19, -32, -15,
			-28, -35, -36, -52, -75, 22, -39, 25,
			-16, -4, 15, -7, 42, 57, 84, 42,
			-20, -9, -9, -21, -12, -33, -18, -10,
			-3, -1, -1, -15, -2, -2, -1, -17,
			-12, 15, 4, 2, 8, 4, 6, -14,
			-13, 8, 16, 25, 27, 53, 41, 16,
			7, -3, 12, 21, 6, -18, -2, -9,
		},
		{
			21, 2, 14, 31, 10, 25, 4, 19,
			7, 24, 41, 76, 76, 51, 22, -40,
			-17, 6, 16, 28, 45, -11, -38, -48,
			1, 15, 28, 58, 48, 68, 31, -3,
			-8, 7, 12, 45, 31, 8, 13, -6,
			-23, -31, 5, 8, 10, 13, -4, -18,
			-9, -15, -25, -25, -21, -74, -76, -51,
			-38, -29, -38, -48, -11, -29, -62, -50,
		},
		{
			100, 91, 57, 65, 94, 75, 37, 63,
			79, 100, 77, 45, 47, 34, -37, 2,
			65, 132, 24, 45, -12, 34, 65, -24,
			13, 33, 36, -29, -38, -28, 16, -21,
			38, 29, -6, -68, -64, -75, -32, -72,
			-7, 24, -53, -70, -78, -48, -7, -30,
			9, -15, -31, -67, -52, -24, 24, 19,
			-35, 19, -1, -81, -17, -43, 34, 21,
		},
		{
			-66, -33, -34, -19, -27, -10, 0, -38,
			-35, -13, -10, -4, -1, 21, 34, 7,
			-22, -8, 8, 1, 15, 21, 27, 18,
			-28, -6, 5, 20, 22, 30, 20, 14,
			-33, -9, 12, 29, 32, 33, 19, 13,
			-23, -5, 19, 30, 35, 29, 11, 5,
			-23, 2, 16, 27, 27, 21, 3, -12,
			-11, -14, -4, 7, -12, 3, -24, -40,
		},
	},
	PieceValues: [2][7]Score{
		{0, 80, 468, 513, 649, 1377, 0},
		{0, 101, 247, 262, 483, 892, 0},
	},
	TempoBonus: [2]Score{37, 26},
	KingAttackPieces: [2][4]Score{
		{3, 2, 3, 3},
		{-2, -2, -2, 5},
	},
	KingAttackCount: [2][7]Score{
		{0, 3, 4, 6, 7, 11, 22},
		{0, 1, 2, 3, 3, 4, 7},
	},
	MobilityKnight: [2][9]Score{
		{-41, -10, 0, 5, 14, 20, 25, 29, 32},
		{-58, -32, -8, -2, 5, 11, 10, 7, -4},
	},
	MobilityBishop: [2][14]Score{
		{-19, -6, 3, 6, 11, 14, 19, 20, 23, 24, 31, 31, 27, 14},
		{-42, -29, -27, -15, -6, 2, 7, 14, 19, 19, 18, 19, 29, 24},
	},
	MobilityRook: [2][11]Score{
		{-15, -8, -3, 4, 7, 13, 9, 17, 24, 34, 49},
		{-20, -17, -11, -9, -2, 2, 8, 10, 13, 15, 8},
	},
	KnightOutpost: [2][40]Score{
		{
			-13, -26, -102, -18, 64, 7, 25, 5,
			-2, -5, -26, 75, 15, 30, -34, 0,
			-25, 5, 79, 23, -80, 18, 38, 80,
			24, 35, 54, 29, 34, 62, 106, 72,
			0, 0, 0, 42, 44, 0, 0, 0,
		},
		{
			-23, 48, -53, -82, 44, -15, 4, 39,
			-1, 6, 1, -38, -30, -10, 25, 74,
			20, 15, -2, 24, 55, 20, 11, 14,
			24, -20, 10, 26, 21, 6, -3, -13,
			0, 0, 0, 20, -1, 0, 0, 0,
		},
	},
	ConnectedRooks:  [2]Score{-1, 3},
	BishopPair:      [2]Score{-25, 46},
	ProtectedPasser: [2]Score{37, 11},
	PasserKingDist:  [2]Score{17, 3},
	PasserRank: [2][6]Score{
		{-6, -28, -32, -11, -10, 36},
		{2, 10, 29, 55, 120, 95},
	},
	LazyMargin: [...]Score{658, 445, 552, 565, 565, 573, 576},
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

	bishopCnt := [2]int{}

	for pType := Pawn; pType <= Queen; pType++ {
		for color := White; color <= Black; color++ {
			cnt := (b.Pieces[pType] & b.Colors[color]).Count()

			if pType == Bishop {
				bishopCnt[color] = cnt
			}

			phase += cnt * Phase[pType]

			mg[color] += T(cnt) * c.PieceValues[0][pType]
			eg[color] += T(cnt) * c.PieceValues[1][pType]
		}
	}

	for color := White; color <= Black; color++ {
		if bishopCnt[color] >= 2 && bishopCnt[color.Flip()] == 0 {
			mg[color] += c.BishopPair[0]
			eg[color] += c.BishopPair[1]
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
		mg[color] += pWise.kingAScore[0][color] * c.KingAttackCount[0][kingACnt]
		eg[color] += pWise.kingAScore[1][color] * c.KingAttackCount[1][kingACnt]
	}

	score = TaperedScore(b, phase, mg[:], eg[:])

	// for i, v := range scoreHist {
	// 	c.LazyMargin[i] = max(c.LazyMargin[i], score-v)
	// }

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
	frontSpan  [2]board.BitBoard
	holes      [2]board.BitBoard
	kingACount [2]int
	kingAScore [2][2]T
	kingSq     [2]Square
}

// the player's side of the board with the extra 2 central squares included at
// enemy side.
var sideOfBoard = [2]board.BitBoard{0x00000018_ffffffff, 0xffffffff_18000000}

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

	// various useful pawn bitboards
	wFrontSpan := frontFill(wP, White) << 8
	bFrontSpan := frontFill(bP, Black) >> 8

	result.frontSpan[White] = wFrontSpan
	result.frontSpan[Black] = bFrontSpan

	// calculate holes in our position, squares that cannot be protected by one
	// of our pawns.
	wCover := ((wFrontSpan & ^board.AFile) >> 1) | ((wFrontSpan & ^board.HFile) << 1)
	bCover := ((bFrontSpan & ^board.HFile) << 1) | ((bFrontSpan & ^board.AFile) >> 1)
	result.holes[White] = sideOfBoard[White] & ^wCover
	result.holes[Black] = sideOfBoard[Black] & ^bCover

	return result
}

func (p *pieceWise[T]) Passers() {
	b := p.b

	for color := White; color <= Black; color++ {
		myPawns := b.Pieces[Pawn] & b.Colors[color]
		myRearFill := frontFill(myPawns, color.Flip())
		theirFrontSpan := p.frontSpan[color.Flip()]

		var myRearSpan board.BitBoard
		switch color {
		case White:
			myRearSpan = myRearFill >> 8

		case Black:
			myRearSpan = myRearFill << 8
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

		// connected rooks
		if attack&p.b.Pieces[Rook]&p.b.Colors[color] != 0 {
			mg[color] += p.c.ConnectedRooks[0]
			eg[color] += p.c.ConnectedRooks[1]
		}

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

		// calculate knight outputs
		if (board.BitBoard(1)<<sq)&p.holes[color.Flip()]&p.pawnCover[color] != 0 {
			// the hole square is from the enemy's perspective, white's in black's territory
			if color == White {
				sq ^= 56
			}
			mg[color] += p.c.KnightOutpost[0][sq]
			eg[color] += p.c.KnightOutpost[1][sq]
		}

	case Pawn:
		pawn := board.BitBoard(1) << sq
		if p.passers&pawn != 0 {
			rank := sq / 8
			if color == Black {
				rank ^= 7
			}

			// if protected passers add protection bonus
			if pawn&p.pawnCover[color] != 0 {
				mg[color] += p.c.ProtectedPasser[0]
				eg[color] += p.c.ProtectedPasser[1]
			}

			mg[color] += p.c.PasserRank[0][rank-1]
			eg[color] += p.c.PasserRank[1][rank-1]

			// KPR, KPNB
			if p.b.Pieces[Knight]|p.b.Pieces[Bishop]|p.b.Pieces[Queen] == 0 || p.b.Pieces[Rook]|p.b.Pieces[Queen] == 0 {
				qSq := sq % 8
				if color == White {
					qSq += 56
				}
				// mid square between the pawn and its queening square
				mSq := (qSq + sq) / 2

				kingDist := Manhattan(mSq, p.kingSq[color.Flip()]) - Manhattan(mSq, p.kingSq[color])

				mg[color] += p.c.PasserKingDist[0] * T(kingDist)
				eg[color] += p.c.PasserKingDist[1] * T(kingDist)
			}

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
