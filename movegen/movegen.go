package movegen

import (
	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"

	. "github.com/paulsonkoly/chess-3/chess"
)

// KingMoves is the bitboard set where the king can move to from from. It does
// not take into accound occupancies or legality.
func KingMoves(from Square) BitBoard { return kingMoves[from] }

// KnightMoves is the bitboard set where a knight can move to from from. It does
// not take into accound occupancies or legality.
func KnightMoves(from Square) BitBoard { return knightMoves[from] }

// BishopMoves is the bitboard set where a bishop can move to from from. It
// does not take into account occupancy for the side to move, (can have bits
// set on STM's pieces), or legality.
func BishopMoves(from Square, occ BitBoard) BitBoard {
	mask := bishopMasks[from]
	magic := bishopMagics[from]
	shift := bishopShifts[from]

	return bishopAttacks[from][((occ&mask)*magic)>>(64-shift)]
}

// RookMoves is the bitbord set where a rook can move to from from. It does not
// take into account occupancy for the side to move, (can have bits set on
// STM's pieces), or legality.
func RookMoves(from Square, occ BitBoard) BitBoard {
	mask := rookMasks[from]
	magic := rookMagics[from]
	shift := rookShifts[from]

	return rookAttacks[from][((occ&mask)*magic)>>(64-shift)]
}

// PawnCaptureMoves is the bitboard set where the pawns of color color can
// capture, from any of the squares set in b.
func PawnCaptureMoves(b BitBoard, color Color) BitBoard {
	return ((((b & ^AFile) << 7) | ((b & ^HFile) << 9)) >> (color << 4)) |
		((((b & ^HFile) >> 7) | ((b & ^AFile) >> 9)) << (color.Flip() << 4))
}

// PawnSinglePushMoves is the bitboard set where the pawns of color color can
// push a single square forward from any of the squares set in b.
func PawnSinglePushMoves(b BitBoard, color Color) BitBoard {
	return ((b)<<8)>>((color)<<4) | ((b)>>8)<<((color^1)<<4)
}

var (
	// SecondRank[stm] is the second rank from stm's perspective.
	mySecondRank = [...]BitBoard{SecondRank, SeventhRank}
	myFourthRank = [...]BitBoard{FourthRank, FifthRank}
	castleMask   = [2][2]BitBoard{
		{(1 << E1) | (1 << F1) | (1 << G1), (1 << E1) | (1 << D1) | (1 << C1)},
		{(1 << E8) | (1 << F8) | (1 << G8), (1 << E8) | (1 << D8) | (1 << C8)},
	}
)

type generator struct {
	self, them, occ BitBoard
}

func (g generator) kingMoves(ms *move.Store, b *board.Board, fromMsk, toMsk BitBoard) {
	// king moves
	if piece := g.self & b.Pieces[King] & fromMsk; piece != 0 {
		from := piece.LowestSet()

		tSqrs := kingMoves[from] & ^g.self & toMsk

		for tSqr := BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()
			m := ms.Alloc()
			m.SetFrom(from)
			m.SetTo(to)
		}
	}
}

func (g generator) knightMoves(ms *move.Store, b *board.Board, fromMsk, toMsk BitBoard) {
	// knight moves
	knights := g.self & b.Pieces[Knight] & fromMsk

	for knight := BitBoard(0); knights != 0; knights ^= knight {
		knight = knights & -knights
		from := knight.LowestSet()

		tSqrs := knightMoves[from] & ^g.self & toMsk

		for tSqr := BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()
			m := ms.Alloc()
			m.SetFrom(from)
			m.SetTo(to)
		}
	}
}

func (g generator) bishopMoves(ms *move.Store, b *board.Board, fromMsk, toMsk BitBoard) {
	// bishop moves
	bishops := g.self & b.Pieces[Bishop] & fromMsk
	for bishop := BitBoard(0); bishops != 0; bishops ^= bishop {
		bishop = bishops & -bishops
		from := bishop.LowestSet()
		mask := bishopMasks[from]
		magic := bishopMagics[from]
		shift := bishopShifts[from]

		tSqrs := bishopAttacks[from][((g.occ&mask)*magic)>>(64-shift)] & ^g.self & toMsk

		for tSqr := BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()
			m := ms.Alloc()
			m.SetFrom(from)
			m.SetTo(to)
		}
	}
}

