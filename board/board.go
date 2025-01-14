package board

import (
	"iter"
	"math/bits"
	"math/rand/v2"

	"github.com/paulsonkoly/chess-3/move"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

type BitBoard uint64

func (bb BitBoard) All() iter.Seq[BitBoard] {
	return func(yield func(BitBoard) bool) {
		for bb != 0 {
			single := bb & -bb
			if !yield(single) {
				return
			}
			bb ^= single
		}
	}
}

func (bb BitBoard) LowestSet() Square {
	return Square(bits.TrailingZeros64(uint64(bb)))
}

func BitBoardFromSquares(squares ...Square) BitBoard {
	var bb BitBoard
	for _, sq := range squares {
		bb |= BitBoard(1 << sq)
	}
	return bb
}

func (bb BitBoard) Count() int {
	return bits.OnesCount64(uint64(bb))
}

const (
	AFile = BitBoard(0x0101010101010101)
	BFile = BitBoard(0x0202020202020202)
	CFile = BitBoard(0x0404040404040404)
	DFile = BitBoard(0x0808080808080808)
	EFile = BitBoard(0x1010101010101010)
	FFile = BitBoard(0x2020202020202020)
	GFile = BitBoard(0x4040404040404040)
	HFile = BitBoard(0x8080808080808080)

	FistRank    = BitBoard(0x00000000000000ff)
	SecondRank  = BitBoard(0x000000000000ff00)
	ThirdRank   = BitBoard(0x0000000000ff0000)
	FourthRank  = BitBoard(0x00000000ff000000)
	FifthRank   = BitBoard(0x000000ff00000000)
	SixRank     = BitBoard(0x0000ff0000000000)
	SeventhRank = BitBoard(0x00ff000000000000)
	EightsRank  = BitBoard(0xff00000000000000)
)

const Full = BitBoard(0xffffffffffffffff)

type Board struct {
	SquaresToPiece [64]Piece
	Pieces         [7]BitBoard
	Colors         [2]BitBoard
	STM            Color
	EnPassant      Square
	CRights        CastlingRights
	Hashes         []Hash
	FiftyCnt       int
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

func (b *Board) MakeMove(m *move.Move) {
	m.FiftyCnt = b.FiftyCnt
	if m.Piece == Pawn || m.CRights != 0 || b.SquaresToPiece[m.To] != NoPiece {
		b.FiftyCnt = 0
	} else {
		b.FiftyCnt++
	}

	hash := b.Hashes[len(b.Hashes)-1]

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

	m.CRights, b.CRights = b.CRights, m.CRights^b.CRights
	m.Captured = b.SquaresToPiece[m.To]
	hash ^= epFileRand[b.EnPassant%8] & hashEnable[1&(b.EnPassant>>3|b.EnPassant>>5)]
	m.EPSq, b.EnPassant = b.EnPassant, m.To&m.EPSq // m.EnPassant is 0xff for double pawn pushes
	hash ^= epFileRand[b.EnPassant%8] & hashEnable[1&(b.EnPassant>>3|b.EnPassant>>5)]

	pm := pieceMask[m.Promo]

	b.Pieces[m.Captured] &= ^(1 << m.To)
	hash ^= piecesRand[b.STM.Flip()][m.Captured][m.To]
	b.Pieces[m.Piece] ^= (1 << m.From) | ((1 << m.To) & ^pm)
	hash ^= piecesRand[b.STM][m.Piece][m.From] ^ (piecesRand[b.STM][m.Piece][m.To] & ^Hash(pm))
	b.Pieces[m.Promo] ^= (1 << m.To) & pm
	hash ^= (piecesRand[b.STM][m.Promo][m.To] & Hash(pm))

	b.Colors[b.STM.Flip()] &= ^(1 << m.To)
	b.Colors[b.STM] ^= (1 << m.From) | (1 << m.To)

	b.SquaresToPiece[m.From] = NoPiece
	promo := Piece(pm & 1)
	b.SquaresToPiece[m.To] = (1-promo)*m.Piece + promo*m.Promo

	castle := castles[m.Castle]
	hash ^= piecesRand[b.STM][Rook][castle.down] & hashEnable[castle.piece>>2]
	hash ^= piecesRand[b.STM][Rook][castle.up] & hashEnable[castle.piece>>2]
	b.SquaresToPiece[castle.down] -= castle.piece
	b.SquaresToPiece[castle.up] += castle.piece
	b.Pieces[Rook] ^= castle.swap
	b.Colors[b.STM] ^= castle.swap

	if b.SquaresToPiece[castle.up] > King {
		panic("oops")
	}

	if b.SquaresToPiece[castle.down] > King {
		panic("oops")
	}

	// if b.Pieces[Knight]|b.Pieces[King] != b.Colors[White]|b.Colors[Black] {
	// 	b.Print(*ansi.NewWriter(os.Stdout))
	// 	fmt.Println(*b)
	// 	fmt.Println(*m)
	// 	panic("board inconsistency")
	// }
	b.STM = b.STM.Flip()

	hash ^= stmRand

	// TODO: optimise thise
	b.Hashes = append(b.Hashes, hash)
}

func (b *Board) UndoMove(m *move.Move) {
	b.FiftyCnt = m.FiftyCnt
	b.STM = b.STM.Flip()

	castle := castles[m.Castle]
	b.Pieces[Rook] ^= castle.swap
	b.Colors[b.STM] ^= castle.swap
	b.SquaresToPiece[castle.down] += castle.piece
	b.SquaresToPiece[castle.up] -= castle.piece

	pm := pieceMask[m.Promo]

	b.Pieces[m.Piece] ^= (1 << m.From) | ((1 << m.To) & ^pm)
	b.Pieces[m.Promo] ^= (1 << m.To) & pm
	b.Colors[b.STM] ^= (1 << m.From) | (1 << m.To)

	b.SquaresToPiece[m.From] = m.Piece
	b.SquaresToPiece[m.To] = m.Captured

	cm := (1 << m.To) & pieceMask[m.Captured]
	b.Pieces[m.Captured] ^= cm
	b.Colors[b.STM.Flip()] ^= cm

	b.CRights = m.CRights
	b.EnPassant = m.EPSq

	epMask := pieceMask[m.EPP]
	ep := Piece(epMask & 1)

	b.SquaresToPiece[b.EnPassant] += Pawn * ep
	b.Pieces[Pawn] |= (1 << b.EnPassant) & epMask
	b.Colors[b.STM.Flip()] |= (1 << b.EnPassant) & epMask

	b.Hashes = b.Hashes[:len(b.Hashes)-1]
}

type Hash uint64

// zobrist hashes
var (
	piecesRand   [2][7][64]Hash
	stmRand      Hash
	castlingRand [4]Hash
	epFileRand   [8]Hash
)

func init() {
	for i := range piecesRand {
		for j := range piecesRand[i] {
			for k := range piecesRand[i][j] {
				if j == int(NoPiece) {
					piecesRand[i][j][k] = 0
				} else {
					piecesRand[i][j][k] = Hash(rand.Uint64())
				}
			}
		}
	}
	stmRand = Hash(rand.Uint64())
	for i := range castlingRand {
		castlingRand[i] = Hash(rand.Uint64())
	}
	for i := range epFileRand {
		epFileRand[i] = Hash(rand.Uint64())
	}
}

func (b *Board) Hash() Hash {
	var hash Hash

	for color := White; color <= Black; color++ {
		occ := b.Colors[color]
		for piece := range occ.All() {
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
