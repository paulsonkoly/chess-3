package movegen

import (
	"github.com/paulsonkoly/chess-3/attacks"
	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"

	. "github.com/paulsonkoly/chess-3/chess"
)

func kingMoves(ms *move.Store, b *board.Board, toMsk BitBoard) {
	if piece := b.Colors[b.STM] & b.Pieces[King]; piece != 0 {
		from := piece.LowestSet()

		tSqrs := attacks.KingMoves(from) & toMsk
		for ; tSqrs != 0; tSqrs &= tSqrs - 1 {
			to := tSqrs.LowestSet()
			ms.Alloc(move.From(from) | move.To(to))
		}
	}
}

func knightMoves(ms *move.Store, b *board.Board, toMsk BitBoard) {
	for knights := b.Colors[b.STM] & b.Pieces[Knight]; knights != 0; knights &= knights - 1 {
		from := knights.LowestSet()

		tSqrs := attacks.KnightMoves(from) & toMsk
		for ; tSqrs != 0; tSqrs &= tSqrs - 1 {
			to := tSqrs.LowestSet()
			ms.Alloc(move.From(from) | move.To(to))
		}
	}
}

func bishopMoves(ms *move.Store, b *board.Board, toMsk BitBoard) {
	occ := b.Colors[White] | b.Colors[Black]
	for bishops := b.Colors[b.STM] & b.Pieces[Bishop]; bishops != 0; bishops &= bishops - 1 {
		from := bishops.LowestSet()

		tSqrs := attacks.BishopMoves(from, occ) & toMsk
		for ; tSqrs != 0; tSqrs &= tSqrs - 1 {
			to := tSqrs.LowestSet()
			ms.Alloc(move.From(from) | move.To(to))
		}
	}
}

func rookMoves(ms *move.Store, b *board.Board, toMsk BitBoard) {
	occ := b.Colors[White] | b.Colors[Black]
	for rooks := b.Colors[b.STM] & b.Pieces[Rook]; rooks != 0; rooks &= rooks - 1 {
		from := rooks.LowestSet()

		tSqrs := attacks.RookMoves(from, occ) & toMsk
		for ; tSqrs != 0; tSqrs &= tSqrs - 1 {
			to := tSqrs.LowestSet()
			ms.Alloc(move.From(from) | move.To(to))
		}
	}
}

func queenMoves(ms *move.Store, b *board.Board, toMsk BitBoard) {
	occ := b.Colors[White] | b.Colors[Black]
	for queens := b.Colors[b.STM] & b.Pieces[Queen]; queens != 0; queens &= queens - 1 {
		from := queens.LowestSet()

		tSqrs := (attacks.BishopMoves(from, occ) | attacks.RookMoves(from, occ)) & toMsk
		for ; tSqrs != 0; tSqrs &= tSqrs - 1 {
			to := tSqrs.LowestSet()
			ms.Alloc(move.From(from) | move.To(to))
		}
	}
}

var shifts = [2]Square{8, -8}

// Note: toMsk should ensure that there is no push onto occupied square.
func singlePushMoves(ms *move.Store, b *board.Board, toMsk BitBoard) {
	shifted := attacks.PawnSinglePushMoves(toMsk, b.STM.Flip())
	pushable := b.Colors[b.STM] & b.Pieces[Pawn] & shifted
	shift := shifts[b.STM]
	mySeventhRank := RankBB(SeventhRank.FromPerspectiveOf(b.STM))

	// single pawn pushes (no promotions)
	for pawns := pushable & ^mySeventhRank; pawns != 0; pawns &= pawns - 1 {
		from := pawns.LowestSet()

		ms.Alloc(move.From(from) | move.To(from+shift))
	}
}

// Note: toMsk should ensure that there is no push onto occupied square.
func promoPushMoves(ms *move.Store, b *board.Board, toMsk BitBoard) {
	shifted := attacks.PawnSinglePushMoves(toMsk, b.STM.Flip())
	pushable := b.Colors[b.STM] & b.Pieces[Pawn] & shifted
	shift := shifts[b.STM]
	mySeventhRank := RankBB(SeventhRank.FromPerspectiveOf(b.STM))

	// promotions pushes
	for pawns := pushable & mySeventhRank; pawns != 0; pawns &= pawns - 1 {
		from := pawns.LowestSet()
		for promo := Queen; promo > Pawn; promo-- {
			ms.Alloc(move.From(from) | move.To(from+shift) | move.Promo(promo))
		}
	}
}

