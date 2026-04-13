package board

import (
	"errors"

	"github.com/paulsonkoly/chess-3/attacks"
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/move"
)

var (
	ErrWrongPieceCount = errors.New("wrong piece count")
	ErrWrongCastle     = errors.New("wrong castle")
	ErrWrongEnPassant  = errors.New("wrong en-passant")
	ErrNSTMInCheck     = errors.New("non side to move in check")
)

// Valid determines if the position is legal (reachable from startpos) in chess.
// Useful for validating a syntactically correct fen which might yield an invalid position.
// Returns error nil if the position is valid.
func (b Board) Valid() error {
	for color := White; color <= Black; color++ {
		if !(b.Colors[color] & b.Pieces[King]).IsPow2() {
			return ErrWrongPieceCount
		}
		knights := b.Counts[color][Knight]
		bishops := b.Counts[color][Bishop]
		rooks := b.Counts[color][Rook]
		queens := b.Counts[color][Queen]
		pawns := b.Counts[color][Pawn]

		// Compute the number of pieces that are guaranteed to be promoted Pawns.
		pknights := max(2, knights) - 2
		pbishops := max(2, bishops) - 2
		prooks := max(2, rooks) - 2
		pqueens := max(1, queens) - 1
		promoted := pknights + pbishops + prooks + pqueens

		pawns += promoted
		if (pawns > 8) || (knights+pawns-pknights > 10) || (bishops+pawns-pbishops > 10) ||
			(rooks+pawns-prooks > 10) || (queens+pawns-pqueens > 9) {
			return ErrWrongPieceCount
		}

		kingSq := SquareAt(EFile, FirstRank.FromPerspectiveOf(color))
		if b.Castles&Castle(color, Short) != 0 {
			rookSq := SquareAt(HFile, FirstRank.FromPerspectiveOf(color))
			if b.Pieces[Rook]&b.Colors[color]&BitBoardFromSquares(rookSq) == 0 ||
				b.Pieces[King]&b.Colors[color]&BitBoardFromSquares(kingSq) == 0 {
				return ErrWrongCastle
			}
		}

		if b.Castles&Castle(color, Long) != 0 {
			rookSq := SquareAt(AFile, FirstRank.FromPerspectiveOf(color))
			if b.Pieces[Rook]&b.Colors[color]&BitBoardFromSquares(rookSq) == 0 ||
				b.Pieces[King]&b.Colors[color]&BitBoardFromSquares(kingSq) == 0 {
				return ErrWrongCastle
			}
		}
	}

	if b.InCheck(b.STM.Flip()) {
		return ErrNSTMInCheck
	}

	if b.EnPassant != 0 {
		if b.EnPassant.Rank() != SixthRank.FromPerspectiveOf(b.STM) {
			return ErrWrongEnPassant
		}

		occ := b.Colors[White] | b.Colors[Black]
		epBB := BitBoardFromSquares(b.EnPassant)
		capturingBB := attacks.PawnCaptureMoves(epBB, b.STM.Flip()) & b.Pieces[Pawn] & b.Colors[b.STM]
		capturedBB := attacks.PawnSinglePushMoves(epBB, b.STM.Flip()) & b.Pieces[Pawn] & b.Colors[b.STM.Flip()]
		if capturingBB == 0 || capturedBB == 0 || epBB&occ != 0 {
			return ErrWrongEnPassant
		}

		hasUnpinned := false
		for capturers := capturingBB; capturers != 0; capturers &= capturers - 1 {
			capturer := capturers & -capturers
			occMod := occ & ^(capturer|capturedBB) | epBB

			if !b.IsAttacked(b.STM.Flip(), occMod, b.Pieces[King]&b.Colors[b.STM]) {
				hasUnpinned = true
				break
			}
		}
		if !hasUnpinned {
			return ErrWrongEnPassant
		}
	}
	return nil
}

// IsPseudoLegal determines if m is pseudo legal in the position.
// Useful for checking hash moves, or validating moves from UCI.
func (b *Board) IsPseudoLegal(m move.Move) bool {
	from := m.From()
	fromBB := BitBoard(1) << from
	to := m.To()
	toBB := BitBoard(1) << to

	if b.Colors[b.STM]&fromBB == 0 {
		return false
	}
	if b.Colors[b.STM]&toBB != 0 {
		return false
	}

	piece := b.SquaresToPiece[from]
	occ := b.Colors[White] | b.Colors[Black]

	if m.Promo() != NoPiece && piece != Pawn {
		return false
	}

	switch piece {

	case Knight:
		if attacks.KnightMoves(from)&toBB == 0 {
			return false
		}

	case Bishop:
		if attacks.BishopMoves(from, occ)&toBB == 0 {
			return false
		}

	case Rook:
		if attacks.RookMoves(from, occ)&toBB == 0 {
			return false
		}

	case Queen:
		if (attacks.RookMoves(from, occ)|attacks.BishopMoves(from, occ))&toBB == 0 {
			return false
		}

	case King:
		switch {

		case from == E1 && to == G1 && b.STM == White:
			if b.Castles&ShortWhite == 0 || BitBoardFromSquares(F1, G1)&occ != 0 ||
				b.IsAttacked(b.STM.Flip(), occ, BitBoardFromSquares(E1, F1, G1)) {
				return false
			}

		case from == E1 && to == C1 && b.STM == White:
			if b.Castles&LongWhite == 0 || BitBoardFromSquares(D1, C1, B1)&occ != 0 ||
				b.IsAttacked(b.STM.Flip(), occ, BitBoardFromSquares(E1, D1, C1)) {
				return false
			}

		case from == E8 && to == G8 && b.STM == Black:
			if b.Castles&ShortBlack == 0 || BitBoardFromSquares(F8, G8)&occ != 0 ||
				b.IsAttacked(b.STM.Flip(), occ, BitBoardFromSquares(E8, F8, G8)) {
				return false
			}

		case from == E8 && to == C8 && b.STM == Black:
			if b.Castles&LongBlack == 0 || BitBoardFromSquares(D8, C8, B8)&occ != 0 ||
				b.IsAttacked(b.STM.Flip(), occ, BitBoardFromSquares(E8, D8, C8)) {
				return false
			}

		default:
			if attacks.KingMoves(from)&toBB == 0 {
				return false
			}
		}

	case Pawn:

		if (from < to && b.STM == Black) || (from > to && b.STM == White) {
			return false
		}

		if RankBB(SeventhRank.FromPerspectiveOf(b.STM))&fromBB != 0 {
			if m.Promo() == NoPiece {
				return false
			}
		}

		switch Abs(from.File() - to.File()) {

		case 0: // pawn pushing

			switch Abs(from.Rank() - to.Rank()) {

			case 1: // single pawn push
				if occ&toBB != 0 {
					return false
				}

			case 2: // double pawn push
				if fromBB&RankBB(SecondRank.FromPerspectiveOf(b.STM)) == 0 {
					return false
				}

				if occ&(toBB|(BitBoard(1)<<((from+to)/2))) != 0 {
					return false
				}

			default:
				return false
			}

		case 1: // pawn capturing
			if Abs(from.Rank()-to.Rank()) != 1 {
				return false
			}

			enPassant := BitBoard(0)
			if b.EnPassant != 0 {
				enPassant = BitBoard(1) << b.EnPassant
			}
			if (b.Colors[b.STM.Flip()]|enPassant)&toBB == 0 {
				return false
			}

		default:
			return false
		}
	}

	return true
}
