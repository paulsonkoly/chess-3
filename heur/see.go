package heur

import (
	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

var capturesStore = [64]Piece{}

// SEE is static exchange evaluation of m.
func SEE(b *board.Board, m *move.Move) Score {
	fromBB, to := board.BitBoard(1)<<m.From, m.To
	toBB := board.BitBoard(1) << m.To

	// attackers of square "to"
	attackers := [2][7]board.BitBoard{
		// White
		{
			0, // NoPiece
			movegen.PawnCaptureMoves(toBB, Black) & b.Pieces[Pawn] & b.Colors[White],
			movegen.KnightMoves(to) & b.Pieces[Knight] & b.Colors[White],
			0, // bishops
			0, // rooks
			0, // queens
			movegen.KingMoves(to) & b.Pieces[King] & b.Colors[White],
		},
		// Black
		{
			0, // NoPiece
			movegen.PawnCaptureMoves(toBB, White) & b.Pieces[Pawn] & b.Colors[Black],
			movegen.KnightMoves(to) & b.Pieces[Knight] & b.Colors[Black],
			0, // bishops
			0, // rooks
			0, // queens
			movegen.KingMoves(to) & b.Pieces[King] & b.Colors[Black],
		},
	}

	captures := capturesStore[:0]

	// piece type of least valueable attacker per side
	start := [2]Piece{Pawn, Pawn}
	piece := m.Piece
	occ := b.Colors[White] | b.Colors[Black]
	stm := b.STM

	captures = append(captures, b.SquaresToPiece[m.To], piece)

	// dummy mkMove
	attackers[stm][piece] &= ^fromBB
	occ &= ^fromBB
	stm = stm.Flip()

	for {
		// least valueable attacker
		switch start[stm] {

		case Pawn:
			fromBB = attackers[stm][start[stm]]
			if fromBB != 0 {
				piece = Pawn
				fromBB &= -fromBB
				break
			}
			fallthrough

		case Knight:
			start[stm] = Knight

			fromBB = attackers[stm][start[stm]]
			if fromBB != 0 {
				piece = Knight
				fromBB &= -fromBB
				break
			}
			fallthrough

		case Bishop:
			start[stm] = Bishop // these attackers can change if occ changes, always restart from Bishops

			if attackers[stm][Bishop] == 0 {
				attackers[stm][Bishop] = movegen.BishopMoves(to, occ) & occ & b.Pieces[Bishop] & b.Colors[stm]
			}

			fromBB = attackers[stm][Bishop]
			if fromBB != 0 {
				piece = Bishop
				fromBB &= -fromBB
				break
			}
			fallthrough

		case Rook:
			if attackers[stm][Rook] == 0 {
				attackers[stm][Rook] = movegen.RookMoves(to, occ) & occ & b.Pieces[Rook] & b.Colors[stm]
			}

			fromBB = attackers[stm][Rook]
			if fromBB != 0 {
				piece = Rook
				fromBB &= -fromBB
				break
			}
			fallthrough

		case Queen:
			if attackers[stm][Queen] == 0 {
				attackers[stm][Queen] = (movegen.BishopMoves(to, occ) | movegen.RookMoves(to, occ)) & occ & b.Pieces[Queen] & b.Colors[stm]
			}

			fromBB = attackers[stm][Queen]
			if fromBB != 0 {
				piece = Queen
				fromBB &= -fromBB
				break
			}
			fallthrough

		case King:
			fromBB = attackers[stm][King]
			if fromBB != 0 {
				piece = King
				fromBB &= -fromBB
				break
			}
			fallthrough

		default:
			value := Score(0)

			// ignore the last piece that's not captured
			for ply := len(captures) - 2; ply >= 0; ply-- {
				value = max(value, 0)
				value = PieceValues[captures[ply]] - value
			}

			return value
		}

		// dummy mkmove
		captures = append(captures, piece)
		attackers[stm][piece] &= ^fromBB
		occ &= ^fromBB
		stm = stm.Flip()
	}
}