func (g generator) rookMoves(ms *move.Store, b *board.Board, fromMsk, toMsk BitBoard) {
	rooks := g.self & b.Pieces[Rook] & fromMsk

	for rook := BitBoard(0); rooks != 0; rooks ^= rook {
		rook = rooks & -rooks
		from := rook.LowestSet()
		mask := rookMasks[from]
		magic := rookMagics[from]
		shift := rookShifts[from]

		tSqrs := rookAttacks[from][((g.occ&mask)*magic)>>(64-shift)] & ^g.self & toMsk

		for tSqr := BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()
			m := ms.Alloc()
			m.SetFrom(from)
			m.SetTo(to)
		}
	}
}

func (g generator) queenMoves(ms *move.Store, b *board.Board, fromMsk, toMsk BitBoard) {
	queens := g.self & b.Pieces[Queen] & fromMsk
	for queen := BitBoard(0); queens != 0; queens ^= queen {
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

		for tSqr := BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()
			m := ms.Alloc()
			m.SetFrom(from)
			m.SetTo(to)
		}
	}
}

var shifts = [2]Square{8, -8}

func (g generator) singlePushMoves(ms *move.Store, b *board.Board, fromMsk BitBoard) {
	occ1 := (g.occ >> 8) << (b.STM << 4)
	pushable := g.self & b.Pieces[Pawn] & ^occ1
	theirSndRank := mySecondRank[b.STM.Flip()]
	shift := shifts[b.STM]

	// single pawn pushes (no promotions)
	for pawns, pawn := pushable&fromMsk & ^theirSndRank, BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()

		m := ms.Alloc()
		m.SetFrom(from)
		m.SetTo(from + shift)
	}
}

func (g generator) promoPushMoves(ms *move.Store, b *board.Board, fromMsk BitBoard) {
	occ1 := ((g.occ >> 8) << (b.STM << 4)) | ((g.occ << 8) >> (b.STM.Flip() << 4))
	pushable := g.self & b.Pieces[Pawn] & ^occ1
	theirSndRank := mySecondRank[b.STM.Flip()]
	shift := shifts[b.STM]

	// promotions pushes
	for pawns, pawn := pushable&fromMsk&theirSndRank, BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()
		for promo := Queen; promo > Pawn; promo-- {
			m := ms.Alloc()
			m.SetFrom(from)
			m.SetTo(from + shift)
			m.SetPromo(promo)
		}
	}
}

func (g generator) doublePushMoves(ms *move.Store, b *board.Board, fromMsk BitBoard) {
	occ1 := (g.occ >> 8) << (b.STM << 4)
	occ2 := (g.occ >> 16) << (b.STM << 5)
	pushable := g.self & b.Pieces[Pawn] & ^occ1
	mySndRank := mySecondRank[b.STM]
	shift := shifts[b.STM]

	// double pawn pushes
	for pawns, pawn := pushable & ^occ2 & fromMsk & mySndRank, BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()

		m := ms.Alloc()
		m.SetFrom(from)
		m.SetTo(from + 2*shift)
		m.SetEnPassant(CanEnPassant(b, m.To()))
	}
}

// CanEnPassant determines if we need to change the en passant state of the
// board after a double pawn push.
//
// This is important in order to have the right hashes for 3-fold repetition.
// If we didn't do this the next turn move generator would take care of things
// and everything would work, apart from we would have the incorrect board en
// passant state.
// https://chess.stackexchange.com/questions/777/rules-en-passant-and-draw-by-triple-repetition
func CanEnPassant(b *board.Board, to Square) bool {
	target := BitBoard(1) << to
	them := b.Colors[b.STM.Flip()]
	shift := shifts[b.STM]
	king := b.Pieces[King] & them
	dest := BitBoard(1) << (to - shift)

	// pawns that are able to en-passant
	ables := ((target & ^AFile >> 1) | (target & ^HFile << 1)) & b.Pieces[Pawn] & them
	for able := BitBoard(0); ables != 0; ables ^= able {
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
		occ1l, occ1r BitBoard
	)
	theirSndRank := mySecondRank[b.STM.Flip()]

	if b.STM == White {
		occ1l = (g.them &^ HFile) >> 7
		occ1r = (g.them &^ AFile) >> 9
	} else {
		occ1l = (g.them &^ AFile) << 7
		occ1r = (g.them &^ HFile) << 9
	}

	pawns := g.self & b.Pieces[Pawn] & ^theirSndRank & (occ1l | occ1r)
	for pawn := BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()

		tSqrs := PawnCaptureMoves(pawn, b.STM) & g.them

		for tSqr := BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()

			m := ms.Alloc()
			m.SetFrom(from)
			m.SetTo(to)
		}
	}
}