// Note: toMsk should ensure that there is no push onto occupied square. We
// ensure that the square between source and target squares is unoccupied.
func doublePushMoves(ms *move.Store, b *board.Board, toMsk BitBoard) {
	occ := b.Colors[White] | b.Colors[Black]
	shifted1 := attacks.PawnSinglePushMoves(toMsk, b.STM.Flip())
	shifted2 := attacks.PawnSinglePushMoves(shifted1 & ^occ, b.STM.Flip())
	pushable := b.Colors[b.STM] & b.Pieces[Pawn] & shifted2
	shift := shifts[b.STM]
	mySecondRank := RankBB(SecondRank.FromPerspectiveOf(b.STM))

	// double pawn pushes
	for pawns := pushable & mySecondRank; pawns != 0; pawns &= pawns - 1 {
		from := pawns.LowestSet()
		ms.Alloc(move.From(from) | move.To(from+2*shift))
	}
}

// Note: toMsk should ensure that there is no capture onto a square not occupied by NSTM.
func pawnCaptureMoves(ms *move.Store, b *board.Board, toMsk BitBoard) {
	shifted := attacks.PawnCaptureMoves(toMsk, b.STM.Flip())
	mySeventhRank := RankBB(SeventhRank.FromPerspectiveOf(b.STM))
	pawns := b.Colors[b.STM] & b.Pieces[Pawn] & ^mySeventhRank & shifted

	for ; pawns != 0; pawns &= pawns - 1 {
		from := pawns.LowestSet()

		pawn := pawns & -pawns
		tSqrs := attacks.PawnCaptureMoves(pawn, b.STM) & toMsk
		for tSqr := BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()
			ms.Alloc(move.From(from) | move.To(to))
		}
	}
}

// Note: toMsk should ensure that there is no capture onto a square not occupied by NSTM.
func pawnCapturePromoMoves(ms *move.Store, b *board.Board, toMsk BitBoard) {
	shifted := attacks.PawnCaptureMoves(toMsk, b.STM.Flip())
	mySeventhRank := RankBB(SeventhRank.FromPerspectiveOf(b.STM))
	pawns := b.Colors[b.STM] & b.Pieces[Pawn] & mySeventhRank & shifted

	// pawn captures with promotions
	for ; pawns != 0; pawns &= pawns - 1 {
		from := pawns.LowestSet()

		pawn := pawns & -pawns
		tSqrs := attacks.PawnCaptureMoves(pawn, b.STM) & toMsk
		for tSqr := BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()

			for promo := Queen; promo > Pawn; promo-- {
				ms.Alloc(move.From(from) | move.To(to) | move.Promo(promo))
			}
		}
	}
}

func enPassant(ms *move.Store, b *board.Board) {
	if b.EnPassant == 0 {
		return
	}

	// en-passant
	ep := attacks.PawnCaptureMoves(1<<b.EnPassant, b.STM.Flip())
	for pawns := ep & b.Colors[b.STM] & b.Pieces[Pawn]; pawns != 0; pawns &= pawns - 1 {
		from := pawns.LowestSet()
		ms.Alloc(move.From(from) | move.To(b.EnPassant))
	}
}

func shortCastle(ms *move.Store, b *board.Board) {
	var castleMask BitBoard
	switch b.STM {
	case White:
		castleMask = BitBoardFromSquares(E1, F1, G1)
	case Black:
		castleMask = BitBoardFromSquares(E8, F8, G8)
	}

	occ := b.Colors[White] | b.Colors[Black]
	king := b.Colors[b.STM] & b.Pieces[King]
	if b.Castles&Castle(b.STM, Short) != 0 && occ&castleMask == king {
		// this isn't quite the right condition, we would need to properly
		// calculate if the rook gives check this condition is simple, and would
		// suffice most of the time
		if castleMask != 0 {
			if !b.IsAttacked(b.STM.Flip(), occ, castleMask) {
				from := king.LowestSet()
				ms.Alloc(move.From(from) | move.To(from+2))
			}
		}
	}
}

