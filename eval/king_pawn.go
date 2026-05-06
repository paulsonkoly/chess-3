package eval

import (
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
)

const PawnKingCacheSize = 16 * 1024

type PawnKingCache struct {
	hash  board.Hash
	accum [Colors]Score
}

func (e *Eval[T]) addStormShelter(b *board.Board, c *CoeffSet[T]) {
	var (
		t     T
		hash  board.Hash
		accum [2]T
	)

	// It is not possible to cache in the tuner as the coeffs are changing, thus
	// score always needs recomputing.
	if _, ok := any(t).(Score); ok {
		wKHash := board.PiecesRand[White][King][e.kings[White].sq]
		bKHash := board.PiecesRand[Black][King][e.kings[Black].sq]
		hash = b.Hashes().Pawn ^ wKHash ^ bKHash

		if e.pawnKingCache[hash%PawnKingCacheSize].hash == hash {
			entry := &e.pawnKingCache[hash%PawnKingCacheSize].accum
			e.kingAttacks[White] += T(entry[White])
			e.kingAttacks[Black] += T(entry[Black])
			return
		}
	}

	for color := range Colors {
		eKing := e.kings[color.Flip()].sq
		kFile := eKing.File()
		kRank := eKing.Rank()

		switch shelterIx, shelterKind := shelter(b, color.Flip()); shelterKind {

		case invalidShelter:
			accum[color] -= c.InvalidShelter

		case smallShelter:
			accum[color] -= c.SmallShelter[shelterIx]

		case normalShelter:
			accum[color] -= c.Shelter[shelterIx]
		}

		frontLine := e.pawns[color].frontline
		theirPawns := b.Colors[color.Flip()] & b.Pieces[Pawn]

		var central, front, side Coord
		if kFile >= EFile {
			central, front, side = kFile-1, kFile, kFile+1
		} else {
			central, front, side = kFile+1, kFile, kFile-1
		}

		// central
		{
			fileBB := FileBB(central & 7)

			if storm := frontLine & fileBB; storm != 0 {
				dist := Abs(kRank - storm.LowestSet().Rank())
				accum[color] += c.KingStorm[0][dist]
			}

			if theirPawns&fileBB == 0 {
				accum[color] += c.KingOpenFile[0]
			}
		}

		// front
		{
			fileBB := FileBB(front & 7)

			if storm := frontLine & fileBB; storm != 0 {
				dist := Abs(kRank - storm.LowestSet().Rank())
				accum[color] += c.KingStorm[1][dist]
			}

			if theirPawns&fileBB == 0 {
				accum[color] += c.KingOpenFile[1]
			}
		}

		// side
		{
			if side < 0 || side >= 8 {
				continue
			}
			fileBB := FileBB(side & 7)

			if storm := frontLine & fileBB; storm != 0 {
				dist := Abs(kRank - storm.LowestSet().Rank())
				accum[color] += c.KingStorm[2][dist]
			}

			if theirPawns&fileBB == 0 {
				accum[color] += c.KingOpenFile[2]
			}
		}
	}

	e.kingAttacks[White] += accum[White]
	e.kingAttacks[Black] += accum[Black]

	if _, ok := any(t).(Score); ok {
		e.pawnKingCache[hash%PawnKingCacheSize].hash = hash
		e.pawnKingCache[hash%PawnKingCacheSize].accum[White] = Score(accum[White])
		e.pawnKingCache[hash%PawnKingCacheSize].accum[Black] = Score(accum[Black])
	}
}

type shelterKind byte

const (
	invalidShelter = shelterKind(iota)
	smallShelter
	normalShelter
)

