package movegen

import (
	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

// KingMoves is the bitboard set where the king can move to from from. It does
// not take into accound occupancies or legality.
func KingMoves(from Square) board.BitBoard { return kingMoves[from] }

// KnightMoves is the bitboard set where a knight can move to from from. It does
// not take into accound occupancies or legality.
func KnightMoves(from Square) board.BitBoard { return knightMoves[from] }

// BishopMoves is the bitboard set where a bishop can move to from from. It
// does not take into account occupancy for the side to move, (can have bits
// set on STM's pieces), or legality.
func BishopMoves(from Square, occ board.BitBoard) board.BitBoard {
	mask := bishopMasks[from]
	magic := bishopMagics[from]
	shift := bishopShifts[from]

	return bishopAttacks[from][((occ&mask)*magic)>>(64-shift)]
}

// RookMoves is the bitbord set where a rook can move to from from. It does not
// take into account occupancy for the side to move, (can have bits set on
// STM's pieces), or legality.
func RookMoves(from Square, occ board.BitBoard) board.BitBoard {
	mask := rookMasks[from]
	magic := rookMagics[from]
	shift := rookShifts[from]

	return rookAttacks[from][((occ&mask)*magic)>>(64-shift)]
}

// PawnCaptureMoves is the bitboard set where the pawns of color color can
// capture, from any of the squares set in b.
func PawnCaptureMoves(b board.BitBoard, color Color) board.BitBoard {
	return ((((b & ^board.AFile) << 7) | ((b & ^board.HFile) << 9)) >> (color << 4)) |
		((((b & ^board.HFile) >> 7) | ((b & ^board.AFile) >> 9)) << (color.Flip() << 4))
}

// PawnSinglePushMoves is the bitboard set where the pawns of color color can
// push a single square forward from any of the squares set in b.
func PawnSinglePushMoves(b board.BitBoard, color Color) board.BitBoard {
	return ((b)<<8)>>((color)<<4) | ((b)>>8)<<((color^1)<<4)
}

var (
	sndRank    = [...]board.BitBoard{board.SecondRank, board.SeventhRank}
	fourthRank = [...]board.BitBoard{board.FourthRank, board.FifthRank}
	castleMask = [2][2]board.BitBoard{
		{(1 << E1) | (1 << F1) | (1 << G1), (1 << E1) | (1 << D1) | (1 << C1)},
		{(1 << E8) | (1 << F8) | (1 << G8), (1 << E8) | (1 << D8) | (1 << C8)},
	}
	kingCRightsUpd = [2]CastlingRights{CRights(ShortWhite, LongWhite), CRights(ShortBlack, LongBlack)}
	rookCRightsUpd = [64]CastlingRights{
		CRights(LongWhite), 0, 0, 0, 0, 0, 0, CRights(ShortWhite),
		0 /**************/, 0, 0, 0, 0, 0, 0, 0,
		0 /**************/, 0, 0, 0, 0, 0, 0, 0,
		0 /**************/, 0, 0, 0, 0, 0, 0, 0,
		0 /**************/, 0, 0, 0, 0, 0, 0, 0,
		0 /**************/, 0, 0, 0, 0, 0, 0, 0,
		0 /**************/, 0, 0, 0, 0, 0, 0, 0,
		CRights(LongBlack), 0, 0, 0, 0, 0, 0, CRights(ShortBlack),
	}
)

type generator struct {
	self, them, occ board.BitBoard
}

func (g generator) kingMoves(ms *move.Store, b *board.Board, fromMsk, toMsk board.BitBoard) {
	// king moves
	if piece := g.self & b.Pieces[King] & fromMsk; piece != 0 {
		from := piece.LowestSet()

		tSqrs := kingMoves[from] & ^g.self & toMsk

		for tSqr := board.BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()
			newC := b.CRights & ^(kingCRightsUpd[b.STM] | rookCRightsUpd[to])
			m := ms.Alloc()
			m.Piece = King
			m.SetFrom(from)
			m.SetTo(to)
			m.CRights = newC ^ b.CRights
			m.Castle = 0
			m.SetPromo(0)
			m.EPP = 0
			m.EPSq = b.EnPassant
		}
	}
}

