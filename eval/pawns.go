package eval

import (
	"github.com/paulsonkoly/chess-3/attacks"
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
)

func (p *Pawns) calc(b *board.Board, color Color) {
	pawns := b.Colors[color] & b.Pieces[Pawn]

	frontfill := frontFill(pawns, color)
	p.frontspan = attacks.PawnSinglePushMoves(frontfill, color)
	rearSpan := attacks.PawnSinglePushMoves(frontFill(pawns, color.Flip()), color.Flip())

	files := pawns | p.frontspan | rearSpan
	p.neighbourF = ((files & ^AFileBB) >> 1) | ((files & ^HFileBB) << 1)

	p.frontline = ^rearSpan & pawns
	p.backmost = ^p.frontspan & pawns
	p.cover = attacks.PawnCaptureMoves(frontfill, color)
}

const PawnCacheSize = 16 * 1024

type PawnCache struct {
	hash  board.Hash
	accum [Colors][Phases]Score
}

const PawnKingCacheSize = 16 * 1024

type PawnKingCache struct {
	hash  board.Hash
	accum [Colors]Score
}

func (e *Eval[T]) addPawns(b *board.Board, c *CoeffSet[T]) {
	e.addPassers(b, c)

	var (
		t     T
		hash  board.Hash
		accum [Colors][Phases]T
	)

	if _, ok := any(t).(Score); ok {
		hash = b.Hashes().Pawn

		if e.pawnCache[hash%PawnCacheSize].hash == hash {
			entry := &e.pawnCache[hash%PawnCacheSize].accum
			e.sp[White][MG] += T(entry[White][MG])
			e.sp[White][EG] += T(entry[White][EG])
			e.sp[Black][MG] += T(entry[Black][MG])
			e.sp[Black][EG] += T(entry[Black][EG])
			return
		}
	}

	for color := range Colors {
		pawns := b.Colors[color] & b.Pieces[Pawn]

		dblCnt := T(e.doubledPawns(pawns, color).Count())
		accum[color][MG] += c.DoubledPawns[MG] * dblCnt
		accum[color][EG] += c.DoubledPawns[EG] * dblCnt

		isoCnt := T(e.isolatedPawns(pawns, color).Count())
		accum[color][MG] += c.IsolatedPawns[MG] * isoCnt
		accum[color][EG] += c.IsolatedPawns[EG] * isoCnt

		for phalanxes := ((pawns & ^AFileBB) >> 1) & pawns; phalanxes != 0; phalanxes &= phalanxes - 1 {
			rank := phalanxes.LowestSet().Rank().FromPerspectiveOf(color)
			accum[color][MG] += c.Phalanx[MG][rank]
			accum[color][EG] += c.Phalanx[EG][rank]
		}

		for passers := e.passers(color); passers != 0; passers &= passers - 1 {
			sq := passers.LowestSet()

			rank := sq / 8
			if color == Black {
				rank ^= 7
			}

			passer := passers & -passers

			// if protected passers add protection bonus
			if passer&e.attacks[color][Pawn] != 0 {
				accum[color][MG] += c.ProtectedPasser[MG]
				accum[color][EG] += c.ProtectedPasser[EG]
			}

			accum[color][MG] += c.PasserRank[0][rank-1]
			accum[color][EG] += c.PasserRank[1][rank-1]
		}

		for pieces := pawns; pieces != 0; pieces &= pieces - 1 {
			sq := pieces.LowestSet()

			if color == White {
				sq ^= 56 // upside down
			}

			accum[color][MG] += c.PSqT[0][sq]
			accum[color][EG] += c.PSqT[1][sq]

			accum[color][MG] += c.PieceValues[MG][Pawn]
			accum[color][EG] += c.PieceValues[EG][Pawn]
		}
	}

	e.sp[White][MG] += accum[White][MG]
	e.sp[White][EG] += accum[White][EG]
	e.sp[Black][MG] += accum[Black][MG]
	e.sp[Black][EG] += accum[Black][EG]

	if _, ok := any(t).(Score); ok {
		e.pawnCache[hash%PawnCacheSize].hash = hash
		e.pawnCache[hash%PawnCacheSize].accum[White][MG] = Score(accum[White][MG])
		e.pawnCache[hash%PawnCacheSize].accum[White][EG] = Score(accum[White][EG])
		e.pawnCache[hash%PawnCacheSize].accum[Black][MG] = Score(accum[Black][MG])
		e.pawnCache[hash%PawnCacheSize].accum[Black][EG] = Score(accum[Black][EG])
	}
}

