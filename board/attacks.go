package board

import (
	"github.com/paulsonkoly/chess-3/attacks"
	. "github.com/paulsonkoly/chess-3/chess"
)

func (b *Board) Attackers(squares BitBoard, occ BitBoard, color Color) BitBoard {
	opp := b.Colors[color]
	var res BitBoard

	for sqrs, sqBB := squares, BitBoard(0); sqrs != 0; sqrs ^= sqBB {
		sqBB = sqrs & -sqrs
		sq := sqBB.LowestSet()

		sub := attacks.KingMoves(sq) & b.Pieces[King]
		sub |= attacks.KnightMoves(sq) & b.Pieces[Knight]
		sub |= attacks.BishopMoves(sq, occ) & (b.Pieces[Bishop] | b.Pieces[Queen])
		sub |= attacks.RookMoves(sq, occ) & (b.Pieces[Rook] | b.Pieces[Queen])

		res |= sub & opp
	}

	res |= attacks.PawnCaptureMoves(squares, color.Flip()) & opp & b.Pieces[Pawn]

	return res
}

var (
	myFourthRank = [...]BitBoard{FourthRank, FifthRank}
)

func (b *Board) Block(squares BitBoard, color Color) BitBoard {
	blockers := b.Colors[color]
	res := BitBoard(0)
	occ := b.Colors[White] | b.Colors[Black]

	for square, eachSquare := BitBoard(0), squares; eachSquare != 0; eachSquare ^= square {
		square = eachSquare & -eachSquare
		sq := square.LowestSet()

		sub := BitBoard(0)

		/* king can't block */
		sub |= attacks.KnightMoves(sq) & b.Pieces[Knight]
		sub |= attacks.BishopMoves(sq, occ) & (b.Pieces[Bishop] | b.Pieces[Queen])
		sub |= attacks.RookMoves(sq, occ) & (b.Pieces[Rook] | b.Pieces[Queen])

		res |= sub & blockers
	}

	// we are making a pawn move backwards, so ignore the pawn in occupancy, as
	// we are moving where the actual pawn is, but don't ignore a blocking pawn
	// otherwise we would jump over it. See:
	// 6k1/8/8/1b6/3PP3/r1PKP3/2PRB3/8 w - - 0 1
	occNoPawn := occ & ^(b.Pieces[Pawn] & blockers)

	/* double pawn push blocking */
	dpawn := myFourthRank[color] & squares
	dpawn = attacks.PawnSinglePushMoves(dpawn, color.Flip()) &^ occ
	dpawn = attacks.PawnSinglePushMoves(dpawn, color.Flip()) &^ occNoPawn

	res |= ((attacks.PawnSinglePushMoves(squares, color.Flip()) & ^occNoPawn) | dpawn) & blockers & b.Pieces[Pawn]

	return res
}

// IsCheckmate determines whether the position is checkmate. The king should be
// in check.
func (b *Board) IsCheckmate() bool {
	king := b.Pieces[King] & b.Colors[b.STM]
	occ := b.Colors[White] | b.Colors[Black]
	opp := b.Colors[b.STM.Flip()]

	attackers := b.Attackers(king, occ, b.STM.Flip())

	// making the king move first
	kingSq := king.LowestSet()
	kMvs := attacks.KingMoves(kingSq) & ^b.Colors[b.STM]

	for to := BitBoard(0); kMvs != 0; kMvs ^= to {
		to = kMvs & -kMvs

		if !b.IsAttacked(b.STM.Flip(), occ&^king, to) {
			return false
		}
	}

	if attackers.Count() > 1 { // double check, and king can't move
		return true
	}

	attacker := attackers // only 1 attacker

	//  see if we can capture the attacker
	defenders := b.Attackers(attacker, occ, b.STM)
	// remove the king, if the king can capture the attacker it would have done
	// in the king moves try
	defenders &= ^king

	// are all my defenders pinned in a way that they can't capture the attacker
	for defender := BitBoard(0); defenders != 0; defenders ^= defender {
		defender = defenders & -defenders
		nocc := occ
		pinned := false

		//  dummy mk move
		nocc &= ^defender
		opp &= ^attacker

		if (attacks.BishopMoves(kingSq, nocc) & (b.Pieces[Bishop] | b.Pieces[Queen]) & opp) != 0 {
			pinned = true
		} else if (attacks.RookMoves(kingSq, nocc) & (b.Pieces[Rook] | b.Pieces[Queen]) & opp) != 0 {
			pinned = true
		}

		if !pinned {
			return false
		}
	}

	// en passant capture
	if b.EnPassant != 0 {
		epPawn := attacks.PawnSinglePushMoves(BitBoard(1)<<b.EnPassant, b.STM.Flip())

		if epPawn == attacker {
			return false
		}
	}

	// block the attacker
	aSq := attacker.LowestSet()
	blocked := attacks.InBetween(kingSq, aSq) & ^(king | attacker)

	defenders = b.Block(blocked, b.STM)

	for defender := BitBoard(0); defenders != 0; defenders ^= defender {
		defender = defenders & -defenders
		nocc := occ
		opp := b.Colors[b.STM.Flip()]
		pinned := false

		//  dummy mk move
		nocc &= ^defender
		// we move somewhere on the blocked squares
		nocc |= blocked

		if (attacks.BishopMoves(kingSq, nocc) & (b.Pieces[Bishop] | b.Pieces[Queen]) & opp) != 0 {
			pinned = true
		} else if attacks.RookMoves(kingSq, nocc)&(b.Pieces[Rook]|b.Pieces[Queen])&opp != 0 {
			pinned = true
		}

		if !pinned {
			return false
		}
	}

	return true
}