// shelter maps the formation of pawns in the kings "porch" area to a unique index.
// In normal case when the king is not on an edge file and not in enemy
// territory these squares comprise of the intersection of the kingfile, king
// neighbouring files with the 2 ranks in front of the king from color
// perspective. The squares where there is a shelter pawn are then packed down
// to a 6-bit index. For the queenside the bit pattern is horizontally
// mirrored. The rank closer to the king is always on less significant bits
// then the rank further away - for black there is a vertical mirror.
//
// Example:
//
//	color: White
//	kingSq: F3
//	porch: E4, F4, G4, E5, F5, G5
//	result 0-5 bit index: 0: E4 1: Fl 2: G4 3: E5 4: F5 5: G5
//
// Example2:
//
//	color: Black
//	kingSq: B5
//	porch: A3, B3, C3, A4, B4, C4
//	result 0-5 bit index: 0: C4 1: B4 2: A4 3: C3 4: B3 5: A3
func shelter(b *board.Board, color Color) (int, shelterKind) {
	king := b.Colors[color] & b.Pieces[King]
	kingSq := king.LowestSet()
	pawns := b.Colors[color] & b.Pieces[Pawn]

	var shelter BitBoard
	var kind shelterKind

	switch color {

	case White:
		if king&(SixthRankBB|SeventhRankBB|EighthRankBB) != 0 {
			return 0, invalidShelter
		}

		switch {
		case king&AFileBB != 0:
			kind = smallShelter
			shelter = ((king << 8) & pawns) >> (kingSq + 7)
			shelter |= ((king << 9) & pawns) >> (kingSq + 9)
			shelter |= ((king << 16) & pawns) >> (kingSq + 13)
			shelter |= ((king << 17) & pawns) >> (kingSq + 15)
		case king&HFileBB != 0:
			kind = smallShelter
			shelter = ((king << 7) & pawns) >> (kingSq + 7)
			shelter |= ((king << 8) & pawns) >> (kingSq + 7)
			shelter |= ((king << 15) & pawns) >> (kingSq + 13)
			shelter |= ((king << 16) & pawns) >> (kingSq + 13)
		default:
			kind = normalShelter
			shelter = ((king << 7) & pawns) >> (kingSq + 7)
			shelter |= ((king << 8) & pawns) >> (kingSq + 7)
			shelter |= ((king << 9) & pawns) >> (kingSq + 7)
			shelter |= ((king << 15) & pawns) >> (kingSq + 12)
			shelter |= ((king << 16) & pawns) >> (kingSq + 12)
			shelter |= ((king << 17) & pawns) >> (kingSq + 12)
		}

	case Black:
		if king&(FirstRankBB|SecondRankBB|ThirdRankBB) != 0 {
			return 0, invalidShelter
		}

		switch {
		case king&AFileBB != 0:
			kind = smallShelter
			shelter = ((king >> 7) & pawns) >> (kingSq - 7)
			shelter |= ((king >> 8) & pawns) >> (kingSq - 9)
			shelter |= ((king >> 15) & pawns) >> (kingSq - 17)
			shelter |= ((king >> 16) & pawns) >> (kingSq - 19)
		case king&HFileBB != 0:
			kind = smallShelter
			shelter = ((king >> 8) & pawns) >> (kingSq - 9)
			shelter |= ((king >> 9) & pawns) >> (kingSq - 9)
			shelter |= ((king >> 16) & pawns) >> (kingSq - 19)
			shelter |= ((king >> 17) & pawns) >> (kingSq - 19)
		default:
			kind = normalShelter
			shelter = ((king >> 7) & pawns) >> (kingSq - 9)
			shelter |= ((king >> 8) & pawns) >> (kingSq - 9)
			shelter |= ((king >> 9) & pawns) >> (kingSq - 9)
			shelter |= ((king >> 15) & pawns) >> (kingSq - 20)
			shelter |= ((king >> 16) & pawns) >> (kingSq - 20)
			shelter |= ((king >> 17) & pawns) >> (kingSq - 20)
		}
	}

	if kind == normalShelter && kingSq.File() < EFile {
		// mirror horizontally so the rank towards the centre moves to LSB
		shelter = ((shelter & 1) << 2) |
			(shelter & 2) |
			((shelter & 4) >> 2) |
			((shelter & 8) << 2) |
			(shelter & 16) |
			((shelter & 32) >> 2)
	}

	return int(shelter), kind
}
