// Package eval gives position evaluation measuerd in centipawns.
package eval

import (
	"github.com/paulsonkoly/chess-3/attacks"
	"github.com/paulsonkoly/chess-3/board"

	. "github.com/paulsonkoly/chess-3/chess"
)

// ScoreType defines the evaluation result type. The engine uses int16 for
// score type, as defined in types. The tuner uses float64.
type ScoreType interface{ Score | float64 }

func Eval[T ScoreType](b *board.Board, c *CoeffSet[T]) T {
	if insufficientMat(b) {
		return 0
	}

	sp := scorePair[T]{}
	phase := phase[T]{}

	for pType := Pawn; pType <= Queen; pType++ {
		wCnt := (b.Pieces[pType] & b.Colors[White]).Count()
		bCnt := (b.Pieces[pType] & b.Colors[Black]).Count()

		phase.addPieces(pType, wCnt+bCnt)
		sp.addPieceValues(White, pType, wCnt, c)
		sp.addPieceValues(Black, pType, bCnt, c)
	}

	// special case checkmate patterns
	if KNBvK(b) { // knight and bishop checkmate
		sp.KNBvK(b, c)

		return sp.endgameScore(b)
	}

	sp.addTempo(b, c)
	sp.addBishopPair(b, c)

	pw := pieceWise{}

	pw.calcPawnAttacks(b)

	pw.calcOccupancy(b)
	pw.calcKingSquares(b)

	pawns := pawns{}
	pawns.calcPawns(b)

	sp.addPassers(b, &pawns, &pw, c)
	sp.addDoubledPawns(&pawns, c)
	sp.addIsolatedPawns(&pawns, c)

	ka := kingAttacks[T]{}

	for color := White; color <= Black; color++ {

		// enemy king neighbourhood
		eKNb := pw.kingNb[color.Flip()]

		// queens
		for pieces := b.Pieces[Queen] & b.Colors[color]; pieces != 0; pieces &= pieces - 1 {
			sq := pieces.LowestSet()

			attacks := pw.calcQueenAttacks(color, sq)

			ka.addAttackPieces(color, Queen, attacks, eKNb, c)

			sp.addPSqT(color, Queen, sq, c)
		}

		// rooks
		for pieces := b.Pieces[Rook] & b.Colors[color]; pieces != 0; pieces &= pieces - 1 {
			sq := pieces.LowestSet()

			attacks := pw.calcRookAttacks(color, sq)

			ka.addAttackPieces(color, Rook, attacks, eKNb, c)
			sp.addRookMobility(b, color, attacks, c)
			sp.addRookFiles(b, color, sq, c)
			sp.addPSqT(color, Rook, sq, c)
		}

		// bishops
		for pieces := b.Pieces[Bishop] & b.Colors[color]; pieces != 0; pieces &= pieces - 1 {
			sq := pieces.LowestSet()

			attacks := pw.calcBishopAttacks(color, sq)

			ka.addAttackPieces(color, Bishop, attacks, eKNb, c)
			sp.addBishopMobility(b, color, attacks, c)
			sp.addPSqT(color, Bishop, sq, c)
		}

		// knights
		outposts := pawns.holes(color.Flip()) & pw.attacks[color][0]
		for pieces := b.Pieces[Knight] & b.Colors[color]; pieces != 0; pieces &= pieces - 1 {
			sq := pieces.LowestSet()

			attacks := pw.calcKnightAttacks(color, sq)

			ka.addAttackPieces(color, Knight, attacks, eKNb, c)
			sp.addKnightMobility(b, color, attacks, pw.attacks[color.Flip()][0], c)
			sp.addKnightOutposts(color, sq, outposts, c)
			sp.addPSqT(color, Knight, sq, c)
		}

		// pawns
		for pieces := b.Pieces[Pawn] & b.Colors[color]; pieces != 0; pieces &= pieces - 1 {
			sq := pieces.LowestSet()

			sp.addPSqT(color, Pawn, sq, c)
		}

		// king
		piece := b.Pieces[King] & b.Colors[color]
		sq := piece.LowestSet()

		sp.addPSqT(color, King, sq, c)
		sp.addPawnlessFlank(color, sq, b.Pieces[Pawn], c)
	}

	pw.calcCover()

	sp.addThreats(b, &pw, c)

	// safe checks
	for color := White; color <= Black; color++ {
		eCover := pw.cover[color.Flip()]

		var safeChecks BitBoard

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
		eKAttack = attacks.KnightMoves(pw.kingSq[color.Flip()])
		safeChecks = pw.attacks[color][Knight-Pawn] & eKAttack & ^eCover & ^b.Colors[color]

		ka.addSafeChecks(color, Knight, safeChecks, c)

		// shelter
		pCnt := (pw.kingNb[color] & b.Colors[color] & b.Pieces[Pawn]).Count()
		penalty := T(max(3-pCnt, 0))

		ka.addShelter(color, penalty, c)
	}

	sp.addKingAttacks(ka, c)

	score := sp.taperedScore(b, phase)
	// drawishness
	fifty := int(b.FiftyCnt)
	if _, ok := ((any)(score)).(Score); ok {
		return T(int(score) * (100 - fifty) / 100)
	}

	return score * T(100-fifty) / 100
}

