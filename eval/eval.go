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
			41, 79, 54, 82, 73, 32, -41, -71,
			18, 36, 58, 61, 70, 99, 85, 35,
			-5, 18, 16, 21, 43, 35, 43, 16,
			-11, 9, 7, 24, 27, 20, 25, 5,
			-12, 5, 0, 5, 20, 10, 40, 10,
			-10, 6, -6, -2, 9, 29, 53, 5,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			0, 0, 0, 0, 0, 0, 0, 0,
			110, 97, 94, 36, 34, 61, 107, 119,
			24, 25, -17, -54, -58, -37, 3, 6,
			12, 3, -17, -35, -39, -28, -9, -11,
			-5, -3, -22, -28, -29, -25, -12, -22,
			-10, -6, -23, -18, -21, -23, -17, -24,
			-8, -5, -19, -12, -9, -21, -19, -26,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			-157, -111, -63, -33, 10, -67, -102, -106,
			-23, 1, 34, 46, 10, 82, -3, 18,
			-6, 27, 33, 33, 89, 64, 46, 13,
			1, 0, 12, 38, 18, 42, 11, 33,
			-12, -3, 9, 11, 22, 21, 21, 3,
			-34, -15, -6, 5, 22, 2, 7, -12,
			-40, -25, -20, 1, 1, 0, -8, -11,
			-73, -35, -35, -17, -17, -9, -29, -40,
		},
		{
			-41, -5, 5, 1, -5, -21, -8, -69,
			2, 8, -1, -1, -5, -18, 5, -21,
			2, 4, 21, 20, -1, -5, -9, -10,
			13, 21, 31, 26, 24, 25, 16, 2,
			23, 19, 37, 29, 34, 27, 16, 16,
			4, 10, 16, 31, 25, 9, 3, 8,
			7, 9, 9, 9, 10, 5, 3, 14,
			7, -6, 9, 12, 13, 1, -6, 0,
		},
		{
			-39, -61, -41, -94, -83, -87, -40, -62,
			-30, 1, -10, -29, 0, -14, -7, -43,
			-12, 9, 10, 26, 10, 54, 29, 17,
			-19, -8, 9, 23, 21, 13, -6, -22,
			-18, -14, -16, 15, 10, -11, -15, -3,
			-12, -5, -3, -10, -5, -1, 1, 4,
			-6, 0, 1, -13, -4, 10, 17, 2,
			-18, 1, -14, -21, -12, -18, 8, -4,
		},
		{
			14, 17, 7, 21, 16, 10, 6, 6,
			4, 6, 6, 7, 0, 4, 9, -1,
			16, 7, 10, -2, 4, 5, 3, 6,
			11, 15, 11, 22, 13, 13, 12, 12,
			6, 13, 19, 16, 15, 14, 9, -9,
			3, 15, 15, 14, 18, 13, 5, -4,
			9, -1, -5, 10, 5, 0, 4, -9,
			0, 8, 6, 10, 11, 13, -8, -11,
		},
		{
			26, 15, 20, 20, 33, 34, 34, 57,
			8, 3, 28, 46, 30, 55, 35, 71,
			-14, 17, 12, 17, 46, 41, 86, 55,
			-17, -7, -7, 5, 7, 7, 15, 18,
			-33, -32, -19, -7, -8, -27, -3, -14,
			-30, -28, -18, -15, -8, -12, 15, -3,
			-34, -23, -8, -4, -2, -1, 12, -21,
			-19, -14, -6, 3, 7, 0, 7, -20,
		},
		{
			21, 28, 33, 27, 23, 25, 20, 13,
			21, 34, 33, 22, 25, 17, 16, 0,
			23, 20, 21, 17, 7, 5, -4, -7,
			25, 21, 28, 21, 9, 7, 5, 1,
			19, 21, 20, 16, 14, 15, 1, 1,
			12, 11, 10, 11, 5, 0, -20, -16,
			12, 11, 9, 9, 3, -3, -11, 2,
			9, 8, 12, 4, -1, 1, -6, -3,
		},
		{
			-48, -28, -8, 16, 13, 23, 20, -27,
			-17, -30, -19, -26, -32, 4, -14, 12,
			-1, -2, -3, 6, 5, 48, 51, 52,
			-17, -8, -11, -16, -6, 3, 6, 0,
			-11, -18, -15, -2, -3, -6, -3, 2,
			-17, -6, -2, -8, 2, 2, 9, 1,
			-8, -4, 3, 13, 10, 18, 23, 35,
			-10, -15, -4, 5, 4, -15, 6, -11,
		},
		{
			40, 26, 41, 40, 41, 33, 1, 45,
			5, 28, 55, 66, 94, 61, 28, 35,
			8, 23, 53, 44, 59, 35, 3, -18,
			26, 33, 38, 54, 55, 41, 39, 29,
			3, 32, 36, 46, 42, 33, 20, 7,
			-5, 2, 21, 24, 25, 18, -10, -18,
			-19, -14, -15, -12, -5, -34, -72, -101,
			-16, -16, -10, -11, -18, -25, -49, -31,
		},
		{
			43, 39, 28, -25, -34, -6, -15, 93,
			-79, -25, -53, 41, 7, -20, -12, -53,
			-74, 19, -29, -40, -10, 50, 2, -32,
			-40, -49, -60, -69, -69, -57, -68, -113,
			-70, -61, -62, -79, -78, -60, -91, -131,
			-30, -15, -50, -56, -43, -50, -25, -56,
			41, 0, -14, -42, -45, -28, 15, 18,
			23, 57, 27, -71, -16, -42, 31, 29,
		},
		{
			-93, -48, -31, 3, -5, -8, -15, -104,
			-2, 25, 35, 16, 33, 48, 44, 11,
			6, 26, 41, 52, 56, 48, 50, 16,
			-6, 28, 46, 54, 57, 53, 46, 20,
			-10, 17, 36, 51, 49, 37, 30, 16,
			-20, 2, 22, 32, 29, 21, 5, -5,
			-36, -8, 4, 12, 15, 7, -13, -35,
			-67, -53, -32, -8, -28, -13, -46, -78,
		},
	},
	PieceValues: [2][7]Score{
		{0, 80, 376, 409, 495, 1044, 0},
		{0, 123, 315, 328, 614, 1186, 0},
	},
	TempoBonus: [2]Score{25, 22},
	KingAttackPieces: [2][4]Score{
		{5, 5, 6, 31},
		{8, 14, 11, -48},
	},
	SafeChecks: [2][4]Score{
		{11, 6, 9, 5},
		{3, 7, 10, 10},
	},
	MobilityKnight: [2][9]Score{
		{-59, -39, -27, -21, -14, -8, 0, 6, 10},
		{-44, -8, 12, 22, 29, 38, 36, 33, 24},
	},
	MobilityBishop: [2][14]Score{
		{-42, -32, -22, -19, -12, -5, 2, 5, 6, 8, 10, 12, 9, 28},
		{-34, -16, -11, 1, 14, 27, 30, 37, 43, 42, 40, 40, 47, 31},
	},
	MobilityRook: [2][11]Score{
		{-27, -20, -17, -12, -7, -2, 4, 11, 12, 16, 25},
		{-2, 3, 11, 14, 22, 26, 28, 29, 38, 43, 37},
	},
	KnightOutpost: [2][40]Score{
		{
			-41, 21, -41, 13, 43, 66, 51, -7,
			38, -7, -25, -34, 13, -37, -13, -25,
			-4, -4, 11, 34, 12, 40, 13, 12,
			1, 26, 36, 35, 50, 54, 74, 14,
			0, 0, 0, 36, 48, 0, 0, 0,
		},
		{
			-72, 128, -10, -16, -28, 49, -28, 44,
			-29, 16, 30, 10, 3, 23, 14, 85,
			25, 12, 23, 23, 32, 31, 22, 27,
			20, 13, 14, 24, 29, 9, 1, 20,
			0, 0, 0, 22, 22, 0, 0, 0,
		},
	},
	ConnectedRooks:  [2]Score{1, 7},
	BishopPair:      [2]Score{-16, 71},
	ProtectedPasser: [2]Score{30, 18},
	PasserKingDist:  [2]Score{9, 3},
	PasserRank: [2][6]Score{
		{-17, -35, -30, -5, -2, 39},
		{2, 4, 27, 54, 121, 85},
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
	result.calcAttacks()
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

		p.kingSq[color] = kingSq
		p.kingNb[color] = king | kingA
	}
}

