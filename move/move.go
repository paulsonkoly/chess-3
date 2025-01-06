package move

//revive:disable-next-line
import . "github.com/paulsonkoly/chess-3/types"

type Move struct {
	Promo    Piece
	Captured Piece
	Piece    Piece
	From, To Square
	EPSq     Square
	EPP      Piece
}

func (m Move) String() string {
	return m.From.String() + m.To.String()
}