func (g generator) pawnCapturePromoMoves(ms *move.Store, b *board.Board) {
	var (
		occ1l, occ1r BitBoard
	)
	theirSndRank := mySecondRank[b.STM.Flip()]

	if b.STM == White {
		occ1l = (g.them &^ HFile) >> 7
		occ1r = (g.them &^ AFile) >> 9
	} else {
		occ1l = (g.them &^ AFile) << 7
		occ1r = (g.them &^ HFile) << 9
	}
	// pawn captures with promotions
	for pawns, pawn := g.self&b.Pieces[Pawn]&theirSndRank&(occ1l|occ1r), BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()

		tSqrs := PawnCaptureMoves(pawn, b.STM) & g.them
		for tSqr := BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()

			for promo := Queen; promo > Pawn; promo-- {
				m := ms.Alloc()
				m.SetFrom(from)
				m.SetTo(to)
				m.SetPromo(promo)
			}
		}
	}
}

func (g generator) enPassant(ms *move.Store, b *board.Board) {
	if b.EnPassant == 0 {
		return
	}

	// en-passant
	ep := PawnCaptureMoves(1<<b.EnPassant, b.STM.Flip())
	for pawns, pawn := ep&g.self&b.Pieces[Pawn], BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()

		m := ms.Alloc()
		m.SetFrom(from)
		m.SetTo(b.EnPassant)
	}
}

func (g generator) shortCastle(ms *move.Store, b *board.Board, rChkMsk BitBoard) {
	// castling short
	if b.Castles&Castle(b.STM, Short) != 0 && g.occ&castleMask[b.STM][Short] == g.self&b.Pieces[King] {
		// this isn't quite the right condition, we would need to properly
		// calculate if the rook gives check this condition is simple, and would
		// suffice most of the time
		if castleMask[b.STM][Short]&rChkMsk != 0 {
			if !IsAttacked(b, b.STM.Flip(), g.occ, castleMask[b.STM][Short]) {
				from := (g.self & b.Pieces[King]).LowestSet()
				m := ms.Alloc()
				m.SetFrom(from)
				m.SetTo(from + 2)
			}
		}
	}
}

