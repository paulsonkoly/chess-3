package board

import (
	"math/bits"

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