// IsStalemate determines whether the position is stalemate. The king shouldn't
// be in check.
func (b *Board) IsStalemate() bool {
	me := b.Colors[b.STM]
	opp := b.Colors[b.STM.Flip()]
	king := b.Pieces[King] & me
	kingSq := king.LowestSet()
	occ := me | opp

	// look at pawns guaranteed not to be pinned first
	maybePinned := (attacks.BishopMoves(kingSq, occ) | attacks.RookMoves(kingSq, occ)) & me

	// this should give an answer 99% of the time we also don't have to bother
	// with double pushes as if there is no single pawn push there can't be a
	// double pawn push
	//
	pieces := b.Pieces[Pawn] & me & ^maybePinned
	if b.STM == White {
		if pieces<<8 & ^occ != 0 {
			return false
		}

		if (((pieces & ^AFile)<<7)|((pieces & ^HFile)<<9))&opp != 0 {
			return false
		}

	} else {
		if pieces>>8 & ^occ != 0 {
			return false
		}

		if (((pieces & ^HFile)>>7)|((pieces & ^AFile)>>9))&opp != 0 {
			return false
		}
	}

	// queens can't be pinned to the extent that they can't move, for instance
	// they can always capture the pinner.
	pieces = b.Pieces[Queen] & me
	for piece := BitBoard(0); pieces != 0; pieces ^= piece {
		piece = pieces & -pieces
		sq := piece.LowestSet()

		if ((attacks.BishopMoves(sq, occ) | attacks.RookMoves(sq, occ)) & ^me) != 0 {
			return false
		}
	}

	// bishop can only be paralyzed by rook or queen but in case of queen not the
	// one it can capture
	pieces = b.Pieces[Bishop] & me
	for piece := BitBoard(0); pieces != 0; pieces ^= piece {
		piece = pieces & -pieces
		sq := piece.LowestSet()
		nocc := occ & ^piece

		if (attacks.RookMoves(kingSq, nocc) & (b.Pieces[Rook] | b.Pieces[Queen]) & opp) == 0 {
			if (attacks.BishopMoves(sq, nocc) & ^me) != 0 {
				return false
			}
		}
	}

	//  rooks can only be paralyzed by bishop or queen but in case of queen not the
	//   one it can capture
	pieces = b.Pieces[Rook] & me
	for piece := BitBoard(0); pieces != 0; pieces ^= piece {
		piece = pieces & -pieces
		sq := piece.LowestSet()
		nocc := occ & ^piece

		if (attacks.BishopMoves(kingSq, nocc) & (b.Pieces[Bishop] | b.Pieces[Queen]) & opp) == 0 {
			if (attacks.RookMoves(sq, nocc) & ^me) != 0 {
				return false
			}
		}
	}

	//  knight move in pins cannot be legal
	pieces = b.Pieces[Knight] & me
	for piece := BitBoard(0); pieces != 0; pieces ^= piece {
		piece = pieces & -pieces
		sq := piece.LowestSet()
		nocc := occ & ^piece
		pinned := false

		if (piece & maybePinned) != 0 {
			if attacks.BishopMoves(kingSq, nocc)&(b.Pieces[Bishop]|b.Pieces[Queen])&opp != 0 {
				pinned = true
			} else if attacks.RookMoves(kingSq, nocc)&(b.Pieces[Rook]|b.Pieces[Queen])&opp != 0 {
				pinned = true
			}
		}

		if !pinned && (attacks.KnightMoves(sq) & ^me != 0) {
			return false
		}
	}

	kMoves := attacks.KingMoves(kingSq) & ^me
	for kMove := BitBoard(0); kMoves != 0; kMoves ^= kMove {
		kMove = kMoves & -kMoves

		if !b.IsAttacked(b.STM.Flip(), occ&^king, kMove) {
			return false
		}
	}

	//  maybe pinned pawns
	pieces = b.Pieces[Pawn] & me & maybePinned
	for piece := BitBoard(0); pieces != 0; pieces ^= piece {
		piece = pieces & -pieces

		targets := attacks.PawnSinglePushMoves(piece, b.STM) & ^occ
		nocc := (occ & ^piece) | targets
		pinned := false

		if (attacks.BishopMoves(kingSq, nocc) & (b.Pieces[Bishop] | b.Pieces[Queen]) & opp) != 0 {
			pinned = true
		} else if (attacks.RookMoves(kingSq, nocc) & (b.Pieces[Rook] | b.Pieces[Queen]) & opp) != 0 {
			pinned = true
		}

		if !pinned && targets != 0 {
			return false
		}

		targets = attacks.PawnCaptureMoves(piece, b.STM) & opp
		nocc = (occ & ^piece) | targets
		pinned = false

		if (attacks.BishopMoves(kingSq, nocc) & (b.Pieces[Bishop] | b.Pieces[Queen]) & ^targets & opp) != 0 {
			pinned = true
		} else if (attacks.RookMoves(kingSq, nocc) & (b.Pieces[Rook] | b.Pieces[Queen]) & opp) != 0 {
			pinned = true
		}

		if !pinned && targets != 0 {
			return false
		}
	}

	//  finally deal with en passant
	if b.EnPassant != 0 {
		enPassantBB := BitBoard(1) << b.EnPassant
		pieces := attacks.PawnCaptureMoves(enPassantBB, b.STM.Flip()) & b.Pieces[Pawn] & me
		remove := attacks.PawnSinglePushMoves(enPassantBB, b.STM.Flip())

		for piece := BitBoard(0); pieces != 0; pieces ^= piece {
			piece = pieces & -pieces
			nocc := (occ & ^piece & ^remove) | enPassantBB
			pinned := false

			if (attacks.RookMoves(kingSq, nocc) & (b.Pieces[Rook] | b.Pieces[Queen]) & opp) != 0 {
				pinned = true
			} else if (attacks.BishopMoves(kingSq, nocc) & (b.Pieces[Bishop] | b.Pieces[Queen]) & opp) != 0 {
				pinned = true
			}

			if !pinned {
				return false
			}
		}
	}

	return true
}

