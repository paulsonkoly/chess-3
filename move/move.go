package move

import "github.com/paulsonkoly/chess-3/types"

type Move struct {
	Captured types.Piece
	Piece    types.Piece
	From, To types.Square
}

func (m Move) String() string {
  return m.From.String() + m.To.String()
}
