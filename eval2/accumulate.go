package eval2

import (
	"github.com/paulsonkoly/chess-3/attacks"
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
)

func (e *Eval[T]) addPieceValues(c *CoeffSet[T]) {
	for color := range Colors {
		for pType := Pawn; pType <= Queen; pType++ {
			e.sp[color][MG] += c.PieceValues[MG][pType] * T(e.pieceCounts[color][pType])
			e.sp[color][EG] += c.PieceValues[EG][pType] * T(e.pieceCounts[color][pType])
		}
	}
}

func (e *Eval[T]) addTempo(b *board.Board, c *CoeffSet[T]) {
	e.sp[b.STM][MG] += c.TempoBonus[MG]
	e.sp[b.STM][EG] += c.TempoBonus[EG]
}

func (e *Eval[T]) addBishopPair(c *CoeffSet[T]) {
	for color := range Colors {
		bishops := e.pieceCounts[color][Bishop]
		pawns := e.pieceCounts[color][Pawn]

		// technically FEN allows more than 8 pawns
		pawns = min(pawns, len(c.BishopPair)-1)

		// this fails in the rare case of having 2 matching colour complex bishops
		if bishops >= 2 {
			e.sp[color][MG] += c.BishopPair[pawns]
			e.sp[color][EG] += c.BishopPair[pawns]
		}
	}
}

func (e *Eval[T]) addKingNBAttack(color Color, pType Piece, attacks BitBoard, kingNB BitBoard, c *CoeffSet[T]) {
	if kingNB&attacks != 0 {
		e.kingAttacks[color] += c.KingAttackPieces[pType-Knight]
	}
}

func (e *Eval[T]) addPSqT(color Color, pType Piece, sq Square, c *CoeffSet[T]) {
	if color == White {
		sq ^= 56 // upside down
	}

	ix := pType - 1

	e.sp[color][MG] += c.PSqT[2*ix][sq]
	e.sp[color][EG] += c.PSqT[2*ix+1][sq]
}

func (e *Eval[T]) addKingAttacks(c *CoeffSet[T]) {
	whiteSgm := sigmoidal(e.kingAttacks[White])
	blackSgm := sigmoidal(e.kingAttacks[Black])

	var t T
	if _, ok := ((any)(t).(Score)); ok {
		e.sp[White][MG] += T(((int)(whiteSgm) * (int)(c.KingAttackMagnitude[MG])) / 64)
		e.sp[Black][MG] += T(((int)(blackSgm) * (int)(c.KingAttackMagnitude[MG])) / 64)
		e.sp[White][EG] += T(((int)(whiteSgm) * (int)(c.KingAttackMagnitude[EG])) / 64)
		e.sp[Black][EG] += T(((int)(blackSgm) * (int)(c.KingAttackMagnitude[EG])) / 64)
		return
	}
	e.sp[White][MG] += (whiteSgm * c.KingAttackMagnitude[MG]) / 64
	e.sp[Black][MG] += (blackSgm * c.KingAttackMagnitude[MG]) / 64
	e.sp[White][EG] += (whiteSgm * c.KingAttackMagnitude[EG]) / 64
	e.sp[Black][EG] += (blackSgm * c.KingAttackMagnitude[EG]) / 64
}

func (e *Eval[T]) addThreats(b *board.Board, c *CoeffSet[T]) {
	for color := range Colors {
		defended := e.cover[color.Flip()]
		undefendedAttacked := ^defended & e.cover[color]

		// special case safe pawn threats
		safe := ^e.cover[color.Flip()] | e.cover[color]
		pawns := b.Colors[color] & b.Pieces[Pawn]
		spThreatened := attacks.PawnCaptureMoves(safe&pawns, color)
		targets := b.Colors[color.Flip()] &^ b.Pieces[Pawn]

		cnt := T((spThreatened & targets).Count())

		e.sp[color][MG] += c.SafePawnThreats[MG] * cnt
		e.sp[color][EG] += c.SafePawnThreats[EG] * cnt

		lesserAttackers := e.attacks[color][Pawn] & ^spThreatened // pawns to start with, but not double counting safe pawns.

		// defended minors are threatened by pawns, undefended minors are theatened by anything
		threatened := (defended & lesserAttackers) | undefendedAttacked
		minors := (b.Pieces[Knight] | b.Pieces[Bishop]) & b.Colors[color.Flip()]
		cnt = T((threatened & minors).Count())

		// defended rooks are threatened by anything but rooks and queens, undefended rooks threatened by anything
		lesserAttackers |= e.attacks[color][Knight] | e.attacks[color][Bishop]
		threatened = (defended & lesserAttackers) | undefendedAttacked
		rooks := b.Pieces[Rook] & b.Colors[color.Flip()]
		cnt += T((threatened & rooks).Count())

		// defended queens are threatened by anything but queens, undefended queens threatened by anything
		lesserAttackers |= e.attacks[color][Rook]
		threatened = (defended & lesserAttackers) | undefendedAttacked
		queens := b.Pieces[Queen] & b.Colors[color.Flip()]
		cnt += T((threatened & queens).Count())

		e.sp[color][MG] += c.Threats[MG] * cnt
		e.sp[color][EG] += c.Threats[EG] * cnt
	}
}

