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
			45, 83, 57, 88, 82, 37, -41, -80,
			20, 37, 61, 63, 71, 100, 87, 34,
			-3, 20, 17, 23, 45, 36, 43, 14,
			-9, 10, 8, 26, 27, 20, 24, 3,
			-10, 7, 1, 6, 20, 9, 38, 8,
			-8, 8, -5, -4, 9, 25, 49, 1,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			0, 0, 0, 0, 0, 0, 0, 0,
			114, 102, 97, 38, 36, 65, 112, 126,
			25, 26, -17, -54, -59, -36, 4, 8,
			13, 3, -17, -36, -39, -29, -9, -11,
			-5, -3, -23, -29, -29, -25, -12, -22,
			-9, -6, -24, -18, -22, -24, -17, -24,
			-8, -5, -20, -13, -10, -22, -20, -27,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			-158, -112, -65, -34, 6, -73, -118, -113,
			-24, 0, 33, 48, 8, 72, -9, 8,
			-7, 28, 34, 34, 84, 60, 41, 9,
			1, 0, 13, 38, 21, 42, 11, 31,
			-11, -2, 9, 11, 22, 22, 23, 4,
			-34, -15, -6, 5, 22, 3, 9, -12,
			-40, -25, -19, 2, 1, 1, -7, -10,
			-72, -36, -36, -17, -17, -9, -29, -40,
		},
		{
			-44, -6, 4, -2, -8, -23, -8, -72,
			3, 7, -1, -3, -7, -18, 4, -21,
			2, 3, 21, 19, -3, -9, -13, -13,
			14, 21, 31, 25, 23, 22, 16, 0,
			24, 19, 39, 30, 35, 26, 15, 17,
			6, 12, 19, 33, 27, 11, 4, 9,
			8, 10, 10, 11, 12, 6, 4, 16,
			10, -4, 10, 13, 14, 2, -4, 1,
		},
		{
			-38, -59, -40, -91, -81, -87, -49, -60,
			-30, 2, -10, -31, 0, -18, -10, -54,
			-11, 9, 11, 27, 7, 51, 24, 16,
			-20, -7, 9, 25, 23, 10, -5, -22,
			-17, -13, -15, 16, 11, -10, -14, -2,
			-11, -3, -1, -9, -3, 0, 4, 4,
			-5, 1, 3, -12, -2, 11, 19, 4,
			-17, 2, -13, -20, -11, -17, 8, -3,
		},
		{
			15, 18, 8, 21, 15, 9, 8, 6,
			4, 7, 8, 8, 1, 6, 11, 3,
			17, 10, 11, 0, 7, 7, 4, 7,
			12, 16, 12, 24, 13, 16, 13, 12,
			6, 14, 20, 17, 17, 15, 10, -10,
			4, 16, 16, 15, 20, 15, 6, -2,
			10, 0, -5, 11, 6, 2, 6, -8,
			0, 9, 8, 11, 12, 14, -7, -10,
		},
		{
			28, 18, 25, 24, 35, 33, 28, 59,
			7, 3, 29, 46, 30, 46, 23, 54,
			-14, 17, 13, 17, 45, 39, 72, 44,
			-16, -6, -6, 7, 9, 7, 11, 17,
			-33, -31, -18, -6, -7, -27, -4, -14,
			-30, -28, -16, -14, -7, -11, 17, -3,
			-34, -23, -7, -3, -1, 0, 16, -21,
			-19, -14, -5, 3, 7, 0, 5, -20,
		},
		{
			21, 28, 33, 27, 24, 26, 22, 15,
			22, 35, 34, 24, 26, 20, 19, 4,
			23, 21, 22, 19, 9, 6, -1, -5,
			26, 23, 29, 22, 10, 7, 6, 2,
			21, 22, 21, 18, 15, 16, 1, 3,
			14, 12, 11, 13, 6, 1, -21, -15,
			14, 13, 11, 10, 4, -2, -12, 3,
			10, 9, 13, 6, 0, 2, -5, -1,
		},
		{
			-49, -28, -6, 15, 14, 24, 16, -27,
			-13, -28, -15, -20, -29, 12, -10, 20,
			-1, -2, -3, 10, 4, 53, 42, 52,
			-17, -9, -7, -11, -2, 9, 9, 4,
			-11, -17, -13, -1, 2, -2, 2, 6,
			-16, -6, -2, -6, 0, 3, 12, 3,
			-8, -4, 4, 12, 9, 17, 23, 34,
			-11, -15, -5, 3, 2, -16, 6, -15,
		},
		{
			43, 30, 48, 44, 49, 39, 3, 49,
			0, 27, 55, 65, 97, 61, 22, 29,
			8, 23, 54, 47, 65, 35, -1, -17,
			26, 34, 44, 57, 60, 42, 39, 30,
			4, 36, 39, 52, 47, 39, 23, 11,
			-3, 5, 25, 29, 34, 26, -4, -12,
			-19, -11, -12, -7, -3, -27, -62, -96,
			-14, -14, -8, -11, -17, -25, -47, -27,
		},
		{
			61, 80, 57, 5, -35, -18, -43, 96,
			-72, 5, -19, 75, 31, -4, -4, -62,
			-52, 43, 7, -7, 24, 82, 8, -22,
			-15, -24, -34, -41, -42, -38, -63, -110,
			-65, -52, -47, -69, -74, -65, -101, -132,
			-41, -37, -62, -60, -45, -72, -46, -68,
			37, -9, -17, -41, -46, -33, 6, 16,
			22, 54, 26, -58, -12, -40, 31, 30,
		},
		{
			-94, -52, -32, 3, -1, -5, -10, -109,
			-2, 23, 33, 14, 33, 48, 46, 13,
			4, 25, 39, 50, 54, 47, 51, 15,
			-9, 26, 44, 53, 56, 52, 45, 19,
			-12, 16, 36, 51, 50, 38, 30, 14,
			-19, 5, 24, 33, 31, 24, 7, -6,
			-38, -7, 5, 14, 17, 7, -14, -38,
			-68, -55, -32, -6, -26, -13, -49, -82,
		},
	},
	PieceValues: [2][7]Score{
		{0, 79, 381, 412, 499, 1030, 0},
		{0, 126, 320, 333, 630, 1241, 0},
	},
	TempoBonus: [2]Score{25, 22},
	KingAttackPieces: [2][4]Score{
		{7, 7, 8, 15},
		{-2, 19, 4, -81},
	},
	SafeChecks: [2][4]Score{
		{10, 9, 9, 5},
		{1, -1, 1, 7},
	},
	KingShelter: [2]Score{7, -1},
	MobilityKnight: [2][9]Score{
		{-59, -39, -27, -21, -14, -8, 0, 6, 10},
		{-43, -5, 15, 25, 32, 41, 39, 34, 24},
	},
	MobilityBishop: [2][14]Score{
		{-41, -31, -22, -19, -12, -4, 2, 5, 6, 8, 10, 12, 9, 25},
		{-30, -12, -8, 5, 17, 30, 33, 40, 47, 44, 43, 41, 49, 33},
	},
	MobilityRook: [2][11]Score{
		{-25, -18, -16, -11, -6, -1, 5, 11, 12, 16, 23},
		{-1, 4, 13, 17, 24, 27, 29, 31, 39, 43, 37},
	},
	KnightOutpost: [2][40]Score{
		{
			-44, 35, -15, 15, 47, 80, 68, -15,
			52, -8, -25, -46, 11, -38, -9, -23,
			-4, -3, 9, 34, 13, 47, 15, 13,
			-1, 26, 36, 36, 49, 53, 69, 13,
			0, 0, 0, 36, 49, 0, 0, 0,
		},
		{
			-82, 138, -4, -1, -34, 62, -38, 43,
			-41, 18, 32, 24, 7, 30, 14, 90,
			25, 12, 24, 25, 36, 31, 28, 31,
			21, 14, 15, 26, 29, 12, 0, 22,
			0, 0, 0, 22, 21, 0, 0, 0,
		},
	},
	ConnectedRooks:  [2]Score{1, 7},
	BishopPair:      [2]Score{-15, 69},
	ProtectedPasser: [2]Score{31, 19},
	PasserKingDist:  [2]Score{10, 3},
	PasserRank: [2][6]Score{
		{-23, -38, -30, -5, -2, 38},
		{2, 4, 27, 55, 124, 88},
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
			// mid square between the pawn and its queening square
			mSq := (qSq + sq) / 2

			kingDist := Manhattan(mSq, p.kingSq[color.Flip()]) - Manhattan(mSq, p.kingSq[color])

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
