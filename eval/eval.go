// package eval gives position evaluation measuerd in centipawns.
package eval

import (
	"math"

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

	// SafeChecks is the bonus per piece type for being able to give a safe check.
	SafeChecks [2][4]T

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
			42, 80, 53, 85, 74, 34, -41, -70,
			18, 36, 58, 61, 70, 100, 86, 35,
			-5, 18, 15, 21, 43, 36, 43, 16,
			-11, 8, 7, 24, 27, 20, 25, 5,
			-12, 5, 0, 5, 20, 11, 40, 10,
			-10, 6, -6, -2, 9, 29, 53, 5,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			0, 0, 0, 0, 0, 0, 0, 0,
			112, 99, 96, 37, 35, 63, 109, 121,
			24, 25, -17, -54, -58, -37, 3, 6,
			12, 3, -17, -35, -39, -29, -9, -11,
			-5, -3, -22, -29, -29, -25, -12, -22,
			-10, -6, -23, -18, -21, -23, -17, -25,
			-9, -5, -19, -12, -9, -21, -19, -27,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			-157, -113, -63, -32, 12, -65, -108, -103,
			-23, 1, 34, 47, 10, 83, -3, 19,
			-6, 27, 33, 33, 89, 63, 46, 13,
			0, -1, 12, 38, 18, 42, 11, 33,
			-12, -3, 9, 11, 21, 21, 21, 3,
			-34, -16, -7, 5, 22, 2, 7, -12,
			-40, -25, -20, 1, 0, 0, -9, -11,
			-74, -36, -36, -17, -17, -10, -29, -41,
		},
		{
			-42, -5, 5, 0, -5, -22, -7, -71,
			2, 8, 0, -1, -5, -18, 5, -22,
			2, 4, 21, 21, -1, -4, -9, -10,
			14, 22, 32, 26, 25, 25, 17, 2,
			23, 19, 38, 30, 35, 27, 16, 17,
			4, 10, 17, 32, 25, 10, 4, 8,
			7, 10, 10, 10, 11, 5, 4, 14,
			8, -6, 9, 12, 13, 1, -6, 1,
		},
		{
			-39, -60, -40, -90, -80, -85, -43, -60,
			-30, 1, -10, -30, 0, -12, -7, -43,
			-12, 9, 10, 27, 10, 54, 30, 17,
			-19, -8, 10, 24, 22, 13, -6, -22,
			-18, -13, -16, 15, 11, -11, -15, -3,
			-12, -5, -3, -10, -4, -1, 1, 4,
			-6, 0, 1, -13, -4, 10, 17, 2,
			-18, 1, -14, -21, -12, -18, 8, -4,
		},
		{
			15, 17, 7, 20, 15, 9, 7, 6,
			4, 7, 7, 7, 0, 4, 9, 0,
			16, 8, 10, -1, 5, 6, 3, 7,
			12, 16, 12, 23, 13, 14, 12, 12,
			6, 13, 20, 17, 16, 15, 10, -9,
			4, 16, 16, 15, 19, 14, 6, -3,
			9, 0, -5, 11, 6, 1, 5, -9,
			0, 9, 7, 10, 12, 14, -8, -11,
		},
		{
			27, 17, 21, 22, 36, 34, 33, 57,
			8, 3, 28, 46, 31, 55, 35, 72,
			-15, 17, 12, 17, 47, 42, 87, 55,
			-18, -8, -7, 5, 7, 7, 15, 19,
			-33, -32, -19, -8, -8, -27, -3, -14,
			-31, -28, -18, -15, -8, -12, 15, -3,
			-34, -24, -8, -5, -2, -1, 12, -21,
			-20, -15, -6, 3, 7, 0, 6, -20,
		},
		{
			21, 28, 33, 27, 23, 26, 21, 14,
			22, 35, 34, 23, 25, 18, 17, 1,
			23, 21, 22, 18, 8, 6, -4, -6,
			26, 22, 29, 21, 10, 8, 6, 2,
			20, 21, 21, 18, 14, 15, 1, 2,
			13, 11, 11, 12, 6, 1, -19, -16,
			13, 12, 10, 10, 4, -2, -11, 2,
			9, 9, 12, 5, -1, 2, -6, -3,
		},
		{
			-49, -28, -8, 14, 10, 23, 20, -30,
			-16, -29, -17, -25, -31, 8, -13, 15,
			0, 1, -1, 7, 5, 50, 54, 52,
			-16, -7, -10, -16, -5, 2, 7, -1,
			-10, -18, -15, -2, -2, -6, -3, 1,
			-16, -6, -2, -8, 2, 2, 9, 0,
			-8, -4, 4, 13, 10, 18, 23, 35,
			-11, -15, -4, 5, 4, -15, 6, -12,
		},
		{
			42, 26, 40, 39, 43, 34, -2, 46,
			5, 25, 53, 64, 92, 60, 27, 34,
			8, 21, 51, 44, 60, 36, -1, -15,
			27, 33, 40, 56, 55, 44, 42, 32,
			5, 34, 37, 47, 45, 36, 23, 11,
			-4, 3, 22, 27, 28, 21, -7, -13,
			-19, -13, -13, -10, -4, -32, -69, -99,
			-16, -15, -9, -9, -17, -24, -46, -28,
		},
		{
			50, 55, 37, -17, -41, -15, -30, 93,
			-83, -20, -45, 50, 13, -19, -14, -61,
			-70, 25, -16, -30, 0, 62, 3, -31,
			-32, -44, -51, -58, -62, -51, -68, -115,
			-70, -58, -61, -75, -78, -64, -94, -131,
			-32, -17, -51, -57, -44, -56, -29, -59,
			40, -1, -14, -42, -44, -29, 15, 18,
			23, 58, 27, -70, -16, -43, 31, 30,
		},
		{
			-93, -49, -32, 2, -3, -6, -12, -107,
			-2, 25, 34, 14, 33, 48, 45, 13,
			6, 26, 40, 50, 55, 47, 50, 16,
			-7, 27, 45, 53, 56, 52, 46, 20,
			-11, 16, 36, 50, 49, 38, 30, 15,
			-20, 3, 22, 32, 30, 23, 6, -4,
			-36, -8, 4, 12, 16, 7, -13, -35,
			-67, -54, -32, -8, -29, -14, -47, -78,
		},
	},
	PieceValues: [2][7]Score{
		{0, 80, 375, 407, 492, 1026, 0},
		{0, 124, 318, 330, 621, 1209, 0},
	},
	TempoBonus: [2]Score{25, 22},
	KingAttackPieces: [2][4]Score{
		{5, 5, 6, 31},
		{8, 15, 11, -76},
	},
	SafeChecks: [2][4]Score{
		{12, 7, 12, 5},
		{3, 6, 9, 10},
	},
	MobilityKnight: [2][9]Score{
		{-60, -40, -28, -22, -14, -8, -1, 5, 9},
		{-43, -6, 14, 23, 31, 40, 38, 35, 26},
	},
	MobilityBishop: [2][14]Score{
		{-42, -32, -22, -19, -13, -5, 1, 5, 6, 7, 10, 12, 8, 27},
		{-33, -14, -10, 3, 16, 29, 32, 39, 45, 43, 42, 41, 49, 32},
	},
	MobilityRook: [2][11]Score{
		{-27, -19, -17, -12, -7, -2, 4, 11, 12, 16, 24},
		{-2, 4, 12, 15, 22, 26, 28, 30, 38, 43, 38},
	},
	KnightOutpost: [2][40]Score{
		{
			-46, 28, -28, 13, 46, 73, 58, -12,
			49, -8, -26, -42, 12, -40, -11, -26,
			-4, -5, 10, 34, 10, 42, 11, 11,
			0, 27, 36, 35, 51, 54, 74, 14,
			0, 0, 0, 36, 48, 0, 0, 0,
		},
		{
			-79, 134, -6, -10, -31, 56, -35, 43,
			-33, 17, 31, 19, 4, 24, 13, 89,
			24, 12, 24, 24, 34, 30, 25, 28,
			19, 13, 15, 25, 29, 10, 1, 20,
			0, 0, 0, 22, 22, 0, 0, 0,
		},
	},
	ConnectedRooks:  [2]Score{1, 7},
	BishopPair:      [2]Score{-16, 71},
	ProtectedPasser: [2]Score{30, 18},
	PasserKingDist:  [2]Score{9, 3},
	PasserRank: [2][6]Score{
		{-17, -35, -29, -5, -2, 38},
		{1, 4, 27, 54, 121, 86},
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

	for pType := Pawn; pType <= King; pType++ {
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
	}

	pWise.calcCover()

	for pType := Knight; pType <= Queen; pType++ {
		for color := White; color <= Black; color++ {
			pWise.safeChecks(pType, color)
		}
	}

	for color := White; color <= Black; color++ {
		mg[color] += sigmoidal(pWise.kingAScore[0][color])
		eg[color] += sigmoidal(pWise.kingAScore[1][color])
	}

	score := TaperedScore(b, phase, mg[:], eg[:])

	return score
}

