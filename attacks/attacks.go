// Package attacks provides low level bitboard pattern calculations like piece type attacks.
package attacks

import (
	. "github.com/paulsonkoly/chess-3/chess"
)

// KingMoves is the bitboard set where the king can move to from from. It does
// not take into account occupancies or legality or castling.
func KingMoves(from Square) BitBoard { return kingMoves[from] }

// KnightMoves is the bitboard set where a knight can move to from from. It does
// not take into account occupancies or legality.
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
	// CastleMask[color][side] is the set of affected cells in castling for color
	// and side of Short / Long.
	CastleMask = [2][2]BitBoard{
		{(1 << E1) | (1 << F1) | (1 << G1), (1 << E1) | (1 << D1) | (1 << C1)},
		{(1 << E8) | (1 << F8) | (1 << G8), (1 << E8) | (1 << D8) | (1 << C8)},
	}
)
