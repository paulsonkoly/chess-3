package board

import (
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/move"
)

// Board is a chess position.
type Board struct {
	SquaresToPiece [64]Piece
	Pieces         [Pieces]BitBoard
	Colors         [2]BitBoard
	hashes         []Hashes
	fullMoves      int
	Counts         [Colors][Pieces]int16
	STM            Color
	EnPassant      Square
	Castles        Castles
	FiftyCnt       Depth
}

func StartPos() *Board {
	return Must(FromFEN(StartPosFEN))
}

// Hashes is the last set of Zobrist hashes in the move history of b.
func (b *Board) Hashes() Hashes {
	return b.hashes[len(b.hashes)-1]
}

// ResetHashes removes all previous hash history and sets it to contain the
// current position hash.
func (b *Board) ResetHashes() {
	if cap(b.hashes) == 0 {
		b.hashes = make([]Hashes, 0, 128)
	} else {
		b.hashes = b.hashes[:0]
	}
	b.hashes = append(b.hashes, b.calculateHash())
}

// ResetFifty resets the fifty move counter.
func (b *Board) ResetFifty() { b.FiftyCnt = 0 }

// IsEnPassant determines if a move is an en-passant pawn capture according to the current board en-passant state.
func (b *Board) IsEnPassant(sm move.Move) bool {
	return b.EnPassant != 0 && b.EnPassant == sm.To() && b.SquaresToPiece[sm.From()] == Pawn
}

// CaptureSq is mostly just the To() square of m except for an en-passant capture it's the square of the captured pawn.
func (b *Board) CaptureSq(m move.Move) Square {
	if b.IsEnPassant(m) {
		return (m.To() & FileMask) | (m.From() & RankMask)
	}
	return m.To()
}

var hashEnable = [2]Hash{0, 0xffffffffffffffff}

// Reverse is move reversing token. After making a move on the board this can
// be used to reverse the move.
type Reverse uint64

const (
	fiftyCntMask        = Reverse(0x00000000000000ff)
	fiftyCntShift       = 0
	castlingChangeMask  = Reverse(0x0000000000000f00)
	castlingChangeShift = 8
	epChangeMask        = Reverse(0x000000000003f000)
	epChangeShift       = 12
	captureMask         = Reverse(0x00000000001c0000)
	captureShift        = 18
)

func (r Reverse) fiftyCnt() Depth       { return Depth((r & fiftyCntMask) >> fiftyCntShift) }
func (r *Reverse) setFiftyCnt(fc Depth) { *r = (*r & ^fiftyCntMask) | Reverse(fc)<<fiftyCntShift }
func (r Reverse) castlingChange() Castles {
	return Castles((r & castlingChangeMask) >> castlingChangeShift)
}
func (r *Reverse) setCastlingChange(c Castles) {
	*r = (*r & ^castlingChangeMask) | Reverse(c)<<castlingChangeShift
}
func (r Reverse) enPassantChange() Square { return Square((r & epChangeMask) >> epChangeShift) }
func (r *Reverse) setEnPassantChange(epc Square) {
	*r = (*r & ^epChangeMask) | Reverse(epc)<<epChangeShift
}
func (r Reverse) capture() Piece { return Piece((r & captureMask) >> captureShift) }
func (r *Reverse) setCapture(p Piece) {
	*r = (*r & ^captureMask) | Reverse(p)<<captureShift
}

