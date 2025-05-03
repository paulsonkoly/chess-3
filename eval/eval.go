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

	// KingShelter is the bonus for damage on the oppoent's king shelter.
	KingShelter [2]T

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
			50, 87, 60, 90, 82, 38, -39, -79,
			21, 38, 61, 63, 71, 101, 89, 34,
			-2, 21, 17, 23, 45, 37, 43, 15,
			-9, 10, 8, 25, 27, 20, 24, 4,
			-10, 7, 1, 5, 20, 9, 38, 8,
			-8, 8, -6, -4, 8, 25, 49, 1,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			0, 0, 0, 0, 0, 0, 0, 0,
			113, 102, 100, 41, 45, 70, 115, 129,
			24, 26, -17, -54, -58, -37, 3, 7,
			10, 2, -18, -35, -38, -29, -9, -12,
			-6, -3, -23, -28, -29, -25, -13, -22,
			-11, -7, -23, -17, -20, -23, -17, -25,
			-10, -5, -20, -12, -9, -21, -20, -27,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			-159, -113, -67, -34, 1, -77, -126, -117,
			-24, 0, 34, 49, 8, 72, -9, 8,
			-7, 29, 35, 34, 85, 61, 41, 10,
			2, 0, 13, 39, 21, 43, 11, 32,
			-11, -2, 10, 12, 23, 23, 24, 4,
			-34, -15, -6, 5, 23, 4, 9, -11,
			-40, -25, -19, 2, 2, 2, -7, -10,
			-72, -36, -36, -17, -16, -9, -29, -40,
		},
		{
			-45, -6, 5, -2, -5, -22, -5, -71,
			3, 8, 0, -3, -8, -18, 4, -21,
			2, 4, 21, 20, -2, -9, -13, -13,
			14, 22, 32, 26, 24, 22, 16, 0,
			24, 20, 39, 31, 36, 27, 15, 18,
			6, 12, 19, 34, 27, 11, 4, 10,
			8, 11, 11, 12, 12, 7, 5, 17,
			10, -4, 11, 14, 15, 2, -4, 2,
		},
		{
			-39, -60, -41, -91, -82, -89, -52, -60,
			-30, 3, -9, -31, 1, -18, -10, -54,
			-11, 10, 12, 28, 7, 53, 25, 16,
			-19, -7, 10, 26, 24, 11, -4, -22,
			-17, -12, -15, 17, 12, -9, -13, -1,
			-11, -3, 0, -8, -3, 1, 5, 5,
			-4, 2, 4, -11, -2, 12, 20, 5,
			-16, 3, -13, -19, -10, -16, 9, -2,
		},
		{
			17, 19, 10, 22, 16, 11, 10, 8,
			6, 8, 8, 9, 1, 7, 12, 4,
			18, 12, 12, 1, 8, 8, 6, 9,
			13, 17, 13, 25, 14, 16, 13, 13,
			7, 15, 21, 18, 17, 16, 11, -9,
			5, 17, 17, 16, 21, 16, 7, -2,
			10, 1, -4, 12, 7, 3, 6, -8,
			1, 10, 9, 12, 13, 15, -6, -10,
		},
		{
			29, 18, 26, 24, 35, 32, 26, 59,
			7, 3, 29, 46, 30, 45, 20, 52,
			-14, 17, 13, 17, 45, 39, 71, 44,
			-16, -6, -6, 7, 9, 6, 11, 16,
			-33, -31, -18, -6, -7, -27, -4, -15,
			-30, -28, -17, -14, -7, -11, 17, -3,
			-34, -23, -7, -3, -1, 0, 16, -21,
			-20, -14, -6, 3, 7, 0, 5, -21,
		},
		{
			22, 29, 34, 28, 25, 28, 24, 16,
			23, 36, 36, 25, 28, 22, 21, 6,
			24, 22, 23, 20, 10, 7, 0, -4,
			27, 24, 30, 23, 11, 9, 6, 4,
			22, 22, 22, 19, 16, 17, 2, 4,
			15, 12, 12, 14, 7, 1, -20, -14,
			14, 13, 11, 11, 5, -2, -12, 4,
			11, 10, 14, 7, 1, 3, -4, 0,
		},
		{
			-49, -29, -7, 16, 16, 25, 14, -26,
			-13, -27, -14, -17, -28, 13, -8, 23,
			-1, -2, -2, 12, 5, 56, 43, 54,
			-17, -8, -6, -10, -1, 10, 10, 6,
			-10, -17, -12, 0, 3, -2, 3, 7,
			-15, -5, -1, -5, 1, 3, 13, 4,
			-7, -3, 5, 13, 10, 18, 24, 35,
			-11, -14, -4, 4, 3, -15, 6, -15,
		},
		{
			45, 34, 51, 46, 50, 41, 6, 51,
			2, 29, 57, 64, 99, 62, 21, 27,
			10, 26, 56, 48, 67, 35, 1, -17,
			28, 36, 45, 59, 62, 43, 41, 31,
			6, 38, 41, 54, 48, 40, 25, 12,
			-2, 7, 27, 31, 36, 28, -3, -10,
			-18, -9, -10, -6, -1, -25, -61, -94,
			-13, -13, -6, -9, -15, -23, -46, -25,
		},
		{
			75, 113, 83, 27, -27, -22, -57, 109,
			-63, 27, 5, 104, 51, 11, 7, -60,
			-38, 58, 25, 7, 41, 96, 16, -12,
			-8, -13, -29, -37, -36, -34, -60, -109,
			-61, -50, -46, -70, -76, -67, -104, -135,
			-45, -38, -63, -61, -46, -73, -47, -69,
			37, -9, -17, -42, -47, -33, 5, 15,
			22, 54, 26, -59, -13, -41, 30, 30,
		},
		{
			-105, -63, -39, -4, -4, -6, -12, -120,
			-7, 17, 28, 8, 29, 44, 40, 11,
			0, 21, 36, 49, 52, 44, 48, 11,
			-10, 25, 45, 54, 56, 53, 45, 18,
			-14, 16, 37, 53, 51, 39, 30, 14,
			-18, 7, 25, 34, 31, 25, 8, -6,
			-38, -7, 5, 14, 17, 8, -14, -38,
			-70, -55, -32, -6, -26, -13, -49, -82,
		},
	},
	PieceValues: [2][7]Score{
		{0, 80, 385, 415, 504, 1034, 0},
		{0, 127, 324, 337, 642, 1269, 0},
	},
	TempoBonus: [2]Score{25, 22},
	KingAttackPieces: [2][4]Score{
		{7, 7, 9, 15},
		{-37, 19, 3, -81},
	},
	SafeChecks: [2][4]Score{
		{10, 9, 9, 5},
		{11, -1, 2, 7},
	},
	KingShelter: [2]Score{7, -1},
	MobilityKnight: [2][9]Score{
		{-60, -40, -28, -21, -14, -8, -1, 6, 10},
		{-42, -3, 18, 28, 36, 44, 42, 38, 27},
	},
	MobilityBishop: [2][14]Score{
		{-42, -31, -22, -19, -12, -4, 2, 5, 6, 8, 10, 13, 9, 22},
		{-28, -10, -5, 7, 20, 33, 36, 43, 50, 48, 46, 45, 53, 38},
	},
	MobilityRook: [2][11]Score{
		{-25, -19, -16, -12, -7, -2, 4, 11, 12, 16, 22},
		{-1, 4, 13, 17, 24, 27, 29, 31, 39, 44, 38},
	},
	KnightOutpost: [2][40]Score{
		{
			-39, 44, 6, 20, 50, 92, 88, -21,
			57, -9, -25, -53, 11, -38, -6, -18,
			-4, -3, 9, 34, 14, 48, 16, 14,
			-2, 26, 36, 36, 50, 53, 69, 12,
			0, 0, 0, 37, 50, 0, 0, 0,
		},
		{
			-88, 140, -4, 16, -38, 70, -43, 43,
			-48, 20, 33, 30, 7, 32, 14, 86,
			24, 13, 25, 25, 35, 31, 27, 32,
			22, 14, 16, 26, 29, 12, 1, 23,
			0, 0, 0, 22, 21, 0, 0, 0,
		},
	},
	ConnectedRooks:  [2]Score{1, 7},
	BishopPair:      [2]Score{-16, 71},
	ProtectedPasser: [2]Score{31, 18},
	PasserKingDist:  [2]Score{7, 9},
	PasserRank: [2][6]Score{
		{-24, -38, -32, -6, -4, 36},
		{4, 5, 30, 58, 129, 88},
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

	for pType := Pawn; pType <= Queen; pType++ {
		for color := White; color <= Black; color++ {
			cnt := (b.Pieces[pType] & b.Colors[color]).Count()

			phase += cnt * Phase[pType]

			mg[color] += T(cnt) * c.PieceValues[0][pType]
			eg[color] += T(cnt) * c.PieceValues[1][pType]
		}
	}

	for color := White; color <= Black; color++ {
		myBishops := b.Colors[color] & b.Pieces[Bishop]
		theirBishops := b.Colors[color.Flip()] & b.Pieces[Bishop]

		if myBishops != 0 && theirBishops == 0 && myBishops&(myBishops-1) != 0 {
			mg[color] += c.BishopPair[0]
			eg[color] += c.BishopPair[1]
		}
	}

	// add up PSqT
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
			}
		}
	}

	pWise := newPieceWise(b, c)

	// evaluate piece wise
	for pType := Knight; pType <= Queen; pType++ {
		for color := White; color <= Black; color++ {

			pieces := b.Pieces[pType] & b.Colors[color]
			for piece := board.BitBoard(0); pieces != 0; pieces ^= piece {
				piece = pieces & -pieces
				sq := piece.LowestSet()
				pWise.Eval(pType, color, sq, mg[:], eg[:])
			}
		}
	}

	for color := White; color <= Black; color++ {
		pWise.pawns(color, mg[:], eg[:])
	}

	pWise.calcCover()

	for pType := Knight; pType <= Queen; pType++ {
		for color := White; color <= Black; color++ {
			pWise.safeChecks(pType, color)
		}
	}

	for color := White; color <= Black; color++ {
		pCnt := (pWise.kingNb[color] & b.Colors[color] & b.Pieces[Pawn]).Count()
		penalty := T(max(3-pCnt, 0))

		pWise.kingAScore[0][color.Flip()] += c.KingShelter[0] * penalty
		pWise.kingAScore[1][color.Flip()] += c.KingShelter[1] * penalty
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

	default:
		return
	}

	if p.kingNb[color.Flip()]&attack != 0 {
		p.kingAScore[0][color] += p.c.KingAttackPieces[0][pType-Knight]
		p.kingAScore[1][color] += p.c.KingAttackPieces[1][pType-Knight]
	}
}