// def f(x) = 600.fdiv(1+Math.exp(-0.2*(x-50)))
//
// 100.times.map { |x| f(x).round }.each_slice(10).to_a
//
// where 600 is the maximal bonus for attack, 0.2 is the steepness of the
// sigmoid, and 50 is the inflection point, implying a 0-100 range for king
// attack score
var sigm = [...]Score{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 1, 1, 1, 1, 1,
	1, 2, 2, 3, 3, 4, 5, 6, 7, 9,
	11, 13, 16, 19, 23, 28, 34, 41, 50, 60,
	72, 85, 101, 119, 139, 161, 186, 213, 241, 270,
	300, 330, 359, 387, 414, 439, 461, 481, 499, 515,
	528, 540, 550, 559, 566, 572, 577, 581, 584, 587,
	589, 591, 593, 594, 595, 596, 597, 597, 598, 598,
	599, 599, 599, 599, 599, 599, 600, 600, 600, 600,
	600, 600, 600, 600, 600, 600, 600, 600, 600, 600,
}

func sigmoidal[T ScoreType](n T) T {
	if _, ok := (any(n)).(Score); ok {
		return T(sigm[Clamp(int(n), 0, len(sigm)-1)])
	}
	return T(600.0 / (1.0 + math.Exp(-0.2*(float64(n)-50.0))))
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
	attacks    [2][6]board.BitBoard
	cover      [2]board.BitBoard
	kingNb     [2]board.BitBoard
	kingRays   [2][2]board.BitBoard
	frontSpan  [2]board.BitBoard
	holes      [2]board.BitBoard
	kingAScore [2][2]T
	kingSq     [2]Square
}

