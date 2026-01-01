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

func (g generator) kingMoves(moves *[]move.Weighted, b *board.Board, fromMsk, toMsk BitBoard) {
	// king moves
	if piece := g.self & b.Pieces[King] & fromMsk; piece != 0 {
		from := piece.LowestSet()

		tSqrs := attacks.KingMoves(from) & ^g.self & toMsk

		for tSqr := BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()
			*moves = append(*moves, move.Weighted{Move: move.New(from, to)})
		}
	}
}

func (g generator) knightMoves(moves *[]move.Weighted, b *board.Board, fromMsk, toMsk BitBoard) {
	// knight moves
	knights := g.self & b.Pieces[Knight] & fromMsk

	for knight := BitBoard(0); knights != 0; knights ^= knight {
		knight = knights & -knights
		from := knight.LowestSet()

		tSqrs := attacks.KnightMoves(from) & ^g.self & toMsk

		for tSqr := BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()
			*moves = append(*moves, move.Weighted{Move: move.New(from, to)})
		}
	}
}

func (g generator) bishopMoves(moves *[]move.Weighted, b *board.Board, fromMsk, toMsk BitBoard) {
	// bishop moves
	bishops := g.self & b.Pieces[Bishop] & fromMsk
	for bishop := BitBoard(0); bishops != 0; bishops ^= bishop {
		bishop = bishops & -bishops
		from := bishop.LowestSet()

		tSqrs := attacks.BishopMoves(from, g.occ) & ^g.self & toMsk

		for tSqr := BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()
			*moves = append(*moves, move.Weighted{Move: move.New(from, to)})
		}
	}
}

func (g generator) rookMoves(moves *[]move.Weighted, b *board.Board, fromMsk, toMsk BitBoard) {
	rooks := g.self & b.Pieces[Rook] & fromMsk

	for rook := BitBoard(0); rooks != 0; rooks ^= rook {
		rook = rooks & -rooks
		from := rook.LowestSet()

		tSqrs := attacks.RookMoves(from, g.occ) & ^g.self & toMsk

		for tSqr := BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()
			*moves = append(*moves, move.Weighted{Move: move.New(from, to)})
		}
	}
}

func (g generator) queenMoves(moves *[]move.Weighted, b *board.Board, fromMsk, toMsk BitBoard) {
	queens := g.self & b.Pieces[Queen] & fromMsk
	for queen := BitBoard(0); queens != 0; queens ^= queen {
		queen = queens & -queens
		from := queen.LowestSet()

		tSqrs := (attacks.BishopMoves(from, g.occ) | attacks.RookMoves(from, g.occ)) & ^g.self & toMsk

		for tSqr := BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()
			*moves = append(*moves, move.Weighted{Move: move.New(from, to)})
		}
	}
}

var shifts = [2]Square{8, -8}

func (g generator) singlePushMoves(moves *[]move.Weighted, b *board.Board, fromMsk BitBoard) {
	occ1 := (g.occ >> 8) << (b.STM << 4)
	pushable := g.self & b.Pieces[Pawn] & ^occ1
	shift := shifts[b.STM]
	mySeventhRank := RankBB(SeventhRank.FromPerspectiveOf(b.STM))

	// single pawn pushes (no promotions)
	for pawns, pawn := pushable&fromMsk & ^mySeventhRank, BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()

		*moves = append(*moves, move.Weighted{Move: move.New(from, from+shift)})
	}
}

func (g generator) promoPushMoves(moves *[]move.Weighted, b *board.Board, fromMsk BitBoard) {
	occ1 := ((g.occ >> 8) << (b.STM << 4)) | ((g.occ << 8) >> (b.STM.Flip() << 4))
	pushable := g.self & b.Pieces[Pawn] & ^occ1
	shift := shifts[b.STM]
	mySeventhRank := RankBB(SeventhRank.FromPerspectiveOf(b.STM))

	// promotions pushes
	for pawns, pawn := pushable&fromMsk&mySeventhRank, BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()
		for promo := Queen; promo > Pawn; promo-- {
			*moves = append(*moves, move.Weighted{Move: move.New(from, from+shift, move.WithPromo(promo))})
		}
	}
}

func (g generator) doublePushMoves(moves *[]move.Weighted, b *board.Board, fromMsk BitBoard) {
	occ1 := (g.occ >> 8) << (b.STM << 4)
	occ2 := (g.occ >> 16) << (b.STM << 5)
	pushable := g.self & b.Pieces[Pawn] & ^occ1
	shift := shifts[b.STM]
	mySecondRank := RankBB(SecondRank.FromPerspectiveOf(b.STM))

	// double pawn pushes
	for pawns, pawn := pushable & ^occ2 & fromMsk & mySecondRank, BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()

		*moves = append(*moves, move.Weighted{Move: move.New(from, from+2*shift)})
	}
}

