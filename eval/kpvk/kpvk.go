package kpvk

import (
	"fmt"
	"iter"

	"github.com/paulsonkoly/chess-3/attacks"
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
)

func Winning(b *board.Board) bool {
	wKingSq := (b.Colors[White] & b.Pieces[King]).LowestSet()
	bKingSq := (b.Colors[Black] & b.Pieces[King]).LowestSet()
	pawn := b.Pieces[Pawn]
	pawnSq := pawn.LowestSet()
	stm := b.STM

	// vertical mirror, swap sides. kpvk only supports white as the strong side.
	if b.Colors[Black]&pawn != 0 {
		wKingSq, bKingSq = bKingSq^56, wKingSq^56
		pawnSq ^= 56
		stm = stm.Flip()
	}

	// horizontal mirror, kpvk only supports pawn on the queen side.
	if pawn&(EFileBB|FFileBB|GFileBB|HFileBB) != 0 {
		wKingSq ^= 7
		bKingSq ^= 7
		pawnSq ^= 7
	}

	p := position{
		whiteKing: wKingSq,
		blackKing: bKingSq,
		pawnFile:  pawnSq.File(),
		pawnRank:  pawnSq.Rank(),
		stm:       stm,
	}

	return lut.Get(&p) == Win
}

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
		wKingCover := attacks.KingMoves(p.whiteKing)
		bKingCover := attacks.KingMoves(p.blackKing)

		switch {

		case Chebishev(p.whiteKing, p.blackKing) <= 1: // kings take each other, or on top of each other
			lut.Set(p, Invalid)
			unknowns--

		case p.whiteKing == pSq || p.blackKing == pSq: // king on top of pawn
			lut.Set(p, Invalid)
			unknowns--

		case p.stm == White && attacks.PawnCaptureMoves(BitBoardFromSquares(pSq), White)&BitBoardFromSquares(p.blackKing) != 0:
			lut.Set(p, Invalid)
			unknowns--

		case Chebishev(p.whiteKing, pSq) > 1 && Chebishev(p.blackKing, pSq) == 1 && p.stm == Black: // pawn can be captured
			lut.Set(p, Draw)
			unknowns--

		case p.blackKing == SquareAt(p.pawnFile, p.pawnRank+1) && p.pawnRank < SeventhRank:
			lut.Set(p, Draw)
			unknowns--

		case p.whiteKing == SquareAt(p.pawnFile, p.pawnRank+1) && p.blackKing == SquareAt(p.pawnFile, p.pawnRank+3) &&
			p.pawnRank < FifthRank && p.stm == White:
			lut.Set(p, Draw)
			unknowns--

		case p.pawnFile == AFile && p.blackKing.File() == AFile && p.blackKing.Rank() > p.pawnRank:
			lut.Set(p, Draw)
			unknowns--

		case p.pawnFile == AFile && p.whiteKing.File() == AFile && p.whiteKing.Rank() > p.pawnRank &&
			p.blackKing.File() == CFile && p.blackKing.Rank() == p.whiteKing.Rank() && p.stm == White:
			lut.Set(p, Draw)
			unknowns--

		case p.whiteKing == A7 && p.blackKing == C8 && p.pawnFile == AFile && p.stm == White:
			lut.Set(p, Draw)
			unknowns--

		case p.pawnRank == SeventhRank && p.stm == White && qSq != p.whiteKing && qSq != p.blackKing &&
			(wKingCover|^bKingCover)&BitBoardFromSquares(qSq) != 0:
			// pawn can queen
			lut.Set(p, Win)
			unknowns--
		}
	}

	for unknowns > 0 {
		fmt.Println("unknowns", unknowns)
		for p := range allPositions() {
			if lut.Get(p) != Unknown {
				continue
			}

			if unknowns == 50 {
				fmt.Println(p.whiteKing, p.blackKing, SquareAt(p.pawnFile, p.pawnRank), p.stm)
			}

			hasAny, hasUnknown, hasWin, hasDraw := false, false, false, false
			for child := range p.children() {
				hasAny = true

				switch lut.Get(child) {

				case Unknown:
					hasUnknown = true

				case Invalid:
					panic("children() generated an invalid position from an unknown position")

				case Win:
					hasWin = true

				case Draw:
					hasDraw = true
				}
			}

			if p.stm == White {
				switch {

				case hasWin:
					lut.Set(p, Win)
					unknowns--

				case !hasAny:
					lut.Set(p, Draw)
					unknowns--

				case !hasUnknown && hasDraw:
					lut.Set(p, Draw)
					unknowns--
				}
			} else {
				switch {

				case hasDraw:
					lut.Set(p, Draw)
					unknowns--

				case !hasAny:
					lut.Set(p, Draw)
					unknowns--

				case !hasUnknown && hasWin:
					lut.Set(p, Win)
					unknowns--
				}
			}
		}
	}
}
