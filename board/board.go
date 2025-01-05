package board

import (
	"iter"
	"math/bits"

	"github.com/paulsonkoly/chess-3/move"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

type BitBoard uint64

func (bb BitBoard) All() iter.Seq[BitBoard] {
	return func(yield func(BitBoard) bool) {
		for bb != 0 {
			single := bb & -bb
			if !yield(single) {
				return
			}
			bb ^= single
		}
	}
}

func (bb BitBoard) LowestSet() Square {
	return Square(bits.TrailingZeros64(uint64(bb)))
}

func BitBoardFromSquares(squares ...Square) BitBoard {
	var bb BitBoard
	for _, sq := range squares {
		bb |= BitBoard(1 << sq)
	}
	return bb
}

const (
	AFile = BitBoard(0x8080808080808080)
	BFile = BitBoard(0x4040404040404040)
	CFile = BitBoard(0x2020202020202020)
	DFile = BitBoard(0x1010101010101010)
	EFile = BitBoard(0x0808080080808080)
	FFile = BitBoard(0x0404040040404040)
	GFile = BitBoard(0x0202020020202020)
	HFile = BitBoard(0x0101010010101010)
)

const Full = BitBoard(0xffffffffffffffff)

type Board struct {
	STM            Color
	SquaresToPiece [64]Piece
	Pieces         [7]BitBoard
	Colors         [2]BitBoard
}

// func New() *Board {
// 	sqTP := [64]Piece{}
// 	sqTP[B1] = Knight
// 	sqTP[G1] = Knight
// 	sqTP[B8] = Knight
// 	sqTP[G8] = Knight
//
// 	sqTP[E1] = King
// 	sqTP[E8] = King
//
// 	return &Board{
// 		SquaresToPiece: sqTP,
// 		Pieces: [7]BitBoard{
// 			Full,
// 			BitBoardFromSquares(E1, E8),
// 			BitBoardFromSquares(B1, G1, B8, G8),
// 		},
// 		Colors: [2]BitBoard{
// 			BitBoardFromSquares(B1, E1, G1),
// 			BitBoardFromSquares(B8, E8, G8),
// 		},
// 	}
// }
func (b *Board) MakeMove(m *move.Move) {

	m.Captured = b.SquaresToPiece[m.To]

	b.Pieces[m.Captured] &= ^(1 << m.To)
	b.Pieces[m.Piece] ^= (1 << m.From) | (1 << m.To)

	b.Colors[b.STM.Flip()] &= ^(1 << m.To)
	b.Colors[b.STM] ^= (1 << m.From) | (1 << m.To)

	b.SquaresToPiece[m.From] = NoPiece
	b.SquaresToPiece[m.To] = m.Piece

	// if b.Pieces[Knight]|b.Pieces[King] != b.Colors[White]|b.Colors[Black] {
	// 	b.Print(*ansi.NewWriter(os.Stdout))
	// 	fmt.Println(*b)
	// 	fmt.Println(*m)
	// 	panic("board inconsistency")
	// }
	b.STM = b.STM.Flip()
}

var captureMask = [...]BitBoard{
  0, Full, Full,  // TODO more piece types
}

func (b *Board) UndoMove(m *move.Move) {
	b.STM = b.STM.Flip()

	b.Pieces[m.Piece] ^= (1 << m.From) | (1 << m.To)
	b.Colors[b.STM] ^= (1 << m.From) | (1 << m.To)
	b.SquaresToPiece[m.To] = NoPiece
	b.SquaresToPiece[m.From] = m.Piece

	b.SquaresToPiece[m.To] = m.Captured

  cm := (1 << m.To) & captureMask[m.Captured]
  b.Pieces[m.Captured] ^= cm
  b.Colors[b.STM.Flip()] ^= cm

	// if b.Pieces[Knight]|b.Pieces[King] != b.Colors[White]|b.Colors[Black] {
	// 	panic("board inconsistency")
	// }
}