func (g generator) knightMoves(ms *move.Store, b *board.Board, fromMsk, toMsk board.BitBoard) {
	// knight moves
	knights := g.self & b.Pieces[Knight] & fromMsk

	for knight := board.BitBoard(0); knights != 0; knights ^= knight {
		knight = knights & -knights
		from := knight.LowestSet()

		tSqrs := knightMoves[from] & ^g.self & toMsk

		for tSqr := board.BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()
			newC := b.CRights & ^(rookCRightsUpd[to])
			m := ms.Alloc()
			m.Piece = Knight
			m.SetFrom(from)
			m.SetTo(to)
			m.CRights = newC ^ b.CRights
			m.Castle = 0
			m.SetPromo(0)
			m.EPP = 0
			m.EPSq = b.EnPassant
		}
	}
}

func (g generator) bishopMoves(ms *move.Store, b *board.Board, fromMsk, toMsk board.BitBoard) {
	// bishop moves
	bishops := g.self & b.Pieces[Bishop] & fromMsk
	for bishop := board.BitBoard(0); bishops != 0; bishops ^= bishop {
		bishop = bishops & -bishops
		from := bishop.LowestSet()
		mask := bishopMasks[from]
		magic := bishopMagics[from]
		shift := bishopShifts[from]

		tSqrs := bishopAttacks[from][((g.occ&mask)*magic)>>(64-shift)] & ^g.self & toMsk

		for tSqr := board.BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()
			newC := b.CRights & ^(rookCRightsUpd[to])
			m := ms.Alloc()
			m.Piece = Bishop
			m.SetFrom(from)
			m.SetTo(to)
			m.CRights = newC ^ b.CRights
			m.Castle = 0
			m.SetPromo(0)
			m.EPP = 0
			m.EPSq = b.EnPassant
		}
	}
}

func (g generator) rookMoves(ms *move.Store, b *board.Board, fromMsk, toMsk board.BitBoard) {
	rooks := g.self & b.Pieces[Rook] & fromMsk

	for rook := board.BitBoard(0); rooks != 0; rooks ^= rook {
		rook = rooks & -rooks
		from := rook.LowestSet()
		mask := rookMasks[from]
		magic := rookMagics[from]
		shift := rookShifts[from]

		tSqrs := rookAttacks[from][((g.occ&mask)*magic)>>(64-shift)] & ^g.self & toMsk

		for tSqr := board.BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()
			// this accounts for flipping the castling rights for the moving side
			// if the rook moves away from castling position and also for the
			// opponent when a rook is capturing a rook in castling position
			newC := b.CRights & ^(rookCRightsUpd[from] | rookCRightsUpd[to])
			m := ms.Alloc()
			m.Piece = Rook
			m.SetFrom(from)
			m.SetTo(to)
			m.CRights = newC ^ b.CRights
			m.Castle = 0
			m.SetPromo(0)
			m.EPP = 0
			m.EPSq = b.EnPassant
		}
	}
}

func (g generator) queenMoves(ms *move.Store, b *board.Board, fromMsk, toMsk board.BitBoard) {
	queens := g.self & b.Pieces[Queen] & fromMsk
	for queen := board.BitBoard(0); queens != 0; queens ^= queen {
		queen = queens & -queens
		from := queen.LowestSet()
		mask := rookMasks[from]
		magic := rookMagics[from]
		shift := rookShifts[from]

		tSqrs := rookAttacks[from][((g.occ&mask)*magic)>>(64-shift)]

		mask = bishopMasks[from]
		magic = bishopMagics[from]
		shift = bishopShifts[from]

		tSqrs |= bishopAttacks[from][((g.occ&mask)*magic)>>(64-shift)]
		tSqrs &= ^g.self & toMsk

		for tSqr := board.BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()
			cNew := b.CRights &^ rookCRightsUpd[to]
			m := ms.Alloc()
			m.Piece = Queen
			m.SetFrom(from)
			m.SetTo(to)
			m.CRights = cNew ^ b.CRights
			m.Castle = 0
			m.SetPromo(0)
			m.EPP = 0
			m.EPSq = b.EnPassant
		}
	}
}