func (b *Board) IsAttacked(by Color, occ, target BitBoard) bool {
	other := b.Colors[by]

	// pawn capture
	if attacks.PawnCaptureMoves(b.Pieces[Pawn]&other, by)&target != 0 {
		return true
	}

	for tSqr := BitBoard(0); target != 0; target ^= tSqr {
		tSqr = target & -target
		sq := tSqr.LowestSet()

		if attacks.KingMoves(sq)&b.Pieces[King]&other != 0 {
			return true
		}

		if attacks.KnightMoves(sq)&b.Pieces[Knight]&other != 0 {
			return true
		}

		// bishop or queen moves
		if attacks.BishopMoves(sq, occ)&(b.Pieces[Queen]|b.Pieces[Bishop])&other != 0 {
			return true
		}

		// rook or queen moves
		if attacks.RookMoves(sq, occ)&(b.Pieces[Rook]|b.Pieces[Queen])&other != 0 {
			return true
		}
	}

	return false
}

func (b *Board) InCheck(who Color) bool {
	return b.IsAttacked(who.Flip(), b.Colors[White]|b.Colors[Black], b.Colors[who]&b.Pieces[King])
}

var shifts = [2]Square{8, -8}

// CanEnPassant determines if we need to change the en passant state of the
// board after a double pawn push.
//
// This is important in order to have the right hashes for 3-fold repetition.
// If we didn't do this the next turn move generator would take care of things
// and everything would work, apart from we would have the incorrect board en
// passant state.
// https://chess.stackexchange.com/questions/777/rules-en-passant-and-draw-by-triple-repetition
func (b *Board) CanEnPassant(to Square) bool {
	target := BitBoard(1) << to
	them := b.Colors[b.STM.Flip()]
	shift := shifts[b.STM]
	king := b.Pieces[King] & them
	dest := BitBoard(1) << (to - shift)

	// pawns that are able to en-passant
	ables := ((target & ^AFile >> 1) | (target & ^HFile << 1)) & b.Pieces[Pawn] & them
	for able := BitBoard(0); ables != 0; ables ^= able {
		able = ables & -ables
		// remove the pawns from the occupancy
		occ := (b.Colors[White] | b.Colors[Black] | dest) &^ (target | able)
		if !b.IsAttacked(b.STM, occ, king) {
			return true
		}
	}
	return false
}
