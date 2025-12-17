package move

import . "github.com/paulsonkoly/chess-3/types"

// Move represents a chess move.
type Move struct {
	SimpleMove
	// Weight is the heuristic weight of the move.
	Weight Score
	// Piece is the type of piece moving.
	Piece Piece
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

// SimpleMove s good enough to identify a move, so it can be stored in
// heuristic stores. It encodes from and to squares and promotion piece.
type SimpleMove uint16

// SimpleMoveOption is an optional argument to NewSimple.
type SimpleMoveOption interface {
	Apply(sm SimpleMove) SimpleMove
}

// Promo is a SimpleMoveOption setting the promotion piece type.
type Promo Piece

func (p Promo) Apply(sm SimpleMove) SimpleMove {
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
	toMsk      = SimpleMove(1<<6 - 1)
	toShift    = 0
	fromMsk    = SimpleMove((1<<6 - 1) << 6)
	fromShift  = 6
	promoMsk   = SimpleMove((1<<3 - 1) << 12)
	promoShift = 12
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

// SetPromo sets the promotion piece ofof  the move.
func (s *SimpleMove) SetPromo(p Piece) { *s = (*s & ^promoMsk) | SimpleMove(p)<<promoShift&promoMsk }

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
