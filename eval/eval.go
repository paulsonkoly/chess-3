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
			43, 76, 54, 83, 67, 30, -43, -71,
			20, 38, 61, 64, 71, 99, 85, 37,
			-4, 20, 18, 23, 45, 39, 50, 21,
			-11, 10, 8, 26, 29, 26, 30, 10,
			-12, 6, 2, 6, 23, 16, 46, 15,
			-10, 7, -5, 0, 11, 34, 58, 9,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			0, 0, 0, 0, 0, 0, 0, 0,
			107, 95, 92, 34, 32, 59, 106, 117,
			23, 24, -18, -54, -57, -36, 2, 5,
			12, 3, -17, -35, -38, -29, -10, -12,
			-5, -3, -22, -28, -28, -25, -13, -23,
			-9, -6, -22, -17, -20, -22, -17, -25,
			-8, -4, -18, -12, -8, -20, -19, -27,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			-169, -110, -63, -39, 5, -77, -96, -107,
			-20, 3, 36, 41, 3, 73, -4, 12,
			-4, 29, 34, 28, 78, 60, 31, 3,
			3, 0, 13, 32, 6, 31, 4, 24,
			-9, -1, 9, 11, 14, 15, 8, -3,
			-30, -12, -3, 8, 25, 5, 10, -10,
			-36, -21, -16, 5, 4, 5, -4, -7,
			-72, -32, -32, -14, -14, -5, -26, -37,
		},
		{
			-37, -7, 5, 3, -3, -16, -8, -66,
			1, 6, -1, 2, -2, -15, 4, -19,
			1, 3, 20, 23, 4, -2, -4, -6,
			11, 20, 30, 29, 30, 31, 20, 6,
			20, 17, 36, 29, 37, 28, 21, 18,
			1, 6, 14, 30, 23, 6, 1, 6,
			4, 7, 7, 7, 9, 3, 2, 12,
			5, -9, 7, 9, 11, -1, -8, -2,
		},
		{
			-39, -61, -42, -106, -91, -88, -38, -72,
			-29, 5, -8, -29, -6, -18, -20, -49,
			-10, 12, 13, 26, 6, 47, 16, 7,
			-18, -5, 10, 22, 17, 7, -13, -28,
			-17, -13, -16, 13, 7, -13, -15, -9,
			-12, -5, -2, -10, -5, 2, 3, 4,
			-6, 0, 1, -13, -1, 13, 19, 3,
			-18, -2, -15, -20, -10, -16, 8, -1,
		},
		{
			14, 18, 8, 23, 18, 10, 6, 9,
			2, 3, 5, 6, 1, 5, 11, 1,
			14, 6, 8, -1, 6, 8, 6, 11,
			10, 14, 11, 24, 15, 16, 14, 15,
			5, 13, 20, 17, 17, 15, 10, -7,
			3, 15, 14, 14, 17, 11, 4, -4,
			9, -1, -6, 9, 3, -2, 3, -10,
			-1, 9, 5, 8, 9, 11, -10, -14,
		},
		{
			24, 15, 15, 14, 26, 35, 32, 61,
			7, 2, 24, 40, 22, 39, 25, 64,
			-15, 13, 6, 4, 29, 18, 62, 36,
			-14, -4, -4, 5, 5, -5, 3, 6,
			-29, -28, -15, -3, -3, -30, -14, -21,
			-26, -24, -13, -11, -3, -14, 7, -8,
			-30, -20, -4, -1, 2, -1, 9, -21,
			-16, -10, -2, 6, 10, 1, 5, -18,
		},
		{
			20, 27, 34, 29, 26, 25, 23, 12,
			21, 34, 34, 24, 28, 21, 19, 3,
			22, 20, 23, 22, 13, 14, 3, -1,
			22, 19, 26, 20, 10, 13, 10, 6,
			16, 17, 17, 14, 12, 17, 6, 4,
			8, 7, 7, 9, 3, 2, -15, -14,
			8, 7, 6, 6, 1, -3, -10, 0,
			6, 5, 8, 2, -3, 0, -6, -3,
		},
		{
			-46, -29, -7, 6, 8, 16, 13, -29,
			-8, -25, -17, -22, -41, 7, -6, 31,
			0, -2, -6, 2, 2, 37, 42, 40,
			-13, -5, -6, -10, -9, -2, -3, -5,
			-8, -14, -11, -4, -3, -10, -3, -4,
			-12, -3, -1, -7, 0, 2, 9, 1,
			-5, -2, 5, 13, 12, 23, 27, 39,
			-7, -13, -4, 6, 5, -12, 7, -8,
		},
		{
			26, 21, 32, 23, 22, 19, -6, 25,
			0, 38, 65, 71, 94, 40, 31, 14,
			1, 22, 54, 48, 54, 23, -14, -39,
			25, 35, 43, 63, 58, 33, 20, 1,
			8, 36, 43, 53, 46, 23, 9, -11,
			-2, 7, 26, 27, 27, 15, -16, -31,
			-18, -12, -13, -10, 2, -35, -73, -110,
			-19, -19, -9, -5, -10, -29, -53, -41,
		},
		{
			41, 24, 16, -32, -18, 11, 14, 91,
			-72, -29, -61, 23, -1, -17, -1, -37,
			-70, 7, -54, -63, -36, 32, 9, -31,
			-53, -63, -86, -117, -106, -80, -75, -115,
			-83, -73, -94, -117, -109, -83, -96, -147,
			-39, -17, -62, -67, -52, -54, -22, -63,
			42, 8, -6, -37, -35, -19, 27, 22,
			25, 63, 26, -72, -16, -42, 33, 28,
		},
		{
			-97, -50, -32, 0, -10, -10, -14, -98,
			-6, 22, 32, 15, 32, 46, 43, 12,
			4, 26, 41, 51, 57, 48, 48, 17,
			-5, 28, 46, 57, 58, 54, 46, 21,
			-9, 17, 38, 52, 50, 39, 30, 20,
			-17, 2, 22, 31, 29, 21, 5, 0,
			-35, -10, 1, 10, 12, 4, -15, -32,
			-66, -53, -31, -9, -29, -13, -44, -73,
		},
	},
	PieceValues: [2][7]Score{
		{0, 80, 374, 408, 495, 1098, 0},
		{0, 121, 316, 328, 610, 1132, 0},
	},
	TempoBonus: [2]Score{25, 22},
	KingAttackPieces: [2][4]Score{
		{3, 2, 4, 2},
		{-1, -1, -2, 7},
	},
	KingAttackCount: [2][7]Score{
		{0, 2, 3, 6, 8, 10, 14},
		{0, 2, 3, 5, 8, 12, 51},
	},
	MobilityKnight: [2][9]Score{
		{-60, -39, -27, -21, -13, -7, 1, 7, 13},
		{-47, -10, 10, 20, 28, 36, 36, 32, 23},
	},
	MobilityBishop: [2][14]Score{
		{-40, -30, -20, -17, -11, -5, 1, 4, 4, 5, 7, 9, 5, 28},
		{-38, -19, -14, -2, 11, 26, 29, 36, 44, 43, 42, 41, 48, 31},
	},
	MobilityRook: [2][11]Score{
		{-24, -17, -15, -10, -6, -3, 1, 7, 9, 13, 17},
		{-4, 2, 10, 12, 20, 25, 29, 32, 40, 45, 42},
	},
	KnightOutpost: [2][40]Score{
		{
			-42, 10, -59, 12, 40, 55, 43, -3,
			22, -6, -27, -19, 15, -33, -13, -29,
			-7, -3, 12, 33, 8, 31, 13, 12,
			1, 28, 37, 36, 52, 58, 75, 18,
			0, 0, 0, 37, 50, 0, 0, 0,
		},
		{
			-65, 118, -18, -26, -20, 37, -20, 44,
			-27, 15, 29, -4, 4, 16, 19, 78,
			27, 10, 21, 23, 36, 36, 20, 23,
			20, 13, 15, 24, 27, 6, 1, 18,
			0, 0, 0, 21, 21, 0, 0, 0,
		},
	},
	ConnectedRooks:  [2]Score{1, 6},
	BishopPair:      [2]Score{-17, 75},
	ProtectedPasser: [2]Score{31, 17},
	PasserKingDist:  [2]Score{10, 3},
	PasserRank: [2][6]Score{
		{-14, -33, -31, -6, -3, 39},
		{1, 5, 27, 53, 119, 84},
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