var shifts = [2]Square{8, -8}

func (g generator) singlePushMoves(ms *move.Store, b *board.Board, fromMsk board.BitBoard) {
	occ1 := (g.occ >> 8) << (b.STM << 4)
	pushable := g.self & b.Pieces[Pawn] & ^occ1
	theirSndRank := sndRank[b.STM.Flip()]
	shift := shifts[b.STM]

	// single pawn pushes (no promotions)
	for pawns, pawn := pushable&fromMsk & ^theirSndRank, board.BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()

		m := ms.Alloc()
		m.Piece = Pawn
		m.SetFrom(from)
		m.SetTo(from + shift)
		m.CRights = 0
		m.Castle = 0
		m.SetPromo(0)
		m.EPP = 0
		m.EPSq = b.EnPassant
	}
}

func (g generator) promoPushMoves(ms *move.Store, b *board.Board, fromMsk board.BitBoard) {
	occ1 := ((g.occ >> 8) << (b.STM << 4)) | ((g.occ << 8) >> (b.STM.Flip() << 4))
	pushable := g.self & b.Pieces[Pawn] & ^occ1
	theirSndRank := sndRank[b.STM.Flip()]
	shift := shifts[b.STM]

	// promotions pushes
	for pawns, pawn := pushable&fromMsk&theirSndRank, board.BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()
		for promo := Queen; promo > Pawn; promo-- {
			m := ms.Alloc()
			m.Piece = Pawn
			m.SetFrom(from)
			m.SetTo(from + shift)
			m.CRights = 0
			m.SetPromo(promo)
			m.Castle = 0
			m.EPP = 0
			m.EPSq = b.EnPassant
		}
	}
}

func (g generator) doublePushMoves(ms *move.Store, b *board.Board, fromMsk board.BitBoard) {
	occ1 := (g.occ >> 8) << (b.STM << 4)
	occ2 := (g.occ >> 16) << (b.STM << 5)
	pushable := g.self & b.Pieces[Pawn] & ^occ1
	mySndRank := sndRank[b.STM]
	shift := shifts[b.STM]

	// double pawn pushes
	for pawns, pawn := pushable & ^occ2 & fromMsk & mySndRank, board.BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()

		m := ms.Alloc()
		m.Piece = Pawn
		m.SetFrom(from)
		m.SetTo(from + 2*shift)
		m.CRights = 0
		m.EPSq = b.EnPassant
		if canEnPassant(b, m.To()) {
			m.EPSq ^= m.To()
		}
		m.Castle = 0
		m.SetPromo(0)
		m.EPP = 0
	}
}

// canEnPassant determines if we need to change the en passant state of the
// board after a double pawn push.
//
// This is important in order to have the right hashes for 3-fold repetation.
// If we didn't do this the next turn move generator would take care of things
// and everything would work, apart from we would have the incorrect board en
// passant state.
// https://chess.stackexchange.com/questions/777/rules-en-passant-and-draw-by-triple-repetition
func canEnPassant(b *board.Board, to Square) bool {
	target := board.BitBoard(1) << to
	them := b.Colors[b.STM.Flip()]
	shift := shifts[b.STM]
	king := b.Pieces[King] & them
	dest := board.BitBoard(1) << (to - shift)

	// pawns that are able to en-passant
	ables := ((target & ^board.AFile >> 1) | (target & ^board.HFile << 1)) & b.Pieces[Pawn] & them
	for able := board.BitBoard(0); ables != 0; ables ^= able {
		able = ables & -ables
		// remove the pawns from the occupancy
		occ := (b.Colors[White] | b.Colors[Black] | dest) &^ (target | able)
		if !IsAttacked(b, b.STM, occ, king) {
			return true
		}
	}
	return false
}