// the player's side of the board with the extra 2 central squares included at
// enemy side.
var sideOfBoard = [2]board.BitBoard{0x00000018_ffffffff, 0xffffffff_18000000}

func newPieceWise[T ScoreType](b *board.Board, c *CoeffSet[T]) pieceWise[T] {
	result := pieceWise[T]{b: b, c: c}
	result.occ = b.Colors[White] | b.Colors[Black]

	// the order of these is important. There are inter-dependencies
	result.calcKingSquares()
	result.calcPawnBitBoards()
	result.calcPassers()

	return result
}

func (p *pieceWise[T]) calcKingSquares() {
	b := p.b

	for color := White; color <= Black; color++ {
		king := b.Colors[color] & b.Pieces[King]
		kingSq := king.LowestSet()
		kingA := movegen.KingMoves(kingSq)

		p.attacks[color][King-Pawn] = kingA
		p.kingRays[color][0] = movegen.BishopMoves(kingSq, p.occ)
		p.kingRays[color][Rook-Bishop] = movegen.RookMoves(kingSq, p.occ)
		p.kingSq[color] = kingSq
		p.kingNb[color] = king | kingA
	}
}

// this has to be called after the per piece loop as we need the attacks to be filled in
func (p *pieceWise[T]) calcCover() {
	for color := White; color <= Black; color++ {
		p.cover[color] = p.attacks[color][0] |
			p.attacks[color][Knight-Pawn] |
			p.attacks[color][Bishop-Pawn] |
			p.attacks[color][Rook-Pawn] |
			p.attacks[color][Queen-Pawn] |
			p.attacks[color][King-Pawn]
	}
}

