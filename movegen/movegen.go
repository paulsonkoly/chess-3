package movegen

import (
	"github.com/paulsonkoly/chess-3/attacks"
	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"

	. "github.com/paulsonkoly/chess-3/chess"
)

type generator struct {
	self, them, occ BitBoard
}

func (g generator) kingMoves(ms *move.Store, b *board.Board, fromMsk, toMsk BitBoard) {
	if piece := g.self & b.Pieces[King] & fromMsk; piece != 0 {
		from := piece.LowestSet()

		tSqrs := attacks.KingMoves(from) & ^g.self & toMsk
		for ; tSqrs != 0; tSqrs &= tSqrs - 1 {
			to := tSqrs.LowestSet()
			ms.Alloc(move.From(from) | move.To(to))
		}
	}
}

func (g generator) knightMoves(ms *move.Store, b *board.Board, fromMsk, toMsk BitBoard) {
	knights := g.self & b.Pieces[Knight] & fromMsk
	for ; knights != 0; knights &= knights - 1 {
		from := knights.LowestSet()

		tSqrs := attacks.KnightMoves(from) & ^g.self & toMsk
		for ; tSqrs != 0; tSqrs &= tSqrs - 1 {
			to := tSqrs.LowestSet()
			ms.Alloc(move.From(from) | move.To(to))
		}
	}
}

func (g generator) bishopMoves(ms *move.Store, b *board.Board, fromMsk, toMsk BitBoard) {
	bishops := g.self & b.Pieces[Bishop] & fromMsk
	for ; bishops != 0; bishops &= bishops - 1 {
		from := bishops.LowestSet()

		tSqrs := attacks.BishopMoves(from, g.occ) & ^g.self & toMsk
		for ; tSqrs != 0; tSqrs &= tSqrs - 1 {
			to := tSqrs.LowestSet()
			ms.Alloc(move.From(from) | move.To(to))
		}
	}
}

func (g generator) rookMoves(ms *move.Store, b *board.Board, fromMsk, toMsk BitBoard) {
	rooks := g.self & b.Pieces[Rook] & fromMsk
	for ; rooks != 0; rooks &= rooks - 1 {
		from := rooks.LowestSet()

		tSqrs := attacks.RookMoves(from, g.occ) & ^g.self & toMsk
		for ; tSqrs != 0; tSqrs &= tSqrs - 1 {
			to := tSqrs.LowestSet()
			ms.Alloc(move.From(from) | move.To(to))
		}
	}
}

func (g generator) queenMoves(ms *move.Store, b *board.Board, fromMsk, toMsk BitBoard) {
	queens := g.self & b.Pieces[Queen] & fromMsk
	for ; queens != 0; queens &= queens - 1 {
		from := queens.LowestSet()

		tSqrs := (attacks.BishopMoves(from, g.occ) | attacks.RookMoves(from, g.occ)) & ^g.self & toMsk
		for ; tSqrs != 0; tSqrs &= tSqrs - 1 {
			to := tSqrs.LowestSet()
			ms.Alloc(move.From(from) | move.To(to))
		}
	}
}

var shifts = [2]Square{8, -8}

func (g generator) singlePushMoves(ms *move.Store, b *board.Board, fromMsk BitBoard) {
	occ1 := (g.occ >> 8) << (b.STM << 4)
	pushable := g.self & b.Pieces[Pawn] & ^occ1
	shift := shifts[b.STM]
	mySeventhRank := RankBB(SeventhRank.FromPerspectiveOf(b.STM))

	// single pawn pushes (no promotions)
	for pawns := pushable & fromMsk & ^mySeventhRank; pawns != 0; pawns &= pawns - 1 {
		from := pawns.LowestSet()

		ms.Alloc(move.From(from) | move.To(from+shift))
	}
}

func (g generator) promoPushMoves(ms *move.Store, b *board.Board, fromMsk BitBoard) {
	occ1 := ((g.occ >> 8) << (b.STM << 4)) | ((g.occ << 8) >> (b.STM.Flip() << 4))
	pushable := g.self & b.Pieces[Pawn] & ^occ1
	shift := shifts[b.STM]
	mySeventhRank := RankBB(SeventhRank.FromPerspectiveOf(b.STM))

	// promotions pushes
	for pawns := pushable & fromMsk & mySeventhRank; pawns != 0; pawns &= pawns - 1 {
		from := pawns.LowestSet()
		for promo := Queen; promo > Pawn; promo-- {
			ms.Alloc(move.From(from) | move.To(from+shift) | move.Promo(promo))
		}
	}
}

func (g generator) doublePushMoves(ms *move.Store, b *board.Board, fromMsk BitBoard) {
	occ1 := (g.occ >> 8) << (b.STM << 4)
	occ2 := (g.occ >> 16) << (b.STM << 5)
	pushable := g.self & b.Pieces[Pawn] & ^occ1
	shift := shifts[b.STM]
	mySecondRank := RankBB(SecondRank.FromPerspectiveOf(b.STM))

	// double pawn pushes
	for pawns := pushable & ^occ2 & fromMsk & mySecondRank; pawns != 0; pawns &= pawns - 1 {
		from := pawns.LowestSet()
		ms.Alloc(move.From(from) | move.To(from+2*shift))
	}
}

