package eval

import (
	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

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

var Coefficients = CoeffSet[Score]{
	PSqT: [12][64]Score{
		{
			0, 0, 0, 0, 0, 0, 0, 0,
			102, 112, 89, 111, 109, 67, -65, -64,
			9, 30, 51, 45, 61, 114, 77, 24,
			4, 16, 18, 31, 49, 49, 38, 13,
			-2, 2, 11, 19, 25, 24, 21, 7,
			-4, -4, 1, 5, 16, 18, 28, 12,
			-4, -6, -9, -4, -1, 29, 39, 7,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			0, 0, 0, 0, 0, 0, 0, 0,
			110, 101, 96, 47, 46, 64, 120, 123,
			37, 30, 1, -31, -38, -21, 1, 10,
			10, 0, -12, -34, -32, -27, -15, -14,
			-7, -7, -22, -26, -26, -22, -22, -23,
			-13, -15, -18, -20, -15, -16, -27, -28,
			-8, -12, -10, -11, -1, -11, -27, -34,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			-185, -130, -91, -25, 24, -104, -114, -141,
			-14, -10, 33, 56, 36, 73, -14, 13,
			-19, 18, 34, 39, 86, 73, 49, 18,
			0, 5, 21, 33, 21, 46, 21, 33,
			-10, 5, 13, 15, 25, 21, 35, 8,
			-26, -8, -2, 13, 19, 7, 7, -7,
			-39, -32, -16, 0, -2, 2, -9, -9,
			-87, -28, -44, -21, -17, -10, -24, -60,
		},
		{
			-34, 11, 22, 8, 6, 30, 12, -62,
			-12, 4, -1, 15, 10, -18, 6, -22,
			-4, 3, 32, 29, 11, 13, -9, -12,
			1, 14, 34, 40, 39, 29, 24, 5,
			10, 19, 45, 40, 45, 44, 22, 19,
			-17, 7, 17, 36, 32, 14, 7, -2,
			-2, 12, 0, 12, 15, -4, 4, 7,
			-13, -27, -2, 9, 10, -3, -14, -16,
		},
		{
			-59, -58, -94, -97, -89, -124, -34, -55,
			-46, -9, -12, -31, -7, -8, -23, -52,
			-9, 6, 26, 22, 28, 52, 38, 10,
			-20, 5, 7, 36, 20, 19, 2, -12,
			-13, -5, -3, 12, 14, -9, -4, -5,
			-13, 2, -2, -2, -4, 0, 1, 5,
			0, 1, 5, -12, -5, 1, 15, 5,
			-2, 2, -19, -28, -23, -13, -9, -9,
		},
		{
			31, 28, 29, 36, 30, 32, 16, 17,
			15, 20, 21, 21, 16, 18, 23, 13,
			17, 22, 18, 12, 14, 23, 18, 14,
			13, 22, 19, 26, 30, 20, 28, 18,
			5, 14, 27, 28, 26, 25, 14, 2,
			6, 15, 19, 21, 25, 15, 9, 3,
			4, -5, -2, 10, 9, 2, 1, -21,
			-10, 3, 6, 6, 10, 9, -3, -12,
		},
		{
			49, 53, 31, 46, 47, 56, 58, 77,
			14, 4, 34, 50, 43, 64, 32, 56,
			-4, 32, 23, 42, 66, 69, 92, 38,
			-13, -1, 8, 27, 20, 26, 29, 19,
			-30, -26, -21, -10, -13, -22, 0, -14,
			-32, -23, -23, -16, -14, -17, 9, -13,
			-48, -23, -13, -8, -8, -2, 11, -34,
			-20, -14, -7, 1, 1, -3, 7, -20,
		},
		{
			33, 36, 48, 40, 44, 44, 41, 36,
			36, 45, 41, 38, 39, 21, 30, 19,
			33, 23, 33, 22, 14, 17, 1, 18,
			29, 30, 33, 27, 24, 20, 13, 19,
			20, 30, 31, 26, 26, 27, 16, 14,
			6, 16, 15, 13, 13, 9, 0, -4,
			9, 5, 10, 7, 5, -2, -10, 4,
			8, 9, 13, 7, 4, 10, 1, 0,
		},
		{
			-30, 0, 10, 23, 31, 57, 42, 15,
			-20, -45, -8, -30, -25, 31, -29, 27,
			-10, 0, -2, 6, 16, 67, 62, 40,
			-5, -5, -7, -5, 2, 9, 23, 15,
			-2, -1, -3, 0, 1, 7, 9, 17,
			-6, 6, 8, 2, 7, 7, 18, 10,
			-1, 7, 11, 14, 13, 23, 26, 19,
			2, 0, 6, 12, 14, -11, 2, -10,
		},
		{
			59, 51, 68, 78, 91, 85, 64, 86,
			32, 61, 59, 97, 133, 99, 81, 54,
			26, 26, 58, 64, 90, 78, 53, 56,
			14, 49, 49, 72, 88, 95, 83, 61,
			0, 37, 42, 67, 60, 58, 39, 35,
			-12, 6, 25, 22, 24, 30, 1, -13,
			-24, -17, -23, -9, -10, -47, -68, -58,
			-23, -34, -26, -16, -33, -33, -52, -12,
		},
		{
			100, 171, 136, 78, 14, 34, -3, 135,
			-34, 82, 73, 146, 84, 75, 58, -61,
			16, 121, 110, 58, 89, 132, 86, -12,
			21, 55, 54, 25, 34, 50, 30, -47,
			-31, 14, 32, -14, 4, 4, -2, -83,
			-39, -20, -30, -30, -17, -34, -17, -50,
			17, -12, -29, -60, -54, -48, -1, 9,
			8, 39, 7, -70, -28, -67, 17, 22,
		},
		{
			-152, -88, -60, -22, -23, -25, -18, -151,
			-19, 8, 12, -9, 10, 18, 34, -8,
			-2, 18, 23, 34, 27, 32, 35, 7,
			-5, 21, 37, 43, 40, 34, 29, 7,
			-15, 15, 31, 47, 40, 31, 16, 3,
			-10, 9, 23, 33, 30, 23, 3, -7,
			-18, -1, 12, 18, 18, 16, -8, -33,
			-65, -40, -21, -11, -32, -3, -36, -85,
		},
	},
	PieceValues: [2][7]Score{
		{0, 77, 394, 405, 521, 1113, 0},
		{0, 133, 346, 369, 675, 1313, 0},
	},
	TempoBonus: [2]Score{14, 13},
	KingAttackPieces: [2][4]Score{
		{8, 8, 10, 16},
		{-55, 16, 4, -81},
	},
	SafeChecks: [2][4]Score{
		{7, 7, 7, 6},
		{12, -5, 3, 7},
	},
	KingShelter: [2]Score{7, -1},
	MobilityKnight: [2][9]Score{
		{-40, -29, -24, -20, -16, -12, -6, 1, 9},
		{-64, -7, 22, 37, 46, 55, 53, 46, 29},
	},
	MobilityBishop: [2][14]Score{
		{-32, -28, -22, -21, -17, -11, -6, -5, -4, -2, 1, 4, 0, 37},
		{-31, -10, 0, 14, 27, 40, 47, 54, 61, 60, 56, 55, 57, 36},
	},
	MobilityRook: [2][11]Score{
		{-20, -14, -14, -10, -7, -1, 5, 12, 14, 18, 36},
		{-3, 4, 16, 23, 34, 39, 43, 45, 50, 54, 43},
	},
	KnightOutpost: [2][40]Score{
		{
			-12, 16, 68, 22, 82, 87, 151, -37,
			69, 18, -11, -24, 22, -33, -2, -9,
			2, -5, 16, 23, 6, 35, 12, 28,
			11, 23, 29, 34, 44, 60, 54, 17,
			0, 0, 0, 33, 39, 0, 0, 0,
		},
		{
			-62, 68, 10, 16, -8, 55, -25, 34,
			-23, 10, 20, 12, 6, 30, 21, 42,
			34, 40, 22, 28, 38, 37, 46, 37,
			19, 16, 22, 28, 32, 17, 11, 20,
			0, 0, 0, 19, 17, 0, 0, 0,
		},
	},
	ConnectedRooks:  [2]Score{2, 14},
	BishopPair:      [9]Score{7, 48, 56, 51, 46, 41, 35, 31, 24},
	ProtectedPasser: [2]Score{20, 5},
	PasserKingDist:  [2]Score{-5, 12},
	PasserRank: [2][6]Score{
		{-14, -22, -16, 11, 37, 72},
		{18, 20, 43, 67, 118, 77},
	},
	DoubledPawns:  [2]Score{-13, -14},
	IsolatedPawns: [2]Score{-13, -14},
}