func (e *Eval[T]) addChecks(b *board.Board, c *CoeffSet[T]) {
	occ := b.Colors[White] | b.Colors[Black]
	// safe checks
	for color := range Colors {
		eKSq := e.kings[color.Flip()].sq
		eCover := e.cover[color.Flip()]
		eKBRays := attacks.BishopMoves(eKSq, occ)
		eKRRays := attacks.RookMoves(eKSq, occ)

		// Queen
		eKAttack := eKBRays | eKRRays
		checks := e.attacks[color][Queen] & eKAttack & ^b.Colors[color]

		e.kingAttacks[color] += c.SafeChecks[Queen-Knight] * T((checks &^ eCover).Count())
		e.kingAttacks[color] += c.UnsafeChecks[Queen-Knight] * T((checks & eCover).Count())

		// Rook
		eKAttack = eKRRays
		checks = e.attacks[color][Rook] & eKAttack & ^b.Colors[color]

		e.kingAttacks[color] += c.SafeChecks[Rook-Knight] * T((checks &^ eCover).Count())
		e.kingAttacks[color] += c.UnsafeChecks[Rook-Knight] * T((checks & eCover).Count())

		// Bishop
		eKAttack = eKBRays
		checks = e.attacks[color][Bishop] & eKAttack & ^b.Colors[color]

		e.kingAttacks[color] += c.SafeChecks[Bishop-Knight] * T((checks &^ eCover).Count())
		e.kingAttacks[color] += c.UnsafeChecks[Bishop-Knight] * T((checks & eCover).Count())

		// Knight
		eKAttack = attacks.KnightMoves(e.kings[color.Flip()].sq)
		checks = e.attacks[color][Knight] & eKAttack & ^b.Colors[color]

		e.kingAttacks[color] += c.SafeChecks[0] * T((checks &^ eCover).Count())
		e.kingAttacks[color] += c.UnsafeChecks[0] * T((checks & eCover).Count())
	}
}

func (e *Eval[T]) addRookMobility(b *board.Board, color Color, attacks BitBoard, c *CoeffSet[T]) {
	mobCnt := (attacks & ^b.Colors[color]).Count()

	e.sp[color][MG] += c.MobilityRook[MG][mobCnt]
	e.sp[color][EG] += c.MobilityRook[EG][mobCnt]

	// connected rooks
	if attacks&b.Pieces[Rook]&b.Colors[color] != 0 {
		e.sp[color][MG] += c.ConnectedRooks[MG]
		e.sp[color][EG] += c.ConnectedRooks[EG]
	}
}

func (e *Eval[T]) addRookFiles(b *board.Board, color Color, sq Square, c *CoeffSet[T]) {
	file := FileBB(sq.File())

	if file&b.Pieces[Pawn] == 0 {
		e.sp[color][MG] += c.RookOnOpen[MG]
		e.sp[color][EG] += c.RookOnOpen[EG]
	} else if file&b.Pieces[Pawn]&b.Colors[color] == 0 {
		e.sp[color][MG] += c.RookOnSemiOpen[MG]
		e.sp[color][EG] += c.RookOnSemiOpen[EG]
	}
}

func (e *Eval[T]) addBishopMobility(b *board.Board, color Color, attacks BitBoard, c *CoeffSet[T]) {

	mobCnt := (attacks & ^b.Colors[color]).Count()
	e.sp[color][MG] += c.MobilityBishop[MG][mobCnt]
	e.sp[color][EG] += c.MobilityBishop[EG][mobCnt]
}

func (e *Eval[T]) addBishopOutposts(color Color, sq Square, outposts BitBoard, c *CoeffSet[T]) {
	rank := sq.Rank().FromPerspectiveOf(color)
	if BitBoard(1)<<sq&outposts != 0 && FourthRank <= rank && rank <= SixthRank {
		e.sp[color][MG] += c.BishopOutpost[MG]
		e.sp[color][EG] += c.BishopOutpost[EG]
	}
}

func (e *Eval[T]) addKnightBehindPawn(b *board.Board, color Color, c *CoeffSet[T]) {
	pawnMask := attacks.PawnSinglePushMoves(b.Pieces[Pawn], color.Flip())
	knights := b.Pieces[Knight] & b.Colors[color]
	cnt := T((pawnMask & knights).Count())

	e.sp[color][MG] += c.KnightBehindPawn[MG] * cnt
	e.sp[color][EG] += c.KnightBehindPawn[EG] * cnt
}

func (e *Eval[T]) addKnightMobility(b *board.Board, color Color, attacks BitBoard, c *CoeffSet[T]) {
	ePCover := e.attacks[color.Flip()][Pawn]
	mobCnt := (attacks & ^b.Colors[color] & ^ePCover).Count()
	e.sp[color][MG] += c.MobilityKnight[MG][mobCnt]
	e.sp[color][EG] += c.MobilityKnight[EG][mobCnt]
}

// the player's side of the board with the extra 2 central squares included at
// enemy side.
var sideOfBoard = [2]BitBoard{0x00000018_ffffffff, 0xffffffff_18000000}

func (e *Eval[T]) addKnightOutposts(color Color, sq Square, outposts BitBoard, c *CoeffSet[T]) {
	outposts &= sideOfBoard[color.Flip()]
	// calculate knight outputs
	if (BitBoard(1)<<sq)&outposts != 0 {
		// the hole square is from the enemy's perspective, white's in black's territory
		if color == White {
			sq ^= 56
		}
		e.sp[color][MG] += c.KnightOutpost[MG][sq]
		e.sp[color][EG] += c.KnightOutpost[EG][sq]
	}
}
