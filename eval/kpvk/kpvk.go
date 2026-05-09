package kpvk

import (
	"fmt"
	"iter"

	"github.com/paulsonkoly/chess-3/attacks"
	. "github.com/paulsonkoly/chess-3/chess"
)

type position struct {
	whiteKing Square
	blackKing Square
	pawnFile  Coord
	pawnRank  Coord
	stm       Color
}

func (p *position) children() iter.Seq[*position] {
	child := position{stm: p.stm.Flip()}

	occ := BitBoardFromSquares(p.whiteKing, p.blackKing, SquareAt(p.pawnFile, p.pawnRank))

	return func(yield func(*position) bool) {

		if p.stm == Black {

			child.whiteKing = p.whiteKing
			child.pawnFile = p.pawnFile
			child.pawnRank = p.pawnRank

			whitePawn := BitBoardFromSquares(SquareAt(p.pawnFile, p.pawnRank))
			whiteCover := attacks.KingMoves(p.whiteKing) | attacks.PawnCaptureMoves(whitePawn, White)
			mask := ^(whiteCover | occ)

			for kingMoves := attacks.KingMoves(p.blackKing) & mask; kingMoves != 0; kingMoves &= kingMoves - 1 {
				child.blackKing = kingMoves.LowestSet()

				if !yield(&child) {
					return
				}
			}
		} else {

			child.blackKing = p.blackKing

			blackCover := attacks.KingMoves(p.blackKing)
			mask := ^(blackCover | occ)

			for kingMoves := attacks.KingMoves(p.whiteKing) & mask; kingMoves != 0; kingMoves &= kingMoves - 1 {
				child.whiteKing = kingMoves.LowestSet()
				child.pawnFile = p.pawnFile
				child.pawnRank = p.pawnRank

				if !yield(&child) {
					return
				}
			}

			child.whiteKing = p.whiteKing

			occ := BitBoardFromSquares(p.whiteKing, p.blackKing)
			thirdSq := SquareAt(p.pawnFile, ThirdRank)
			fourthSq := SquareAt(p.pawnFile, FourthRank)

			switch p.pawnRank {

			case SecondRank:
				if BitBoardFromSquares(thirdSq, fourthSq)&occ == 0 {
					child.pawnFile = p.pawnFile
					child.pawnRank = FourthRank
					if !yield(&child) {
						return
					}
				}
				fallthrough

			case ThirdRank, FourthRank, FifthRank, SixthRank:
				if BitBoardFromSquares(SquareAt(p.pawnFile, p.pawnRank+1))&occ == 0 {
					child.pawnFile = p.pawnFile
					child.pawnRank = p.pawnRank + 1
					if !yield(&child) {
						return
					}
				}

			case SeventhRank:
				// already queening
			}
		}
	}
}

func allPositions() iter.Seq[*position] {
	var p position
	return func(yield func(*position) bool) {

		for stm := range Colors {
			for wK := range Squares {
				for bK := range Squares {
					for pF := range Coord(4) {
						for pR := SecondRank; pR <= SeventhRank; pR++ {
							p.stm = stm
							p.whiteKing = wK
							p.blackKing = bK
							p.pawnFile = pF
							p.pawnRank = pR

							if !yield(&p) {
								return
							}
						}
					}
				}
			}
		}
	}
}

type Kind byte

const (
	Unknown = Kind(iota)
	Invalid
	Draw
	Win
)

const (
	count = int(Colors) * int(Squares) * int(Squares) * 4 * 6
	// size is the byte size of the LUT. 2 bits per kind, fitted in an 8 bit byte => 4 entries per byte.
	size = count / 4
)

type table [size]Kind

var lut = table{}

func (t *table) Set(p *position, k Kind) {
	index := index(p)

	t[index/4] &= ^(3 << (2 * (index & 3)))
	t[index/4] |= k << (2 * (index & 3))
}

func (t *table) Get(p *position) Kind {
	index := index(p)
	return (t[index/4] >> (2 * (index & 3))) & 3
}

func index(p *position) int {
	return int(p.stm)*int(Squares)*int(Squares)*4*6 +
		int(p.whiteKing)*int(Squares)*4*6 +
		int(p.blackKing)*4*6 +
		int(p.pawnFile)*6 +
		int(p.pawnRank-1)
}

func init() {
	unknowns := count
	for p := range allPositions() {
		pSq := SquareAt(p.pawnFile, p.pawnRank)
		qSq := SquareAt(p.pawnFile, EighthRank)

		if p.whiteKing == 8 && p.blackKing == 2 && p.pawnFile == 2 && p.pawnRank == 2 && p.stm == 1 {
			fmt.Printf("%v %v\n", *p, lut.Get(p))
		}

		switch {

		case Chebishev(p.whiteKing, p.blackKing) <= 1: // kings take each other
			lut.Set(p, Invalid)
			unknowns--

		case p.whiteKing == pSq || p.blackKing == pSq: // king on top of pawn
			lut.Set(p, Invalid)
			unknowns--

		case Chebishev(p.whiteKing, pSq) > 1 && Chebishev(p.blackKing, pSq) == 1 && p.stm == Black: // pawn can be captured
			lut.Set(p, Draw)
			unknowns--

		case p.pawnRank == SeventhRank && (Chebishev(p.whiteKing, qSq) == 1 || Chebishev(p.blackKing, qSq) > 1):
			// pawn can queen
			lut.Set(p, Win)
			unknowns--
		}

		fmt.Printf("%v %v\n", *p, lut.Get(p))
	}

	iter := 0
	for unknowns > 0 {
		fmt.Printf("iter: %d unknowns: %d\n", iter, unknowns)
		iter++
		for p := range allPositions() {
			if lut.Get(p) != Unknown {
				continue
			}

			k := Draw

			for child := range p.children() {
				switch lut.Get(child) {

				case Unknown:
					// if the child is unknown this can be a draw or a win, if we see an
					// other child that is win, we can say it's a win. Otherwise stays
					// unknown.
					k = Unknown

				case Invalid:
					fmt.Println(child)
					panic("children() generated an invalid position from an unknown position")

				case Win:
					k = Win
					goto End

				case Draw:
					// if all are drawn then this a draw.
				}
			}

		End:
			if k != Unknown {
				lut.Set(p, k)
				unknowns--
			}
		}
	}

	invCnt, drawCnt, winCnt := 0, 0, 0

	for p := range allPositions() {
		switch lut.Get(p) {
		case Invalid:
			invCnt++
		case Draw:
			drawCnt++
		case Win:
			winCnt++
		}
	}
	fmt.Printf("inv: %d draw: %d win: %d\n", invCnt, drawCnt, winCnt)
}
