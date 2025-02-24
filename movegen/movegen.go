package movegen

import (
	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

var kingMoves = [64]board.BitBoard{
	0x0000000000000302, 0x0000000000000705, 0x0000000000000e0a, 0x0000000000001c14,
	0x0000000000003828, 0x0000000000007050, 0x000000000000e0a0, 0x000000000000c040,
	0x0000000000030203, 0x0000000000070507, 0x00000000000e0a0e, 0x00000000001c141c,
	0x0000000000382838, 0x0000000000705070, 0x0000000000e0a0e0, 0x0000000000c040c0,
	0x0000000003020300, 0x0000000007050700, 0x000000000e0a0e00, 0x000000001c141c00,
	0x0000000038283800, 0x0000000070507000, 0x00000000e0a0e000, 0x00000000c040c000,
	0x0000000302030000, 0x0000000705070000, 0x0000000e0a0e0000, 0x0000001c141c0000,
	0x0000003828380000, 0x0000007050700000, 0x000000e0a0e00000, 0x000000c040c00000,
	0x0000030203000000, 0x0000070507000000, 0x00000e0a0e000000, 0x00001c141c000000,
	0x0000382838000000, 0x0000705070000000, 0x0000e0a0e0000000, 0x0000c040c0000000,
	0x0003020300000000, 0x0007050700000000, 0x000e0a0e00000000, 0x001c141c00000000,
	0x0038283800000000, 0x0070507000000000, 0x00e0a0e000000000, 0x00c040c000000000,
	0x0302030000000000, 0x0705070000000000, 0x0e0a0e0000000000, 0x1c141c0000000000,
	0x3828380000000000, 0x7050700000000000, 0xe0a0e00000000000, 0xc040c00000000000,
	0x0203000000000000, 0x0507000000000000, 0x0a0e000000000000, 0x141c000000000000,
	0x2838000000000000, 0x5070000000000000, 0xa0e0000000000000, 0x40c0000000000000,
}

func KingMoves(from Square) board.BitBoard { return kingMoves[from] }

var knightMoves = [64]board.BitBoard{
	0x0000000000020400, 0x0000000000050800, 0x00000000000a1100, 0x0000000000142200,
	0x0000000000284400, 0x0000000000508800, 0x0000000000a01000, 0x0000000000402000,
	0x0000000002040004, 0x0000000005080008, 0x000000000a110011, 0x0000000014220022,
	0x0000000028440044, 0x0000000050880088, 0x00000000a0100010, 0x0000000040200020,
	0x0000000204000402, 0x0000000508000805, 0x0000000a1100110a, 0x0000001422002214,
	0x0000002844004428, 0x0000005088008850, 0x000000a0100010a0, 0x0000004020002040,
	0x0000020400040200, 0x0000050800080500, 0x00000a1100110a00, 0x0000142200221400,
	0x0000284400442800, 0x0000508800885000, 0x0000a0100010a000, 0x0000402000204000,
	0x0002040004020000, 0x0005080008050000, 0x000a1100110a0000, 0x0014220022140000,
	0x0028440044280000, 0x0050880088500000, 0x00a0100010a00000, 0x0040200020400000,
	0x0204000402000000, 0x0508000805000000, 0x0a1100110a000000, 0x1422002214000000,
	0x2844004428000000, 0x5088008850000000, 0xa0100010a0000000, 0x4020002040000000,
	0x0400040200000000, 0x0800080500000000, 0x1100110a00000000, 0x2200221400000000,
	0x4400442800000000, 0x8800885000000000, 0x100010a000000000, 0x2000204000000000,
	0x0004020000000000, 0x0008050000000000, 0x00110a0000000000, 0x0022140000000000,
	0x0044280000000000, 0x0088500000000000, 0x0010a00000000000, 0x0020400000000000,
}

func KnightMoves(from Square) board.BitBoard { return knightMoves[from] }

var bishopMasks = [64]board.BitBoard{
	0x0040201008040200, 0x0000402010080400, 0x0000004020100a00, 0x0000000040221400,
	0x0000000002442800, 0x0000000204085000, 0x0000020408102000, 0x0002040810204000,
	0x0020100804020000, 0x0040201008040000, 0x00004020100a0000, 0x0000004022140000,
	0x0000000244280000, 0x0000020408500000, 0x0002040810200000, 0x0004081020400000,
	0x0010080402000200, 0x0020100804000400, 0x004020100a000a00, 0x0000402214001400,
	0x0000024428002800, 0x0002040850005000, 0x0004081020002000, 0x0008102040004000,
	0x0008040200020400, 0x0010080400040800, 0x0020100a000a1000, 0x0040221400142200,
	0x0002442800284400, 0x0004085000500800, 0x0008102000201000, 0x0010204000402000,
	0x0004020002040800, 0x0008040004081000, 0x00100a000a102000, 0x0022140014224000,
	0x0044280028440200, 0x0008500050080400, 0x0010200020100800, 0x0020400040201000,
	0x0002000204081000, 0x0004000408102000, 0x000a000a10204000, 0x0014001422400000,
	0x0028002844020000, 0x0050005008040200, 0x0020002010080400, 0x0040004020100800,
	0x0000020408102000, 0x0000040810204000, 0x00000a1020400000, 0x0000142240000000,
	0x0000284402000000, 0x0000500804020000, 0x0000201008040200, 0x0000402010080400,
	0x0002040810204000, 0x0004081020400000, 0x000a102040000000, 0x0014224000000000,
	0x0028440200000000, 0x0050080402000000, 0x0020100804020000, 0x0040201008040200,
}

var rookMasks = [64]board.BitBoard{
	0x000101010101017e, 0x000202020202027c, 0x000404040404047a, 0x0008080808080876,
	0x001010101010106e, 0x002020202020205e, 0x004040404040403e, 0x008080808080807e,
	0x0001010101017e00, 0x0002020202027c00, 0x0004040404047a00, 0x0008080808087600,
	0x0010101010106e00, 0x0020202020205e00, 0x0040404040403e00, 0x0080808080807e00,
	0x00010101017e0100, 0x00020202027c0200, 0x00040404047a0400, 0x0008080808760800,
	0x00101010106e1000, 0x00202020205e2000, 0x00404040403e4000, 0x00808080807e8000,
	0x000101017e010100, 0x000202027c020200, 0x000404047a040400, 0x0008080876080800,
	0x001010106e101000, 0x002020205e202000, 0x004040403e404000, 0x008080807e808000,
	0x0001017e01010100, 0x0002027c02020200, 0x0004047a04040400, 0x0008087608080800,
	0x0010106e10101000, 0x0020205e20202000, 0x0040403e40404000, 0x0080807e80808000,
	0x00017e0101010100, 0x00027c0202020200, 0x00047a0404040400, 0x0008760808080800,
	0x00106e1010101000, 0x00205e2020202000, 0x00403e4040404000, 0x00807e8080808000,
	0x007e010101010100, 0x007c020202020200, 0x007a040404040400, 0x0076080808080800,
	0x006e101010101000, 0x005e202020202000, 0x003e404040404000, 0x007e808080808000,
	0x7e01010101010100, 0x7c02020202020200, 0x7a04040404040400, 0x7608080808080800,
	0x6e10101010101000, 0x5e20202020202000, 0x3e40404040404000, 0x7e80808080808000,
}

var bishopShifts = [64]byte{
	6, 5, 5, 5, 5, 5, 5, 6,
	5, 5, 5, 5, 5, 5, 5, 5,
	5, 5, 7, 7, 7, 7, 5, 5,
	5, 5, 7, 9, 9, 7, 5, 5,
	5, 5, 7, 9, 9, 7, 5, 5,
	5, 5, 7, 7, 7, 7, 5, 5,
	5, 5, 5, 5, 5, 5, 5, 5,
	6, 5, 5, 5, 5, 5, 5, 6,
}

var rookShifts = [64]byte{
	12, 11, 11, 11, 11, 11, 11, 12,
	11, 10, 10, 10, 10, 10, 10, 11,
	11, 10, 10, 10, 10, 10, 10, 11,
	11, 10, 10, 10, 10, 10, 10, 11,
	11, 10, 10, 10, 10, 10, 10, 11,
	11, 10, 10, 10, 10, 10, 10, 11,
	11, 10, 10, 10, 10, 10, 10, 11,
	12, 11, 11, 11, 11, 11, 11, 12,
}

var bishopMagics = [64]board.BitBoard{
	0x01442b1002020240, 0x23280a482202e891, 0x685843810202408c, 0x287820d044d95020,
	0x5a6b1041408282c3, 0x01482c0c60008813, 0x024a43083841f042, 0x24a0460150080c00,
	0x4602412408220640, 0x2300a00142420042, 0x0100281806408228, 0x41100405020c1c62,
	0x03053a0210002808, 0x0400a202302c4465, 0x054504012c022018, 0x000a908405051012,
	0x0c600c901403a800, 0x082202e018910103, 0x529000260c007160, 0x1004041824049080,
	0x2a0400b82208020a, 0x0a42c0020100a000, 0x10070024040b5461, 0x4031011c49082111,
	0x43d0c12408820400, 0x0844549006082810, 0x2000680230018264, 0x000a018008008042,
	0x4481011109004000, 0x451014800c0a010a, 0x41180350408a2802, 0x1163020001028082,
	0x1108200400281820, 0x0622312423101031, 0x4022a0f000680488, 0x0991404800018200,
	0x280508060001a200, 0x3b821202023c4802, 0x00011e1a00040100, 0x089112020483a304,
	0x0804211828874060, 0x14c209842428a002, 0x0042010041041822, 0x40900c2104002040,
	0x1020441d04080609, 0x40844810010010c8, 0x2028231122042c0a, 0x40508200410428c1,
	0x1424473011102926, 0x0518410c10064a30, 0x04b1824044501029, 0x204051c120880641,
	0x5012f41060622031, 0x2875604722460049, 0x61a0202409829031, 0x400714080481000c,
	0x401201041a024221, 0x62140200421e1004, 0x009b828826051057, 0x2440370004209814,
	0x404601a741028600, 0x004fd42024091201, 0x262c888a14444400, 0x4208652400820202,
}

var rookMagics = [64]board.BitBoard{
	0x0200128102002041, 0x64c0002002100440, 0x0100200008c05102, 0x0200120020043840,
	0x1b00030008008410, 0x4200120054211008, 0x24000d5000880a04, 0x610000270000c886,
	0x06ca801040022580, 0x0c0040045000a001, 0x0021001102c06000, 0x08d1001000a19b00,
	0x2b17005800841100, 0x1206000201300894, 0x3004000813041012, 0x681a00056106008c,
	0x2d60628002c00483, 0x0020014000505012, 0x4026848020029000, 0x010452000862c200,
	0x1308818008002400, 0x0c1b080120104004, 0x1089240010863108, 0x0211320004008051,
	0x2000800300210842, 0x0620810600220ac1, 0x0404420200201482, 0x0042500300096100,
	0x2801a21200093200, 0x2c12000600043019, 0x4042000200580479, 0x40040d02000a815c,
	0x2900804000800030, 0x0303420102002c83, 0x004200a142003084, 0x2660520062000940,
	0x026200a00a001084, 0x3802001d42001830, 0x0808662104001088, 0x110222a52a0003a4,
	0x00d88b2140028007, 0x735950002000400b, 0x34c5820020160040, 0x0206003820420011,
	0x684a002830c60020, 0x405e001410a20009, 0x0440221011040008, 0x211009014096000c,
	0x129c421502208200, 0x6a42042052850200, 0x2f09014a6002f100, 0x0a06191042012200,
	0x58f0715500780100, 0x090201904c081200, 0x310e002528041200, 0x484da14408810200,
	0x4d3a416133008001, 0x4406802436400101, 0x0043c17912200101, 0x1240615000890015,
	0x4c4600302004c822, 0x0122001008410462, 0x04be821803102084, 0x5202884c0500826a,
}

var bishopAttacks [64][512]board.BitBoard
var rookAttacks [64][4096]board.BitBoard

func BishopMoves(from Square, occ board.BitBoard) board.BitBoard {
	mask := bishopMasks[from]
	magic := bishopMagics[from]
	shift := bishopShifts[from]

	return bishopAttacks[from][((occ&mask)*magic)>>(64-shift)]
}

func RookMoves(from Square, occ board.BitBoard) board.BitBoard {
	mask := rookMasks[from]
	magic := rookMagics[from]
	shift := rookShifts[from]

	return rookAttacks[from][((occ&mask)*magic)>>(64-shift)]
}

func init() {
	for sq := Square(0); sq < 64; sq++ {
		mask := bishopMasks[sq]
		magic := bishopMagics[sq]
		shift := bishopShifts[sq]
		occ := mask

		for {
			attacks := calcBishopAttacks(sq, occ)
			bishopAttacks[sq][(occ*magic)>>(64-shift)] = attacks
			occ = (occ - mask) & mask

			if occ == mask {

				break
			}
		}
	}
	for sq := Square(0); sq < 64; sq++ {
		mask := rookMasks[sq]
		magic := rookMagics[sq]
		shift := rookShifts[sq]
		occ := mask

		for {
			attacks := calcRookAttacks(sq, occ)
			rookAttacks[sq][(occ*magic)>>(64-shift)] = attacks
			occ = (occ - mask) & mask

			if occ == mask {

				break
			}
		}
	}
}

func calcBishopAttacks(sq Square, occ board.BitBoard) board.BitBoard {
	result := board.BitBoard(0)

	r := int(sq / 8)
	f := int(sq % 8)

	for rr, ff := r+1, f+1; rr <= 7 && ff <= 7; {
		result |= (1 << (ff + rr*8))
		if occ&(1<<(ff+rr*8)) != 0 {
			break
		}
		rr++
		ff++
	}
	for rr, ff := r+1, f-1; rr <= 7 && ff >= 0; {
		result |= (1 << (ff + rr*8))
		if occ&(1<<(ff+rr*8)) != 0 {
			break
		}
		rr++
		ff--
	}
	for rr, ff := r-1, f+1; rr >= 0 && ff <= 7; {
		result |= (1 << (ff + rr*8))
		if occ&(1<<(ff+rr*8)) != 0 {
			break
		}
		rr--
		ff++
	}

	for rr, ff := r-1, f-1; rr >= 0 && ff >= 0; {
		result |= 1 << (ff + rr*8)
		if occ&(1<<(ff+rr*8)) != 0 {
			break
		}
		rr--
		ff--
	}

	return result
}

func calcRookAttacks(sq Square, occ board.BitBoard) board.BitBoard {
	result := board.BitBoard(0)

	r := int(sq / 8)
	f := int(sq % 8)

	for rr := r + 1; rr <= 7; rr++ {
		result |= (1 << (f + rr*8))
		if occ&(1<<(f+rr*8)) != 0 {
			break
		}
	}
	for rr := r - 1; rr >= 0; rr-- {
		result |= (1 << (f + rr*8))
		if occ&(1<<(f+rr*8)) != 0 {
			break
		}
	}
	for ff := f + 1; ff <= 7; ff++ {
		result |= (1 << (ff + r*8))
		if occ&(1<<(ff+r*8)) != 0 {
			break
		}
	}
	for ff := f - 1; ff >= 0; ff-- {
		result |= (1 << (ff + r*8))
		if occ&(1<<(ff+r*8)) != 0 {
			break
		}
	}

	return result
}

func PawnCaptureMoves(from Square, color Color) board.BitBoard {
	bb := board.BitBoard(1) << from

	if color == White {
		return ((bb & ^board.AFile) << 7) | ((bb & ^board.HFile) << 9)
	}
	return ((bb & ^board.HFile) >> 7) | ((bb & ^board.AFile) >> 9)
}

var (
	sndRank    = [...]board.BitBoard{board.SecondRank, board.SeventhRank}
	fourthRank = [...]board.BitBoard{board.FourthRank, board.FifthRank}
	castleMask = [2][2]board.BitBoard{
		{(1 << E1) | (1 << F1) | (1 << G1), (1 << E1) | (1 << D1) | (1 << C1)},
		{(1 << E8) | (1 << F8) | (1 << G8), (1 << E8) | (1 << D8) | (1 << C8)},
	}
	kingCRightsUpd = [2]CastlingRights{CRights(ShortWhite, LongWhite), CRights(ShortBlack, LongBlack)}
	rookCRightsUpd = [64]CastlingRights{
		CRights(LongWhite), 0, 0, 0, 0, 0, 0, CRights(ShortWhite),
		0 /**************/, 0, 0, 0, 0, 0, 0, 0,
		0 /**************/, 0, 0, 0, 0, 0, 0, 0,
		0 /**************/, 0, 0, 0, 0, 0, 0, 0,
		0 /**************/, 0, 0, 0, 0, 0, 0, 0,
		0 /**************/, 0, 0, 0, 0, 0, 0, 0,
		0 /**************/, 0, 0, 0, 0, 0, 0, 0,
		CRights(LongBlack), 0, 0, 0, 0, 0, 0, CRights(ShortBlack),
	}
)

func GenMoves(ms *move.Store, b *board.Board, target board.BitBoard) {

	self := b.Colors[b.STM]
	them := b.Colors[b.STM.Flip()]

	if b.FiftyCnt >= 100 {
		return
	}

	// king moves
	if piece := self & b.Pieces[King]; piece != 0 {
		from := piece.LowestSet()

		tSqrs := kingMoves[from] & ^self & target

		for tSqr := board.BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()
			newC := b.CRights & ^(kingCRightsUpd[b.STM] | rookCRightsUpd[to])
			m := ms.Alloc()
			m.Piece = King
			m.From = from
			m.To = to
			m.CRights = newC ^ b.CRights
			m.Castle = 0
			m.Promo = 0
			m.EPP = 0
			m.EPSq = b.EnPassant
		}
	}

	// knight moves
	knights := self & b.Pieces[Knight]
	for knight := board.BitBoard(0); knights != 0; knights ^= knight {
		knight = knights & -knights
		from := knight.LowestSet()

		tSqrs := knightMoves[from] & ^self & target

		for tSqr := board.BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()
			newC := b.CRights & ^(rookCRightsUpd[to])
			m := ms.Alloc()
			m.Piece = Knight
			m.From = from
			m.To = to
			m.CRights = newC ^ b.CRights
			m.Castle = 0
			m.Promo = 0
			m.EPP = 0
			m.EPSq = b.EnPassant
		}
	}

	occ := b.Colors[White] | b.Colors[Black]

	// bishop moves
	bishops := self & b.Pieces[Bishop]
	for bishop := board.BitBoard(0); bishops != 0; bishops ^= bishop {
		bishop = bishops & -bishops
		from := bishop.LowestSet()
		mask := bishopMasks[from]
		magic := bishopMagics[from]
		shift := bishopShifts[from]

		tSqrs := bishopAttacks[from][((occ&mask)*magic)>>(64-shift)] & ^self & target

		for tSqr := board.BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()
			newC := b.CRights & ^(rookCRightsUpd[to])
			m := ms.Alloc()
			m.Piece = Bishop
			m.From = from
			m.To = to
			m.CRights = newC ^ b.CRights
			m.Castle = 0
			m.Promo = 0
			m.EPP = 0
			m.EPSq = b.EnPassant
		}
	}

	// rook moves
	rooks := self & b.Pieces[Rook]
	for rook := board.BitBoard(0); rooks != 0; rooks ^= rook {
		rook = rooks & -rooks
		from := rook.LowestSet()
		mask := rookMasks[from]
		magic := rookMagics[from]
		shift := rookShifts[from]

		tSqrs := rookAttacks[from][((occ&mask)*magic)>>(64-shift)] & ^self & target

		for tSqr := board.BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()
			// this accounts for flipping the castling rights for the moving side
			// if the rook moves away from castling position and also for the
			// opponent when a rook is capturing a rook in castling position
			newC := b.CRights & ^(rookCRightsUpd[from] | rookCRightsUpd[to])
			m := ms.Alloc()
			m.Piece = Rook
			m.From = from
			m.To = to
			m.CRights = newC ^ b.CRights
			m.Castle = 0
			m.Promo = 0
			m.EPP = 0
			m.EPSq = b.EnPassant
		}
	}

	// queen moves
	queens := self & b.Pieces[Queen]
	for queen := board.BitBoard(0); queens != 0; queens ^= queen {
		queen = queens & -queens
		from := queen.LowestSet()
		mask := rookMasks[from]
		magic := rookMagics[from]
		shift := rookShifts[from]

		tSqrs := rookAttacks[from][((occ&mask)*magic)>>(64-shift)]

		mask = bishopMasks[from]
		magic = bishopMagics[from]
		shift = bishopShifts[from]

		tSqrs |= bishopAttacks[from][((occ&mask)*magic)>>(64-shift)]
		tSqrs &= ^self & target

		for tSqr := board.BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()
			cNew := b.CRights &^ rookCRightsUpd[to]
			m := ms.Alloc()
			m.Piece = Queen
			m.From = from
			m.To = to
			m.CRights = cNew ^ b.CRights
			m.Castle = 0
			m.Promo = 0
			m.EPP = 0
			m.EPSq = b.EnPassant
		}
	}

	mySndRank := sndRank[b.STM]
	theirSndRank := sndRank[b.STM.Flip()]

	// since shifting by negative is illegal, I bite the bullet and branch on STM
	var (
		occ1, occ2   board.BitBoard
		occ1l, occ1r board.BitBoard
		tgt1, tgt2   board.BitBoard
		shift        int
	)
	if b.STM == White {
		occ1 = occ >> 8
		tgt1 = target >> 8
		occ1l = (them &^ board.HFile) >> 7
		occ1r = (them &^ board.AFile) >> 9
		occ2 = occ >> 16
		tgt2 = target >> 16
		shift = 8
	} else {
		occ1 = (occ | ^target) << 8
		tgt1 = target << 8
		occ1l = (them &^ board.AFile) << 7
		occ1r = (them &^ board.HFile) << 9
		occ2 = occ << 16
		tgt2 = target << 16
		shift = -8
	}

	pushable := self & b.Pieces[Pawn] & ^occ1

	// single pawn pushes (no promotions)
	for pawns, pawn := pushable&tgt1 & ^theirSndRank, board.BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()

		m := ms.Alloc()
		m.Piece = Pawn
		m.From = from
		m.To = Square(int(from) + shift)
		m.CRights = 0
		m.Castle = 0
		m.Promo = 0
		m.EPP = 0
		m.EPSq = b.EnPassant
	}

	// promotions pushes
	for pawns, pawn := pushable&tgt1&theirSndRank, board.BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()
		for promo := Queen; promo > Pawn; promo-- {
			m := ms.Alloc()
			m.Piece = Pawn
			m.From = from
			m.To = Square(int(from) + shift)
			m.CRights = 0
			m.Promo = promo
			m.Castle = 0
			m.EPP = 0
			m.EPSq = b.EnPassant
		}
	}

	// double pawn pushes
	for pawns, pawn := pushable&tgt2&mySndRank & ^occ2, board.BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()

		m := ms.Alloc()
		m.Piece = Pawn
		m.From = from
		m.To = Square(int(from) + 2*shift)
		m.CRights = 0
		m.EPSq = m.To ^ b.EnPassant
		m.Castle = 0
		m.Promo = 0
		m.EPP = 0
	}

	// pawn captures (no promotions)
	for pawns, pawn := self&b.Pieces[Pawn] & ^theirSndRank & (occ1l|occ1r), board.BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()
		var bb board.BitBoard

		if b.STM == White {
			bb = ((pawn & ^board.AFile) << 7) | ((pawn & ^board.HFile) << 9)
		} else {
			bb = ((pawn & ^board.HFile) >> 7) | ((pawn & ^board.AFile) >> 9)
		}

		tSqrs := bb & target & them

		for tSqr := board.BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()

			m := ms.Alloc()
			m.Piece = Pawn
			m.From = from
			m.To = to
			m.CRights = 0
			m.Castle = 0
			m.Promo = 0
			m.EPP = 0
			m.EPSq = b.EnPassant
		}
	}

	// pawn captures with promotions
	for pawns, pawn := self&b.Pieces[Pawn]&theirSndRank&(occ1l|occ1r), board.BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()
		var bb board.BitBoard

		if b.STM == White {
			bb = ((pawn & ^board.AFile) << 7) | ((pawn & ^board.HFile) << 9)
		} else {
			bb = ((pawn & ^board.HFile) >> 7) | ((pawn & ^board.AFile) >> 9)
		}

		tSqrs := bb & target & them

		for tSqr := board.BitBoard(0); tSqrs != 0; tSqrs ^= tSqr {
			tSqr = tSqrs & -tSqrs
			to := tSqr.LowestSet()
			cNew := b.CRights &^ rookCRightsUpd[to]

			for promo := Queen; promo > Pawn; promo-- {
				m := ms.Alloc()
				m.Piece = Pawn
				m.From = from
				m.To = to
				m.Promo = promo
				m.CRights = cNew ^ b.CRights
				m.Castle = 0
				m.EPP = 0
				m.EPSq = b.EnPassant
			}
		}
	}

	// en-passant
	ep := (((1 << b.EnPassant) << 1) | ((1 << b.EnPassant) >> 1)) & fourthRank[b.STM.Flip()]
	for pawns, pawn := ep&self&b.Pieces[Pawn], board.BitBoard(0); pawns != 0; pawns ^= pawn {
		pawn = pawns & -pawns
		from := pawn.LowestSet()

		m := ms.Alloc()
		m.Piece = Pawn
		m.From = from
		m.To = Square(int(b.EnPassant) + shift)
		m.EPP = Pawn
		m.CRights = 0
		m.Castle = 0
		m.Promo = 0
		m.EPSq = b.EnPassant
	}

	// castling short
	if b.CRights&CRights(C(b.STM, Short)) != 0 && occ&castleMask[b.STM][Short] == self&b.Pieces[King] {
		if !IsAttacked(b, b.STM.Flip(), castleMask[b.STM][Short]) {
			from := (self & b.Pieces[King]).LowestSet()
			newC := b.CRights & ^kingCRightsUpd[b.STM]
			m := ms.Alloc()
			m.Piece = King
			m.From = from
			m.To = from + 2
			m.Castle = C(b.STM, Short)
			m.CRights = b.CRights ^ newC
			m.Promo = 0
			m.EPP = 0
			m.EPSq = b.EnPassant
		}
	}

	// castle long
	if b.CRights&CRights(C(b.STM, Long)) != 0 && occ&(castleMask[b.STM][Long]>>1) == 0 {
		if !IsAttacked(b, b.STM.Flip(), castleMask[b.STM][Long]) {
			from := (self & b.Pieces[King]).LowestSet()
			newC := b.CRights & ^kingCRightsUpd[b.STM]
			m := ms.Alloc()
			m.Piece = King
			m.From = from
			m.To = from - 2
			m.Castle = C(b.STM, Long)
			m.CRights = b.CRights ^ newC
			m.Promo = 0
			m.EPP = 0
			m.EPSq = b.EnPassant
		}
	}

}

