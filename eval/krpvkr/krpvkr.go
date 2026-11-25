// Package krpvkr provides win/draw classification for krpvkr positions.
package krpvkr

import (
	"github.com/paulsonkoly/chess-3/attacks"
	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
)

// ScaleFactor returns the drawishness of a krpvkr position in 0..MaxScaleFactor range.
func ScaleFactor[T ScoreType](b *board.Board) T {
	// feature classifier with a depth-12 decesion tree
	// https://en.wikipedia.org/wiki/Decision_tree_learning
	f := features(b)

	node := nodeIx(0)
	for {
		n := nodes[node]
		if n.featureIx < 0 {
			return T(scaleFactors[node])
		}
		if f[n.featureIx] <= n.cutoff {
			node = n.left
		} else {
			node = n.right
		}
	}
}

const (
	featureCount = 20
)

const (
	pawnFile = iota
	pawnRank
	stm
	strongKPFileD
	strongKPRankD
	strongRPFileD
	strongRPRankD
	weakKPFileD
	weakKPRankD
	weakRPFileD
	weakRPRankD
	weakRRank
	strongRR
	weakRAttacked
	rookFileMatch
	rookRankMatch
	strongKQSqD
	weakKQSqD
	strongKPD
	weakKPD
)

type feature = int8

func features(b *board.Board) [featureCount]feature {
	pawn := b.Pieces[Pawn]

	strongSide := White
	if pawn&b.Colors[Black] != 0 {
		strongSide = Black
	}
	weakSide := strongSide.Flip()

	pawnSq := pawn.LowestSet()
	strongKSq := (b.Pieces[King] & b.Colors[strongSide]).LowestSet()
	strongRSq := (b.Pieces[Rook] & b.Colors[strongSide]).LowestSet()
	weakKSq := (b.Pieces[King] & b.Colors[weakSide]).LowestSet()
	weakRSq := (b.Pieces[Rook] & b.Colors[weakSide]).LowestSet()

	// vertical flip, krpvkr only supports if white is pushing for a win
	if strongSide == Black {
		pawnSq ^= 56
		strongKSq ^= 56
		strongRSq ^= 56
		weakKSq ^= 56
		weakRSq ^= 56
	}

	// horizontal flip, krpvkr only support pawn on a-d files
	if pawnSq.File() >= EFile {
		pawnSq ^= 7
		strongKSq ^= 7
		strongRSq ^= 7
		weakKSq ^= 7
		weakRSq ^= 7
	}

	pawn = BitBoard(1) << BitBoard(pawnSq)
	weakR := BitBoard(1) << BitBoard(weakRSq)

	// occupoancy from the normalised board view
	occ := pawn | BitBoardFromSquares(strongKSq, strongRSq, weakKSq, weakRSq)

	strongAttacks := attacks.KingMoves(strongKSq) |
		attacks.PawnCaptureMoves(pawn, White) |
		attacks.RookMoves(strongRSq, occ)

	queeningSq := SquareAt(pawnSq.File(), EighthRank)

	var f [featureCount]feature
	f[pawnFile] = feature(pawnSq.File())
	f[pawnRank] = feature(pawnSq.Rank())
	f[stm] = bToF(b.STM == weakSide)
	f[strongKPFileD] = feature(strongKSq.File()) - int8(pawnSq.File())
	f[strongKPRankD] = feature(strongKSq.Rank()) - int8(pawnSq.Rank())
	f[strongRPFileD] = feature(strongRSq.File()) - int8(pawnSq.File())
	f[strongRPRankD] = feature(strongRSq.Rank()) - int8(pawnSq.Rank())
	f[weakKPFileD] = feature(weakKSq.File()) - int8(pawnSq.File())
	f[weakKPRankD] = feature(weakKSq.Rank()) - int8(pawnSq.Rank())
	f[weakRPFileD] = feature(weakRSq.File()) - int8(pawnSq.File())
	f[weakRPRankD] = feature(weakRSq.Rank()) - int8(pawnSq.Rank())
	f[weakRRank] = feature(weakRSq.Rank())
	f[strongRR] = feature(strongRSq.Rank())
	f[weakRAttacked] = bToF(strongAttacks&weakR != 0)
	f[rookFileMatch] = bToF(strongRSq.File() == weakRSq.File())
	f[rookRankMatch] = bToF(strongRSq.Rank() == weakRSq.Rank())
	f[strongKQSqD] = feature(Chebyshev(strongKSq, queeningSq))
	f[weakKQSqD] = feature(Chebyshev(weakKSq, queeningSq))
	f[strongKPD] = feature(Chebyshev(strongKSq, pawnSq))
	f[weakKPD] = feature(Chebyshev(weakKSq, pawnSq))

	return f
}

func bToF(b bool) feature {
	if b {
		return 1
	}
	return 0
}
