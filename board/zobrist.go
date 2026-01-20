package board

import (
	"math/rand/v2"

	. "github.com/paulsonkoly/chess-3/chess"
)

// Hash is a chess position Zobrist hash.
type Hash uint64

// Zobrist hashes.
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
func (b Board) calculateHash() Hash {
	var hash Hash

	for color := White; color <= Black; color++ {

		for occ := b.Colors[color]; occ != 0; occ &= occ - 1 {
			sq := occ.LowestSet()

			hash ^= piecesRand[color][b.SquaresToPiece[sq]][sq]
		}
	}

	if b.STM == Black {
		hash ^= stmRand
	}

	for i, r := range castlingRand {
		if b.Castles&(1<<i) != 0 {
			hash ^= r
		}
	}

	if b.EnPassant != 0 {
		hash ^= epFileRand[b.EnPassant%8]
	}

	return hash
}
