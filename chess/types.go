package chess

import (
	"fmt"
	"iter"

	"golang.org/x/exp/constraints"
)

const StartPosFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

type Depth int8

const MaxPlies = 64

type Score int16

const (
	Inf = Score(10_000)  // Inf is the checkmate score.
	Inv = Score(-11_000) // Inv is an invalid score. It is guaranteed to be less than any valid scores.
)

func (s Score) String() string {
	if s == Inv {
		return "Inv"
	}

	if Abs(s) >= Inf-MaxPlies {
		diff := Inf - s

		if s < 0 {
			diff = -s - Inf
		}

		return fmt.Sprintf("mate %d", (diff+1)/2)
	}

	return fmt.Sprintf("cp %d", s)
}

type Side byte

const (
	Short = Side(iota)
	Long
)

// Castles is bitmap encoding of possible castlings.
type Castles byte

const (
	ShortWhite = Castles(1 << (2*int(White) + int(Short)))
	LongWhite  = Castles(1 << (2*int(White) + int(Long)))
	ShortBlack = Castles(1 << (2*int(Black) + int(Short)))
	LongBlack  = Castles(1 << (2*int(Black) + int(Long)))
)

func Castle(c Color, s Side) Castles { return 1 << (2*int(c) + int(s)) }

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