func (g generator) longCastle(ms *move.Store, b *board.Board, rChkMsk BitBoard) {
	// castle long
	if b.Castles&Castle(b.STM, Long) != 0 && g.occ&(castleMask[b.STM][Long]>>1) == 0 {
		if castleMask[b.STM][Long]&rChkMsk != 0 {
			if !IsAttacked(b, b.STM.Flip(), g.occ, castleMask[b.STM][Long]) {
				from := (g.self & b.Pieces[King]).LowestSet()
				m := ms.Alloc()
				m.SetFrom(from)
				m.SetTo(from - 2)
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

	gen.kingMoves(ms, b, Full, Full)
	gen.knightMoves(ms, b, Full, Full)
	gen.bishopMoves(ms, b, Full, Full)
	gen.rookMoves(ms, b, Full, Full)
	gen.queenMoves(ms, b, Full, Full)

	gen.singlePushMoves(ms, b, Full)
	gen.promoPushMoves(ms, b, Full)
	gen.doublePushMoves(ms, b, Full)
	gen.pawnCaptureMoves(ms, b)
	gen.pawnCapturePromoMoves(ms, b)
	gen.enPassant(ms, b)

	gen.shortCastle(ms, b, Full)
	gen.longCastle(ms, b, Full)
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

	gen.kingMoves(ms, b, Full, them)
	gen.knightMoves(ms, b, Full, them)
	gen.bishopMoves(ms, b, Full, them)
	gen.rookMoves(ms, b, Full, them)
	gen.queenMoves(ms, b, Full, them)

	gen.promoPushMoves(ms, b, Full)
	gen.pawnCaptureMoves(ms, b)
	gen.pawnCapturePromoMoves(ms, b)
	gen.enPassant(ms, b)

}

func Attackers(b *board.Board, squares BitBoard, occ BitBoard, color Color) BitBoard {
	opp := b.Colors[color]
	var res BitBoard

	for sqrs, sqBB := squares, BitBoard(0); sqrs != 0; sqrs ^= sqBB {
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

func Block(b *board.Board, squares BitBoard, color Color) BitBoard {
	blockers := b.Colors[color]
	res := BitBoard(0)
	occ := b.Colors[White] | b.Colors[Black]

	for square, eachSquare := BitBoard(0), squares; eachSquare != 0; eachSquare ^= square {
		square = eachSquare & -eachSquare
		sq := square.LowestSet()

		sub := BitBoard(0)

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
	dpawn := myFourthRank[color] & squares
	dpawn = PawnSinglePushMoves(dpawn, color.Flip()) &^ occ
	dpawn = PawnSinglePushMoves(dpawn, color.Flip()) &^ occNoPawn

	res |= ((PawnSinglePushMoves(squares, color.Flip()) & ^occNoPawn) | dpawn) & blockers & b.Pieces[Pawn]

	return res
}

// IsCheckmate determines whether the position is checkmate. The king should be
// in check.
func IsCheckmate(b *board.Board) bool {
	king := b.Pieces[King] & b.Colors[b.STM]
	occ := b.Colors[White] | b.Colors[Black]
	opp := b.Colors[b.STM.Flip()]

	attackers := Attackers(b, king, occ, b.STM.Flip())

	// making the king move first
	kingSq := king.LowestSet()
	kMvs := KingMoves(kingSq) & ^b.Colors[b.STM]

	for to := BitBoard(0); kMvs != 0; kMvs ^= to {
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
	for defender := BitBoard(0); defenders != 0; defenders ^= defender {
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
		epPawn := PawnSinglePushMoves(BitBoard(1)<<b.EnPassant, b.STM.Flip())

		if epPawn == attacker {
			return false
		}
	}

	// block the attacker
	aSq := attacker.LowestSet()
	blocked := inBetween[kingSq][aSq] & ^(king | attacker)

	defenders = Block(b, blocked, b.STM)

	for defender := BitBoard(0); defenders != 0; defenders ^= defender {
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

// IsStalemate determines whether the position is stalemate. The king shouldn't
// be in check.
func IsStalemate(b *board.Board) bool {
	me := b.Colors[b.STM]
	opp := b.Colors[b.STM.Flip()]
	king := b.Pieces[King] & me
	kingSq := king.LowestSet()
	occ := me | opp

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

		if (((pieces & ^AFile)<<7)|((pieces & ^HFile)<<9))&opp != 0 {
			return false
		}

	} else {
		if pieces>>8 & ^occ != 0 {
			return false
		}

		if (((pieces & ^HFile)>>7)|((pieces & ^AFile)>>9))&opp != 0 {
			return false
		}
	}

	// queens can't be pinned to the extent that they can't move, for instance
	// they can always capture the pinner.
	pieces = b.Pieces[Queen] & me
	for piece := BitBoard(0); pieces != 0; pieces ^= piece {
		piece = pieces & -pieces
		sq := piece.LowestSet()

		if ((BishopMoves(sq, occ) | RookMoves(sq, occ)) & ^me) != 0 {
			return false
		}
	}

	// bishop can only be paralyzed by rook or queen but in case of queen not the
	// one it can capture
	pieces = b.Pieces[Bishop] & me
	for piece := BitBoard(0); pieces != 0; pieces ^= piece {
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
	for piece := BitBoard(0); pieces != 0; pieces ^= piece {
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
	for piece := BitBoard(0); pieces != 0; pieces ^= piece {
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
	for kMove := BitBoard(0); kMoves != 0; kMoves ^= kMove {
		kMove = kMoves & -kMoves

		if !IsAttacked(b, b.STM.Flip(), occ&^king, kMove) {
			return false
		}
	}

	//  maybe pinned pawns
	pieces = b.Pieces[Pawn] & me & maybePinned
	for piece := BitBoard(0); pieces != 0; pieces ^= piece {
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
		enPassantBB := BitBoard(1) << b.EnPassant
		pieces := PawnCaptureMoves(enPassantBB, b.STM.Flip()) & b.Pieces[Pawn] & me
		remove := PawnSinglePushMoves(enPassantBB, b.STM.Flip())

		for piece := BitBoard(0); pieces != 0; pieces ^= piece {
			piece = pieces & -pieces
			nocc := (occ & ^piece & ^remove) | enPassantBB
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

func IsAttacked(b *board.Board, by Color, occ, target BitBoard) bool {
	other := b.Colors[by]

	// pawn capture
	if PawnCaptureMoves(b.Pieces[Pawn]&other, by)&target != 0 {
		return true
	}

	for tSqr := BitBoard(0); target != 0; target ^= tSqr {
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