func (p *pieceWise[T]) calcPawnBitBoards() {
	b := p.b

	wP := b.Pieces[Pawn] & b.Colors[White]
	bP := b.Pieces[Pawn] & b.Colors[Black]

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

func (p *pieceWise[T]) calcAttacks() {
	b := p.b

	for color := White; color <= Black; color++ {
		pawns := b.Pieces[Pawn] & b.Colors[color]
		p.attacks[color][0] = movegen.PawnCaptureMoves(pawns, color)
	}

	for color := White; color <= Black; color++ {
		accum := board.BitBoard(0)
		pieces := b.Pieces[Knight] & b.Colors[color]

		for piece := board.BitBoard(0); pieces != 0; pieces ^= piece {
			piece = pieces & -pieces
			sq := piece.LowestSet()

			accum |= movegen.KnightMoves(sq)
		}
		p.attacks[color][Knight-Pawn] = accum

		accum = board.BitBoard(0)
		pieces = b.Pieces[Bishop] & b.Colors[color]

		for piece := board.BitBoard(0); pieces != 0; pieces ^= piece {
			piece = pieces & -pieces
			sq := piece.LowestSet()

			accum |= movegen.BishopMoves(sq, p.occ)
		}
		p.attacks[color][Bishop-Pawn] = accum

		accum = board.BitBoard(0)
		pieces = b.Pieces[Rook] & b.Colors[color]

		for piece := board.BitBoard(0); pieces != 0; pieces ^= piece {
			piece = pieces & -pieces
			sq := piece.LowestSet()

			accum |= movegen.RookMoves(sq, p.occ)
		}
		p.attacks[color][Rook-Pawn] = accum

		accum = board.BitBoard(0)
		pieces = b.Pieces[Queen] & b.Colors[color]

		for piece := board.BitBoard(0); pieces != 0; pieces ^= piece {
			piece = pieces & -pieces
			sq := piece.LowestSet()

			accum |= movegen.RookMoves(sq, p.occ) | movegen.BishopMoves(sq, p.occ)
		}
		p.attacks[color][Queen-Pawn] = accum

		kingSq := p.kingSq[color]
		p.attacks[color][King-Pawn] = movegen.KingMoves(kingSq)

		p.cover[color] = p.attacks[color][0] |
			p.attacks[color][Knight-Pawn] |
			p.attacks[color][Bishop-Pawn] |
			p.attacks[color][Rook-Pawn] |
			p.attacks[color][Queen-Pawn] |
			p.attacks[color][King-Pawn]
	}
}

func (p *pieceWise[T]) Eval(pType Piece, color Color, sq Square, mg, eg []T) {

	occ := p.occ

	var (
		attack     board.BitBoard
		safeChecks board.BitBoard
	)

	switch pType {

	case Queen:
		attack = movegen.BishopMoves(sq, occ) | movegen.RookMoves(sq, occ)

		// calculate safe checks
		eKing := p.kingSq[color.Flip()]
		eKAttack := movegen.BishopMoves(eKing, occ) | movegen.RookMoves(eKing, occ)
		eCover := p.cover[color.Flip()]

		safeChecks = attack & eKAttack & ^eCover

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

		// calculate safe checks
		eKing := p.kingSq[color.Flip()]
		eKAttack := movegen.RookMoves(eKing, occ)
		eCover := p.cover[color.Flip()]

		safeChecks = attack & eKAttack & ^eCover

	case Bishop:
		attack = movegen.BishopMoves(sq, occ)

		mobCnt := (attack & ^p.b.Colors[color]).Count()
		mg[color] += p.c.MobilityBishop[0][mobCnt]
		eg[color] += p.c.MobilityBishop[1][mobCnt]

		// calculate safe checks
		eKing := p.kingSq[color.Flip()]
		eKAttack := movegen.BishopMoves(eKing, occ)
		eCover := p.cover[color.Flip()]

		safeChecks = attack & eKAttack & ^eCover

	case Knight:
		attack = movegen.KnightMoves(sq)

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

		// calculate safe checks
		eKing := p.kingSq[color.Flip()]
		eKAttack := movegen.KnightMoves(eKing)
		eCover := p.cover[color.Flip()]

		safeChecks = attack & eKAttack & ^eCover

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

	p.kingAScore[0][color] += p.c.SafeChecks[0][pType-Knight] * T(safeChecks.Count())
	p.kingAScore[1][color] += p.c.SafeChecks[1][pType-Knight] * T(safeChecks.Count())
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