func (g generator) pawnCaptureMoves(ms *move.Store, b *board.Board) {
	var (
		occ1l, occ1r BitBoard
	)

	if b.STM == White {
		occ1l = (g.them &^ HFileBB) >> 7
		occ1r = (g.them &^ AFileBB) >> 9
	} else {
		occ1l = (g.them &^ AFileBB) << 7
		occ1r = (g.them &^ HFileBB) << 9
	}

	mySeventhRank := RankBB(SeventhRank.FromPerspectiveOf(b.STM))

	pawns := g.self & b.Pieces[Pawn] & ^mySeventhRank & (occ1l | occ1r)
	for ; pawns != 0; pawns &= pawns - 1 {
		from := pawns.LowestSet()

		pawn := pawns & -pawns
		tSqrs := attacks.PawnCaptureMoves(pawn, b.STM) & g.them
		for tSqr := BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()
			ms.Alloc(move.From(from) | move.To(to))
		}
	}
}

func (g generator) pawnCapturePromoMoves(ms *move.Store, b *board.Board) {
	var (
		occ1l, occ1r BitBoard
	)

	if b.STM == White {
		occ1l = (g.them &^ HFileBB) >> 7
		occ1r = (g.them &^ AFileBB) >> 9
	} else {
		occ1l = (g.them &^ AFileBB) << 7
		occ1r = (g.them &^ HFileBB) << 9
	}
	mySeventhRank := RankBB(SeventhRank.FromPerspectiveOf(b.STM))
	// pawn captures with promotions
	for pawns := g.self & b.Pieces[Pawn] & mySeventhRank & (occ1l | occ1r); pawns != 0; pawns &= pawns - 1 {
		from := pawns.LowestSet()

		pawn := pawns & -pawns
		tSqrs := attacks.PawnCaptureMoves(pawn, b.STM) & g.them
		for tSqr := BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()

			for promo := Queen; promo > Pawn; promo-- {
				ms.Alloc(move.From(from) | move.To(to) | move.Promo(promo))
			}
		}
	}
}

func (g generator) enPassant(ms *move.Store, b *board.Board) {
	if b.EnPassant == 0 {
		return
	}

	// en-passant
	ep := attacks.PawnCaptureMoves(1<<b.EnPassant, b.STM.Flip())
	for pawns := ep & g.self & b.Pieces[Pawn]; pawns != 0; pawns &= pawns - 1 {
		from := pawns.LowestSet()
		ms.Alloc(move.From(from) | move.To(b.EnPassant))
	}
}

func (g generator) shortCastle(ms *move.Store, b *board.Board, rChkMsk BitBoard) {
	// castling short
	if b.Castles&Castle(b.STM, Short) != 0 && g.occ&attacks.CastleMask[b.STM][Short] == g.self&b.Pieces[King] {
		// this isn't quite the right condition, we would need to properly
		// calculate if the rook gives check this condition is simple, and would
		// suffice most of the time
		if attacks.CastleMask[b.STM][Short]&rChkMsk != 0 {
			if !b.IsAttacked(b.STM.Flip(), g.occ, attacks.CastleMask[b.STM][Short]) {
				from := (g.self & b.Pieces[King]).LowestSet()
				ms.Alloc(move.From(from) | move.To(from+2))
			}
		}
	}
}

func (g generator) longCastle(ms *move.Store, b *board.Board, rChkMsk BitBoard) {
	// castle long
	if b.Castles&Castle(b.STM, Long) != 0 && g.occ&(attacks.CastleMask[b.STM][Long]>>1) == 0 {
		if attacks.CastleMask[b.STM][Long]&rChkMsk != 0 {
			if !b.IsAttacked(b.STM.Flip(), g.occ, attacks.CastleMask[b.STM][Long]) {
				from := (g.self & b.Pieces[King]).LowestSet()
				ms.Alloc(move.From(from) | move.To(from-2))
			}
		}
	}
}

// GenNoisy generates all noisy (captures/promotions) pseudo-legal moves in the
// position.
func GenNoisy(ms *move.Store, b *board.Board) {

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

// GenNotNoisy generates all psudo legal moves not generated by GenNoisy.
func GenNotNoisy(ms *move.Store, b *board.Board) {

	self := b.Colors[b.STM]
	them := b.Colors[b.STM.Flip()]
	occ := b.Colors[White] | b.Colors[Black]

	gen := generator{
		self: self,
		them: them,
		occ:  occ,
	}

	gen.kingMoves(ms, b, Full, ^them)
	gen.knightMoves(ms, b, Full, ^them)
	gen.bishopMoves(ms, b, Full, ^them)
	gen.rookMoves(ms, b, Full, ^them)
	gen.queenMoves(ms, b, Full, ^them)

	gen.singlePushMoves(ms, b, Full)
	gen.doublePushMoves(ms, b, Full)

	gen.shortCastle(ms, b, Full)
	gen.longCastle(ms, b, Full)
}
