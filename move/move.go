package move

//revive:disable-next-line
import . "github.com/paulsonkoly/chess-3/types"

// Move represents a chess move.
type Move struct {
	SimpleMove
	// Piece is the type of piece moving.
	Piece Piece
	// Weight is the heiristic weight of the move.
	Weight Score
	// Captured is the captured piece type. Filled in by making a move, value is
	// not set by the move generator.
	Captured Piece
	// EPP is NoPiece for non en-passant moves, Pawn otherwise.
	EPP Piece
	// EPSq is the bit-change in the boards en-passant state.
	EPSq Square
	// Castle is the castling type in case of a castling move.
	Castle Castle
	// CRights is the bit-change in the boards castling state.
	CRights CastlingRights
	// FiftyCnt is the board's 50 move counter. Filled in by making the move on the board.
	FiftyCnt Depth
}

// SimpleMove s good enough to identify a move, so it can be stored in heuristic stores.
type SimpleMove struct {
	// From and To are the origin and destination squares for the move.
	From, To Square
	// Promo is either NoPiece, or a non-Pawn Piece type, for pawn promotion.
	Promo Piece
}

// SimpleMove determines if a Move m matches a SimpleMove s.
func (s SimpleMove) Matches(m *Move) bool {
	return s == m.SimpleMove
}

func (s SimpleMove) String() string {
	if s.From|s.To == 0 {
		return "0000"
	}
	promo := ""
	if s.Promo != NoPiece {
		promo = s.Promo.String()
	}
	return s.From.String() + s.To.String() + promo
}
