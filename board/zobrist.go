package board

import (
	"math/rand/v2"

	. "github.com/paulsonkoly/chess-3/chess"
)

// Hash is a chess position Zobrist hash.
type Hash uint64

// Hashes are a pair of separate pawns and non-panw Zobrist hashes.
type Hashes struct {
	Pawn    Hash // Pawn is unique per pawn placement
	NonPawn Hash // NonPawn is unique per non-pawn piece placement + other board states.
}

// Full is combined Zobrist hash.
func (h Hashes) Full() Hash { return h.Pawn ^ h.NonPawn }

func (h *Hashes) Xor(pt Piece, val Hash) {
	if pt == Pawn {
		h.Pawn ^= val
	} else {
		h.NonPawn ^= val
	}
}

// Zobrist hashes.
var (
	PiecesRand   [2][7][64]Hash // Piece constituent of Zobrist.
	stmRand      Hash
	castlingRand [4]Hash
	epFileRand   [8]Hash
)

var r rand.Source = rand.NewPCG(0xdeadbeeff0dbaad, 0xbaadf00ddeadbeef)

func init() {
	for i := range PiecesRand {
		for j := range PiecesRand[i] {
			for k := range PiecesRand[i][j] {
				if j == int(NoPiece) {
					PiecesRand[i][j][k] = 0
				} else {
					PiecesRand[i][j][k] = Hash(r.Uint64())
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
func (b Board) calculateHash() Hashes {
	var hashes Hashes

	for color := range Colors {
		for pawns := b.Pieces[Pawn] & b.Colors[color]; pawns != 0; pawns &= pawns - 1 {
			sq := pawns.LowestSet()

			hashes.Pawn ^= PiecesRand[color][Pawn][sq]
		}
		for pType := Knight; pType <= King; pType++ {
			for pieces := b.Pieces[pType] & b.Colors[color]; pieces != 0; pieces &= pieces - 1 {
				sq := pieces.LowestSet()

				hashes.NonPawn ^= PiecesRand[color][pType][sq]
			}
		}
	}

	if b.STM == Black {
		hashes.NonPawn ^= stmRand
	}

	for i, r := range castlingRand {
		if b.Castles&(1<<i) != 0 {
			hashes.NonPawn ^= r
		}
	}

	if b.EnPassant != 0 {
		hashes.NonPawn ^= epFileRand[b.EnPassant%8]
	}

	return hashes
}
