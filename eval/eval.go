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
			25, 53, 29, 55, 94, 65, -3, -41,
			9, 44, 69, 62, 73, 114, 49, 29,
			5, 29, 22, 46, 59, 37, 47, 11,
			-8, 18, 18, 42, 43, 39, 42, 10,
			-6, 13, 18, 19, 38, 19, 72, 26,
			-7, 27, 4, 17, 25, 59, 88, 22,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			0, 0, 0, 0, 0, 0, 0, 0,
			110, 102, 84, 50, 35, 53, 103, 98,
			22, 8, -11, -39, -45, -41, -13, -4,
			8, 0, -18, -37, -40, -28, -14, -12,
			-7, -8, -26, -33, -30, -29, -19, -21,
			-12, -11, -24, -23, -24, -21, -28, -27,
			-6, -13, -11, -13, -11, -20, -28, -28,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			-252, -22, -40, -99, 3, -180, -88, -140,
			-50, -16, 89, 10, 95, 56, 22, -15,
			16, 38, 18, 57, 92, 108, 55, 2,
			20, 11, -4, 38, 7, 26, -9, 28,
			0, 7, 17, 15, 17, 28, 4, -6,
			-14, 1, 15, 20, 35, 15, 29, -1,
			-8, -11, -3, 23, 22, 32, 14, 18,
			-40, -16, -36, 2, -1, 1, -15, -39,
		},
		{
			26, -36, -10, 1, -27, -10, -11, -71,
			-1, 3, -26, 0, -29, -24, -20, -24,
			-26, -18, 17, 7, -13, -9, -24, -11,
			-1, 22, 24, 27, 22, 22, 8, -4,
			1, 3, 24, 21, 30, 15, 1, 10,
			-9, 3, -6, 18, 12, -3, -7, -14,
			-14, -5, -4, -5, -4, -6, -5, -13,
			-22, -13, 3, 9, -5, -5, -22, -1,
		},
		{
			-44, -44, -47, -89, -118, -90, -78, -83,
			-34, 13, -19, -21, 2, 51, 1, -1,
			-22, 3, 55, 20, 53, 33, 34, 23,
			-24, -6, 5, 16, 9, -7, -2, -25,
			-9, 4, -8, 27, 23, -5, -4, -33,
			11, 10, 19, -5, 8, 18, 7, 3,
			-3, 30, 3, 8, 9, 40, 41, 19,
			5, -9, 1, -7, -3, -14, -15, 1,
		},
		{
			-3, -5, -1, 14, 10, 7, -2, 14,
			-6, -5, 8, -9, -3, -16, 2, -29,
			9, -2, -11, 5, -9, 7, -5, -7,
			9, 9, 5, 10, 13, 2, -1, -2,
			-2, -3, 9, 2, 3, 5, -5, 3,
			-10, 3, -3, 5, 7, -4, -4, -8,
			1, -19, -9, -3, -4, -13, -13, -24,
			-9, 4, -7, 5, 2, 3, -12, -14,
		},
		{
			41, 47, 41, 37, 65, 70, 19, 33,
			-3, -16, 24, 57, 10, 41, 43, 15,
			-15, -4, -8, 13, 30, 31, 80, 29,
			-25, -15, 14, 16, 7, 4, 4, 27,
			-28, -28, -16, -12, 8, -16, -15, -8,
			-23, -11, -6, 7, 8, -9, 20, 15,
			-27, -16, -9, 8, 4, 2, 9, -47,
			-9, -3, 11, 19, 28, 14, -14, 0,
		},
		{
			8, 2, 3, 3, -5, -7, 7, 4,
			10, 23, 10, -5, 6, 8, 2, 0,
			12, 10, 10, 4, -2, -1, -8, -8,
			9, 7, 2, 2, -1, 4, 3, -10,
			7, 7, 7, 8, 1, 3, 0, -6,
			-1, -2, -3, -2, -4, -1, -19, -18,
			3, 1, 8, 0, -6, 0, -4, 2,
			3, 1, -1, -4, -13, -6, -1, -24,
		},
		{
			-45, 6, -4, -19, -4, -28, -39, -23,
			-26, -33, -38, -53, -77, 19, -36, 22,
			-18, -3, 16, -7, 39, 57, 81, 47,
			-19, -10, -9, -18, -15, -31, -21, -9,
			-3, -6, -2, -16, 2, 2, 0, -18,
			-16, 13, 5, 4, 8, 5, 6, -13,
			-10, 11, 17, 25, 30, 49, 40, 16,
			13, -2, 11, 22, 5, -14, -6, -9,
		},
		{
			15, 0, 15, 36, 11, 30, 6, 20,
			8, 23, 45, 75, 77, 51, 25, -41,
			-16, 6, 19, 31, 43, -13, -39, -51,
			4, 18, 34, 57, 45, 67, 26, -3,
			-9, 9, 14, 48, 28, 5, 8, -7,
			-22, -30, 5, 11, 11, 12, -3, -21,
			-9, -19, -29, -25, -26, -69, -76, -53,
			-41, -29, -30, -48, -8, -28, -62, -51,
		},
		{
			114, 92, 58, 65, 101, 87, 44, 69,
			75, 106, 78, 48, 52, 39, -34, -2,
			67, 130, 25, 54, -14, 34, 62, -20,
			8, 34, 31, -20, -37, -19, 22, -23,
			37, 28, -2, -66, -57, -76, -33, -76,
			1, 22, -56, -72, -71, -48, -5, -27,
			9, -11, -30, -80, -58, -24, 26, 20,
			-37, 21, -4, -87, -17, -45, 33, 22,
		},
		{
			-67, -34, -36, -19, -27, -11, -6, -40,
			-35, -10, -13, -3, -1, 19, 31, 8,
			-25, -8, 6, 1, 15, 21, 26, 18,
			-28, -6, 6, 20, 23, 29, 18, 12,
			-34, -7, 13, 30, 32, 33, 19, 13,
			-21, -6, 19, 32, 35, 30, 10, 4,
			-22, 2, 15, 31, 29, 21, 3, -14,
			-13, -18, -6, 10, -12, 4, -26, -44,
		},
	},
	PieceValues: [2][7]Score{
		{0, 80, 468, 513, 653, 1379, 0},
		{0, 101, 245, 260, 478, 886, 0},
	},
	TempoBonus: [2]Score{35, 25},
	KingAttackPieces: [2][4]Score{
		{3, 2, 3, 3},
		{-1, -1, -1, 4},
	},
	KingAttackCount: [2][7]Score{
		{0, 2, 3, 5, 7, 11, 22},
		{0, 2, 3, 4, 7, 12, 7},
	},
	MobilityKnight: [2][9]Score{
		{-43, -10, -1, 4, 14, 20, 25, 30, 32},
		{-52, -32, -11, -4, 3, 9, 9, 6, -4},
	},
	MobilityBishop: [2][14]Score{
		{-21, -7, 3, 6, 11, 14, 20, 20, 23, 25, 30, 34, 35, 16},
		{-43, -30, -28, -16, -8, 2, 5, 14, 18, 19, 18, 19, 32, 25},
	},
	MobilityRook: [2][11]Score{
		{-15, -8, -3, 4, 7, 12, 10, 19, 24, 34, 51},
		{-21, -19, -12, -9, -2, 2, 7, 8, 13, 15, 7},
	},
	KnightOutpost: [2][40]Score{
		{
			-14, -28, -112, -13, 71, 8, 24, 5,
			-6, -3, -25, 82, 16, 34, -39, -5,
			-47, -1, 74, 18, -74, 12, 45, 86,
			16, 36, 54, 32, 30, 60, 114, 76,
			0, 0, 0, 43, 43, 0, 0, 0,
		},
		{
			-25, 56, -57, -85, 48, -17, -2, 42,
			-8, 8, 5, -48, -32, -7, 31, 71,
			9, 16, -11, 29, 60, 23, 16, 5,
			18, -23, 7, 18, 23, 5, -3, -16,
			0, 0, 0, 20, 1, 0, 0, 0,
		},
	},
	ConnectedRooks:  [2]Score{-1, 4},
	BishopPair:      [2]Score{-30, 53},
	ProtectedPasser: [2]Score{35, 11},
	PasserKingDist:  [2]Score{15, 3},
	PasserRank: [2][6]Score{
		{-5, -25, -30, -9, -11, 41},
		{1, 8, 26, 52, 117, 89},
	},
	LazyMargin: [...]Score{718, 495, 602, 615, 615, 626, 626},
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