func (e *Eval[T]) addPassers(b *board.Board, c *CoeffSet[T]) {
	// KPR, KPNB
	if b.Pieces[Knight]|b.Pieces[Bishop]|b.Pieces[Queen] == 0 || b.Pieces[Rook]|b.Pieces[Queen] == 0 {

		for color := range Colors {

			// if there is a sole passer
			passers := e.passers(color)
			if passers != 0 && passers&(passers-1) == 0 {
				sq := passers.LowestSet()

				qSq := sq % 8
				if color == White {
					qSq += 56
				}

				kingDist := Chebishev(qSq, e.kings[color.Flip()].sq) - Chebishev(qSq, e.kings[color].sq)

				e.sp[color][MG] += c.PasserKingDist[MG] * T(kingDist)
				e.sp[color][EG] += c.PasserKingDist[EG] * T(kingDist)
			}
		}
	}
}

func Chebishev(a, b Square) int {
	ax, ay, bx, by := int(a%8), int(a/8), int(b%8), int(b/8)
	return max(Abs(ax-bx), Abs(ay-by))
}

// passers are pawns not stoppable by enemy pawns without them changing file.
func (e *Eval[T]) passers(color Color) BitBoard {
	return e.pawns[color].frontline & ^(e.pawns[color.Flip()].frontspan | (e.pawns[color.Flip()].cover))
}

// doubledPawns are pawns that have a friendly further advanced pawn on the same file.
func (e *Eval[T]) doubledPawns(pawns BitBoard, color Color) BitBoard {
	return pawns &^ e.pawns[color].frontline
}

// isolatedPawns are pawns not having any friendly pawn on adjacent files.
func (e *Eval[T]) isolatedPawns(pawns BitBoard, color Color) BitBoard {
	return pawns &^ e.pawns[color].neighbourF
}

// outposts are squares defended by our pawns and not attackable by any enemy pawn.
func (e *Eval[T]) outposts(color Color) BitBoard {
	return ^e.pawns[color.Flip()].cover & e.attacks[color][Pawn]
}

func (e *Eval[T]) addPawnlessFlank(b *board.Board, c *CoeffSet[T]) {
	pawns := b.Pieces[Pawn]
	for color := range Colors {
		if FileCluster(e.kings[color].sq.File())&pawns == 0 {
			e.sp[color][MG] += c.PawnlessFlank[MG]
			e.sp[color][EG] += c.PawnlessFlank[EG]
		}
	}
}

func (e *Eval[T]) addStormShelter(b *board.Board, c *CoeffSet[T]) {
	var (
		t     T
		hash  board.Hash
		accum [2]T
	)

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

		onFlank := 0
		if b.Colors[color.Flip()]&b.Pieces[King]&(DFileBB|EFileBB) == 0 {
			onFlank = 1
		}

		kFile := eKing.File()
		kRank := eKing.Rank()

		frontLine := e.pawns[color].frontline
		backMost := e.pawns[color.Flip()].backmost

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
				accum[color] += c.KingStorm[onFlank][0][dist]
			}

			if shelter := backMost & fileBB; shelter != 0 {
				dist := Abs(kRank - shelter.LowestSet().Rank())
				accum[color] -= c.KingShelter[onFlank][0][dist]
			} else {
				accum[color] += c.KingOpenFile[0]
			}
		}

		// front
		{
			fileBB := FileBB(front & 7)

			if storm := frontLine & fileBB; storm != 0 {
				dist := Abs(kRank - storm.LowestSet().Rank())
				accum[color] += c.KingStorm[onFlank][1][dist]
			}

			if shelter := backMost & fileBB; shelter != 0 {
				dist := Abs(kRank - shelter.LowestSet().Rank())
				accum[color] -= c.KingShelter[onFlank][1][dist]
			} else {
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
				accum[color] += c.KingStorm[onFlank][2][dist]
			}

			if shelter := backMost & fileBB; shelter != 0 {
				dist := Abs(kRank - shelter.LowestSet().Rank())
				accum[color] -= c.KingShelter[onFlank][2][dist]
			} else {
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

func frontFill(b BitBoard, color Color) BitBoard {
	switch color {
	case White:
		b |= b << 8
		b |= b << 16
		b |= b << 32

	case Black:
		b |= b >> 8
		b |= b >> 16
		b |= b >> 32
	}

	return b
}
