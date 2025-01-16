package types

import "iter"

type Depth int8

type Score int16

// Inf is the checkmate score.
const Inf = Score(10_000)

const (
	Short = 0
	Long  = 1
)

type Castle byte

const (
	NoCastle   = 0
	ShortWhite = 1
	LongWhite  = 2
	ShortBlack = 3
	LongBlack  = 4
)

func C(c Color, typ int) Castle {
	return Castle(int(c)*2 + typ + 1)
}

type CastlingRights byte

func CRights(castles ...Castle) CastlingRights {
	result := CastlingRights(0)
	for _, c := range castles {
		result |= 1 << int(c-1)
	}
	return result
}

type Color byte

const (
	White = Color(iota)
	Black
)

func (c Color) Flip() Color { return c ^ 1 }

type Piece byte

const (
	NoPiece = Piece(iota)
	Pawn
	Knight
	Bishop
	Rook
	Queen
	King
)

func AllPieces() iter.Seq[Piece] {
	return func(yield func(Piece) bool) {
		for piece := Pawn; piece <= King; piece++ {
			if !yield(piece) {
				return
			}
		}
	}
}

func (p Piece) String() string {
	return string(" pnbrqk"[p])
}
