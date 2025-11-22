package types

import (
	"iter"

	"golang.org/x/exp/constraints"
)

type Depth int8

const MaxPlies = 64

type Score int16

const (
	Inf = Score(10_000) // Inf is the checkmate score.
	Inv = Score(11_000) // Inv is an invalid score.
)

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

	Colors
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
	if p == NoPiece {
		return ""
	}
	return string(" pnbrqk"[p])
}

func Abs[T constraints.Signed](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

func Signum[T constraints.Signed](x T) T {
	switch {
	case x < 0:
		return -1
	case x > 0:
		return 1
	}
	return 0
}

// Clamp clamps the value x between a and b.
func Clamp[T constraints.Signed](x, a, b T) T {
	return min(b, max(x, a))
}

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
