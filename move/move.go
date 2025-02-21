package move

//revive:disable-next-line
import . "github.com/paulsonkoly/chess-3/types"

// Move represents a chess move.
type Move struct {
  // SEE is the SEE score of the move for Quiessence search. (filled in by Quiessence move ranking).
	SEE      Score
  // Weight is the heiristic weight of the move.
	Weight   Score
  // Promo is either NoPiece, or a non-Pawn Piece type, for pawn promotion.
	Promo    Piece
  // Captured is the captured piece type. Filled in by making a move, value is
  // not set by the move generator.
	Captured Piece
  // Piece is the type of piece moving.
	Piece    Piece
  // EPP is NoPiece for non en-passant moves, Pawn otherwise.
	EPP      Piece
  // From and To are the origin and destination squares for the move.
	From, To Square
  // EPSq is the bit-change in the boards en-passant state.
	EPSq     Square
  // Castle is the castling type in case of a castling move.
	Castle   Castle
  // CRights is the bit-change in the boards castling state.
	CRights  CastlingRights
  // FiftyCnt is the board's 50 move counter. Filled in by making the move on the board. 
	FiftyCnt Depth
}

func (m Move) String() string {
	if m.From|m.To == 0 {
		return "0000"
	}
	promo := ""
	if m.Promo != NoPiece {
		promo = m.Promo.String()
	}
	return m.From.String() + m.To.String() + promo
}
