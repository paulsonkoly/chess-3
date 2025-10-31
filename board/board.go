package board

import (
	"math/bits"
	"math/rand/v2"

	"github.com/paulsonkoly/chess-3/move"
	. "github.com/paulsonkoly/chess-3/types"
)

// A BitBoard is a 64 bit bitmap with one bit for each square of a chessboard.
type BitBoard uint64

// LowestSet is the first Square where bb has a '1' bit. Square order is from A1 to H8.
func (bb BitBoard) LowestSet() Square {
	return Square(bits.TrailingZeros64(uint64(bb)))
}

// BitBoardFromSquares returns a BitBoard with squares from squares set to '1'.
func BitBoardFromSquares(squares ...Square) BitBoard {
	var bb BitBoard
	for _, sq := range squares {
		bb |= BitBoard(1 << sq)
	}
	return bb
}

// Count is the number of bits set in bb.
func (bb BitBoard) Count() int {
	return bits.OnesCount64(uint64(bb))
}

// IsPow2 determines if exactly 1 bit is set in bb.
func (bb BitBoard) IsPow2() bool {
	return bb&(bb-1) == 0 && bb != 0
}

const (
	AFile = BitBoard(0x0101010101010101) // AFile is a BitBoard with bits set for the A file.
	BFile = BitBoard(0x0202020202020202) // BFile is a BitBoard with bits set for the B file.
	CFile = BitBoard(0x0404040404040404) // CFile is a BitBoard with bits set for the C file.
	DFile = BitBoard(0x0808080808080808) // DFile is a BitBoard with bits set for the D file.
	EFile = BitBoard(0x1010101010101010) // EFile is a BitBoard with bits set for the E file.
	FFile = BitBoard(0x2020202020202020) // FFile is a BitBoard with bits set for the F file.
	GFile = BitBoard(0x4040404040404040) // GFile is a BitBoard with bits set for the G file.
	HFile = BitBoard(0x8080808080808080) // HFile is a BitBoard with bits set for the H file.

	FistRank    = BitBoard(0x00000000000000ff) // FistRank is a BitBoard with bits set for the fist rank.
	SecondRank  = BitBoard(0x000000000000ff00) // SecondRank is a BitBoard with bits set for the second rank.
	ThirdRank   = BitBoard(0x0000000000ff0000) // ThirdRank is a BitBoard with bits set for the third rank.
	FourthRank  = BitBoard(0x00000000ff000000) // FourthRank is a BitBoard with bits set for the fourth rank.
	FifthRank   = BitBoard(0x000000ff00000000) // FifthRank is a BitBoard with bits set for the fifth rank.
	SixRank     = BitBoard(0x0000ff0000000000) // SixRank is a BitBoard with bits set for the six rank.
	SeventhRank = BitBoard(0x00ff000000000000) // SeventhRank is a BitBoard with bits set for the seventh rank.
	EightsRank  = BitBoard(0xff00000000000000) // EightsRank is a BitBoard with bits set for the eights rank.
)

// Full is a BitBoard with all 64 bits set.
const Full = BitBoard(0xffffffffffffffff)

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

// Hash is the last Zobrist hash in the move history of b.
func (b *Board) Hash() Hash {
	return b.hashes[len(b.hashes)-1]
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

// Hash is a chess position Zobrist hash.
type Hash uint64

// zobrist hashes
var (
	piecesRand   [2][7][64]Hash
	stmRand      Hash
	castlingRand [4]Hash
	epFileRand   [8]Hash
)

var r rand.Source = rand.NewPCG(0xdeadbeeff0dbaad, 0xbaadf00ddeadbeef)

func init() {
	for i := range piecesRand {
		for j := range piecesRand[i] {
			for k := range piecesRand[i][j] {
				if j == int(NoPiece) {
					piecesRand[i][j][k] = 0
				} else {
					piecesRand[i][j][k] = Hash(r.Uint64())
				}
			}
		}
	}
	stmRand = Hash(r.Uint64())
	for i := range castlingRand {
		castlingRand[i] = Hash(r.Uint64())
	}
	for i := range epFileRand {
		epFileRand[i] = Hash(r.Uint64())
	}
}

// CalculateHash calculates the Zobrist hash for b from scratch. Normally it
// should not be used, b.Hash would give you a cached value of the same if b is
// obtained by making moves on b.
func (b *Board) CalculateHash() Hash {
	var hash Hash

	for color := White; color <= Black; color++ {
		occ := b.Colors[color]

		for piece := BitBoard(0); occ != 0; occ = occ ^ piece {
			piece = occ & -occ

			sq := piece.LowestSet()

			hash ^= piecesRand[color][b.SquaresToPiece[sq]][sq]
		}
	}

	if b.STM == Black {
		hash ^= stmRand
	}

	for i, r := range castlingRand {
		if b.CRights&(1<<i) != 0 {
			hash ^= r
		}
	}

	if b.EnPassant != 0 {
		hash ^= epFileRand[b.EnPassant%8]
	}

	return hash
}

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
