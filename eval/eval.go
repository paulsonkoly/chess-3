package eval

import (
	"github.com/paulsonkoly/chess-3/attacks"
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
)

// ScoreType defines the evaluation result type. The engine uses int16 for
// score type, as defined in types. The tuner uses float64.
type ScoreType interface{ Score | float64 }

type Phase byte

const (
	MG = Phase(iota)
	EG

	Phases
)

type Eval[T ScoreType] struct {
	sp            [Colors][Phases]T
	scaleFactor   T
	kingAttacks   [Colors]T
	attacks       [Colors][Pieces]BitBoard
	cover         [Colors]BitBoard
	pawns         [Colors]Pawns
	kings         [Colors]Kings
	pawnCache     []PawnCache
	pawnKingCache []PawnKingCache
}

type Pawns struct {
	cover      BitBoard
	frontline  BitBoard
	backmost   BitBoard
	frontspan  BitBoard
	neighbourF BitBoard
}

type Kings struct {
	nb BitBoard
	sq Square
}

func New[T ScoreType]() *Eval[T] {
	return &Eval[T]{
		pawnCache:     make([]PawnCache, PawnCacheSize),
		pawnKingCache: make([]PawnKingCache, PawnKingCacheSize),
	}
}

func (e *Eval[T]) Clear() {
	for i := range e.pawnCache {
		e.pawnCache[i] = PawnCache{}
	}

	for i := range e.pawnKingCache {
		e.pawnKingCache[i] = PawnKingCache{}
	}
}

func (e *Eval[T]) Score(b *board.Board, c *CoeffSet[T]) T {
	e.sp = [Colors][Phases]T{}
	e.kingAttacks = [Colors]T{}
	e.attacks = [Colors][Pieces]BitBoard{}

	if insufficient(b) {
		return 0
	}

	if isKNBvK(b) {
		e.knbvk(b, c)
		return e.endgameScore(b)
	}
	_ = e.scaleBKvBK(b, c) || e.scaleFifty(b)

	for color := range Colors {
		e.pawns[color].calc(b, color)
		e.kings[color].calc(b, color)
	}

	for color := range Colors {
		pawns := b.Colors[color] & b.Pieces[Pawn]
		attacked := attacks.PawnCaptureMoves(pawns, color)
		e.attacks[color][Pawn] = attacked
		e.cover[color] = attacked
		attacked = attacks.KingMoves(e.kings[color].sq)
		e.attacks[color][King] = attacked
		e.cover[color] |= attacked
	}

	occ := b.Colors[White] | b.Colors[Black]

	for color := range Colors {
		// enemy king neighbourhood
		eKNb := e.kings[color.Flip()].nb

		// queens
		for pieces := b.Pieces[Queen] & b.Colors[color]; pieces != 0; pieces &= pieces - 1 {
			sq := pieces.LowestSet()

			attacks := attacks.BishopMoves(sq, occ) | attacks.RookMoves(sq, occ)
			e.attacks[color][Queen] |= attacks
			e.cover[color] |= attacks

			e.addKingNBAttack(color, Queen, attacks, eKNb, c)
			e.addPSqT(color, Queen, sq, c)
			e.addPieceValue(color, Queen, c)
		}

		// rooks
		for pieces := b.Pieces[Rook] & b.Colors[color]; pieces != 0; pieces &= pieces - 1 {
			sq := pieces.LowestSet()

			attacks := attacks.RookMoves(sq, occ)
			e.attacks[color][Rook] |= attacks
			e.cover[color] |= attacks

			e.addKingNBAttack(color, Rook, attacks, eKNb, c)
			e.addRookMobility(b, color, attacks, c)
			e.addRookFiles(b, color, sq, c)
			e.addPSqT(color, Rook, sq, c)
			e.addPieceValue(color, Rook, c)
		}

		outposts := e.outposts(color)

		// bishops
		for pieces := b.Pieces[Bishop] & b.Colors[color]; pieces != 0; pieces &= pieces - 1 {
			sq := pieces.LowestSet()

			attacks := attacks.BishopMoves(sq, occ)
			e.attacks[color][Bishop] |= attacks
			e.cover[color] |= attacks

			e.addKingNBAttack(color, Bishop, attacks, eKNb, c)
			e.addBishopMobility(b, color, attacks, c)
			e.addBishopOutposts(color, sq, outposts, c)
			e.addPSqT(color, Bishop, sq, c)
			e.addPieceValue(color, Bishop, c)
		}

		// knights
		e.addKnightBehindPawn(b, color, c)
		for pieces := b.Pieces[Knight] & b.Colors[color]; pieces != 0; pieces &= pieces - 1 {
			sq := pieces.LowestSet()

			attacks := attacks.KnightMoves(sq)
			e.attacks[color][Knight] |= attacks
			e.cover[color] |= attacks

			e.addKingNBAttack(color, Knight, attacks, eKNb, c)
			e.addKnightMobility(b, color, attacks, c)
			e.addKnightOutposts(color, sq, outposts, c)
			e.addPSqT(color, Knight, sq, c)
			e.addPieceValue(color, Knight, c)
		}

		// king
		e.addPSqT(color, King, e.kings[color].sq, c)
	}

	e.addTempo(b, c)
	e.addBishopPair(b, c)
	e.addPawns(b, c)
	e.addPawnlessFlank(b, c)
	e.addThreats(b, c)
	e.addChecks(b, c)
	e.addStormShelter(b, c)

	e.addKingAttacks(c)

	return e.taperedScore(b)
}
