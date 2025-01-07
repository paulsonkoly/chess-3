package board

import (
	"fmt"
	"strings"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

var cToP = map[byte]Piece{
	'p': Pawn, 'r': Rook, 'n': Knight, 'b': Bishop, 'q': Queen, 'k': King,
	'P': Pawn, 'R': Rook, 'N': Knight, 'B': Bishop, 'Q': Queen, 'K': King,
}

func FromFEN(fen string) *Board {
	b := Board{}

	rank := 7
	file := 0

	var (
		ix int
		c  byte
	)

	for ix, c = range []byte(fen) {
		sq := (8 * rank) + file
		bb := BitBoard(1 << sq)

		switch c {

		case '1', '2', '3', '4', '5', '6', '7', '8':
			file += int(c - '0')

		case '/':
			file = 0
			rank--

		case 'p', 'r', 'n', 'b', 'q', 'k', 'P', 'R', 'N', 'B', 'Q', 'K':
			color := White
			piece := cToP[c]
			if c > 'a' && c < 'z' {
				color = Black
			}
			b.Pieces[piece] |= bb
			b.Colors[color] |= bb
			b.SquaresToPiece[sq] = piece
			file++

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
	ix++

	for fen[ix] == ' ' {
		ix++
	}

	for fen[ix] != ' ' {
		switch fen[ix] {
		case 'K':
			b.CRights |= CRights(ShortWhite)
		case 'Q':
			b.CRights |= CRights(LongWhite)
		case 'k':
			b.CRights |= CRights(ShortBlack)
		case 'q':
			b.CRights |= CRights(LongBlack)
		}

		ix++
	}

	for fen[ix] == ' ' {
		ix++
	}

	if fen[ix] != '-' {
		file := fen[ix] - 'a'
		rank := fen[ix+1] - '1'
		if rank == 2 {
			rank = 3
		} else if rank == 5 {
			rank = 4
		}
		b.EnPassant = Square(rank*8 + file)
	}

	// /* TODO: move counter */
	// board->halfmovecnt = 0;
	// board->history[0].hash = calculate_hash(board);
	// board->history[0].flags = 0;
	//
	// return board;
	return &b
}

func (b Board) FEN() string {
	s := " PNBRQK pnbrqk"
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

				sb.WriteByte(s[int(7*c)+int(p)])
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

	sb.WriteString(fmt.Sprintf(" %c ", "wb"[b.STM]))

	if b.CRights&CRights(ShortWhite) != 0 {
		sb.WriteString("K")
	}

	if b.CRights&CRights(LongWhite) != 0 {
		sb.WriteString("Q")
	}

	if b.CRights&CRights(ShortBlack) != 0 {
		sb.WriteString("k")
	}

	if b.CRights&CRights(LongBlack) != 0 {
		sb.WriteString("q")
	}
	if b.CRights == 0 {
		sb.WriteString("-")
	}
	sb.WriteString(" ")

	if b.EnPassant == 0 {
		sb.WriteString("- 0 1")
	} else {
		var ep Square
		if b.EnPassant <= 31 {
			ep = b.EnPassant - 8
		} else {
			ep = b.EnPassant + 8
		}
		sb.WriteString(fmt.Sprintf("%s 0 1", ep))
	}
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