func (g generator) pawnCaptureMoves(moves *[]move.Weighted, b *board.Board) {
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
	for pawn := BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()

		tSqrs := attacks.PawnCaptureMoves(pawn, b.STM) & g.them

		for tSqr := BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()

			*moves = append(*moves, move.Weighted{Move: move.New(from, to)})
		}
	}
}

func (g generator) pawnCapturePromoMoves(moves *[]move.Weighted, b *board.Board) {
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
	for pawns, pawn := g.self&b.Pieces[Pawn]&mySeventhRank&(occ1l|occ1r), BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()

		tSqrs := attacks.PawnCaptureMoves(pawn, b.STM) & g.them
		for tSqr := BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()

			for promo := Queen; promo > Pawn; promo-- {
				*moves = append(*moves, move.Weighted{Move: move.New(from, to, move.WithPromo(promo))})
			}
		}
	}
}

func (g generator) enPassant(moves *[]move.Weighted, b *board.Board) {
	if b.EnPassant == 0 {
		return
	}

	// en-passant
	ep := attacks.PawnCaptureMoves(1<<b.EnPassant, b.STM.Flip())
	for pawns, pawn := ep&g.self&b.Pieces[Pawn], BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()
		*moves = append(*moves, move.Weighted{Move: move.New(from, b.EnPassant)})
	}
}

func (g generator) shortCastle(moves *[]move.Weighted, b *board.Board, rChkMsk BitBoard) {
	// castling short
	if b.Castles&Castle(b.STM, Short) != 0 && g.occ&attacks.CastleMask[b.STM][Short] == g.self&b.Pieces[King] {
		// this isn't quite the right condition, we would need to properly
		// calculate if the rook gives check this condition is simple, and would
		// suffice most of the time
		if attacks.CastleMask[b.STM][Short]&rChkMsk != 0 {
			if !b.IsAttacked(b.STM.Flip(), g.occ, attacks.CastleMask[b.STM][Short]) {
				from := (g.self & b.Pieces[King]).LowestSet()
				*moves = append(*moves, move.Weighted{Move: move.New(from, from+2)})
			}
		}
	}
}

func (g generator) longCastle(moves *[]move.Weighted, b *board.Board, rChkMsk BitBoard) {
	// castle long
	if b.Castles&Castle(b.STM, Long) != 0 && g.occ&(attacks.CastleMask[b.STM][Long]>>1) == 0 {
		if attacks.CastleMask[b.STM][Long]&rChkMsk != 0 {
			if !b.IsAttacked(b.STM.Flip(), g.occ, attacks.CastleMask[b.STM][Long]) {
				from := (g.self & b.Pieces[King]).LowestSet()
				*moves = append(*moves, move.Weighted{Move: move.New(from, from-2)})
			}
		}
	}
}

// GenMoves generates all pseudo-legal moves in the position.
func GenMoves(moves *[]move.Weighted, b *board.Board) {

	gen := generator{
		self: b.Colors[b.STM],
		them: b.Colors[b.STM.Flip()],
		occ:  b.Colors[White] | b.Colors[Black],
	}

	gen.kingMoves(moves, b, Full, Full)
	gen.knightMoves(moves, b, Full, Full)
	gen.bishopMoves(moves, b, Full, Full)
	gen.rookMoves(moves, b, Full, Full)
	gen.queenMoves(moves, b, Full, Full)

	gen.singlePushMoves(moves, b, Full)
	gen.promoPushMoves(moves, b, Full)
	gen.doublePushMoves(moves, b, Full)
	gen.pawnCaptureMoves(moves, b)
	gen.pawnCapturePromoMoves(moves, b)
	gen.enPassant(moves, b)

	gen.shortCastle(moves, b, Full)
	gen.longCastle(moves, b, Full)
}

// GenForcing generates all forcing pseudo-legal moves in the position. We do
// not guarantee that a generated move is forcing, just that all forcing moves
// are generated. But we make our best efforts to avoid quiet moves.
func GenForcing(moves *[]move.Weighted, b *board.Board) {

	self := b.Colors[b.STM]
	them := b.Colors[b.STM.Flip()]
	occ := b.Colors[White] | b.Colors[Black]

	gen := generator{
		self: self,
		them: them,
		occ:  occ,
	}

	gen.kingMoves(moves, b, Full, them)
	gen.knightMoves(moves, b, Full, them)
	gen.bishopMoves(moves, b, Full, them)
	gen.rookMoves(moves, b, Full, them)
	gen.queenMoves(moves, b, Full, them)

	gen.promoPushMoves(moves, b, Full)
	gen.pawnCaptureMoves(moves, b)
	gen.pawnCapturePromoMoves(moves, b)
	gen.enPassant(moves, b)
}