// MakeMove plays out a move m on the board b. It returns a Reverse token that
// can be used in UndoMove().
func (b *Board) MakeMove(m move.Move) Reverse {
	var r Reverse

	b.fullMoves += int(b.STM)

	hashes := b.Hashes()

	piece := b.SquaresToPiece[m.From()]
	canEnPassant := piece == Pawn && Abs(m.From()-m.To()) == 16 && b.CanEnPassant(m.To())
	captureSq := b.CaptureSq(m)
	capture := b.SquaresToPiece[captureSq]

	castlingChange := b.Castles ^ b.NewCastles(m)

	r.setFiftyCnt(b.FiftyCnt)
	if piece == Pawn || capture != NoPiece {
		b.FiftyCnt = 0
	} else {
		b.FiftyCnt++
	}

	hashes.Xor(NoPiece, castlingRand[0]&hashEnable[(castlingChange>>0)&1])
	hashes.Xor(NoPiece, castlingRand[1]&hashEnable[(castlingChange>>1)&1])
	hashes.Xor(NoPiece, castlingRand[2]&hashEnable[(castlingChange>>2)&1])
	hashes.Xor(NoPiece, castlingRand[3]&hashEnable[(castlingChange>>3)&1])

	b.Castles ^= castlingChange
	r.setCastlingChange(castlingChange)
	r.setCapture(capture)

	putPiece := piece
	if m.Promo() != NoPiece {
		putPiece = m.Promo()
	}

	hashes.Xor(capture, b.removePiece(b.STM.Flip(), capture, captureSq))
	hashes.Xor(piece, b.removePiece(b.STM, piece, m.From()))
	hashes.Xor(putPiece, b.addPiece(b.STM, putPiece, m.To()))

	if b.EnPassant != 0 {
		hashes.NonPawn ^= epFileRand[b.EnPassant.File()] // remove old enPassant
	}

	newEnPassant := Square(0)
	if canEnPassant {
		newEnPassant = (m.From() + m.To()) / 2
		hashes.NonPawn ^= epFileRand[newEnPassant.File()]
	}

	r.setEnPassantChange(b.EnPassant ^ newEnPassant)
	b.EnPassant = newEnPassant

	if piece == King {
		switch {

		case m.From() == E1 && m.To() == G1:
			hashes.Xor(NoPiece, b.removePiece(b.STM, Rook, H1))
			hashes.Xor(NoPiece, b.addPiece(b.STM, Rook, F1))

		case m.From() == E1 && m.To() == C1:
			hashes.Xor(NoPiece, b.removePiece(b.STM, Rook, A1))
			hashes.Xor(NoPiece, b.addPiece(b.STM, Rook, D1))

		case m.From() == E8 && m.To() == G8:
			hashes.Xor(NoPiece, b.removePiece(b.STM, Rook, H8))
			hashes.Xor(NoPiece, b.addPiece(b.STM, Rook, F8))

		case m.From() == E8 && m.To() == C8:
			hashes.Xor(NoPiece, b.removePiece(b.STM, Rook, A8))
			hashes.Xor(NoPiece, b.addPiece(b.STM, Rook, D8))
		}
	}

	b.STM = b.STM.Flip()
	hashes.NonPawn ^= stmRand

	b.hashes = append(b.hashes, hashes)

	// b.consistencyCheck()

	return r
}

// UndoMove reverses a move m that's already played on the board b, using the
// reversing token r.
func (b *Board) UndoMove(m move.Move, r Reverse) {
	b.hashes = b.hashes[:len(b.hashes)-1]

	b.STM = b.STM.Flip()

	rmPiece := b.SquaresToPiece[m.To()]
	piece := rmPiece
	if m.Promo() != NoPiece {
		piece = Pawn
	}

	if piece == King {
		switch {

		case m.From() == E1 && m.To() == G1:
			b.removePiece(b.STM, Rook, F1)
			b.addPiece(b.STM, Rook, H1)

		case m.From() == E1 && m.To() == C1:
			b.removePiece(b.STM, Rook, D1)
			b.addPiece(b.STM, Rook, A1)

		case m.From() == E8 && m.To() == G8:
			b.removePiece(b.STM, Rook, F8)
			b.addPiece(b.STM, Rook, H8)

		case m.From() == E8 && m.To() == C8:
			b.removePiece(b.STM, Rook, D8)
			b.addPiece(b.STM, Rook, A8)
		}
	}

	b.EnPassant ^= r.enPassantChange()

	b.removePiece(b.STM, rmPiece, m.To())
	b.addPiece(b.STM, piece, m.From())
	b.addPiece(b.STM.Flip(), r.capture(), b.CaptureSq(m))

	b.Castles ^= r.castlingChange()
	b.FiftyCnt = r.fiftyCnt()
	b.fullMoves -= int(b.STM)

	// b.consistencyCheck()
}