func (g generator) pawnCaptureMoves(ms *move.Store, b *board.Board) {
	var (
		occ1l, occ1r board.BitBoard
	)
	theirSndRank := sndRank[b.STM.Flip()]

	if b.STM == White {
		occ1l = (g.them &^ board.HFile) >> 7
		occ1r = (g.them &^ board.AFile) >> 9
	} else {
		occ1l = (g.them &^ board.AFile) << 7
		occ1r = (g.them &^ board.HFile) << 9
	}

	for pawns, pawn := g.self&b.Pieces[Pawn] & ^theirSndRank & (occ1l|occ1r), board.BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()

		tSqrs := PawnCaptureMoves(pawn, b.STM) & g.them

		for tSqr := board.BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()

			m := ms.Alloc()
			m.Piece = Pawn
			m.SetFrom(from)
			m.SetTo(to)
			m.CRights = 0
			m.Castle = 0
			m.SetPromo(0)
			m.EPP = 0
			m.EPSq = b.EnPassant
		}
	}
}

func (g generator) pawnCapturePromoMoves(ms *move.Store, b *board.Board) {
	var (
		occ1l, occ1r board.BitBoard
	)
	theirSndRank := sndRank[b.STM.Flip()]

	if b.STM == White {
		occ1l = (g.them &^ board.HFile) >> 7
		occ1r = (g.them &^ board.AFile) >> 9
	} else {
		occ1l = (g.them &^ board.AFile) << 7
		occ1r = (g.them &^ board.HFile) << 9
	}
	// pawn captures with promotions
	for pawns, pawn := g.self&b.Pieces[Pawn]&theirSndRank&(occ1l|occ1r), board.BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()

		tSqrs := PawnCaptureMoves(pawn, b.STM) & g.them
		for tSqr := board.BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()
			cNew := b.CRights &^ rookCRightsUpd[to]

			for promo := Queen; promo > Pawn; promo-- {
				m := ms.Alloc()
				m.Piece = Pawn
				m.SetFrom(from)
				m.SetTo(to)
				m.SetPromo(promo)
				m.CRights = cNew ^ b.CRights
				m.Castle = 0
				m.EPP = 0
				m.EPSq = b.EnPassant
			}
		}
	}
}

func (g generator) enPassant(ms *move.Store, b *board.Board) {
	shift := shifts[b.STM]

	// en-passant
	ep := (((1 << b.EnPassant) << 1) | ((1 << b.EnPassant) >> 1)) & fourthRank[b.STM.Flip()]
	for pawns, pawn := ep&g.self&b.Pieces[Pawn], board.BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()

		m := ms.Alloc()
		m.Piece = Pawn
		m.SetFrom(from)
		m.SetTo(b.EnPassant + shift)
		m.EPP = Pawn
		m.CRights = 0
		m.Castle = 0
		m.SetPromo(0)
		m.EPSq = b.EnPassant
	}
}

func (g generator) shortCastle(ms *move.Store, b *board.Board, rChkMsk board.BitBoard) {
	// castling short
	if b.CRights&CRights(C(b.STM, Short)) != 0 && g.occ&castleMask[b.STM][Short] == g.self&b.Pieces[King] {
		// this isn't quite the right condition, we would need to properly
		// calculate if the rook gives check this condition is simple, and would
		// suffice most of the time
		if castleMask[b.STM][Short]&rChkMsk != 0 {
			if !IsAttacked(b, b.STM.Flip(), g.occ, castleMask[b.STM][Short]) {
				from := (g.self & b.Pieces[King]).LowestSet()
				newC := b.CRights & ^kingCRightsUpd[b.STM]
				m := ms.Alloc()
				m.Piece = King
				m.SetFrom(from)
				m.SetTo(from + 2)
				m.Castle = C(b.STM, Short)
				m.CRights = b.CRights ^ newC
				m.SetPromo(0)
				m.EPP = 0
				m.EPSq = b.EnPassant
			}
		}
	}
}

