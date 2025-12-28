package board

import (
	"github.com/paulsonkoly/chess-3/attacks"
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/move"
)

// Board is a chess position.
type Board struct {
	SquaresToPiece [64]Piece
	Pieces         [7]BitBoard
	Colors         [2]BitBoard
	STM            Color
	EnPassant      Square
	Castles        Castles
	hashes         []Hash
	FiftyCnt       Depth
}

func StartPos() *Board {
	return Must(FromFEN(StartPosFEN))
}

// Hash is the last Zobrist hash in the move history of b.
func (b *Board) Hash() Hash {
	return b.hashes[len(b.hashes)-1]
}

// ResetHash removes all previous hash history and sets it to contain the
// current position hash.
func (b *Board) ResetHash() {
	if cap(b.hashes) == 0 {
		b.hashes = make([]Hash, 0, 128)
	} else {
		b.hashes = b.hashes[:0]
	}
	b.hashes = append(b.hashes, b.calculateHash())
}

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

	hash := b.Hash()

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

	hash ^= castlingRand[0] & hashEnable[(castlingChange>>0)&1]
	hash ^= castlingRand[1] & hashEnable[(castlingChange>>1)&1]
	hash ^= castlingRand[2] & hashEnable[(castlingChange>>2)&1]
	hash ^= castlingRand[3] & hashEnable[(castlingChange>>3)&1]

	b.Castles ^= castlingChange
	r.setCastlingChange(castlingChange)
	r.setCapture(capture)

	putPiece := piece
	if m.Promo() != NoPiece {
		putPiece = m.Promo()
	}

	hash ^= b.removePiece(b.STM.Flip(), capture, captureSq)
	hash ^= b.removePiece(b.STM, piece, m.From())
	hash ^= b.addPiece(b.STM, putPiece, m.To())

	if b.EnPassant != 0 {
		hash ^= epFileRand[b.EnPassant.File()] // remove old enPassant
	}

	newEnPassant := Square(0)
	if canEnPassant {
		newEnPassant = (m.From() + m.To()) / 2
		hash ^= epFileRand[newEnPassant.File()]
	}

	r.setEnPassantChange(b.EnPassant ^ newEnPassant)
	b.EnPassant = newEnPassant

	if piece == King {
		switch {

		case m.From() == E1 && m.To() == G1:
			hash ^= b.removePiece(b.STM, Rook, H1)
			hash ^= b.addPiece(b.STM, Rook, F1)

		case m.From() == E1 && m.To() == C1:
			hash ^= b.removePiece(b.STM, Rook, A1)
			hash ^= b.addPiece(b.STM, Rook, D1)

		case m.From() == E8 && m.To() == G8:
			hash ^= b.removePiece(b.STM, Rook, H8)
			hash ^= b.addPiece(b.STM, Rook, F8)

		case m.From() == E8 && m.To() == C8:
			hash ^= b.removePiece(b.STM, Rook, A8)
			hash ^= b.addPiece(b.STM, Rook, D8)
		}
	}

	b.STM = b.STM.Flip()
	hash ^= stmRand

	b.hashes = append(b.hashes, hash)

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
	b.Colors[c] |= BitBoard(1) << sq
	b.Pieces[p] |= BitBoard(1) << sq
	b.SquaresToPiece[sq] = p

	return piecesRand[c][p][sq]
}

func (b *Board) removePiece(c Color, p Piece, sq Square) Hash {
	if p == NoPiece {
		return 0
	}
	b.Colors[c] &= ^(BitBoard(1) << sq)
	b.Pieces[p] &= ^(BitBoard(1) << sq)
	b.SquaresToPiece[sq] = NoPiece

	return piecesRand[c][p][sq]
}