func (p *pieceWise[T]) calcPawnBitBoards() {
	b := p.b

	wP := b.Pieces[Pawn] & b.Colors[White]
	p.attacks[White][0] = movegen.PawnCaptureMoves(wP, White)
	bP := b.Pieces[Pawn] & b.Colors[Black]
	p.attacks[Black][0] = movegen.PawnCaptureMoves(bP, Black)

	// various useful pawn bitboards
	wFrontSpan := frontFill(wP, White) << 8
	bFrontSpan := frontFill(bP, Black) >> 8

	p.frontSpan[White] = wFrontSpan
	p.frontSpan[Black] = bFrontSpan

	// calculate holes in our position, squares that cannot be protected by one
	// of our pawns.
	wCover := ((wFrontSpan & ^board.AFile) >> 1) | ((wFrontSpan & ^board.HFile) << 1)
	bCover := ((bFrontSpan & ^board.HFile) << 1) | ((bFrontSpan & ^board.AFile) >> 1)
	p.holes[White] = sideOfBoard[White] & ^wCover
	p.holes[Black] = sideOfBoard[Black] & ^bCover
}

func (p *pieceWise[T]) calcPassers() {
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
		p.attacks[color][Queen-Pawn] |= attack

	case Rook:
		attack = movegen.RookMoves(sq, occ)
		p.attacks[color][Rook-Pawn] |= attack

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
		p.attacks[color][Bishop-Pawn] |= attack

		mobCnt := (attack & ^p.b.Colors[color]).Count()
		mg[color] += p.c.MobilityBishop[0][mobCnt]
		eg[color] += p.c.MobilityBishop[1][mobCnt]

	case Knight:
		attack = movegen.KnightMoves(sq)
		p.attacks[color][Knight-Pawn] |= attack

		mobCnt := (attack & ^p.b.Colors[color] & ^p.attacks[color.Flip()][0]).Count()
		mg[color] += p.c.MobilityKnight[0][mobCnt]
		eg[color] += p.c.MobilityKnight[1][mobCnt]

		// calculate knight outputs
		if (board.BitBoard(1)<<sq)&p.holes[color.Flip()]&p.attacks[color][0] != 0 {
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
			if pawn&p.attacks[color][0] != 0 {
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

	if p.kingNb[color.Flip()]&attack != 0 {
		p.kingAScore[0][color] += p.c.KingAttackPieces[0][pType-Knight]
		p.kingAScore[1][color] += p.c.KingAttackPieces[1][pType-Knight]
	}
}

func (p *pieceWise[T]) safeChecks(pType Piece, color Color) {
	eCover := p.cover[color.Flip()]

	var safeChecks board.BitBoard

	switch pType {

	case Queen:
		eKAttack := p.kingRays[color.Flip()][0] | p.kingRays[color.Flip()][Rook-Bishop]
		safeChecks = p.attacks[color][Queen-Pawn] & eKAttack & ^eCover & ^p.b.Colors[color]

	case Rook:
		eKAttack := p.kingRays[color.Flip()][Rook-Bishop]
		safeChecks = p.attacks[color][Rook-Pawn] & eKAttack & ^eCover & ^p.b.Colors[color]

	case Bishop:
		eKAttack := p.kingRays[color.Flip()][0]
		safeChecks = p.attacks[color][Bishop-Pawn] & eKAttack & ^eCover & ^p.b.Colors[color]

	case Knight:
		eKAttack := movegen.KnightMoves(p.kingSq[color.Flip()])
		safeChecks = p.attacks[color][Knight-Pawn] & eKAttack & ^eCover & ^p.b.Colors[color]
	}

	p.kingAScore[0][color] += p.c.SafeChecks[0][pType-Knight] * T(safeChecks.Count())
	p.kingAScore[1][color] += p.c.SafeChecks[1][pType-Knight] * T(safeChecks.Count())
}

func Manhattan(a, b Square) int {
	ax, ay, bx, by := int(a%8), int(a/8), int(b%8), int(b/8)
	return max(Abs(ax-bx), Abs(ay-by))
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
