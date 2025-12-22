package move

import . "github.com/paulsonkoly/chess-3/chess"

// Move represents a chess move, it contains the to and from squares and the
// promotion piece type. Additionally it contains an en-passant flag indicating
// that the move is a double pawn-push, that should assign new en-passant state
// to the board.
type Move uint16

// MoveOption is an optional argument to New.
type MoveOption interface {
	Apply(sm Move) Move
}

// WithPromo is a MoveOption setting the promotion piece type.
type WithPromo Piece

func (p WithPromo) Apply(sm Move) Move {
	sm.SetPromo(Piece(p))
	return sm
}


// New creates a new move with to and from squares and additional options.
func New(from, to Square, opts ...MoveOption) Move {
	sm := (Move(to) << toShift & toMsk) | (Move(from) << fromShift & fromMsk)
	for _, opt := range opts {
		sm = opt.Apply(sm)
	}
	return sm
}

const (
	toMsk        = Move(1<<6 - 1)
	toShift      = 0
	fromMsk      = Move((1<<6 - 1) << 6)
	fromShift    = 6
	promoMsk     = Move((1<<3 - 1) << 12)
	promoShift   = 12
)

// To is the target square of the move.
func (s Move) To() Square { return Square((s & toMsk) >> toShift) }

// SetTo sets the target square of the move.
func (s *Move) SetTo(sq Square) { *s = (*s & ^toMsk) | (Move(sq) << toShift & toMsk) }

// From is the source square of the move.
func (s Move) From() Square { return Square((s & fromMsk) >> fromShift) }

// SetFrom sets the source square of the move.
func (s *Move) SetFrom(sq Square) { *s = (*s & ^fromMsk) | Move(sq)<<fromShift&fromMsk }

// Promo is the promotion piece of the move.
func (s Move) Promo() Piece { return Piece((s & promoMsk) >> promoShift) }

// SetPromo sets the promotion piece of the move.
func (s *Move) SetPromo(p Piece) { *s = (*s & ^promoMsk) | Move(p)<<promoShift&promoMsk }

// String representation of s, following uci move notation.
func (s Move) String() string {
	if s == 0 {
		return "0000"
	}
	return s.From().String() + s.To().String() + s.Promo().String()
}
