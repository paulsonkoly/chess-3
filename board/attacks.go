package board

import (
	"github.com/paulsonkoly/chess-3/attacks"
	. "github.com/paulsonkoly/chess-3/chess"
)

func (b *Board) Attackers(squares BitBoard, occ BitBoard, color Color) BitBoard {
	opp := b.Colors[color]
	var res BitBoard

	for sqrs := squares; sqrs != 0; sqrs &= sqrs - 1 {
		sq := sqrs.LowestSet()

		sub := attacks.KingMoves(sq) & b.Pieces[King]
		sub |= attacks.KnightMoves(sq) & b.Pieces[Knight]
		sub |= attacks.BishopMoves(sq, occ) & (b.Pieces[Bishop] | b.Pieces[Queen])
		sub |= attacks.RookMoves(sq, occ) & (b.Pieces[Rook] | b.Pieces[Queen])

		res |= sub & opp
	}

	res |= attacks.PawnCaptureMoves(squares, color.Flip()) & opp & b.Pieces[Pawn]

	return res
}

// IsStalemate determines whether the position is stalemate. The king shouldn't
// be in check.
func (b *Board) IsStalemate() bool {
	me := b.Colors[b.STM]
	opp := b.Colors[b.STM.Flip()]
	king := b.Pieces[King] & me
	kingSq := king.LowestSet()
	occ := me | opp

	// look at pawns guaranteed not to be pinned first
	maybePinned := (attacks.BishopMoves(kingSq, occ) | attacks.RookMoves(kingSq, occ)) & me

	// this should give an answer 99% of the time we also don't have to bother
	// with double pushes as if there is no single pawn push there can't be a
	// double pawn push
	//
	pawns := b.Pieces[Pawn] & me & ^maybePinned
	if b.STM == White {
		if pawns<<8 & ^occ != 0 {
			return false
		}

		if (((pawns & ^AFileBB)<<7)|((pawns & ^HFileBB)<<9))&opp != 0 {
			return false
		}

	} else {
		if pawns>>8 & ^occ != 0 {
			return false
		}

		if (((pawns & ^HFileBB)>>7)|((pawns & ^AFileBB)>>9))&opp != 0 {
			return false
		}
	}

	// queens can't be pinned to the extent that they can't move, for instance
	// they can always capture the pinner.
	for pieces := b.Pieces[Queen] & me; pieces != 0; pieces &= pieces - 1 {
		sq := pieces.LowestSet()

		if ((attacks.BishopMoves(sq, occ) | attacks.RookMoves(sq, occ)) & ^me) != 0 {
			return false
		}
	}

	// bishop can only be paralyzed by rook or queen but in case of queen not the
	// one it can capture
	for pieces := b.Pieces[Bishop] & me; pieces != 0; pieces &= pieces - 1 {
		sq := pieces.LowestSet()
		piece := pieces & -pieces
		nocc := occ & ^piece

		if (attacks.RookMoves(kingSq, nocc) & (b.Pieces[Rook] | b.Pieces[Queen]) & opp) == 0 {
			if (attacks.BishopMoves(sq, nocc) & ^me) != 0 {
				return false
			}
		}
	}

	//  rooks can only be paralyzed by bishop or queen but in case of queen not the
	//   one it can capture
	for pieces := b.Pieces[Rook] & me; pieces != 0; pieces &= pieces - 1 {
		sq := pieces.LowestSet()
		piece := pieces & -pieces
		nocc := occ & ^piece

		if (attacks.BishopMoves(kingSq, nocc) & (b.Pieces[Bishop] | b.Pieces[Queen]) & opp) == 0 {
			if (attacks.RookMoves(sq, nocc) & ^me) != 0 {
				return false
			}
		}
	}

	//  knight move in pins cannot be legal
	for pieces := b.Pieces[Knight] & me; pieces != 0; pieces &= pieces - 1 {
		sq := pieces.LowestSet()
		piece := pieces & -pieces
		nocc := occ & ^piece
		pinned := false

		if (piece & maybePinned) != 0 {
			if attacks.BishopMoves(kingSq, nocc)&(b.Pieces[Bishop]|b.Pieces[Queen])&opp != 0 {
				pinned = true
			} else if attacks.RookMoves(kingSq, nocc)&(b.Pieces[Rook]|b.Pieces[Queen])&opp != 0 {
				pinned = true
			}
		}

		if !pinned && (attacks.KnightMoves(sq) & ^me != 0) {
			return false
		}
	}

	for kMoves := attacks.KingMoves(kingSq) & ^me; kMoves != 0; kMoves &= kMoves - 1 {
		kMove := kMoves & -kMoves

		if !b.IsAttacked(b.STM.Flip(), occ&^king, kMove) {
			return false
		}
	}

	//  maybe pinned pawns

	for pawns := b.Pieces[Pawn] & me & maybePinned; pawns != 0; pawns &= pawns - 1 {
		piece := pawns & -pawns

		targets := attacks.PawnSinglePushMoves(piece, b.STM) & ^occ
		nocc := (occ & ^piece) | targets
		pinned := false

		if (attacks.BishopMoves(kingSq, nocc) & (b.Pieces[Bishop] | b.Pieces[Queen]) & opp) != 0 {
			pinned = true
		} else if (attacks.RookMoves(kingSq, nocc) & (b.Pieces[Rook] | b.Pieces[Queen]) & opp) != 0 {
			pinned = true
		}

		if !pinned && targets != 0 {
			return false
		}

		targets = attacks.PawnCaptureMoves(piece, b.STM) & opp
		nocc = (occ & ^piece) | targets
		pinned = false

		if (attacks.BishopMoves(kingSq, nocc) & (b.Pieces[Bishop] | b.Pieces[Queen]) & ^targets & opp) != 0 {
			pinned = true
		} else if (attacks.RookMoves(kingSq, nocc) & (b.Pieces[Rook] | b.Pieces[Queen]) & opp) != 0 {
			pinned = true
		}

		if !pinned && targets != 0 {
			return false
		}
	}

	//  finally deal with en passant
	if b.EnPassant != 0 {
		enPassantBB := BitBoard(1) << b.EnPassant
		pawns := attacks.PawnCaptureMoves(enPassantBB, b.STM.Flip()) & b.Pieces[Pawn] & me
		remove := attacks.PawnSinglePushMoves(enPassantBB, b.STM.Flip())

		for ; pawns != 0; pawns &= pawns - 1 {
			pawn := pawns & -pawns
			nocc := (occ & ^pawn & ^remove) | enPassantBB
			pinned := false

			if (attacks.RookMoves(kingSq, nocc) & (b.Pieces[Rook] | b.Pieces[Queen]) & opp) != 0 {
				pinned = true
			} else if (attacks.BishopMoves(kingSq, nocc) & (b.Pieces[Bishop] | b.Pieces[Queen]) & opp) != 0 {
				pinned = true
			}

			if !pinned {
				return false
			}
		}
	}

	return true
}

