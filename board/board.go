package board

import (
	"github.com/paulsonkoly/chess-3/move"
	. "github.com/paulsonkoly/chess-3/types"
)

// Board is a chess position.
type Board struct {
	SquaresToPiece [64]Piece
	Pieces         [7]BitBoard
	Colors         [2]BitBoard
	STM            Color
	EnPassant      Square
	CRights        CastlingRights
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

type castle struct {
	piece Piece
	swap  BitBoard
	up    Square
	down  Square
}

var (
	pieceMask = [...]BitBoard{0, Full, Full, Full, Full, Full, Full}
	castles   = [5]castle{
		{piece: 0, swap: 0, up: 0, down: 0},
		{piece: Rook, swap: (1 << F1) | (1 << H1), up: F1, down: H1},
		{piece: Rook, swap: (1 << A1) | (1 << D1), up: D1, down: A1},
		{piece: Rook, swap: (1 << F8) | (1 << H8), up: F8, down: H8},
		{piece: Rook, swap: (1 << A8) | (1 << D8), up: D8, down: A8},
	}
)

var hashEnable = [2]Hash{0, 0xffffffffffffffff}

// MakeMove executes the move m on b. It updates where pieces are, en-passant
// state, move counters, Zobrist hash history etc.
func (b *Board) MakeMove(m *move.Move) {
	m.FiftyCnt = b.FiftyCnt
	if m.Piece == Pawn || m.CRights != 0 || b.SquaresToPiece[m.To()] != NoPiece {
		b.FiftyCnt = 0
	} else {
		b.FiftyCnt++
	}

	hash := b.hashes[len(b.hashes)-1]

	epMask := pieceMask[m.EPP]
	ep := Piece(epMask & 1)

	b.SquaresToPiece[b.EnPassant] -= Pawn * ep
	b.Pieces[m.EPP] &= ^((1 << b.EnPassant) & epMask)
	b.Colors[b.STM.Flip()] &= ^((1 << b.EnPassant) & epMask)
	hash ^= (piecesRand[b.STM.Flip()][Pawn][b.EnPassant] & Hash(epMask))

	hash ^= castlingRand[0] & hashEnable[(m.CRights>>0)&1]
	hash ^= castlingRand[1] & hashEnable[(m.CRights>>1)&1]
	hash ^= castlingRand[2] & hashEnable[(m.CRights>>2)&1]
	hash ^= castlingRand[3] & hashEnable[(m.CRights>>3)&1]

	b.CRights = m.CRights ^ b.CRights
	m.Captured = b.SquaresToPiece[m.To()]
	hash ^= epFileRand[b.EnPassant%8] & hashEnable[1&(b.EnPassant>>3|b.EnPassant>>5)]
	b.EnPassant ^= m.EPSq
	hash ^= epFileRand[b.EnPassant%8] & hashEnable[1&(b.EnPassant>>3|b.EnPassant>>5)]

	pm := pieceMask[m.Promo()]

	b.Pieces[m.Captured] &= ^(1 << m.To())
	hash ^= piecesRand[b.STM.Flip()][m.Captured][m.To()]
	b.Pieces[m.Piece] ^= (1 << m.From()) | ((1 << m.To()) & ^pm)
	hash ^= piecesRand[b.STM][m.Piece][m.From()] ^ (piecesRand[b.STM][m.Piece][m.To()] & ^Hash(pm))
	b.Pieces[m.Promo()] ^= (1 << m.To()) & pm
	hash ^= (piecesRand[b.STM][m.Promo()][m.To()] & Hash(pm))

	b.Colors[b.STM.Flip()] &= ^(1 << m.To())
	b.Colors[b.STM] ^= (1 << m.From()) | (1 << m.To())

	b.SquaresToPiece[m.From()] = NoPiece
	promo := Piece(pm & 1)
	b.SquaresToPiece[m.To()] = (1-promo)*m.Piece + promo*m.Promo()

	castle := castles[m.Castle]
	hash ^= piecesRand[b.STM][Rook][castle.down] & hashEnable[castle.piece>>2]
	hash ^= piecesRand[b.STM][Rook][castle.up] & hashEnable[castle.piece>>2]
	b.SquaresToPiece[castle.down] -= castle.piece
	b.SquaresToPiece[castle.up] += castle.piece
	b.Pieces[Rook] ^= castle.swap
	b.Colors[b.STM] ^= castle.swap

	b.STM = b.STM.Flip()

	hash ^= stmRand

	b.hashes = append(b.hashes, hash)

	// b.consistencyCheck()
}

// UndoMove undoes the effect of MakeMove(m). The board will be in the original
// state after MakeMove and UndoMove.
func (b *Board) UndoMove(m *move.Move) {
	b.FiftyCnt = m.FiftyCnt
	b.STM = b.STM.Flip()

	castle := castles[m.Castle]
	b.Pieces[Rook] ^= castle.swap
	b.Colors[b.STM] ^= castle.swap
	b.SquaresToPiece[castle.down] += castle.piece
	b.SquaresToPiece[castle.up] -= castle.piece

	pm := pieceMask[m.Promo()]

	b.Pieces[m.Piece] ^= (1 << m.From()) | ((1 << m.To()) & ^pm)
	b.Pieces[m.Promo()] ^= (1 << m.To()) & pm
	b.Colors[b.STM] ^= (1 << m.From()) | (1 << m.To())

	b.SquaresToPiece[m.From()] = m.Piece
	b.SquaresToPiece[m.To()] = m.Captured

	cm := (1 << m.To()) & pieceMask[m.Captured]
	b.Pieces[m.Captured] ^= cm
	b.Colors[b.STM.Flip()] ^= cm

	b.CRights ^= m.CRights
	b.EnPassant ^= m.EPSq

	epMask := pieceMask[m.EPP]
	ep := Piece(epMask & 1)

	b.SquaresToPiece[b.EnPassant] += Pawn * ep
	b.Pieces[Pawn] |= (1 << b.EnPassant) & epMask
	b.Colors[b.STM.Flip()] |= (1 << b.EnPassant) & epMask

	b.hashes = b.hashes[:len(b.hashes)-1]

	// b.consistencyCheck()
}

// MakeNullMove makes a null move on b. Passes to the opponent. It returns enP
// which needs to be passed to UndoNullMove unchanged.
func (b *Board) MakeNullMove() (enP Square) {
	enP, b.EnPassant = b.EnPassant, 0
	hash := b.hashes[len(b.hashes)-1]
	b.STM = b.STM.Flip()
	hash ^= epFileRand[enP%8] & hashEnable[1&(enP>>3|enP>>5)]
	hash ^= stmRand
	b.hashes = append(b.hashes, hash)
	// b.consistencyCheck()
	return
}

// UndoNullMove undoes the effect of MakeNullMove. The board will be in the
// original state after executing MakeNullMove and UndoNullMove.
func (b *Board) UndoNullMove(enP Square) {
	b.STM = b.STM.Flip()
	b.EnPassant = enP
	b.hashes = b.hashes[:len(b.hashes)-1]
	// b.consistencyCheck()
}

// func (b *Board) consistencyCheck() {
//   if b.hashes[len(b.hashes)-1]!= b.Hash() {
//     panic("inconsistent hash")
//   }
//
//   if b.Pieces[Pawn] | b.Pieces[Rook] | b.Pieces[Knight] | b.Pieces[Bishop] | b.Pieces[Queen] | b.Pieces[King] !=
//      b.Colors[White] | b.Colors[Black] {
//     panic("inconsistent pieces")
//   }
//
//   for piece := Pawn; piece <= King; piece++ {
//     bb := BitBoard(0)
//     for sq := A1; sq <= H8 ; sq ++ {
//       if b.SquaresToPiece[sq] == piece {
//         bb |= 1 << sq
//       }
//     }
//
//     if bb != b.Pieces[piece] {
//       panic("inconsistent bitboard")
//     }
//   }
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
