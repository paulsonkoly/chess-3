package eval

type CoeffSet[T ScoreType] struct {
	// PSqT is tapered piece square tables.
	PSqT [12][64]T

	// PieceValues is tapered piece values between middle game and end game.
	PieceValues [2][7]T

	// TempoBonus is the advantage of the side to move.
	TempoBonus [2]T

	// KingAttackPieces is the bonus per piece type if piece is attacking a square in the enemy king's neighborhood.
	KingAttackPieces [4]T
	// SafeChecks is the bonus per piece type for being able to give a safe check.
	SafeChecks [4]T
	// UnsafeChecks is the bonus per piece type for being able to give an unsafe check.
	UnsafeChecks [4]T
	// KingOpenFile is the bonus for enemy king missing shelter pawn, per 3 files.
	KingOpenFile [3]T
	// KingShelter is the bonus for damage on the opponent's king shelter, per 3 files per pawn distance.
	KingShelter [3][7]T
	// KingStorm is the bonus pawn storming the opponent's king, per 3 files per pawn distance.
	KingStorm [3][7]T
	// KingAttackMagnitude encodes the importance of attacking the enemy king.
	KingAttackMagnitude [2]T

	// PawnlessFlank is the penalty for a king being on a pawnless flank.
	PawnlessFlank [2]T

	// Mobility* is per piece mobility bonus.
	MobilityKnight [2][9]T
	MobilityBishop [2][14]T
	MobilityRook   [2][15]T

	// KnightOutpost is a per square bonus for a knight being on an outpost, only
	// counting the 5 ranks covering sideOfBoard.
	KnightOutpost [2][40]T
	// Knight is behind an either enemy, or friendly pawn.
	KnightBehindPawn [2]T

	// BishopPair is the bonus for bishop pair per friendly pawn count.
	BishopPair [9]T
	// Bishop outpost is the bonus for a bishop being on an outpost square.
	BishopOutpost [2]T
	// OppositeColoredBishops is the scale factor for opposite colored bishop drawishness.
	// Indexed by other piece presence: nothing, pair of knights, rooks, queens; and pawn count difference.
	OppositeColoredBishops [3][4]T

	// ConnectedRooks is the bonus if rooks are connected.
	ConnectedRooks [2]T
	// RookOnOpen is the bonus for rook on open file.
	RookOnOpen [2]T
	// RookOnSemiOpen is the bonus for rook on semi-open file.
	RookOnSemiOpen [2]T

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
	// Phalanx is the per rank bonus for phalanx pawns.
	Phalanx [2][7]T

	// SafePawnThreats is the bonus for a safe - either unattacked or defended pawn attacking an enemy (non-pawn) piece.
	SafePawnThreats [2]T
	//Threats is the bonus for threatening an enemy piece.
	Threats [2]T

	// InsufficientKnight is the score reduction factor for the strong side only having a knight.
	InsufficientKnight T
	// InsufficientBishop is the score reduction factor for the strong side only having a bishop.
	InsufficientBishop T
	// KRNvKR is the score reduction factor for the side of RN versus R.
	KRNvKR T
	// KRBvKR is the score reduction factor for the side of RB versus R.
	KRBvKR T
}