// MakeNullMove makes a null move on b. Passes to the opponent. It returns enP
// which needs to be passed to UndoNullMove unchanged.
func (b *Board) MakeNullMove() Reverse {
	var r Reverse
	hash := b.hashes[len(b.hashes)-1]

	if b.EnPassant != 0 {
		r.setEnPassantChange(b.EnPassant)
		hash ^= epFileRand[b.EnPassant.File()]
		b.EnPassant = 0
	}

	b.STM = b.STM.Flip()
	hash ^= stmRand

	b.hashes = append(b.hashes, hash)
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
// 	if b.Hash() != b.calculateHash() {
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

// InvalidPieceCount determines if the position is legally reachable in chess.
// Returns true if it's not based on the piece counts.
func (b Board) InvalidPieceCount() bool {
	for color := White; color <= Black; color++ {
		if !(b.Colors[color] & b.Pieces[King]).IsPow2() {
			return true
		}
		knights := (b.Colors[color] & b.Pieces[Knight]).Count()
		bishops := (b.Colors[color] & b.Pieces[Bishop]).Count()
		rooks := (b.Colors[color] & b.Pieces[Rook]).Count()
		queens := (b.Colors[color] & b.Pieces[Queen]).Count()
		pawns := (b.Colors[color] & b.Pieces[Pawn]).Count()

		// Compute the number of pieces that are guaranteed to be promoted Pawns.
		pknights := max(2, knights) - 2
		pbishops := max(2, bishops) - 2
		prooks := max(2, rooks) - 2
		pqueens := max(1, queens) - 1
		promoted := pknights + pbishops + prooks + pqueens

		pawns += promoted
		if (pawns > 8) || (knights+pawns-pknights > 10) || (bishops+pawns-pbishops > 10) ||
			(rooks+pawns-prooks > 10) || (queens+pawns-pqueens > 9) {
			return true
		}
	}
	return false
}

func (b *Board) IsPseudoLegal(m move.Move) bool {
	from := m.From()
	fromBB := BitBoard(1) << from
	to := m.To()
	toBB := BitBoard(1) << to

	if b.Colors[b.STM]&fromBB == 0 {
		return false
	}
	if b.Colors[b.STM]&toBB != 0 {
		return false
	}

	piece := b.SquaresToPiece[from]
	occ := b.Colors[White] | b.Colors[Black]

	switch piece {

	case Knight:
		if attacks.KnightMoves(from)&toBB == 0 {
			return false
		}

	case Bishop:
		if attacks.BishopMoves(from, occ)&toBB == 0 {
			return false
		}

	case Rook:
		if attacks.RookMoves(from, occ)&toBB == 0 {
			return false
		}

	case Queen:
		if (attacks.RookMoves(from, occ)|attacks.BishopMoves(from, occ))&toBB == 0 {
			return false
		}

	case King:
		switch {

		case from == E1 && to == G1 && b.STM == White:
			if b.Castles&ShortWhite == 0 || BitBoardFromSquares(F1, G1)&occ != 0 ||
				b.IsAttacked(b.STM.Flip(), occ, BitBoardFromSquares(E1, F1, G1)) {
				return false
			}

		case from == E1 && to == C1 && b.STM == White:
			if b.Castles&LongWhite == 0 || BitBoardFromSquares(D1, C1, B1)&occ != 0 ||
				b.IsAttacked(b.STM.Flip(), occ, BitBoardFromSquares(E1, D1, C1)) {
				return false
			}

		case from == E8 && to == G8 && b.STM == Black:
			if b.Castles&ShortBlack == 0 || BitBoardFromSquares(F8, G8)&occ != 0 ||
				b.IsAttacked(b.STM.Flip(), occ, BitBoardFromSquares(E8, F8, G8)) {
				return false
			}

		case from == E8 && to == C8 && b.STM == Black:
			if b.Castles&LongBlack == 0 || BitBoardFromSquares(D8, C8, B8)&occ != 0 ||
				b.IsAttacked(b.STM.Flip(), occ, BitBoardFromSquares(E8, D8, C8)) {
				return false
			}

		default:
			if attacks.KingMoves(from)&toBB == 0 {
				return false
			}
		}

	case Pawn:

		if (from < to && b.STM == Black) || (from > to && b.STM == White) {
			return false
		}

		if RankBB(SeventhRank.FromPerspectiveOf(b.STM))&fromBB != 0 {
			if m.Promo() == NoPiece {
				return false
			}
		}

		switch Abs(from.File() - to.File()) {

		case 0: // pawn pushing

			switch Abs(from.Rank() - to.Rank()) {

			case 1: // single pawn push
				if occ&toBB != 0 {
					return false
				}

			case 2: // double pawn push
				if fromBB&RankBB(SecondRank.FromPerspectiveOf(b.STM)) == 0 {
					return false
				}

				if occ&(toBB|(BitBoard(1)<<((from+to)/2))) != 0 {
					return false
				}

			default:
				return false
			}

		case 1: // pawn capturing
			if Abs(from.Rank()-to.Rank()) != 1 {
				return false
			}

			enPassant := BitBoard(0)
			if b.EnPassant != 0 {
				enPassant = BitBoard(1) << BitBoard(b.EnPassant)
			}
			if (b.Colors[b.STM.Flip()]|enPassant)&toBB == 0 {
				return false
			}

		default:
			return false
		}
	}

	return true
}
