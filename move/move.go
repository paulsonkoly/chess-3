package move

import . "github.com/paulsonkoly/chess-3/chess"

// Weighted represents a weighted chess move.
type Weighted struct {
	Move
	// Weight is the heuristic weight of the move.
	Weight Score
}

// Move represents a chess move, it contains the to and from squares and the
// promotion piece type. Additionally it contains an en-passant flag indicating
// that the move is a double pawn-push, that should assign new en-passant state
// to the board.
type Move uint16

const (
	toMsk      = Move(1<<6 - 1)
	toShift    = 0
	fromMsk    = Move((1<<6 - 1) << 6)
	fromShift  = 6
	promoMsk   = Move((1<<3 - 1) << 12)
	promoShift = 12
)

// From constructs a move that has the from square set.
func From(from Square) Move { return (Move(from) << fromShift) & fromMsk }

// To constructs a move that has the to square set.
func To(to Square) Move { return (Move(to) << toShift) & toMsk }

// Promo constructs a move that has the promotion piece type set.
func Promo(p Piece) Move { return (Move(p) << promoShift) & promoMsk }

// To is the target square of the move.
func (s Move) To() Square { return Square((s & toMsk) >> toShift) }

// From is the source square of the move.
func (s Move) From() Square { return Square((s & fromMsk) >> fromShift) }

// Promo is the promotion piece of the move.
func (s Move) Promo() Piece { return Piece((s & promoMsk) >> promoShift) }

// Matches determines if a Move m matches s.
func (s Move) Matches(m *Weighted) bool {
	return s == m.Move
}

// String representation of s, following uci move notation.
func (s Move) String() string {
	if s == 0 {
		return "0000"
	}
	return s.From().String() + s.To().String() + s.Promo().String()
}
