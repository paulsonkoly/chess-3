package movegen

import (
	"github.com/paulsonkoly/chess-3/attacks"
	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"

	. "github.com/paulsonkoly/chess-3/chess"
)

type generator struct {
	self, them, occ         BitBoard
	mySndRank, theirSndRank BitBoard
}

func (g generator) kingMoves(ms *move.Store, b *board.Board, fromMsk, toMsk BitBoard) {
	// king moves
	if piece := g.self & b.Pieces[King] & fromMsk; piece != 0 {
		from := piece.LowestSet()

		tSqrs := attacks.KingMoves(from) & ^g.self & toMsk

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

		tSqrs := attacks.KnightMoves(from) & ^g.self & toMsk

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

		tSqrs := attacks.BishopMoves(from, g.occ) & ^g.self & toMsk

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

		tSqrs := attacks.RookMoves(from, g.occ) & ^g.self & toMsk

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

		tSqrs := (attacks.BishopMoves(from, g.occ) | attacks.RookMoves(from, g.occ)) & ^g.self & toMsk

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
	shift := shifts[b.STM]

	// single pawn pushes (no promotions)
	for pawns, pawn := pushable&fromMsk & ^g.theirSndRank, BitBoard(0); pawns != 0; pawns ^= pawn {
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
	shift := shifts[b.STM]

	// promotions pushes
	for pawns, pawn := pushable&fromMsk&g.theirSndRank, BitBoard(0); pawns != 0; pawns ^= pawn {
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
	shift := shifts[b.STM]

	// double pawn pushes
	for pawns, pawn := pushable & ^occ2 & fromMsk & g.mySndRank, BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()

		m := ms.Alloc()
		m.SetFrom(from)
		m.SetTo(from + 2*shift)
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

	pawns := g.self & b.Pieces[Pawn] & ^g.theirSndRank & (occ1l | occ1r)
	for pawn := BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()

		tSqrs := attacks.PawnCaptureMoves(pawn, b.STM) & g.them

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

	if b.STM == White {
		occ1l = (g.them &^ HFileBB) >> 7
		occ1r = (g.them &^ AFileBB) >> 9
	} else {
		occ1l = (g.them &^ AFileBB) << 7
		occ1r = (g.them &^ HFileBB) << 9
	}
	// pawn captures with promotions
	for pawns, pawn := g.self&b.Pieces[Pawn]&g.theirSndRank&(occ1l|occ1r), BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()

		tSqrs := attacks.PawnCaptureMoves(pawn, b.STM) & g.them
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
	ep := attacks.PawnCaptureMoves(1<<b.EnPassant, b.STM.Flip())
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
	if b.Castles&Castle(b.STM, Short) != 0 && g.occ&attacks.CastleMask[b.STM][Short] == g.self&b.Pieces[King] {
		// this isn't quite the right condition, we would need to properly
		// calculate if the rook gives check this condition is simple, and would
		// suffice most of the time
		if attacks.CastleMask[b.STM][Short]&rChkMsk != 0 {
			if !b.IsAttacked(b.STM.Flip(), g.occ, attacks.CastleMask[b.STM][Short]) {
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
	if b.Castles&Castle(b.STM, Long) != 0 && g.occ&(attacks.CastleMask[b.STM][Long]>>1) == 0 {
		if attacks.CastleMask[b.STM][Long]&rChkMsk != 0 {
			if !b.IsAttacked(b.STM.Flip(), g.occ, attacks.CastleMask[b.STM][Long]) {
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
		self:         b.Colors[b.STM],
		them:         b.Colors[b.STM.Flip()],
		occ:          b.Colors[White] | b.Colors[Black],
		mySndRank:    RankBB(SecondRank.FromPerspectiveOf(b.STM)),
		theirSndRank: RankBB(SecondRank.FromPerspectiveOf(b.STM.Flip())),
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
		self:         self,
		them:         them,
		occ:          occ,
		mySndRank:    RankBB(SecondRank.FromPerspectiveOf(b.STM)),
		theirSndRank: RankBB(SecondRank.FromPerspectiveOf(b.STM.Flip())),
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