func IsAttacked(b *board.Board, by Color, target board.BitBoard) bool {
	other := b.Colors[by]
	occ := b.Colors[White] | b.Colors[Black]

	// pawn capture
	{
		var bb board.BitBoard

		if by == White {
			bb = ((target & ^board.HFile) >> 7) | ((target & ^board.AFile) >> 9)
		} else {
			bb = ((target & ^board.AFile) << 7) | ((target & ^board.HFile) << 9)
		}

		if bb&b.Pieces[Pawn]&other != 0 {
			return true
		}
	}

	for tSqr := board.BitBoard(0); target != 0; target ^= tSqr {
		tSqr = target & -target
		sq := tSqr.LowestSet()

		if KingMoves(sq)&b.Pieces[King]&other != 0 {
			return true
		}

		if KnightMoves(sq)&b.Pieces[Knight]&other != 0 {
			return true
		}

		// bishop or queen moves
		if BishopMoves(sq, occ)&(b.Pieces[Queen]|b.Pieces[Bishop])&other != 0 {
			return true
		}

		// rook or queen moves
		if RookMoves(sq, occ)&(b.Pieces[Rook]|b.Pieces[Queen])&other != 0 {
			return true
		}
	}

	return false
}

func FromSimple(b *board.Board, sm move.SimpleMove) move.Move {
	pType := b.SquaresToPiece[sm.From]
	result := move.Move{SimpleMove: sm, Piece: pType, EPSq: b.EnPassant}

	switch pType {

	case King:
		if sm.From-sm.To == 2 || sm.To-sm.From == 2 {
			newC := b.CRights & ^kingCRightsUpd[b.STM]
			result.CRights = newC ^ b.CRights
			result.Castle = C(b.STM, int(((sm.From-sm.To)+2)/4))
		} else {
			newC := b.CRights & ^(kingCRightsUpd[b.STM] | rookCRightsUpd[sm.To])
			result.CRights = newC ^ b.CRights
		}

	case Knight, Bishop, Queen:
		newC := b.CRights & ^(rookCRightsUpd[sm.To])
		result.CRights = newC ^ b.CRights

	case Rook:
		newC := b.CRights & ^(rookCRightsUpd[sm.From] | rookCRightsUpd[sm.To])
		result.CRights = newC ^ b.CRights

	case Pawn:
		if sm.From-sm.To == 16 || sm.To-sm.From == 16 {
			result.EPSq ^= sm.To
		}
		if (sm.From-sm.To)&1 != 0 && b.SquaresToPiece[sm.To] == NoPiece { // en-passant capture
			result.EPP = Pawn
		}
		newC := b.CRights &^ rookCRightsUpd[sm.To]
		result.CRights = newC ^ b.CRights
	}

	return result
}
