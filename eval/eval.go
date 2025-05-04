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
	// DoubledPawns is the penalty per doubled pawn (count of non-frontline pawns ie. the pawns in the pawn rearspan).
	DoubledPawns [2]T
}

var Coefficients = CoeffSet[Score]{
	PSqT: [12][64]Score{
		{
			0, 0, 0, 0, 0, 0, 0, 0,
			50, 90, 60, 88, 79, 42, -29, -77,
			21, 39, 65, 63, 74, 107, 93, 34,
			-3, 20, 19, 24, 46, 42, 44, 12,
			-11, 9, 9, 25, 26, 23, 24, 0,
			-12, 6, 3, 5, 21, 14, 39, 4,
			-11, 6, -5, -5, 8, 27, 48, -3,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			0, 0, 0, 0, 0, 0, 0, 0,
			113, 103, 101, 41, 45, 71, 116, 128,
			24, 29, -14, -53, -58, -33, 7, 7,
			10, 4, -15, -33, -36, -25, -7, -13,
			-7, -3, -21, -29, -29, -22, -12, -24,
			-13, -6, -20, -16, -19, -19, -17, -27,
			-11, -5, -15, -12, -8, -18, -20, -29,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			-161, -112, -69, -37, 0, -78, -130, -121,
			-24, -3, 31, 49, 7, 70, -10, 9,
			-8, 28, 34, 33, 84, 60, 41, 11,
			1, 0, 14, 39, 21, 43, 11, 31,
			-11, -3, 10, 12, 23, 23, 24, 4,
			-34, -15, -5, 6, 23, 4, 9, -11,
			-39, -25, -19, 2, 2, 2, -7, -10,
			-71, -36, -35, -17, -16, -8, -30, -39,
		},
		{
			-47, -6, 7, -2, -5, -22, -6, -71,
			3, 9, -1, -4, -7, -19, 3, -21,
			3, 4, 21, 20, -3, -10, -13, -13,
			15, 21, 31, 25, 23, 21, 16, 1,
			24, 20, 39, 31, 36, 27, 15, 18,
			6, 12, 18, 33, 27, 11, 4, 10,
			9, 11, 12, 12, 12, 7, 5, 18,
			10, -6, 10, 14, 15, 3, -4, 5,
		},
		{
			-40, -61, -42, -92, -83, -91, -52, -62,
			-30, 0, -11, -31, 0, -19, -12, -52,
			-10, 9, 11, 26, 6, 52, 25, 17,
			-19, -6, 10, 25, 23, 11, -4, -21,
			-15, -13, -13, 17, 13, -9, -13, 0,
			-10, -1, 0, -7, -3, 2, 5, 5,
			-3, 2, 5, -11, -1, 11, 20, 5,
			-16, 4, -13, -18, -11, -16, 9, -1,
		},
		{
			16, 20, 9, 21, 16, 11, 10, 7,
			5, 7, 8, 9, 1, 6, 11, 4,
			18, 12, 11, 1, 8, 9, 6, 9,
			12, 17, 13, 24, 14, 16, 13, 12,
			7, 16, 22, 19, 18, 17, 11, -9,
			6, 17, 18, 18, 22, 15, 7, -2,
			11, 0, -4, 12, 7, 3, 5, -7,
			-1, 10, 10, 12, 13, 16, -5, -10,
		},
		{
			27, 17, 24, 22, 33, 30, 24, 57,
			6, 1, 26, 44, 28, 44, 19, 51,
			-14, 16, 12, 14, 42, 41, 72, 44,
			-16, -7, -7, 5, 7, 8, 13, 15,
			-34, -31, -18, -8, -8, -26, -2, -15,
			-30, -27, -16, -14, -6, -10, 18, -3,
			-34, -23, -7, -3, 0, 1, 17, -22,
			-20, -14, -5, 3, 8, 0, 6, -21,
		},
		{
			23, 30, 35, 29, 25, 28, 24, 17,
			24, 36, 36, 25, 28, 22, 21, 6,
			24, 23, 23, 20, 10, 6, 0, -4,
			27, 24, 30, 22, 10, 8, 7, 5,
			22, 23, 22, 18, 15, 16, 2, 4,
			15, 13, 12, 13, 6, 2, -19, -14,
			14, 14, 12, 10, 4, -2, -12, 3,
			12, 11, 14, 6, 1, 3, -2, 0,
		},
		{
			-49, -29, -9, 15, 16, 25, 14, -28,
			-12, -29, -16, -19, -30, 11, -9, 24,
			-1, -3, -2, 10, 4, 56, 43, 54,
			-17, -8, -6, -10, -1, 10, 10, 6,
			-10, -16, -10, 0, 3, -1, 3, 7,
			-14, -4, 0, -4, 2, 4, 14, 4,
			-7, -3, 6, 13, 10, 19, 24, 34,
			-10, -13, -3, 4, 3, -14, 7, -14,
		},
		{
			46, 35, 52, 47, 51, 41, 7, 52,
			3, 29, 58, 66, 100, 61, 22, 29,
			12, 28, 55, 48, 67, 35, 1, -16,
			29, 36, 44, 59, 61, 42, 41, 31,
			5, 38, 40, 54, 49, 40, 24, 13,
			-1, 6, 27, 31, 36, 28, -3, -11,
			-18, -10, -10, -6, -1, -25, -61, -94,
			-12, -11, -6, -9, -15, -23, -46, -24,
		},
		{
			83, 125, 91, 36, -24, -21, -58, 117,
			-56, 34, 13, 112, 56, 16, 12, -57,
			-33, 62, 28, 7, 45, 99, 20, -7,
			-4, -12, -26, -35, -34, -34, -60, -108,
			-58, -48, -45, -70, -75, -67, -105, -134,
			-45, -36, -61, -60, -48, -73, -48, -70,
			35, -10, -17, -41, -47, -33, 4, 14,
			22, 53, 25, -59, -13, -41, 30, 29,
		},
		{
			-107, -66, -41, -5, -5, -4, -11, -122,
			-8, 16, 26, 7, 28, 42, 39, 10,
			-1, 21, 34, 48, 51, 43, 48, 11,
			-10, 24, 44, 53, 55, 52, 45, 18,
			-14, 15, 36, 52, 50, 39, 31, 14,
			-19, 6, 25, 34, 32, 26, 8, -6,
			-38, -8, 5, 14, 17, 8, -14, -37,
			-70, -56, -33, -7, -26, -14, -49, -81,
		},
	},
	PieceValues: [2][7]Score{
		{0, 83, 385, 417, 505, 1029, 0},
		{0, 132, 324, 338, 641, 1273, 0},
	},
	TempoBonus: [2]Score{25, 22},
	KingAttackPieces: [2][4]Score{
		{7, 7, 8, 15},
		{-42, 19, 4, -81},
	},
	SafeChecks: [2][4]Score{
		{10, 9, 9, 5},
		{11, -1, 2, 7},
	},
	KingShelter: [2]Score{7, -1},
	MobilityKnight: [2][9]Score{
		{-60, -40, -28, -22, -15, -8, -1, 6, 9},
		{-41, -3, 18, 28, 36, 45, 43, 38, 28},
	},
	MobilityBishop: [2][14]Score{
		{-42, -32, -22, -19, -12, -4, 2, 5, 6, 8, 10, 12, 7, 22},
		{-28, -8, -4, 8, 21, 34, 37, 43, 50, 48, 46, 45, 52, 37},
	},
	MobilityRook: [2][11]Score{
		{-26, -19, -17, -12, -8, -1, 7, 13, 13, 16, 22},
		{-5, 2, 11, 16, 24, 27, 29, 32, 40, 45, 39},
	},
	KnightOutpost: [2][40]Score{
		{
			-36, 48, 15, 20, 53, 95, 99, -24,
			58, -11, -25, -56, 4, -40, -6, -11,
			-2, -3, 8, 33, 14, 47, 16, 13,
			-3, 26, 36, 35, 48, 51, 69, 13,
			0, 0, 0, 35, 49, 0, 0, 0,
		},
		{
			-91, 140, -7, 20, -37, 70, -44, 43,
			-51, 21, 31, 30, 7, 33, 15, 85,
			25, 14, 27, 26, 35, 34, 27, 33,
			23, 15, 16, 26, 27, 12, 0, 23,
			0, 0, 0, 20, 18, 0, 0, 0,
		},
	},
	ConnectedRooks:  [2]Score{1, 7},
	BishopPair:      [2]Score{-16, 71},
	ProtectedPasser: [2]Score{30, 17},
	PasserKingDist:  [2]Score{7, 9},
	PasserRank: [2][6]Score{
		{-25, -40, -32, -5, -2, 37},
		{-2, -1, 26, 54, 124, 85},
	},
	DoubledPawns: [2]Score{-20, -33},
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

	pWise.pawns(mg[:], eg[:])

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

func (p *pieceWise[T]) pawns(mg, eg []T) {
	b := p.b

	ps := [...]board.BitBoard{b.Pieces[Pawn] & b.Colors[White], b.Pieces[Pawn] & b.Colors[Black]}

	p.attacks[White][0] = movegen.PawnCaptureMoves(ps[White], White)
	p.attacks[Black][0] = movegen.PawnCaptureMoves(ps[Black], Black)

	// various useful pawn bitboards
	frontSpan := [...]board.BitBoard{frontFill(ps[White], White) << 8, frontFill(ps[Black], Black) >> 8}
	wRearSpan := frontFill(ps[White], Black) >> 8
	bRearSpan := frontFill(ps[Black], White) << 8

	// calculate holes in our position, squares that cannot be protected by one
	// of our pawns.
	cover := [...]board.BitBoard{
		((frontSpan[White] & ^board.AFile) >> 1) | ((frontSpan[White] & ^board.HFile) << 1),
		((frontSpan[Black] & ^board.HFile) << 1) | ((frontSpan[Black] & ^board.AFile) >> 1),
	}
	p.holes[White] = sideOfBoard[White] & ^cover[White]
	p.holes[Black] = sideOfBoard[Black] & ^cover[Black]

	frontLine := [...]board.BitBoard{^wRearSpan & ps[White], ^bRearSpan & ps[Black]}

	for color := White; color <= Black; color++ {
		passers := frontLine[color] & ^(frontSpan[color.Flip()] | cover[color.Flip()])

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

		// doubled pawns
		mg[color] += p.c.DoubledPawns[0] * T((ps[color] &^ frontLine[color]).Count())
		eg[color] += p.c.DoubledPawns[1] * T((ps[color] &^ frontLine[color]).Count())
	}
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
	attacks    [2][6]board.BitBoard
	cover      [2]board.BitBoard
	kingNb     [2]board.BitBoard
	kingRays   [2][2]board.BitBoard
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
