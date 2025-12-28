package chess

import (
	"math/bits"
)

// A BitBoard is a 64 bit bitmap with one bit for each square of a chessboard.
type BitBoard uint64

// LowestSet is the first Square where bb has a '1' bit. Square order is from A1 to H8.
// In case there is no bits set in bb, the result is 64, outside of the chess board.
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
	AFileBB= BitBoard(0x0101010101010101) // AFileBB is a BitBoard with bits set for the A file.
	BFileBB= BitBoard(0x0202020202020202) // BFileBB is a BitBoard with bits set for the B file.
	CFileBB= BitBoard(0x0404040404040404) // CFileBB is a BitBoard with bits set for the C file.
	DFileBB= BitBoard(0x0808080808080808) // DFileBB is a BitBoard with bits set for the D file.
	EFileBB= BitBoard(0x1010101010101010) // EFileBB is a BitBoard with bits set for the E file.
	FFileBB= BitBoard(0x2020202020202020) // FFileBB is a BitBoard with bits set for the F file.
	GFileBB= BitBoard(0x4040404040404040) // GFileBB is a BitBoard with bits set for the G file.
	HFileBB= BitBoard(0x8080808080808080) // HFileBB is a BitBoard with bits set for the H file.

	FirstRankBB   = BitBoard(0x00000000000000ff) // FirstRankBB is a BitBoard with bits set for the fist rank.
	SecondRankBB  = BitBoard(0x000000000000ff00) // SecondRankBB is a BitBoard with bits set for the second rank.
	ThirdRankBB   = BitBoard(0x0000000000ff0000) // ThirdRankBB is a BitBoard with bits set for the third rank.
	FourthRankBB  = BitBoard(0x00000000ff000000) // FourthRankBB is a BitBoard with bits set for the fourth rank.
	FifthRankBB   = BitBoard(0x000000ff00000000) // FifthRankBB is a BitBoard with bits set for the fifth rank.
	SixthRankBB   = BitBoard(0x0000ff0000000000) // SixthRankBB is a BitBoard with bits set for the sixth rank.
	SeventhRankBB = BitBoard(0x00ff000000000000) // SeventhRankBB is a BitBoard with bits set for the seventh rank.
	EighthRankBB  = BitBoard(0xff00000000000000) // EighthRankBB is a BitBoard with bits set for the eighth rank.
)

var ranks = [...]BitBoard{
	FirstRankBB, SecondRankBB, ThirdRankBB, FourthRankBB, FifthRankBB, SixthRankBB, SeventhRankBB, EighthRankBB,
}

// RankBB is the BitBoard with bits set for the rankth rank from White's perspective.
func RankBB(rank Square) BitBoard { return ranks[rank] }

const (
	Full  = BitBoard(0xffffffffffffffff) // Full is a BitBoard with all 64 bits set.
)