func (g generator) longCastle(ms *move.Store, b *board.Board, rChkMsk board.BitBoard) {
	// castle long
	if b.CRights&CRights(C(b.STM, Long)) != 0 && g.occ&(castleMask[b.STM][Long]>>1) == 0 {
		if castleMask[b.STM][Long]&rChkMsk != 0 {
			if !IsAttacked(b, b.STM.Flip(), g.occ, castleMask[b.STM][Long]) {
				from := (g.self & b.Pieces[King]).LowestSet()
				newC := b.CRights & ^kingCRightsUpd[b.STM]
				m := ms.Alloc()
				m.Piece = King
				m.SetFrom(from)
				m.SetTo(from - 2)
				m.Castle = C(b.STM, Long)
				m.CRights = b.CRights ^ newC
				m.SetPromo(0)
				m.EPP = 0
				m.EPSq = b.EnPassant
			}
		}
	}
}

// GenMoves generates all pseudo-legal moves in the position.
func GenMoves(ms *move.Store, b *board.Board) {

	gen := generator{
		self: b.Colors[b.STM],
		them: b.Colors[b.STM.Flip()],
		occ:  b.Colors[White] | b.Colors[Black],
	}

	gen.kingMoves(ms, b, board.Full, board.Full)
	gen.knightMoves(ms, b, board.Full, board.Full)
	gen.bishopMoves(ms, b, board.Full, board.Full)
	gen.rookMoves(ms, b, board.Full, board.Full)
	gen.queenMoves(ms, b, board.Full, board.Full)

	gen.singlePushMoves(ms, b, board.Full)
	gen.promoPushMoves(ms, b, board.Full)
	gen.doublePushMoves(ms, b, board.Full)
	gen.pawnCaptureMoves(ms, b)
	gen.pawnCapturePromoMoves(ms, b)
	gen.enPassant(ms, b)

	gen.shortCastle(ms, b, board.Full)
	gen.longCastle(ms, b, board.Full)
}

// GenForcing generates all forcing pseudo-legal moves in the position. We do
// not guarantee that a generated move is forcing, just that all forcing moves
// are generated. But we make our best efforts to avoid quiet moves.
func GenForcing(ms *move.Store, b *board.Board) {

	self := b.Colors[b.STM]
	them := b.Colors[b.STM.Flip()]
	occ := b.Colors[White] | b.Colors[Black]

	gen := generator{
		self: self,
		them: them,
		occ:  occ,
	}

	gen.kingMoves(ms, b, board.Full, them)
	gen.knightMoves(ms, b, board.Full, them)
	gen.bishopMoves(ms, b, board.Full, them)
	gen.rookMoves(ms, b, board.Full, them)
	gen.queenMoves(ms, b, board.Full, them)

	gen.promoPushMoves(ms, b, board.Full)
	gen.pawnCaptureMoves(ms, b)
	gen.pawnCapturePromoMoves(ms, b)
	gen.enPassant(ms, b)

}

func Attackers(b *board.Board, squares board.BitBoard, occ board.BitBoard, color Color) board.BitBoard {
	opp := b.Colors[color]
	var res board.BitBoard

	for sqrs, sqBB := squares, board.BitBoard(0); sqrs != 0; sqrs ^= sqBB {
		sqBB = sqrs & -sqrs
		sq := sqBB.LowestSet()

		sub := KingMoves(sq) & b.Pieces[King]
		sub |= KnightMoves(sq) & b.Pieces[Knight]
		sub |= BishopMoves(sq, occ) & (b.Pieces[Bishop] | b.Pieces[Queen])
		sub |= RookMoves(sq, occ) & (b.Pieces[Rook] | b.Pieces[Queen])

		res |= sub & opp
	}

	res |= PawnCaptureMoves(squares, color.Flip()) & opp & b.Pieces[Pawn]

	return res
}

