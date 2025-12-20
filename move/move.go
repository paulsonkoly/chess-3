package move

import . "github.com/paulsonkoly/chess-3/types"

// Move represents a chess move.
type Move struct {
	SimpleMove
	// Weight is the heuristic weight of the move.
	Weight Score
}

// SimpleMove s good enough to identify a move, so it can be stored in
// heuristic stores. It encodes from and to squares and promotion piece.
type SimpleMove uint16

// SimpleMoveOption is an optional argument to NewSimple.
type SimpleMoveOption interface {
	Apply(sm SimpleMove) SimpleMove
}

// WithPromo is a SimpleMoveOption setting the promotion piece type.
type WithPromo Piece

func (p WithPromo) Apply(sm SimpleMove) SimpleMove {
	sm.SetPromo(Piece(p))
	return sm
}

// NewSimple creates a new simple move with to and from squares and additional options.
func NewSimple(from, to Square, opts ...SimpleMoveOption) SimpleMove {
	sm := (SimpleMove(to) << toShift & toMsk) | (SimpleMove(from) << fromShift & fromMsk)
	for _, opt := range opts {
		sm = opt.Apply(sm)
	}
	return sm
}

const (
	toMsk        = SimpleMove(1<<6 - 1)
	toShift      = 0
	fromMsk      = SimpleMove((1<<6 - 1) << 6)
	fromShift    = 6
	promoMsk     = SimpleMove((1<<3 - 1) << 12)
	promoShift   = 12
	enPassantMsk = SimpleMove((1<<1 - 1) << 15)
)

// To is the target square of the move.
func (s SimpleMove) To() Square { return Square((s & toMsk) >> toShift) }

// SetTo sets the target square of the move.
func (s *SimpleMove) SetTo(sq Square) { *s = (*s & ^toMsk) | (SimpleMove(sq) << toShift & toMsk) }

// From is the source square of the move.
func (s SimpleMove) From() Square { return Square((s & fromMsk) >> fromShift) }

// SetFrom sets the source square of the move.
func (s *SimpleMove) SetFrom(sq Square) { *s = (*s & ^fromMsk) | SimpleMove(sq)<<fromShift&fromMsk }

// Promo is the promotion piece of the move.
func (s SimpleMove) Promo() Piece { return Piece((s & promoMsk) >> promoShift) }

// SetPromo sets the promotion piece of the move.
func (s *SimpleMove) SetPromo(p Piece) { *s = (*s & ^promoMsk) | SimpleMove(p)<<promoShift&promoMsk }

// EnPassant indicates that this is a double pawn push changing the en passant state.
func (s SimpleMove) EnPassant() bool { return (s & enPassantMsk) != 0 }

// SetEnPassant sets the en passant flag.
func (s *SimpleMove) SetEnPassant(ep bool) {
	flag := SimpleMove(0)
	if ep {
		flag = enPassantMsk
	}
	*s = (*s & ^enPassantMsk) | flag
}

// Matches determines if a Move m matches a SimpleMove s.
func (s SimpleMove) Matches(m *Move) bool {
	return s == m.SimpleMove
}

// String representation of s, following uci move notation.
func (s SimpleMove) String() string {
	if s == 0 {
		return "0000"
	}
	return s.From().String() + s.To().String() + s.Promo().String()
}
