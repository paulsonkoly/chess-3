package board

import (
	"fmt"
	"iter"
	"math/bits"
	"strings"

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
	Pieces         [3]BitBoard
	Colors         [2]BitBoard
}

func New() *Board {
	sqTP := [64]Piece{}
	sqTP[B1] = Knight
	sqTP[G1] = Knight
	sqTP[B8] = Knight
	sqTP[G8] = Knight

	sqTP[E1] = King
	sqTP[E8] = King

	return &Board{
		SquaresToPiece: sqTP,
		Pieces: [3]BitBoard{
			Full,
			BitBoardFromSquares(E1, E8),
			BitBoardFromSquares(B1, G1, B8, G8),
		},
		Colors: [2]BitBoard{
			BitBoardFromSquares(B1, E1, G1),
			BitBoardFromSquares(B8, E8, G8),
		},
	}
}

func FromFEN(fen string) *Board {
	b := Board{}

	rank := 7
	file := 0

	var (
		ix int
		c  rune
	)

	for ix, c = range fen {
		sq := (8 * rank) + file
		bb := BitBoard(1 << sq)

		switch c {

		case '1', '2', '3', '4', '5', '6', '7', '8':
			file += int(c - '0')

		case '/':
			file = 0
			rank--

		// case 'R': board->rooks   |= flag; board->by_colour.whitepieces |= flag; f++; break;
		// case 'r': board->rooks   |= flag; board->by_colour.blackpieces |= flag; f++; break;
		case 'N':
			b.Pieces[Knight] |= bb
			b.Colors[White] |= bb
			b.SquaresToPiece[sq] = Knight
			file++

		case 'n':
			b.Pieces[Knight] |= bb
			b.Colors[Black] |= bb
			b.SquaresToPiece[sq] = Knight
			file++

		// case 'B': board->bishops |= flag; board->by_colour.whitepieces |= flag; f++; break;
		// case 'b': board->bishops |= flag; board->by_colour.blackpieces |= flag; f++; break;
		// case 'Q': board->queens  |= flag; board->by_colour.whitepieces |= flag; f++; break;
		// case 'q': board->queens  |= flag; board->by_colour.blackpieces |= flag; f++; break;

		case 'K':
			b.Pieces[King] |= bb
			b.Colors[White] |= bb
			b.SquaresToPiece[sq] = King
			file++

		case 'k':
			b.Pieces[King] |= bb
			b.Colors[Black] |= bb
			b.SquaresToPiece[sq] = King
			file++

		// case 'P': board->pawns   |= flag; board->by_colour.whitepieces |= flag; f++; break;
		// case 'p': board->pawns   |= flag; board->by_colour.blackpieces |= flag; f++; break;
		default:
			goto out
		}
	}
out:

	for fen[ix] == ' ' {
		ix++
	}

	switch fen[ix] {
	case 'w':
		b.STM = White
	case 'b':
		b.STM = Black
	}

	// for (; *ptr == ' '; ++ptr);

	// switch (*ptr++) {
	//   case 'w': board->next = WHITE; break;
	//   case 'b': board->next = BLACK; break;
	// }
	// for (; *ptr == ' '; ++ptr);
	//
	// for (; *ptr != ' '; ++ptr) {
	//   switch (*ptr) {
	//     case 'K': board->castle |= CALC_CASTLE(WHITE, SHORT_CASTLE); break;
	//     case 'Q': board->castle |= CALC_CASTLE(WHITE, LONG_CASTLE); break;
	//     case 'k': board->castle |= CALC_CASTLE(BLACK, SHORT_CASTLE); break;
	//     case 'q': board->castle |= CALC_CASTLE(BLACK, LONG_CASTLE); break;
	//   }
	// }
	// for (; *ptr == ' '; ++ptr);
	//
	// if (*ptr != '-') {
	//   f = *ptr++ - 'a';
	//   r = *ptr - '1';
	//   board->en_passant = 1ULL << (r * 8 + f);
	// }
	//
	// /* TODO: move counter */
	// board->halfmovecnt = 0;
	// board->history[0].hash = calculate_hash(board);
	// board->history[0].flags = 0;
	//
	// return board;
	return &b
}

func (b Board) FEN() string {
	s := " KN kn"
	sb := strings.Builder{}

	count := 0

	for rank := 7; rank >= 0; rank-- {
		for file := range 8 {
			sq := Square(rank*8 + file)
			p := b.SquaresToPiece[sq]

			if p != NoPiece {
				c := Black

				if b.Colors[White]&(1<<sq) != 0 {
					c = White
				}

				if count > 0 {
					sb.WriteString(fmt.Sprint(count))
					count = 0
				}

				sb.WriteByte(s[int(3*c)+int(p)]) // TODO add pieces
			} else {
				count++
			}
		}
		if count > 0 {
			sb.WriteString(fmt.Sprint(count))
			count = 0
		}
		if rank != 0 {
			sb.WriteString("/")
		}
	}

	sb.WriteString(fmt.Sprintf(" %c - - 0 1", "wb"[b.STM]))

	// if (board->castle & CALC_CASTLE(WHITE, SHORT_CASTLE)) printf("K");
	// if (board->castle & CALC_CASTLE(WHITE, LONG_CASTLE)) printf("Q");
	// if (board->castle & CALC_CASTLE(BLACK, SHORT_CASTLE)) printf("k");
	// if (board->castle & CALC_CASTLE(BLACK, LONG_CASTLE)) printf("q");
	// if (board->castle == 0) printf("-");
	//
	// if (board->en_passant) {
	//   SQUARE ep = __builtin_ctzll(board->en_passant);
	//   SQUARE f = (ep & 7), r = (ep >> 3);
	//
	//   printf(" %c%c 0 1 ", 'a' + f, '1' + r);
	// }
	// else {
	//   printf(" - 0 1 ");
	// }

	return sb.String()
}

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