func Block(b *board.Board, squares board.BitBoard, color Color) board.BitBoard {
	blockers := b.Colors[color]
	res := board.BitBoard(0)
	occ := b.Colors[White] | b.Colors[Black]

	for square, eachSquare := board.BitBoard(0), squares; eachSquare != 0; eachSquare ^= square {
		square = eachSquare & -eachSquare
		sq := square.LowestSet()

		sub := board.BitBoard(0)

		/* king can't block */
		sub |= KnightMoves(sq) & b.Pieces[Knight]
		sub |= BishopMoves(sq, occ) & (b.Pieces[Bishop] | b.Pieces[Queen])
		sub |= RookMoves(sq, occ) & (b.Pieces[Rook] | b.Pieces[Queen])

		res |= sub & blockers
	}

	// we are making a pawn move backwards, so ignore the pawn in occupancy, as
	// we are moving where the actual pawn is, but don't ignore a blocking pawn
	// otherwise we would jump over it. See:
	// 6k1/8/8/1b6/3PP3/r1PKP3/2PRB3/8 w - - 0 1
	occNoPawn := occ & ^(b.Pieces[Pawn] & blockers)

	/* double pawn push blocking */
	dpawn := fourthRank[color] & squares
	dpawn = PawnSinglePushMoves(dpawn, color.Flip()) &^ occ
	dpawn = PawnSinglePushMoves(dpawn, color.Flip()) &^ occNoPawn

	res |= ((PawnSinglePushMoves(squares, color.Flip()) & ^occNoPawn) | dpawn) & blockers & b.Pieces[Pawn]

	return res
}

func IsCheckmate(b *board.Board) bool {
	king := b.Pieces[King] & b.Colors[b.STM]
	occ := b.Colors[White] | b.Colors[Black]
	opp := b.Colors[b.STM.Flip()]

	attackers := Attackers(b, king, occ, b.STM.Flip())

	if attackers == 0 {
		return false
	}

	// making the king move first
	kingSq := king.LowestSet()
	kMvs := KingMoves(kingSq) & ^b.Colors[b.STM]

	for to := board.BitBoard(0); kMvs != 0; kMvs ^= to {
		to = kMvs & -kMvs

		if !IsAttacked(b, b.STM.Flip(), occ&^king, to) {
			return false
		}
	}

	if attackers.Count() > 1 { // double check, and king can't move
		return true
	}

	attacker := attackers // only 1 attacker

	//  see if we can capture the attacker
	defenders := Attackers(b, attacker, occ, b.STM)
	// remove the king, if the king can capture the attacker it would have done
	// in the king moves try
	defenders &= ^king

	// are all my defenders pinned in a way that they can't capture the attacker
	for defender := board.BitBoard(0); defenders != 0; defenders ^= defender {
		defender = defenders & -defenders
		nocc := occ
		pinned := false

		//  dummy mk move
		nocc &= ^defender
		opp &= ^attacker

		if (BishopMoves(kingSq, nocc) & (b.Pieces[Bishop] | b.Pieces[Queen]) & opp) != 0 {
			pinned = true
		} else if (RookMoves(kingSq, nocc) & (b.Pieces[Rook] | b.Pieces[Queen]) & opp) != 0 {
			pinned = true
		}

		if !pinned {
			return false
		}
	}

	// en passant capture
	if b.EnPassant != 0 {
		epPawn := (board.BitBoard(1) << b.EnPassant)

		if epPawn == attacker {
			return false
		}
	}

	// block the attacker
	aSq := attacker.LowestSet()
	blocked := inBetween[kingSq][aSq] & ^(king | attacker)

	defenders = Block(b, blocked, b.STM)

	for defender := board.BitBoard(0); defenders != 0; defenders ^= defender {
		defender = defenders & -defenders
		nocc := occ
		opp := b.Colors[b.STM.Flip()]
		pinned := false

		//  dummy mk move
		nocc &= ^defender
		// we move somewhere on the blocked squares
		nocc |= blocked

		if (BishopMoves(kingSq, nocc) & (b.Pieces[Bishop] | b.Pieces[Queen]) & opp) != 0 {
			pinned = true
		} else if RookMoves(kingSq, nocc)&(b.Pieces[Rook]|b.Pieces[Queen])&opp != 0 {
			pinned = true
		}

		if !pinned {
			return false
		}
	}

	return true
}

