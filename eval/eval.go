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

func Eval[T ScoreType](b *board.Board, c *CoeffSet[T]) T {
	if insuffientMat(b) {
		return 0
	}

	sp := scorePair[T]{}

	sp.addPieceValues(b, c)
	sp.addTempo(b, c)
	sp.addBishopPair(b, c)

	pw := pieceWise[T]{}

	pw.calcOccupancy(b)
	pw.calcKingSquares(b)
	pw.calcPawnStructure(b)

	sp.addPassers(b, pw, c)
	sp.addDoubledPawns(pw, c)
	sp.addIsolatedPawns(pw, c)

	ka := kingAttacks[T]{}

	for color := White; color <= Black; color++ {

		// enemy king neighbourhood
		eKNb := pw.kingNb[color.Flip()]

		// queens
		pieces := b.Pieces[Queen] & b.Colors[color]
		for piece := board.BitBoard(0); pieces != 0; pieces ^= piece {
			piece = pieces & -pieces
			sq := piece.LowestSet()

			attacks := pw.calcQueenAttacks(color, sq)

			ka.addAttackPieces(color, Queen, attacks, eKNb, c)

			sp.addPSqT(color, Queen, sq, c)
		}

		// rooks
		pieces = b.Pieces[Rook] & b.Colors[color]
		for piece := board.BitBoard(0); pieces != 0; pieces ^= piece {
			piece = pieces & -pieces
			sq := piece.LowestSet()

			attacks := pw.calcRookAttacks(color, sq)

			ka.addAttackPieces(color, Rook, attacks, eKNb, c)
			sp.addRookMobility(b, color, sq, attacks, c)
			sp.addPSqT(color, Rook, sq, c)
		}

		// bishops
		pieces = b.Pieces[Bishop] & b.Colors[color]
		for piece := board.BitBoard(0); pieces != 0; pieces ^= piece {
			piece = pieces & -pieces
			sq := piece.LowestSet()

			attacks := pw.calcBishopAttacks(color, sq)

			ka.addAttackPieces(color, Bishop, attacks, eKNb, c)
			sp.addBishopMobility(b, color, attacks, c)
			sp.addPSqT(color, Bishop, sq, c)
		}

		// knights
		pieces = b.Pieces[Knight] & b.Colors[color]
		for piece := board.BitBoard(0); pieces != 0; pieces ^= piece {
			piece = pieces & -pieces
			sq := piece.LowestSet()

			attacks := pw.calcKnightAttacks(color, sq)

			ka.addAttackPieces(color, Knight, attacks, eKNb, c)
			sp.addKnightMobility(b, color, attacks, pw.attacks[color.Flip()][0], c)
			sp.addKnightOutposts(color, piece, sq, pw.holes[color.Flip()]&pw.attacks[color][0], c)
			sp.addPSqT(color, Knight, sq, c)
		}

		// pawns
		pieces = b.Pieces[Pawn] & b.Colors[color]
		for piece := board.BitBoard(0); pieces != 0; pieces ^= piece {
			piece = pieces & -pieces
			sq := piece.LowestSet()

			sp.addPSqT(color, Pawn, sq, c)
		}

		// king
		piece := b.Pieces[King] & b.Colors[color]
		sq := piece.LowestSet()

		sp.addPSqT(color, King, sq, c)
	}

	pw.calcCover()

	// safe checks
	for color := White; color <= Black; color++ {
		eCover := pw.cover[color.Flip()]

		var safeChecks board.BitBoard

		// Queen
		eKAttack := pw.kingRays[color.Flip()][0] | pw.kingRays[color.Flip()][Rook-Bishop]
		safeChecks = pw.attacks[color][Queen-Pawn] & eKAttack & ^eCover & ^b.Colors[color]

		ka.addSafeChecks(color, Queen, safeChecks, c)

		// Rook
		eKAttack = pw.kingRays[color.Flip()][Rook-Bishop]
		safeChecks = pw.attacks[color][Rook-Pawn] & eKAttack & ^eCover & ^b.Colors[color]

		ka.addSafeChecks(color, Rook, safeChecks, c)

		// Bishop
		eKAttack = pw.kingRays[color.Flip()][0]
		safeChecks = pw.attacks[color][Bishop-Pawn] & eKAttack & ^eCover & ^b.Colors[color]

		ka.addSafeChecks(color, Bishop, safeChecks, c)

		// Knight
		eKAttack = movegen.KnightMoves(pw.kingSq[color.Flip()])
		safeChecks = pw.attacks[color][Knight-Pawn] & eKAttack & ^eCover & ^b.Colors[color]

		ka.addSafeChecks(color, Knight, safeChecks, c)

		// shelter
		pCnt := (pw.kingNb[color] & b.Colors[color] & b.Pieces[Pawn]).Count()
		penalty := T(max(3-pCnt, 0))

		ka.addShelter(color, penalty, c)
	}

	sp.addKingAttacks(ka)

	return sp.taperedScore(b)
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

// Phase is game phase.
var Phase = [...]int{0, 0, 1, 1, 2, 4, 0}

type scorePair[T ScoreType] struct {
	mg [2]T
	eg [2]T

	phase int
}

func (sp *scorePair[T]) addPieceValues(b *board.Board, c *CoeffSet[T]) {
	for pType := Pawn; pType <= Queen; pType++ {
		for color := White; color <= Black; color++ {
			cnt := (b.Pieces[pType] & b.Colors[color]).Count()

			sp.phase += cnt * Phase[pType]

			sp.mg[color] += T(cnt) * c.PieceValues[0][pType]
			sp.eg[color] += T(cnt) * c.PieceValues[1][pType]
		}
	}
}

func (sp *scorePair[T]) addTempo(b *board.Board, c *CoeffSet[T]) {
	sp.mg[b.STM] += c.TempoBonus[0]
	sp.eg[b.STM] += c.TempoBonus[1]
}

func (sp *scorePair[T]) addBishopPair(b *board.Board, c *CoeffSet[T]) {
	for color := White; color <= Black; color++ {
		myBishops := b.Colors[color] & b.Pieces[Bishop]
		theirBishops := b.Colors[color.Flip()] & b.Pieces[Bishop]

		if myBishops != 0 && theirBishops == 0 && myBishops&(myBishops-1) != 0 {
			sp.mg[color] += c.BishopPair[0]
			sp.eg[color] += c.BishopPair[1]
		}
	}
}

func (sp *scorePair[T]) addPSqT(color Color, pType Piece, sq Square, c *CoeffSet[T]) {
	// add up PSqT

	sqIx := sq

	if color == White {
		sqIx ^= 56 // upside down
	}

	ix := pType - 1

	sp.mg[color] += c.PSqT[2*ix][sqIx]
	sp.eg[color] += c.PSqT[2*ix+1][sqIx]
}

func (sp *scorePair[T]) taperedScore(b *board.Board) T {
	fifty := b.FiftyCnt

	mgScore := sp.mg[b.STM] - sp.mg[b.STM.Flip()]
	egScore := sp.eg[b.STM] - sp.eg[b.STM.Flip()]

	mgPhase := min(sp.phase, 24)
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
	occ           board.BitBoard
	attacks       [2][6]board.BitBoard
	kingRays      [2][2]board.BitBoard
	kingSq        [2]Square
	kingNb        [2]board.BitBoard
	holes         [2]board.BitBoard
	passers       [2]board.BitBoard
	doubledPawns  [2]board.BitBoard
	isolatedPawns [2]board.BitBoard
	cover         [2]board.BitBoard
}

func (pw *pieceWise[T]) calcOccupancy(b *board.Board) {
	pw.occ = b.Colors[White] | b.Colors[Black]
}

func (pw *pieceWise[T]) calcKingSquares(b *board.Board) {
	for color := White; color <= Black; color++ {
		king := b.Colors[color] & b.Pieces[King]
		kingSq := king.LowestSet()
		kingA := movegen.KingMoves(kingSq)

		pw.attacks[color][King-Pawn] = kingA
		pw.kingRays[color][0] = movegen.BishopMoves(kingSq, pw.occ)
		pw.kingRays[color][Rook-Bishop] = movegen.RookMoves(kingSq, pw.occ)
		pw.kingSq[color] = kingSq
		pw.kingNb[color] = king | kingA
	}
}

// the player's side of the board with the extra 2 central squares included at
// enemy side.
var sideOfBoard = [2]board.BitBoard{0x00000018_ffffffff, 0xffffffff_18000000}

func (pw *pieceWise[T]) calcPawnStructure(b *board.Board) {

	ps := [...]board.BitBoard{b.Pieces[Pawn] & b.Colors[White], b.Pieces[Pawn] & b.Colors[Black]}

	pw.attacks[White][0] = movegen.PawnCaptureMoves(ps[White], White)
	pw.attacks[Black][0] = movegen.PawnCaptureMoves(ps[Black], Black)

	// various useful pawn bitboards
	frontSpan := [...]board.BitBoard{frontFill(ps[White], White) << 8, frontFill(ps[Black], Black) >> 8}
	rearSpan := [...]board.BitBoard{frontFill(ps[White], Black) >> 8, frontFill(ps[Black], White) << 8}

	// calculate holes in our position, squares that cannot be protected by one
	// of our pawns.
	cover := [...]board.BitBoard{
		((frontSpan[White] & ^board.AFile) >> 1) | ((frontSpan[White] & ^board.HFile) << 1),
		((frontSpan[Black] & ^board.HFile) << 1) | ((frontSpan[Black] & ^board.AFile) >> 1),
	}
	pw.holes[White] = sideOfBoard[White] & ^cover[White]
	pw.holes[Black] = sideOfBoard[Black] & ^cover[Black]

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

		pw.passers[color] = passers
		pw.doubledPawns[color] = ps[color] &^ frontLine[color]
		pw.isolatedPawns[color] = ps[color] &^ neighbourF[color]
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

func (sp *scorePair[T]) addPassers(b *board.Board, pw pieceWise[T], c *CoeffSet[T]) {

	for color := White; color <= Black; color++ {

		passers := pw.passers[color]

		// if there is a sole passer
		if passers != 0 && passers&(passers-1) == 0 {
			sq := passers.LowestSet()

			// KPR, KPNB
			if b.Pieces[Knight]|b.Pieces[Bishop]|b.Pieces[Queen] == 0 || b.Pieces[Rook]|b.Pieces[Queen] == 0 {
				qSq := sq % 8
				if color == White {
					qSq += 56
				}

				kingDist := Chebishev(qSq, pw.kingSq[color.Flip()]) - Chebishev(qSq, pw.kingSq[color])

				sp.mg[color] += c.PasserKingDist[0] * T(kingDist)
				sp.eg[color] += c.PasserKingDist[1] * T(kingDist)
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
			if passer&pw.attacks[color][0] != 0 { // Pawn - Pawn
				sp.mg[color] += c.ProtectedPasser[0]
				sp.eg[color] += c.ProtectedPasser[1]
			}

			sp.mg[color] += c.PasserRank[0][rank-1]
			sp.eg[color] += c.PasserRank[1][rank-1]
		}
	}
}

func Chebishev(a, b Square) int {
	ax, ay, bx, by := int(a%8), int(a/8), int(b%8), int(b/8)
	return max(Abs(ax-bx), Abs(ay-by))
}

func (sp *scorePair[T]) addDoubledPawns(pw pieceWise[T], c *CoeffSet[T]) {
	for color := White; color <= Black; color++ {
		sp.mg[color] += c.DoubledPawns[0] * T(pw.doubledPawns[color].Count())
		sp.eg[color] += c.DoubledPawns[1] * T(pw.doubledPawns[color].Count())
	}
}

func (sp *scorePair[T]) addIsolatedPawns(pw pieceWise[T], c *CoeffSet[T]) {

	for color := White; color <= Black; color++ {
		sp.mg[color] += c.IsolatedPawns[0] * T(pw.isolatedPawns[color].Count())
		sp.eg[color] += c.IsolatedPawns[1] * T(pw.isolatedPawns[color].Count())
	}
}

type kingAttacks[T ScoreType] struct {
	score [2][2]T
}

func (ka *kingAttacks[T]) addAttackPieces(color Color, pType Piece, attacks board.BitBoard, kingNB board.BitBoard, c *CoeffSet[T]) {

	if kingNB&attacks != 0 {
		ka.score[0][color] += c.KingAttackPieces[0][pType-Knight]
		ka.score[1][color] += c.KingAttackPieces[1][pType-Knight]
	}
}

func (ka *kingAttacks[T]) addSafeChecks(color Color, pType Piece, safeChecks board.BitBoard, c *CoeffSet[T]) {
	ka.score[0][color] += c.SafeChecks[0][pType-Knight] * T(safeChecks.Count())
	ka.score[1][color] += c.SafeChecks[1][pType-Knight] * T(safeChecks.Count())
}

func (ka *kingAttacks[T]) addShelter(color Color, penalty T, c *CoeffSet[T]) {
	ka.score[0][color.Flip()] += c.KingShelter[0] * penalty
	ka.score[1][color.Flip()] += c.KingShelter[1] * penalty
}

func (ka kingAttacks[T]) sigmoidal(phase int, color Color) T {
	return sigmoidal(ka.score[phase][color])
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

func (pw *pieceWise[T]) calcQueenAttacks(color Color, sq Square) board.BitBoard {
	attacks := movegen.BishopMoves(sq, pw.occ) | movegen.RookMoves(sq, pw.occ)

	pw.attacks[color][Queen-Pawn] |= attacks
	return attacks
}

func (pw *pieceWise[T]) calcRookAttacks(color Color, sq Square) board.BitBoard {
	attacks := movegen.RookMoves(sq, pw.occ)

	pw.attacks[color][Rook-Pawn] |= attacks
	return attacks
}

func (sp *scorePair[T]) addRookMobility(b *board.Board, color Color, sq Square, attacks board.BitBoard, c *CoeffSet[T]) {

	rank := board.BitBoard(0xff) << (sq & 56)
	hmob := (attacks & rank & ^b.Colors[color]).Count()
	vmob := (attacks & ^rank & ^b.Colors[color]).Count()

	// count vertical mobility 2x compared to horizontal mobility
	mobCnt := (2*vmob + hmob) / 2

	sp.mg[color] += c.MobilityRook[0][mobCnt]
	sp.eg[color] += c.MobilityRook[1][mobCnt]

	// connected rooks
	if attacks&b.Pieces[Rook]&b.Colors[color] != 0 {
		sp.mg[color] += c.ConnectedRooks[0]
		sp.eg[color] += c.ConnectedRooks[1]
	}
}

func (pw *pieceWise[T]) calcBishopAttacks(color Color, sq Square) board.BitBoard {
	attacks := movegen.BishopMoves(sq, pw.occ)

	pw.attacks[color][Bishop-Pawn] |= attacks
	return attacks
}

func (sp *scorePair[T]) addBishopMobility(b *board.Board, color Color, attacks board.BitBoard, c *CoeffSet[T]) {

	mobCnt := (attacks & ^b.Colors[color]).Count()
	sp.mg[color] += c.MobilityBishop[0][mobCnt]
	sp.eg[color] += c.MobilityBishop[1][mobCnt]
}

func (pw *pieceWise[T]) calcKnightAttacks(color Color, sq Square) board.BitBoard {
	attacks := movegen.KnightMoves(sq)

	pw.attacks[color][Knight-Pawn] |= attacks
	return attacks
}

func (sp *scorePair[T]) addKnightMobility(b *board.Board, color Color, attacks board.BitBoard, pawnCover board.BitBoard, c *CoeffSet[T]) {

	mobCnt := (attacks & ^b.Colors[color] & ^pawnCover).Count()
	sp.mg[color] += c.MobilityKnight[0][mobCnt]
	sp.eg[color] += c.MobilityKnight[1][mobCnt]

}

func (sp *scorePair[T]) addKnightOutposts(color Color, knightBB board.BitBoard, sq Square, holes board.BitBoard, c *CoeffSet[T]) {

	// calculate knight outputs
	if (knightBB)&holes != 0 {
		// the hole square is from the enemy's perspective, white's in black's territory
		if color == White {
			sq ^= 56
		}
		sp.mg[color] += c.KnightOutpost[0][sq]
		sp.eg[color] += c.KnightOutpost[1][sq]
	}
}

func (pw *pieceWise[T]) calcCover() {
	for color := White; color <= Black; color++ {
		pw.cover[color] = pw.attacks[color][0] |
			pw.attacks[color][Knight-Pawn] |
			pw.attacks[color][Bishop-Pawn] |
			pw.attacks[color][Rook-Pawn] |
			pw.attacks[color][Queen-Pawn] |
			pw.attacks[color][King-Pawn]
	}
}

func (sp *scorePair[T]) addKingAttacks(ka kingAttacks[T]) {
	sp.mg[White] += ka.sigmoidal(0, White)
	sp.mg[Black] += ka.sigmoidal(0, Black)

	sp.eg[White] += ka.sigmoidal(1, White)
	sp.eg[Black] += ka.sigmoidal(1, Black)
}
