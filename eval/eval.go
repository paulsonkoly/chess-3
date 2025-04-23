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
}

var Coefficients = CoeffSet[Score]{
	PSqT: [12][64]Score{
		{
			0, 0, 0, 0, 0, 0, 0, 0,
			24, 53, 27, 54, 98, 64, -7, -37,
			11, 44, 69, 64, 74, 115, 51, 30,
			6, 29, 22, 47, 60, 38, 48, 11,
			-7, 18, 19, 43, 43, 39, 42, 11,
			-5, 13, 18, 19, 38, 20, 72, 26,
			-6, 28, 4, 17, 25, 59, 89, 23,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			0, 0, 0, 0, 0, 0, 0, 0,
			110, 103, 84, 52, 34, 53, 104, 96,
			21, 8, -11, -40, -45, -41, -14, -4,
			8, 0, -17, -37, -39, -28, -14, -11,
			-7, -8, -25, -33, -30, -29, -19, -21,
			-11, -11, -24, -23, -24, -21, -28, -27,
			-6, -12, -11, -12, -11, -20, -28, -28,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			-254, -18, -41, -98, 4, -177, -87, -145,
			-50, -17, 88, 9, 96, 55, 21, -15,
			15, 40, 17, 57, 92, 110, 53, -3,
			20, 11, -4, 37, 7, 26, -9, 28,
			0, 9, 17, 15, 17, 28, 4, -6,
			-14, 2, 15, 20, 36, 15, 30, -1,
			-7, -11, -3, 23, 23, 32, 15, 19,
			-38, -16, -35, 2, 0, 1, -15, -38,
		},
		{
			29, -37, -11, 1, -28, -11, -9, -68,
			-1, 3, -25, 1, -30, -24, -19, -24,
			-25, -18, 18, 7, -14, -9, -25, -9,
			-2, 22, 24, 28, 22, 21, 9, -4,
			0, 1, 24, 21, 29, 15, 1, 10,
			-9, 2, -6, 18, 11, -3, -8, -13,
			-15, -6, -5, -5, -4, -6, -6, -13,
			-22, -14, 3, 10, -5, -5, -22, -2,
		},
		{
			-44, -44, -45, -89, -115, -88, -78, -88,
			-34, 12, -19, -21, 2, 49, 0, -1,
			-22, 4, 55, 20, 52, 33, 34, 22,
			-24, -6, 5, 16, 9, -7, -1, -25,
			-8, 5, -8, 28, 23, -5, -4, -32,
			11, 10, 19, -5, 8, 19, 8, 3,
			-2, 31, 4, 8, 10, 40, 42, 20,
			5, -8, 2, -6, -2, -13, -14, 2,
		},
		{
			-2, -5, -1, 14, 9, 6, -1, 16,
			-6, -5, 8, -8, -3, -16, 1, -29,
			9, -2, -11, 5, -8, 7, -6, -7,
			10, 8, 6, 9, 13, 2, -1, -1,
			-3, -3, 8, 2, 3, 6, -4, 3,
			-10, 3, -3, 6, 7, -4, -5, -7,
			1, -19, -10, -4, -5, -13, -14, -25,
			-8, 3, -7, 5, 2, 3, -13, -15,
		},
		{
			46, 50, 44, 40, 67, 69, 17, 30,
			-3, -16, 24, 58, 10, 43, 42, 14,
			-14, -4, -9, 14, 30, 31, 82, 33,
			-25, -14, 15, 17, 7, 4, 5, 26,
			-29, -28, -16, -11, 7, -15, -18, -6,
			-23, -12, -7, 7, 9, -8, 18, 15,
			-26, -15, -9, 8, 4, 1, 7, -47,
			-8, -3, 11, 19, 28, 14, -14, 0,
		},
		{
			6, 1, 1, 3, -6, -8, 8, 4,
			9, 23, 10, -6, 6, 8, 2, 1,
			12, 10, 11, 3, -2, -1, -9, -9,
			8, 6, 2, 2, 0, 4, 3, -9,
			7, 8, 7, 8, 1, 3, 1, -6,
			-1, -3, -4, -3, -4, -2, -20, -18,
			2, 2, 8, -1, -6, 1, -4, 2,
			3, 1, -2, -5, -13, -6, -1, -24,
		},
		{
			-43, 4, -5, -22, -4, -36, -44, -26,
			-26, -33, -39, -53, -77, 18, -36, 22,
			-20, -3, 16, -7, 40, 58, 80, 48,
			-21, -11, -10, -19, -15, -31, -22, -9,
			-3, -6, -2, -16, 2, 3, 0, -18,
			-17, 13, 5, 3, 8, 4, 5, -12,
			-10, 12, 17, 25, 30, 49, 41, 16,
			15, -2, 11, 22, 5, -14, -7, -8,
		},
		{
			12, 2, 16, 39, 12, 35, 9, 22,
			7, 24, 47, 75, 78, 52, 26, -42,
			-15, 6, 19, 31, 44, -13, -38, -52,
			4, 19, 35, 57, 46, 67, 25, -3,
			-10, 9, 13, 49, 27, 4, 7, -6,
			-22, -31, 4, 12, 11, 13, -2, -22,
			-10, -21, -30, -26, -27, -69, -75, -54,
			-44, -29, -29, -49, -9, -28, -60, -52,
		},
		{
			125, 93, 60, 65, 106, 94, 51, 74,
			72, 107, 78, 51, 55, 42, -30, -5,
			69, 129, 26, 59, -15, 34, 62, -17,
			5, 35, 29, -16, -35, -15, 24, -23,
			36, 26, 0, -66, -54, -76, -33, -76,
			4, 21, -58, -73, -71, -48, -5, -26,
			8, -11, -30, -80, -58, -25, 25, 19,
			-38, 21, -5, -87, -18, -45, 33, 21,
		},
		{
			-68, -34, -35, -19, -28, -13, -7, -41,
			-34, -11, -13, -3, -1, 18, 30, 9,
			-25, -9, 6, 1, 15, 21, 25, 17,
			-28, -6, 6, 20, 23, 28, 17, 11,
			-33, -7, 13, 30, 32, 33, 19, 14,
			-22, -6, 19, 32, 35, 30, 10, 4,
			-22, 2, 15, 31, 29, 21, 3, -14,
			-13, -18, -5, 10, -11, 4, -26, -43,
		},
	},
	PieceValues: [2][7]Score{
		{0, 80, 469, 514, 654, 1381, 0},
		{0, 100, 245, 259, 476, 881, 0},
	},
	TempoBonus: [2]Score{35, 24},
	KingAttackPieces: [2][4]Score{
		{3, 2, 3, 3},
		{-2, -1, -2, 4},
	},
	KingAttackCount: [2][7]Score{
		{0, 2, 3, 5, 7, 11, 22},
		{0, 1, 3, 3, 5, 9, 7},
	},
	MobilityKnight: [2][9]Score{
		{-44, -10, -1, 4, 14, 20, 24, 30, 32},
		{-52, -34, -11, -4, 3, 8, 8, 6, -4},
	},
	MobilityBishop: [2][14]Score{
		{-20, -7, 3, 6, 12, 15, 20, 21, 23, 25, 30, 34, 36, 15},
		{-44, -31, -28, -16, -8, 1, 5, 13, 18, 19, 17, 18, 32, 25},
	},
	MobilityRook: [2][11]Score{
		{-15, -8, -3, 4, 7, 12, 10, 19, 24, 34, 51},
		{-21, -19, -12, -9, -2, 2, 7, 8, 13, 15, 7},
	},
	KnightOutpost: [2][40]Score{
		{
			-15, -32, -120, -8, 77, 9, 24, 5,
			-8, -1, -23, 90, 14, 38, -46, -9,
			-58, -3, 75, 13, -73, 9, 48, 94,
			13, 35, 53, 33, 31, 60, 116, 80,
			0, 0, 0, 42, 43, 0, 0, 0,
		},
		{
			-27, 59, -59, -85, 51, -18, -7, 44,
			-12, 10, 7, -51, -34, -5, 34, 70,
			7, 17, -12, 32, 61, 25, 18, 0,
			18, -25, 7, 17, 24, 5, -4, -19,
			0, 0, 0, 20, 1, 0, 0, 0,
		},
	},
	ConnectedRooks:  [2]Score{-1, 4},
	BishopPair:      [2]Score{-30, 52},
	ProtectedPasser: [2]Score{35, 11},
	PasserKingDist:  [2]Score{15, 3},
	PasserRank: [2][6]Score{
		{-6, -25, -30, -9, -12, 42},
		{1, 8, 26, 52, 117, 89},
	},
}

// Phase is game phase.
var Phase = [...]int{0, 0, 1, 1, 2, 4, 0}

func Eval[T ScoreType](b *board.Board, _, _ T, c *CoeffSet[T]) T {
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

		if pType == Knight {
			pWise.Passers()
		}
	}

	for color := White; color <= Black; color++ {
		kingACnt := min(len(c.KingAttackCount[0])-1, pWise.kingACount[color])
		mg[color] += pWise.kingAScore[0][color] * c.KingAttackCount[0][kingACnt]
		eg[color] += pWise.kingAScore[1][color] * c.KingAttackCount[1][kingACnt]
	}

	score := TaperedScore(b, phase, mg[:], eg[:])

	return score
}

func TaperedScore[T ScoreType](b *board.Board, phase int, mg, eg []T) T {
	mgScore := mg[b.STM] - mg[b.STM.Flip()]
	egScore := eg[b.STM] - eg[b.STM.Flip()]

	mgPhase := min(phase, 24)
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