func IsStalemate(b *board.Board) bool {
	me := b.Colors[b.STM]
	opp := b.Colors[b.STM.Flip()]
	king := b.Pieces[King] & me
	kingSq := king.LowestSet()
	occ := me | opp

	if IsAttacked(b, b.STM.Flip(), occ, king) {
		return false
	}

	// look at pawns guaranteed not to be pinned first
	maybePinned := (BishopMoves(kingSq, occ) | RookMoves(kingSq, occ)) & me

	// this should give an answer 99% of the time we also don't have to bother
	// with double pushes as if there is no single pawn push there can't be a
	// double pawn push
	//
	pieces := b.Pieces[Pawn] & me & ^maybePinned
	if b.STM == White {
		if pieces<<8 & ^occ != 0 {
			return false
		}

		if (((pieces & ^board.AFile)<<7)|((pieces & ^board.HFile)<<9))&opp != 0 {
			return false
		}

	} else {
		if pieces>>8 & ^occ != 0 {
			return false
		}

		if (((pieces & ^board.HFile)>>7)|((pieces & ^board.AFile)>>9))&opp != 0 {
			return false
		}
	}

	// queens can't be pinned to the extent that they can't move, for instance
	// they can always capture the pinner.
	pieces = b.Pieces[Queen] & me
	for piece := board.BitBoard(0); pieces != 0; pieces ^= piece {
		piece = pieces & -pieces
		sq := piece.LowestSet()

		if ((BishopMoves(sq, occ) | RookMoves(sq, occ)) & ^me) != 0 {
			return false
		}
	}

	// bishop can only be paralyzed by rook or queen but in case of queen not the
	// one it can capture
	pieces = b.Pieces[Bishop] & me
	for piece := board.BitBoard(0); pieces != 0; pieces ^= piece {
		piece = pieces & -pieces
		sq := piece.LowestSet()
		nocc := occ & ^piece

		if (RookMoves(kingSq, nocc) & (b.Pieces[Rook] | b.Pieces[Queen]) & opp) == 0 {
			if (BishopMoves(sq, nocc) & ^me) != 0 {
				return false
			}
		}
	}

	//  rooks can only be paralyzed by bishop or queen but in case of queen not the
	//   one it can capture
	pieces = b.Pieces[Rook] & me
	for piece := board.BitBoard(0); pieces != 0; pieces ^= piece {
		piece = pieces & -pieces
		sq := piece.LowestSet()
		nocc := occ & ^piece

		if (BishopMoves(kingSq, nocc) & (b.Pieces[Bishop] | b.Pieces[Queen]) & opp) == 0 {
			if (RookMoves(sq, nocc) & ^me) != 0 {
				return false
			}
		}
	}

	//  knight move in pins cannot be legal
	pieces = b.Pieces[Knight] & me
	for piece := board.BitBoard(0); pieces != 0; pieces ^= piece {
		piece = pieces & -pieces
		sq := piece.LowestSet()
		nocc := occ & ^piece
		pinned := false

		if (piece & maybePinned) != 0 {
			if BishopMoves(kingSq, nocc)&(b.Pieces[Bishop]|b.Pieces[Queen])&opp != 0 {
				pinned = true
			} else if RookMoves(kingSq, nocc)&(b.Pieces[Rook]|b.Pieces[Queen])&opp != 0 {
				pinned = true
			}
		}

		if !pinned && (KnightMoves(sq) & ^me != 0) {
			return false
		}
	}

	kMoves := KingMoves(kingSq) & ^me
	for kMove := board.BitBoard(0); kMoves != 0; kMoves ^= kMove {
		kMove = kMoves & -kMoves

		if !IsAttacked(b, b.STM.Flip(), occ&^king, kMove) {
			return false
		}
	}

	//  maybe pinned pawns
	pieces = b.Pieces[Pawn] & me & maybePinned
	for piece := board.BitBoard(0); pieces != 0; pieces ^= piece {
		piece = pieces & -pieces

		targets := PawnSinglePushMoves(piece, b.STM) & ^occ
		nocc := (occ & ^piece) | targets
		pinned := false

		if (BishopMoves(kingSq, nocc) & (b.Pieces[Bishop] | b.Pieces[Queen]) & opp) != 0 {
			pinned = true
		} else if (RookMoves(kingSq, nocc) & (b.Pieces[Rook] | b.Pieces[Queen]) & opp) != 0 {
			pinned = true
		}

		if !pinned && targets != 0 {
			return false
		}

		targets = PawnCaptureMoves(piece, b.STM) & opp
		nocc = (occ & ^piece) | targets
		pinned = false

		if (BishopMoves(kingSq, nocc) & (b.Pieces[Bishop] | b.Pieces[Queen]) & ^targets & opp) != 0 {
			pinned = true
		} else if (RookMoves(kingSq, nocc) & (b.Pieces[Rook] | b.Pieces[Queen]) & opp) != 0 {
			pinned = true
		}

		if !pinned && targets != 0 {
			return false
		}
	}

	//  finally deal with en passant
	if b.EnPassant != 0 {
		pieces := (((1 << b.EnPassant) << 1) | ((1 << b.EnPassant) >> 1)) & b.Pieces[Pawn] & me
		remove := board.BitBoard(1) << b.EnPassant

		for piece := board.BitBoard(0); pieces != 0; pieces ^= piece {
			piece = pieces & -pieces
			nocc := (occ & ^piece & ^remove) | PawnSinglePushMoves(remove, b.STM)
			pinned := false

			if (RookMoves(kingSq, nocc) & (b.Pieces[Rook] | b.Pieces[Queen]) & opp) != 0 {
				pinned = true
			} else if (BishopMoves(kingSq, nocc) & (b.Pieces[Bishop] | b.Pieces[Queen]) & opp) != 0 {
				pinned = true
			}

			if !pinned {
				return false
			}
		}
	}

	return true
}

