package heur

import (
	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"

	. "github.com/paulsonkoly/chess-3/types"
)

// SEE determines if the static exchange evaluation is at least the threshold of some move m.
//
// Some of this code is derived from the algorithm found in stockfish.
func SEE(b *board.Board, m move.SimpleMove, threshold Score) bool {
	from := m.From()
	to := m.To()
	fromBB := board.BitBoard(1) << from
	toBB := board.BitBoard(1) << to

	occ := (b.Colors[White] | b.Colors[Black]) ^ fromBB

	captured := b.Captured(m)
	if b.IsEnPassant(m) {
		occ &= ^(board.BitBoard(1) << b.EnPassant)
	}

	var promoVal Score
	if m.Promo() != NoPiece {
		promoVal = PieceValues[m.Promo()] - PieceValues[Pawn]
	}

	swap := PieceValues[captured] + promoVal - threshold
	if swap < 0 {
		return false
	}

	moved := b.Moved(m)

	swap = PieceValues[moved] + promoVal - swap
	if swap <= 0 {
		return true
	}

	stm := b.STM
	// Pawn capture depends on color, do pawn captures backwards. More attacks
	// will be added for sliding pieces as they are discovered with changing
	// occupancy.
	attackers :=
		(movegen.PawnCaptureMoves(toBB, Black) & b.Pieces[Pawn] & b.Colors[White]) |
			(movegen.PawnCaptureMoves(toBB, White) & b.Pieces[Pawn] & b.Colors[Black]) |
			(movegen.KnightMoves(to) & b.Pieces[Knight]) |
			(movegen.BishopMoves(to, occ) & (b.Pieces[Bishop] | b.Pieces[Queen])) |
			(movegen.RookMoves(to, occ) & (b.Pieces[Rook] | b.Pieces[Queen])) |
			(movegen.KingMoves(to) & b.Pieces[King])

	res := Score(1)

	start := [2]Piece{Pawn, Pawn}

	for {
		// dummy mkMove
		stm = stm.Flip()

		attackers &= occ
		stmAttackers := attackers & b.Colors[stm]

		if stmAttackers == 0 {
			break
		}

		res ^= 1

		// least valuable attacker
		switch start[stm] {

		case Pawn:
			fromBB = stmAttackers & b.Pieces[Pawn]
			if fromBB != 0 {
				swap = PieceValues[Pawn] - swap
				if swap < res {
					return res == 1
				}
				occ &= ^(fromBB & -fromBB)
				attackers |= (movegen.BishopMoves(to, occ) & (b.Pieces[Bishop] | b.Pieces[Queen]))
				break
			}
			fallthrough

		case Knight:
			start[stm] = Knight // no more pawns for stm

			fromBB = stmAttackers & b.Pieces[Knight]
			if fromBB != 0 {
				swap = PieceValues[Knight] - swap
				if swap < res {
					return res == 1
				}
				occ &= ^(fromBB & -fromBB)
				break
			}
			fallthrough

		case Bishop:
			// no more pawns and knights for stm
			// attackers from bishop onwards can change if occ changes, always
			// restart from Bishops
			start[stm] = Bishop

			fromBB = stmAttackers & b.Pieces[Bishop]
			if fromBB != 0 {
				swap = PieceValues[Bishop] - swap
				if swap < res {
					return res == 1
				}
				occ &= ^(fromBB & -fromBB)
				attackers |= (movegen.BishopMoves(to, occ) & (b.Pieces[Bishop] | b.Pieces[Queen]))
				break
			}

			fromBB = stmAttackers & b.Pieces[Rook]
			if fromBB != 0 {
				swap = PieceValues[Rook] - swap
				if swap < res {
					return res == 1
				}
				occ &= ^(fromBB & -fromBB)
				attackers |= (movegen.RookMoves(to, occ) & (b.Pieces[Rook] | b.Pieces[Queen]))
				break
			}

			fromBB = stmAttackers & b.Pieces[Queen]
			if fromBB != 0 {
				swap = PieceValues[Queen] - swap
				if swap < res {
					return res == 1
				}
				occ &= ^(fromBB & -fromBB)
				attackers |= (movegen.BishopMoves(to, occ) & (b.Pieces[Bishop] | b.Pieces[Queen])) |
					(movegen.RookMoves(to, occ) & (b.Pieces[Rook] | b.Pieces[Queen]))
				break
			}

			if attackers & ^b.Colors[stm] != 0 {
				return res == 0
			}
			return res == 1
		}
	}
	return res == 1
}
