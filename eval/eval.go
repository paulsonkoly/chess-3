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
	// IsolatedPawns is the penalty per isolated pawns.
	IsolatedPawns [2]T
}

var Coefficients = CoeffSet[Score]{
	PSqT: [12][64]Score{
		{
			0, 0, 0, 0, 0, 0, 0, 0,
			63, 89, 62, 95, 81, 42, -39, -67,
			26, 32, 65, 65, 74, 108, 86, 40,
			1, 11, 21, 26, 48, 43, 35, 18,
			-7, -2, 11, 28, 29, 24, 14, 5,
			-9, -4, 5, 7, 22, 15, 29, 9,
			-8, -5, -4, -4, 9, 28, 38, 1,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			0, 0, 0, 0, 0, 0, 0, 0,
			122, 107, 108, 46, 49, 75, 120, 135,
			28, 29, -13, -54, -58, -34, 4, 9,
			12, 1, -14, -33, -35, -26, -12, -12,
			-6, -8, -20, -27, -28, -23, -18, -22,
			-12, -11, -18, -17, -18, -19, -24, -26,
			-10, -12, -14, -12, -7, -19, -28, -27,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			-164, -114, -70, -38, -1, -80, -134, -123,
			-26, -5, 31, 50, 8, 70, -14, 8,
			-8, 28, 35, 34, 85, 61, 43, 11,
			2, 0, 14, 40, 22, 44, 12, 32,
			-11, -2, 11, 13, 24, 23, 25, 5,
			-34, -14, -5, 7, 24, 5, 10, -11,
			-39, -25, -18, 2, 2, 2, -7, -10,
			-71, -37, -34, -16, -16, -8, -30, -39,
		},
		{
			-48, -6, 6, -2, -5, -23, -5, -72,
			3, 8, -1, -4, -8, -19, 3, -23,
			3, 4, 22, 20, -2, -10, -14, -14,
			15, 21, 31, 26, 24, 21, 16, 0,
			23, 20, 40, 32, 36, 27, 15, 18,
			7, 13, 19, 34, 28, 12, 4, 11,
			9, 10, 13, 12, 12, 7, 5, 17,
			9, -6, 11, 13, 15, 2, -4, 4,
		},
		{
			-42, -61, -43, -91, -83, -93, -53, -64,
			-30, -2, -11, -30, 1, -19, -16, -51,
			-9, 9, 12, 28, 8, 53, 26, 19,
			-19, -5, 10, 26, 24, 11, -3, -21,
			-15, -12, -13, 18, 14, -8, -12, 0,
			-10, 0, 0, -6, -2, 2, 6, 5,
			-2, 2, 6, -11, 0, 12, 20, 6,
			-16, 6, -12, -18, -10, -15, 9, -1,
		},
		{
			17, 20, 9, 21, 17, 11, 11, 7,
			5, 8, 9, 10, 1, 8, 11, 3,
			18, 11, 12, 2, 8, 9, 5, 8,
			12, 17, 13, 25, 15, 17, 13, 13,
			8, 15, 22, 20, 18, 17, 11, -10,
			5, 17, 18, 19, 23, 16, 7, -2,
			11, 0, -3, 12, 8, 3, 6, -8,
			0, 10, 10, 12, 12, 16, -6, -11,
		},
		{
			25, 17, 25, 22, 33, 29, 23, 56,
			6, -1, 26, 44, 27, 44, 18, 50,
			-11, 16, 12, 15, 44, 40, 73, 45,
			-15, -9, -7, 6, 7, 6, 12, 15,
			-34, -32, -18, -8, -8, -27, -2, -15,
			-31, -27, -17, -14, -6, -10, 18, -3,
			-35, -23, -7, -3, 0, 1, 18, -23,
			-20, -14, -6, 3, 8, 0, 6, -21,
		},
		{
			24, 31, 35, 30, 26, 28, 25, 18,
			24, 38, 37, 26, 29, 22, 22, 7,
			24, 23, 23, 20, 10, 6, 0, -3,
			28, 26, 31, 23, 11, 9, 8, 6,
			24, 24, 23, 19, 16, 18, 2, 5,
			16, 13, 12, 13, 7, 2, -19, -14,
			14, 14, 12, 11, 5, -2, -12, 3,
			12, 12, 15, 7, 1, 4, -2, 0,
		},
		{
			-50, -28, -10, 16, 16, 26, 16, -28,
			-12, -32, -16, -19, -30, 12, -11, 26,
			0, -3, -3, 11, 6, 57, 45, 56,
			-17, -7, -5, -9, 0, 10, 11, 6,
			-9, -15, -11, 1, 3, 0, 4, 8,
			-14, -3, 0, -3, 3, 5, 15, 5,
			-7, -2, 7, 14, 11, 20, 25, 35,
			-10, -13, -3, 5, 4, -13, 8, -14,
		},
		{
			47, 36, 53, 48, 51, 42, 9, 54,
			3, 32, 59, 69, 102, 62, 23, 27,
			12, 29, 57, 50, 68, 35, 1, -14,
			30, 38, 44, 60, 62, 43, 42, 32,
			6, 37, 43, 55, 51, 42, 24, 14,
			-2, 7, 29, 32, 36, 29, -3, -10,
			-19, -10, -10, -4, 0, -25, -62, -94,
			-12, -11, -5, -9, -15, -24, -45, -24,
		},
		{
			88, 133, 95, 41, -21, -18, -60, 125,
			-52, 36, 14, 116, 59, 20, 16, -55,
			-31, 64, 28, 9, 47, 101, 24, -6,
			-3, -11, -26, -35, -34, -33, -60, -108,
			-58, -47, -44, -69, -76, -67, -105, -135,
			-46, -36, -62, -61, -48, -74, -49, -71,
			36, -9, -18, -42, -49, -34, 3, 14,
			23, 54, 25, -59, -14, -42, 30, 30,
		},
		{
			-108, -68, -42, -7, -5, -4, -11, -121,
			-8, 16, 26, 6, 29, 43, 40, 10,
			-2, 21, 35, 48, 52, 44, 48, 11,
			-11, 24, 44, 53, 56, 52, 45, 18,
			-14, 15, 35, 52, 51, 39, 31, 14,
			-20, 5, 24, 34, 32, 26, 8, -7,
			-40, -9, 5, 13, 17, 7, -14, -38,
			-71, -57, -35, -7, -27, -14, -50, -83,
		},
	},
	PieceValues: [2][7]Score{
		{0, 78, 389, 421, 512, 1044, 0},
		{0, 126, 326, 342, 648, 1287, 0},
	},
	TempoBonus: [2]Score{25, 23},
	KingAttackPieces: [2][4]Score{
		{7, 7, 8, 15},
		{-43, 19, 4, -81},
	},
	SafeChecks: [2][4]Score{
		{10, 9, 9, 5},
		{11, -1, 2, 7},
	},
	KingShelter: [2]Score{7, -1},
	MobilityKnight: [2][9]Score{
		{-59, -39, -27, -21, -14, -8, 0, 6, 9},
		{-39, 0, 20, 30, 38, 46, 44, 40, 29},
	},
	MobilityBishop: [2][14]Score{
		{-41, -31, -21, -18, -11, -4, 3, 6, 7, 8, 11, 12, 8, 21},
		{-24, -7, -2, 9, 23, 35, 38, 44, 51, 48, 47, 45, 52, 38},
	},
	MobilityRook: [2][11]Score{
		{-26, -19, -17, -12, -8, -1, 7, 14, 13, 16, 23},
		{-5, 2, 12, 17, 25, 28, 30, 32, 41, 45, 39},
	},
	KnightOutpost: [2][40]Score{
		{
			-33, 50, 23, 20, 55, 97, 109, -27,
			60, -9, -24, -56, 3, -40, -3, -6,
			-1, -3, 10, 34, 14, 49, 16, 13,
			0, 25, 38, 36, 49, 52, 68, 16,
			0, 0, 0, 32, 48, 0, 0, 0,
		},
		{
			-91, 140, -9, 23, -38, 71, -46, 43,
			-53, 24, 31, 33, 8, 36, 17, 82,
			25, 14, 28, 26, 36, 34, 30, 36,
			23, 13, 17, 27, 28, 15, -1, 23,
			0, 0, 0, 20, 17, 0, 0, 0,
		},
	},
	ConnectedRooks:  [2]Score{1, 7},
	BishopPair:      [2]Score{-16, 71},
	ProtectedPasser: [2]Score{22, 5},
	PasserKingDist:  [2]Score{7, 8},
	PasserRank: [2][6]Score{
		{-14, -26, -20, 4, 5, 40},
		{11, 15, 41, 68, 138, 95},
	},
	DoubledPawns:  [2]Score{-8, -17},
	IsolatedPawns: [2]Score{-15, -11},
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
	rearSpan := [...]board.BitBoard{frontFill(ps[White], Black) >> 8, frontFill(ps[Black], White) << 8}

	// calculate holes in our position, squares that cannot be protected by one
	// of our pawns.
	cover := [...]board.BitBoard{
		((frontSpan[White] & ^board.AFile) >> 1) | ((frontSpan[White] & ^board.HFile) << 1),
		((frontSpan[Black] & ^board.HFile) << 1) | ((frontSpan[Black] & ^board.AFile) >> 1),
	}
	p.holes[White] = sideOfBoard[White] & ^cover[White]
	p.holes[Black] = sideOfBoard[Black] & ^cover[Black]

	// neighbour files, files adjacent to files with pawns
	wFiles := ps[White] | frontSpan[White] | rearSpan[White]
	bFiles := ps[Black] | frontSpan[Black] | rearSpan[Black]
	neighbourF := [...]board.BitBoard{
		((wFiles & ^board.AFile) >> 1) | ((wFiles & ^board.HFile) << 1),
		((bFiles & ^board.HFile) << 1) | ((bFiles & ^board.AFile) >> 1),
	}

	frontLine := [...]board.BitBoard{^rearSpan[White] & ps[White], ^rearSpan[Black] & ps[Black]}

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

		// isolated pawns
		mg[color] += p.c.IsolatedPawns[0] * T((ps[color] &^ neighbourF[color]).Count())
		eg[color] += p.c.IsolatedPawns[1] * T((ps[color] &^ neighbourF[color]).Count())
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

	fifty := b.FiftyCnt

	mgScore := mg[b.STM] - mg[b.STM.Flip()]
	egScore := eg[b.STM] - eg[b.STM.Flip()]

	mgPhase := min(phase, 24)
	egPhase := 24 - mgPhase

	if _, ok := (any(mgScore)).(Score); ok {
		v := int(mgScore)*mgPhase + int(egScore)*egPhase
		v *= int(100 - fifty)

		return T(v / 2400)
	}

	// The training set has all '0's for the halfmove counter. If that wasn't the
	// case we should blend the fifty counter in like in real score.
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