func IsAttacked(b *board.Board, by Color, occ, target board.BitBoard) bool {
	other := b.Colors[by]

	// pawn capture
	if PawnCaptureMoves(b.Pieces[Pawn]&other, by)&target != 0 {
		return true
	}

	for tSqr := board.BitBoard(0); target != 0; target ^= tSqr {
		tSqr = target & -target
		sq := tSqr.LowestSet()

		if KingMoves(sq)&b.Pieces[King]&other != 0 {
			return true
		}

		if KnightMoves(sq)&b.Pieces[Knight]&other != 0 {
			return true
		}

		// bishop or queen moves
		if BishopMoves(sq, occ)&(b.Pieces[Queen]|b.Pieces[Bishop])&other != 0 {
			return true
		}

		// rook or queen moves
		if RookMoves(sq, occ)&(b.Pieces[Rook]|b.Pieces[Queen])&other != 0 {
			return true
		}
	}

	return false
}

func InCheck(b *board.Board, who Color) bool {
	return IsAttacked(b, who.Flip(), b.Colors[White]|b.Colors[Black], b.Colors[who]&b.Pieces[King])
}

func FromSimple(b *board.Board, sm move.SimpleMove) move.Move {
	pType := b.SquaresToPiece[sm.From()]
	result := move.Move{SimpleMove: sm, Piece: pType, EPSq: b.EnPassant}

	switch pType {

	case King:
		if sm.From()-sm.To() == 2 || sm.To()-sm.From() == 2 {
			newC := b.CRights & ^kingCRightsUpd[b.STM]
			result.CRights = newC ^ b.CRights
			result.Castle = C(b.STM, int(((sm.From()-sm.To())+2)/4))
		} else {
			newC := b.CRights & ^(kingCRightsUpd[b.STM] | rookCRightsUpd[sm.To()])
			result.CRights = newC ^ b.CRights
		}

	case Knight, Bishop, Queen:
		newC := b.CRights & ^(rookCRightsUpd[sm.To()])
		result.CRights = newC ^ b.CRights

	case Rook:
		newC := b.CRights & ^(rookCRightsUpd[sm.From()] | rookCRightsUpd[sm.To()])
		result.CRights = newC ^ b.CRights

	case Pawn:
		if sm.From()-sm.To() == 16 || sm.To()-sm.From() == 16 {
			if canEnPassant(b, sm.To()) {
				result.EPSq ^= sm.To()
			}
		}
		if (sm.From()-sm.To())&1 != 0 && b.SquaresToPiece[sm.To()] == NoPiece { // en-passant capture
			result.EPP = Pawn
		}
		newC := b.CRights &^ rookCRightsUpd[sm.To()]
		result.CRights = newC ^ b.CRights
	}

	return result
}
