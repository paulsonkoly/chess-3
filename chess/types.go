package chess

import (
	"fmt"
	"iter"
)

const StartPosFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

type Depth int8

const MaxPlies = 64

type Score int16

const (
	Inf = Score(10_000)  // Inf is the checkmate score.
	Inv = Score(-11_000) // Inv is an invalid score. It is guaranteed to be less than any valid scores.
)

func (s Score) IsMate() bool {
	return s <= -Inf+MaxPlies || s >= Inf-MaxPlies
}

func (s Score) String() string {
	if s == Inv {
		return "Inv"
	}

	a := Abs(s)
	if a >= Inf-MaxPlies {
		diff := Inf - a
		sign := ""

		if s < 0 {
			sign = "-"
		}

		return fmt.Sprintf("mate %s%d", sign, (diff+1)/2)
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