func longCastle(ms *move.Store, b *board.Board) {
	var castleMask BitBoard
	switch b.STM {
	case White:
		castleMask = BitBoardFromSquares(E1, D1, C1)
	case Black:
		castleMask = BitBoardFromSquares(E8, D8, C8)
	}

	occ := b.Colors[White] | b.Colors[Black]
	king := b.Colors[b.STM] & b.Pieces[King]
	if b.Castles&Castle(b.STM, Long) != 0 && occ&(castleMask>>1) == 0 {
		if castleMask != 0 {
			if !b.IsAttacked(b.STM.Flip(), occ, castleMask) {
				from := king.LowestSet()
				ms.Alloc(move.From(from) | move.To(from-2))
			}
		}
	}
}

// Noisy generates all noisy (captures/promotions) pseudo-legal moves in the
// position.
func Noisy(ms *move.Store, b *board.Board) {
	them := b.Colors[b.STM.Flip()]
	occ := b.Colors[White] | b.Colors[Black]

	kingMoves(ms, b, them)
	knightMoves(ms, b, them)
	bishopMoves(ms, b, them)
	rookMoves(ms, b, them)
	queenMoves(ms, b, them)

	promoPushMoves(ms, b, ^occ)
	pawnCaptureMoves(ms, b, them)
	pawnCapturePromoMoves(ms, b, them)
	enPassant(ms, b)
}

// Quiet generates all pseudo-legal moves not generated by Noisy.
func Quiet(ms *move.Store, b *board.Board) {
	occ := b.Colors[White] | b.Colors[Black]

	kingMoves(ms, b, ^occ)
	knightMoves(ms, b, ^occ)
	bishopMoves(ms, b, ^occ)
	rookMoves(ms, b, ^occ)
	queenMoves(ms, b, ^occ)

	singlePushMoves(ms, b, ^occ)
	doublePushMoves(ms, b, ^occ)

	shortCastle(ms, b)
	longCastle(ms, b)
}

// NoisyEvasions generates all pseudo legal noisy moves when in check by pieces
// located at checkers.
// Note: this function should be called with either 1 or 2 checkers.
func NoisyEvasions(ms *move.Store, b *board.Board, checkers BitBoard) {
	them := b.Colors[b.STM.Flip()]

	switch {
	case checkers.One():
		occ := b.Colors[White] | b.Colors[Black]

		kingMoves(ms, b, them)
		knightMoves(ms, b, checkers)
		bishopMoves(ms, b, checkers)
		rookMoves(ms, b, checkers)
		queenMoves(ms, b, checkers)

		promoPushMoves(ms, b, ^occ)
		pawnCaptureMoves(ms, b, checkers)
		pawnCapturePromoMoves(ms, b, checkers)
		enPassant(ms, b)
	case checkers.Many():
		kingMoves(ms, b, them)
	}
}

// QuietEvasions generates all pseudo legal quiet moves when in check by pieces
// located at checkers.
// Note: this function should be called with either 1 or 2 checkers.
func QuietEvasions(ms *move.Store, b *board.Board, checkers BitBoard) {
	occ := b.Colors[White] | b.Colors[Black]

	switch {
	case checkers.One():
		kingMoves(ms, b, ^occ)

		// a knight or a pawn check cannot be blocked
		if checkers&(b.Pieces[Pawn]|b.Pieces[Knight]) != 0 {
			return
		}
		checkerSq := checkers.LowestSet()
		kingSq := (b.Colors[b.STM] & b.Pieces[King]).LowestSet()
		ray := attacks.InBetween[checkerSq][kingSq]
		mask := ^occ & ray

		knightMoves(ms, b, mask)
		bishopMoves(ms, b, mask)
		rookMoves(ms, b, mask)
		queenMoves(ms, b, mask)

		singlePushMoves(ms, b, mask)
		doublePushMoves(ms, b, mask)

	case checkers.Many():
		kingMoves(ms, b, ^occ)
	}
}
