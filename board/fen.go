package board

import (
	"errors"
	"fmt"
	"strings"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

var cToP = map[byte]Piece{
	'p': Pawn, 'r': Rook, 'n': Knight, 'b': Bishop, 'q': Queen, 'k': King,
	'P': Pawn, 'R': Rook, 'N': Knight, 'B': Bishop, 'Q': Queen, 'K': King,
}

func FromFEN(fen string) (*Board, error) {
	b := Board{}

	l := len(fen)

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

		case ' ':
			goto out

		default:
			return nil, fmt.Errorf("invalid char %c", c)
		}
	}
out:

	if ix >= l-1 {
		return nil, errors.New("premature end of fen")
	}

	for ix < l && fen[ix] == ' ' {
		ix++
	}

	if ix >= l {
		return nil, errors.New("premature end of fen")
	}

	switch fen[ix] {
	case 'w':
		b.STM = White
	case 'b':
		b.STM = Black
	default:
		return nil, fmt.Errorf("w or b expected, got %c", fen[ix])
	}
	ix++

	if ix >= l {
		return nil, errors.New("premature end of fen")
	}

	for ix < l && fen[ix] == ' ' {
		ix++
	}

	if ix >= l {
		return nil, errors.New("premature end of fen")
	}

	for ix < l && fen[ix] != ' ' {
		switch fen[ix] {
		case 'K':
			b.CRights |= CRights(ShortWhite)
		case 'Q':
			b.CRights |= CRights(LongWhite)
		case 'k':
			b.CRights |= CRights(ShortBlack)
		case 'q':
			b.CRights |= CRights(LongBlack)
		case '-':

		default:
			return nil, fmt.Errorf("K, Q, k, q or - expected got %c", fen[ix])
		}

		ix++
	}

	for ix < l && fen[ix] == ' ' {
		ix++
	}

	if ix >= l {
		return nil, errors.New("premature end of fen")
	}

	if fen[ix] != '-' {
		if fen[ix] < 'a' || fen[ix] > 'h' || fen[ix+1] < '1' || fen[ix+1] > '8' {
			return nil, fmt.Errorf("square expected got %c%c", fen[ix], fen[ix+1])
		}
		file := fen[ix] - 'a'
		rank := fen[ix+1] - '1'
		if rank == 2 {
			rank = 3
		} else if rank == 5 {
			rank = 4
		}
		b.EnPassant = Square(rank*8 + file)
		ix++
	}
	ix++

	if ix >= l {
		return nil, errors.New("premature end of fen")
	}

	for ix < l && fen[ix] == ' ' {
		ix++
	}

	if ix >= l {
		return nil, errors.New("premature end of fen")
	}

	cnt := 0
	for ix < l && fen[ix] != ' ' {
		if fen[ix] < '0' || fen[ix] > '9' {
			return nil, fmt.Errorf("digit expected got %c", fen[ix])
		}
		cnt *= 10
		cnt += int(fen[ix] - '0')
		ix++
	}

	b.FiftyCnt = Depth(cnt)

	b.hashes = append(b.hashes, b.CalculateHash())

	// /* TODO: move counter */
	// board->halfmovecnt = 0;
	// board->history[0].hash = calculate_hash(board);
	// board->history[0].flags = 0;
	//
	// return board;
	return &b, nil
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
		sb.WriteString("-")
	} else {
		var ep Square
		if b.EnPassant <= 31 {
			ep = b.EnPassant - 8
		} else {
			ep = b.EnPassant + 8
		}
		sb.WriteString(ep.String())
	}
	sb.WriteString(" ")
	sb.WriteString(fmt.Sprintf("%d 1", b.FiftyCnt))

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