func (p *pieceWise[T]) pawns(color Color, mg, eg []T) {
	b := p.b
	passers := p.passers & b.Colors[color]

	// if there is a sole passer
	if passers != 0 && passers&(passers-1) == 0 {
		sq := passers.LowestSet()

		// KPR, KPNB
		if p.b.Pieces[Knight]|p.b.Pieces[Bishop]|p.b.Pieces[Queen] == 0 || p.b.Pieces[Rook]|p.b.Pieces[Queen] == 0 {
			qSq := sq % 8
			if color == White {
				qSq += 56
			}

			kingDist := Manhattan(qSq, p.kingSq[color.Flip()]) - Manhattan(qSq, p.kingSq[color])

			mg[color] += p.c.PasserKingDist[0] * T(kingDist)
			eg[color] += p.c.PasserKingDist[1] * T(kingDist)
		}
	}

	for passer := board.BitBoard(0); passers != 0; passers ^= passer {
		passer = passers & -passers
		sq := passer.LowestSet()

		rank := sq / 8
		if color == Black {
			rank ^= 7
		}

		// if protected passers add protection bonus
		if passer&p.attacks[color][0] != 0 {
			mg[color] += p.c.ProtectedPasser[0]
			eg[color] += p.c.ProtectedPasser[1]
		}

		mg[color] += p.c.PasserRank[0][rank-1]
		eg[color] += p.c.PasserRank[1][rank-1]
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
