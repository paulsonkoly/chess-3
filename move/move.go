package move

//revive:disable-next-line
import . "github.com/paulsonkoly/chess-3/types"

type Move struct {
	Promo    Piece
	Captured Piece
	Piece    Piece
	EPP      Piece
	From, To Square
	EPSq     Square
	Castle   Castle
	CRights  CastlingRights
	Weight   int
}

func (m Move) String() string {
	return m.From.String() + m.To.String() + m.Promo.String()
}