func insufficientMat(b *board.Board) bool {
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

func KNBvK(b *board.Board) bool {
	whiteN := b.Pieces[Knight] & b.Colors[White]
	blackN := b.Pieces[Knight] & b.Colors[Black]
	whiteB := b.Pieces[Bishop] & b.Colors[White]
	blackB := b.Pieces[Bishop] & b.Colors[Black]

	return b.Pieces[Pawn]|b.Pieces[Rook]|b.Pieces[Queen] == 0 &&
		((whiteN.IsPow2() && whiteB.IsPow2() && (blackN|blackB) == 0) ||
			(blackN.IsPow2() && blackB.IsPow2() && (whiteN|whiteB) == 0))
}

type scorePair[T ScoreType] struct {
	mg [2]T
	eg [2]T
}

// KBCorners are knight-bishop checkmate corners based on parity of square.
var KBCorners = [2][2]Square{{A1, H8}, {H1, A8}}

func (sp *scorePair[T]) KNBvK(b *board.Board, c *CoeffSet[T]) {
	bishopSq := b.Pieces[Bishop].LowestSet()
	knightSq := b.Pieces[Knight].LowestSet()

	victim := White
	if b.Pieces[Bishop]&b.Colors[White] != 0 {
		victim = Black
	}
	victimKSq := (b.Pieces[King] & b.Colors[victim]).LowestSet()
	attackKSq := (b.Pieces[King] & b.Colors[victim.Flip()]).LowestSet()

	sp.addPSqT(victim, King, victimKSq, c)
	sp.addPSqT(victim.Flip(), King, attackKSq, c)
	sp.addPSqT(victim.Flip(), Knight, knightSq, c)
	sp.addPSqT(victim.Flip(), Bishop, bishopSq, c)

	parity := (bishopSq.File() + bishopSq.Rank()) & 1

	cornerDist := min(Chebishev(victimKSq, KBCorners[parity][0]), Chebishev(victimKSq, KBCorners[parity][1]))
	cornerDist = 7 - cornerDist
	cornerDist *= cornerDist

	sp.eg[victim.Flip()] += T(cornerDist) * 30
}

func (sp *scorePair[T]) addPieceValues(color Color, pType Piece, cnt int, c *CoeffSet[T]) {
	sp.mg[color] += T(cnt) * c.PieceValues[0][pType]
	sp.eg[color] += T(cnt) * c.PieceValues[1][pType]
}

func (sp *scorePair[T]) addTempo(b *board.Board, c *CoeffSet[T]) {
	sp.mg[b.STM] += c.TempoBonus[0]
	sp.eg[b.STM] += c.TempoBonus[1]
}

func (sp *scorePair[T]) addBishopPair(b *board.Board, c *CoeffSet[T]) {
	for color := White; color <= Black; color++ {
		myBishops := b.Colors[color] & b.Pieces[Bishop]
		myPawnCnt := (b.Colors[color] & b.Pieces[Pawn]).Count()

		// technically FEN allows more than 8 pawns
		myPawnCnt = min(myPawnCnt, len(c.BishopPair)-1)

		// this fails in the rare case of having 2 matching colour complex bishops
		if myBishops&(myBishops-1) != 0 {
			sp.mg[color] += c.BishopPair[myPawnCnt]
			sp.eg[color] += c.BishopPair[myPawnCnt]
		}
	}
}

func (sp *scorePair[T]) addPSqT(color Color, pType Piece, sq Square, c *CoeffSet[T]) {
	if color == White {
		sq ^= 56 // upside down
	}

	ix := pType - 1

	sp.mg[color] += c.PSqT[2*ix][sq]
	sp.eg[color] += c.PSqT[2*ix+1][sq]
}

func (sp *scorePair[T]) addKingAttacks(ka kingAttacks[T], c *CoeffSet[T]) {
	whiteSgm := ka.sigmoidal(White)
	blackSgm := ka.sigmoidal(Black)
	var t T
	if _, ok := ((any)(t).(Score)); ok {
		sp.mg[White] += T(((int)(whiteSgm) * (int)(c.KingAttackMagnitude[0])) / 64)
		sp.mg[Black] += T(((int)(blackSgm) * (int)(c.KingAttackMagnitude[0])) / 64)
		sp.eg[White] += T(((int)(whiteSgm) * (int)(c.KingAttackMagnitude[1])) / 64)
		sp.eg[Black] += T(((int)(blackSgm) * (int)(c.KingAttackMagnitude[1])) / 64)
		return
	}
	sp.mg[White] += (whiteSgm * c.KingAttackMagnitude[0]) / 64
	sp.mg[Black] += (blackSgm * c.KingAttackMagnitude[0]) / 64
	sp.eg[White] += (whiteSgm * c.KingAttackMagnitude[1]) / 64
	sp.eg[Black] += (blackSgm * c.KingAttackMagnitude[1]) / 64
}

func (sp *scorePair[T]) taperedScore(b *board.Board, phase phase[T]) T {
	mgScore := sp.mg[b.STM] - sp.mg[b.STM.Flip()]
	egScore := sp.eg[b.STM] - sp.eg[b.STM.Flip()]

	return phase.blend(mgScore, egScore)
}

func (sp *scorePair[T]) endgameScore(b *board.Board) T {
	return sp.eg[b.STM] - sp.eg[b.STM.Flip()]
}

type pieceWise struct {
	occ      BitBoard
	attacks  [2][6]BitBoard
	kingRays [2][2]BitBoard
	kingSq   [2]Square
	kingNb   [2]BitBoard
	cover    [2]BitBoard
}

func (pw *pieceWise) calcOccupancy(b *board.Board) {
	pw.occ = b.Colors[White] | b.Colors[Black]
}

func (pw *pieceWise) calcKingSquares(b *board.Board) {
	for color := White; color <= Black; color++ {
		king := b.Colors[color] & b.Pieces[King]
		kingSq := king.LowestSet()
		kingA := attacks.KingMoves(kingSq)

		pw.attacks[color][King-Pawn] = kingA
		pw.kingRays[color][0] = attacks.BishopMoves(kingSq, pw.occ)
		pw.kingRays[color][Rook-Bishop] = attacks.RookMoves(kingSq, pw.occ)
		pw.kingSq[color] = kingSq
		pw.kingNb[color] = king | kingA
	}
}

func (sp *scorePair[T]) addPassers(b *board.Board, pawns *pawns, pw *pieceWise, c *CoeffSet[T]) {
	for color := White; color <= Black; color++ {

		passers := pawns.passers(color)

		// if there is a sole passer
		if passers.IsPow2() {
			sq := passers.LowestSet()

			// KPR, KPNB
			if b.Pieces[Knight]|b.Pieces[Bishop]|b.Pieces[Queen] == 0 || b.Pieces[Rook]|b.Pieces[Queen] == 0 {
				queeningSq := Square(EighthRank.FromPerspectiveOf(color)*8 + sq.File())

				kingDist := Chebishev(queeningSq, pw.kingSq[color.Flip()]) - Chebishev(queeningSq, pw.kingSq[color])

				sp.mg[color] += c.PasserKingDist[0] * T(kingDist)
				sp.eg[color] += c.PasserKingDist[1] * T(kingDist)
			}
		}

		protectedCnt := T((passers & pw.attacks[color][0]).Count())
		sp.mg[color] += c.ProtectedPasser[0] * protectedCnt
		sp.eg[color] += c.ProtectedPasser[1] * protectedCnt

		for ; passers != 0; passers &= passers - 1 {
			sq := passers.LowestSet()

			rank := sq.Rank().FromPerspectiveOf(color)

			sp.mg[color] += c.PasserRank[0][rank-1]
			sp.eg[color] += c.PasserRank[1][rank-1]
		}
	}
}

func Chebishev(a, b Square) int {
	ax, ay, bx, by := int(a%8), int(a/8), int(b%8), int(b/8)
	return max(Abs(ax-bx), Abs(ay-by))
}

func (sp *scorePair[T]) addDoubledPawns(pawns *pawns, c *CoeffSet[T]) {
	for color := White; color <= Black; color++ {
		dblCnt := T(pawns.doubledPawns(color).Count())
		sp.mg[color] += c.DoubledPawns[0] * dblCnt
		sp.eg[color] += c.DoubledPawns[1] * dblCnt
	}
}

func (sp *scorePair[T]) addIsolatedPawns(pawns *pawns, c *CoeffSet[T]) {

	for color := White; color <= Black; color++ {
		isoCnt := T(pawns.isolatedPawns(color).Count())
		sp.mg[color] += c.IsolatedPawns[0] * isoCnt
		sp.eg[color] += c.IsolatedPawns[1] * isoCnt
	}
}

func (sp *scorePair[T]) addThreats(b *board.Board, pw *pieceWise, c *CoeffSet[T]) {
	for color := White; color <= Black; color++ {
		defended := pw.cover[color.Flip()]
		undefendedAttacked := ^defended & pw.cover[color]

		// special case safe pawn threats
		safe := ^pw.cover[color.Flip()] | pw.cover[color]
		pawns := b.Colors[color] & b.Pieces[Pawn]
		spThreatened := attacks.PawnCaptureMoves(safe&pawns, color)
		targets := b.Colors[color.Flip()] &^ b.Pieces[Pawn]

		cnt := T((spThreatened & targets).Count())

		sp.mg[color] += c.SafePawnThreats[0] * cnt
		sp.eg[color] += c.SafePawnThreats[1] * cnt

		lesserAttackers := pw.attacks[color][0] & ^spThreatened // pawns to start with, but not double counting safe pawns.

		// defended minors are threatened by pawns, undefended minors are theatened by anything
		threatened := (defended & lesserAttackers) | undefendedAttacked
		minors := (b.Pieces[Knight] | b.Pieces[Bishop]) & b.Colors[color.Flip()]
		cnt = T((threatened & minors).Count())

		// defended rooks are threatened by anything but rooks and queens, undefended rooks threatened by anything
		lesserAttackers |= pw.attacks[color][Knight-Pawn] | pw.attacks[color][Bishop-Pawn]
		threatened = (defended & lesserAttackers) | undefendedAttacked
		rooks := b.Pieces[Rook] & b.Colors[color.Flip()]
		cnt += T((threatened & rooks).Count())

		// defended queens are threatened by anything but queens, undefended queens threatened by anything
		lesserAttackers |= pw.attacks[color][Rook-Pawn]
		threatened = (defended & lesserAttackers) | undefendedAttacked
		queens := b.Pieces[Queen] & b.Colors[color.Flip()]
		cnt += T((threatened & queens).Count())

		sp.mg[color] += c.Threats[0] * cnt
		sp.eg[color] += c.Threats[1] * cnt
	}
}

func (pw *pieceWise) calcPawnAttacks(b *board.Board) {
	ps := [...]BitBoard{b.Pieces[Pawn] & b.Colors[White], b.Pieces[Pawn] & b.Colors[Black]}

	pw.attacks[White][0] = attacks.PawnCaptureMoves(ps[White], White)
	pw.attacks[Black][0] = attacks.PawnCaptureMoves(ps[Black], Black)
}

func (pw *pieceWise) calcQueenAttacks(color Color, sq Square) BitBoard {
	attacks := attacks.BishopMoves(sq, pw.occ) | attacks.RookMoves(sq, pw.occ)

	pw.attacks[color][Queen-Pawn] |= attacks
	return attacks
}

func (pw *pieceWise) calcRookAttacks(color Color, sq Square) BitBoard {
	attacks := attacks.RookMoves(sq, pw.occ)

	pw.attacks[color][Rook-Pawn] |= attacks
	return attacks
}

func (sp *scorePair[T]) addRookMobility(b *board.Board, color Color, attacks BitBoard, c *CoeffSet[T]) {
	mobCnt := (attacks & ^b.Colors[color]).Count()

	sp.mg[color] += c.MobilityRook[0][mobCnt]
	sp.eg[color] += c.MobilityRook[1][mobCnt]

	// connected rooks
	if attacks&b.Pieces[Rook]&b.Colors[color] != 0 {
		sp.mg[color] += c.ConnectedRooks[0]
		sp.eg[color] += c.ConnectedRooks[1]
	}
}

func (sp *scorePair[T]) addRookFiles(b *board.Board, color Color, sq Square, c *CoeffSet[T]) {
	file := FileBB(sq.File())

	if file&b.Pieces[Pawn] == 0 {
		sp.mg[color] += c.RookOnOpen[0]
		sp.eg[color] += c.RookOnOpen[1]
	} else if file&b.Pieces[Pawn]&b.Colors[color] == 0 {
		sp.mg[color] += c.RookOnSemiOpen[0]
		sp.eg[color] += c.RookOnSemiOpen[1]
	}
}

func (pw *pieceWise) calcBishopAttacks(color Color, sq Square) BitBoard {
	attacks := attacks.BishopMoves(sq, pw.occ)

	pw.attacks[color][Bishop-Pawn] |= attacks
	return attacks
}

func (sp *scorePair[T]) addBishopMobility(b *board.Board, color Color, attacks BitBoard, c *CoeffSet[T]) {

	mobCnt := (attacks & ^b.Colors[color]).Count()
	sp.mg[color] += c.MobilityBishop[0][mobCnt]
	sp.eg[color] += c.MobilityBishop[1][mobCnt]
}

func (pw *pieceWise) calcKnightAttacks(color Color, sq Square) BitBoard {
	attacks := attacks.KnightMoves(sq)

	pw.attacks[color][Knight-Pawn] |= attacks
	return attacks
}

func (sp *scorePair[T]) addKnightMobility(
	b *board.Board,
	color Color,
	attacks BitBoard,
	pawnCover BitBoard,
	c *CoeffSet[T],
) {

	mobCnt := (attacks & ^b.Colors[color] & ^pawnCover).Count()
	sp.mg[color] += c.MobilityKnight[0][mobCnt]
	sp.eg[color] += c.MobilityKnight[1][mobCnt]

}

func (sp *scorePair[T]) addKnightOutposts(color Color, sq Square, holes BitBoard, c *CoeffSet[T]) {

	// calculate knight outputs
	if (BitBoard(1)<<sq)&holes != 0 {
		// the hole square is from the enemy's perspective, white's in black's territory
		if color == White {
			sq ^= 56
		}
		sp.mg[color] += c.KnightOutpost[0][sq]
		sp.eg[color] += c.KnightOutpost[1][sq]
	}
}

func (pw *pieceWise) calcCover() {
	for color := White; color <= Black; color++ {
		pw.cover[color] = pw.attacks[color][0] |
			pw.attacks[color][Knight-Pawn] |
			pw.attacks[color][Bishop-Pawn] |
			pw.attacks[color][Rook-Pawn] |
			pw.attacks[color][Queen-Pawn] |
			pw.attacks[color][King-Pawn]
	}
}

func (sp *scorePair[T]) addPawnlessFlank(color Color, sq Square, pawns BitBoard, c *CoeffSet[T]) {
	if FileCluster(sq.File())&pawns == 0 {
		sp.mg[color] += c.PawnlessFlank[0]
		sp.eg[color] += c.PawnlessFlank[1]
	}
}
