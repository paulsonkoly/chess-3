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
	Weight   Score
	FiftyCnt int
}

func (m Move) String() string {
  promo := ""
  if m.Promo != NoPiece {
    promo = m.Promo.String()
  }
	return m.From.String() + m.To.String() + promo
}