func (b *Board) IsAttacked(by Color, occ, target BitBoard) bool {
	other := b.Colors[by]

	// pawn capture
	if attacks.PawnCaptureMoves(b.Pieces[Pawn]&other, by)&target != 0 {
		return true
	}

	for ; target != 0; target &= target - 1 {
		sq := target.LowestSet()

		if attacks.KingMoves(sq)&b.Pieces[King]&other != 0 {
			return true
		}

		if attacks.KnightMoves(sq)&b.Pieces[Knight]&other != 0 {
			return true
		}

		// bishop or queen moves
		if attacks.BishopMoves(sq, occ)&(b.Pieces[Queen]|b.Pieces[Bishop])&other != 0 {
			return true
		}

		// rook or queen moves
		if attacks.RookMoves(sq, occ)&(b.Pieces[Rook]|b.Pieces[Queen])&other != 0 {
			return true
		}
	}

	return false
}

// InCheck determines if side is in check on b.
func (b *Board) InCheck(side Color) bool {
	return b.IsAttacked(side.Flip(), b.Colors[White]|b.Colors[Black], b.Colors[side]&b.Pieces[King])
}

// Checkers returns a BitBoard with bits set on squares with pieces giving check.
// Note: by the rules of chess the pop count in a BitBoard returned by Checkers
// should be between 0 and 2.
func (b *Board) Checkers() BitBoard {
	king := b.Colors[b.STM] & b.Pieces[King]
	return b.Attackers(king, b.Colors[White]|b.Colors[Black], b.STM.Flip())
}

var shifts = [2]Square{8, -8}

// CanEnPassant determines if we need to change the en passant state of the
// board after a double pawn push.
//
// This is important in order to have the right hashes for 3-fold repetition.
// If we didn't do this the next turn move generator would take care of things
// and everything would work, apart from we would have the incorrect board en
// passant state.
// https://chess.stackexchange.com/questions/777/rules-en-passant-and-draw-by-triple-repetition
func (b *Board) CanEnPassant(to Square) bool {
	target := BitBoard(1) << to
	them := b.Colors[b.STM.Flip()]
	shift := shifts[b.STM]
	king := b.Pieces[King] & them
	dest := BitBoard(1) << (to - shift)

	// pawns that are able to en-passant
	ables := ((target & ^AFileBB >> 1) | (target & ^HFileBB << 1)) & b.Pieces[Pawn] & them
	for ; ables != 0; ables &= ables - 1 {
		able := ables & -ables
		// remove the pawns from the occupancy
		occ := (b.Colors[White] | b.Colors[Black] | dest) &^ (target | able)
		if !b.IsAttacked(b.STM, occ, king) {
			return true
		}
	}
	return false
}
