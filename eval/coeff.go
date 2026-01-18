package eval

type CoeffSet[T ScoreType] struct {
	// PSqT is tapered piece square tables.
	PSqT [12][64]T

	// PieceValues is tapered piece values between middle game and end game.
	PieceValues [2][7]T

	// TempoBonus is the advantage of the side to move.
	TempoBonus [2]T

	// KingAttackPieces is the bonus per piece type if piece is attacking a square in the enemy king's neighborhood.
	KingAttackPieces [2][4]T

	// SafeChecks is the bonus per piece type for being able to give a safe check.
	SafeChecks [2][4]T

	// KingShelter is the bonus for damage on the opponent's king shelter.
	KingShelter [2]T

	// Mobility* is per piece mobility bonus.
	MobilityKnight [2][9]T
	MobilityBishop [2][14]T
	MobilityRook   [2][11]T

	// KnightOutpost is a per square bonus for a knight being on an outpost, only
	// counting the 5 ranks covering sideOfBoard.
	KnightOutpost [2][40]T

	// ConnectedRooks is a bonus if rooks are connected.
	ConnectedRooks [2]T

	// BishopPair is the bonus for bishop pair per friendly pawn count.
	BishopPair [9]T

	// OpBishop is the drawishness of an opposite coloured bishop endgame.
	OpBishops               [1]T
	// OpBishopsOutsidePassers is the reduction on drawishness if the winning
	// side has outside passed pawns on both flanks.
	OpBishopsOutsidePassers [1]T
	// OpBishopsPawnDelta is the reduction on drawishness based on the pawn delta
	// between the two players. Indexed by delta.
	OpBishopsPawnDelta      [4]T

	// ProtectedPasser is the bonus for each protected passed pawn.
	ProtectedPasser [2]T
	// PasserKingDist is the bonus for our king being close / enemy king being far from passed pawn.
	PasserKingDist [2]T
	// PasserRank is the bonus for the passed pawn being on a specific rank.
	PasserRank [2][6]T
	// DoubledPawns is the penalty per doubled pawn (count of non-frontline pawns ie. the pawns in the pawn rearspan).
	DoubledPawns [2]T
	// IsolatedPawns is the penalty per isolated pawns.
	IsolatedPawns [2]T
}