func (b *Board) NewCastles(m move.Move) Castles {
	var affected Castles
	piece := b.SquaresToPiece[m.From()]

	if piece == King {
		affected |= Castle(b.STM, Short) | Castle(b.STM, Long)
	}

	if m.From() == A1 || m.To() == A1 {
		affected |= LongWhite
	}

	if m.From() == H1 || m.To() == H1 {
		affected |= ShortWhite
	}

	if m.From() == A8 || m.To() == A8 {
		affected |= LongBlack
	}

	if m.From() == H8 || m.To() == H8 {
		affected |= ShortBlack
	}

	return b.Castles & ^affected
}

// addPiece adds a piece to the board fields returning the change in hash.
func (b *Board) addPiece(c Color, p Piece, sq Square) Hash {
	if p == NoPiece {
		return 0
	}
	b.Counts[c][p]++
	b.Colors[c] |= BitBoard(1) << sq
	b.Pieces[p] |= BitBoard(1) << sq
	b.SquaresToPiece[sq] = p

	return PiecesRand[c][p][sq]
}

func (b *Board) removePiece(c Color, p Piece, sq Square) Hash {
	if p == NoPiece {
		return 0
	}
	b.Counts[c][p]--
	b.Colors[c] &= ^(BitBoard(1) << sq)
	b.Pieces[p] &= ^(BitBoard(1) << sq)
	b.SquaresToPiece[sq] = NoPiece

	return PiecesRand[c][p][sq]
}

// MakeNullMove makes a null move on b. Passes to the opponent. It returns enP
// which needs to be passed to UndoNullMove unchanged.
func (b *Board) MakeNullMove() Reverse {
	var r Reverse
	hashes := b.hashes[len(b.hashes)-1]

	if b.EnPassant != 0 {
		r.setEnPassantChange(b.EnPassant)
		hashes.Xor(NoPiece, epFileRand[b.EnPassant.File()])
		b.EnPassant = 0
	}

	b.STM = b.STM.Flip()
	hashes.Xor(NoPiece, stmRand)

	b.hashes = append(b.hashes, hashes)
	// b.consistencyCheck()
	return r
}

// UndoNullMove undoes the effect of MakeNullMove. The board will be in the
// original state after executing MakeNullMove and UndoNullMove.
func (b *Board) UndoNullMove(r Reverse) {
	b.STM = b.STM.Flip()
	b.EnPassant = r.enPassantChange()
	b.hashes = b.hashes[:len(b.hashes)-1]
	// b.consistencyCheck()
}

// func (b *Board) consistencyCheck() {
// 	if b.Hashes().Full() != b.calculateHash().Full() {
// 		panic("inconsistent hash")
// 	}
//
// 	if b.Pieces[Pawn]|b.Pieces[Rook]|b.Pieces[Knight]|b.Pieces[Bishop]|b.Pieces[Queen]|b.Pieces[King] !=
// 		b.Colors[White]|b.Colors[Black] {
// 		panic("inconsistent pieces")
// 	}
//
// 	for piece := Pawn; piece <= King; piece++ {
// 		bb := BitBoard(0)
// 		for sq := A1; sq <= H8; sq++ {
// 			if b.SquaresToPiece[sq] == piece {
// 				bb |= 1 << sq
// 			}
// 		}
//
// 		if bb != b.Pieces[piece] {
// 			panic("inconsistent bitboard")
// 		}
// 	}
//
// 	for color := range Colors {
// 		for pType := Pawn; pType <= Queen; pType++ {
// 			if (b.Colors[color] & b.Pieces[pType]).Count() != int(b.Counts[color][pType]) {
// 				panic("inconsistent counts")
// 			}
// 		}
// 	}
// }

// Threefold is the repetition count of the current position in its history.
func (b *Board) Threefold() Depth {
	cnt := Depth(1)
	if len(b.hashes) > 0 {
		hash := b.hashes[len(b.hashes)-1]
		for ix := len(b.hashes) - 5; ix >= 0; ix -= 2 {
			if b.hashes[ix] == hash {
				cnt++
				if cnt >= 3 {
					return cnt
				}
			}
		}
	}
	return cnt
}
