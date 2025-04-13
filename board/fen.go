package board

import (
	"errors"
	"fmt"
	"strings"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

func FromFEN(fen string) (*Board, error) {
	p := fenParser{fen: fen, l: len(fen)}

	if err := p.seq(
		p.position,
		p.stm,
		p.cRights,
		p.enPassant,
		p.fifty,
	); err != nil {
		return nil, err
	}

	b := &p.b

	b.hashes = append(b.hashes, b.CalculateHash())

	return b, nil
}

var cToP = map[byte]Piece{
	'p': Pawn, 'r': Rook, 'n': Knight, 'b': Bishop, 'q': Queen, 'k': King,
	'P': Pawn, 'R': Rook, 'N': Knight, 'B': Bishop, 'Q': Queen, 'K': King,
}

type fenParser struct {
	fen string
	ix  int
	l   int
	b   Board
}

func (fp *fenParser) seq(parsers ...func() error) error {
	first := true
	for _, parser := range parsers {

		if first {
			first = false
		} else {
			for fp.ix < fp.l && fp.fen[fp.ix] == ' ' {
				fp.ix++
			}

			if fp.ix >= fp.l {
				return errors.New("premature end of fen")
			}
		}

		if err := parser(); err != nil {
			return err
		}
	}
	return nil
}

func (fp *fenParser) position() error {
	rank := 7
	file := 0

	for fp.ix < fp.l {
		sq := (8 * rank) + file
		c := fp.fen[fp.ix]

		switch c {

		case '1', '2', '3', '4', '5', '6', '7', '8':
			file += int(c - '0')

		case '/':
			file = 0
			rank--

		case 'p', 'r', 'n', 'b', 'q', 'k', 'P', 'R', 'N', 'B', 'Q', 'K':

			if sq < 0 || sq > 63 {
				return errors.New("invalid position")
			}
			bb := BitBoard(1 << sq)

			color := White
			piece := cToP[c]
			if c > 'a' && c < 'z' {
				color = Black
			}
			fp.b.Pieces[piece] |= bb
			fp.b.Colors[color] |= bb
			fp.b.SquaresToPiece[sq] = piece
			file++

		case ' ':
			return nil

		default:
			return fmt.Errorf("invalid char %c", c)
		}

		fp.ix++
	}

	fp.ix++
	return nil
}

func (fp *fenParser) stm() error {
	switch fp.fen[fp.ix] {
	case 'w':
		fp.b.STM = White
	case 'b':
		fp.b.STM = Black
	default:
		return fmt.Errorf("w or b expected, got %c", fp.fen[fp.ix])
	}
	fp.ix++

	return nil
}

func (fp *fenParser) cRights() error {
	for fp.ix < fp.l && fp.fen[fp.ix] != ' ' {
		switch fp.fen[fp.ix] {
		case 'K':
			fp.b.CRights |= CRights(ShortWhite)
		case 'Q':
			fp.b.CRights |= CRights(LongWhite)
		case 'k':
			fp.b.CRights |= CRights(ShortBlack)
		case 'q':
			fp.b.CRights |= CRights(LongBlack)
		case '-':

		default:
			return fmt.Errorf("K, Q, k, q or - expected got %c", fp.fen[fp.ix])
		}

		fp.ix++
	}
	return nil
}

func (fp *fenParser) enPassant() error {
	if fp.fen[fp.ix] != '-' {
		if fp.fen[fp.ix] < 'a' || fp.fen[fp.ix] > 'h' || fp.fen[fp.ix+1] < '1' || fp.fen[fp.ix+1] > '8' {
			return fmt.Errorf("square expected got %c%c", fp.fen[fp.ix], fp.fen[fp.ix+1])
		}
		file := fp.fen[fp.ix] - 'a'
		rank := fp.fen[fp.ix+1] - '1'
		if rank == 2 {
			rank = 3
		} else if rank == 5 {
			rank = 4
		}
		fp.b.EnPassant = Square(rank*8 + file)
		fp.ix++
	}
	fp.ix++
	return nil
}

func (fp *fenParser) fifty() error {
	cnt := 0
	for fp.ix < fp.l && fp.fen[fp.ix] != ' ' {
		if fp.fen[fp.ix] < '0' || fp.fen[fp.ix] > '9' {
			return fmt.Errorf("digit expected got %c", fp.fen[fp.ix])
		}
		cnt *= 10
		cnt += int(fp.fen[fp.ix] - '0')
		fp.ix++
	}

	fp.b.FiftyCnt = Depth(cnt)
	return nil
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
