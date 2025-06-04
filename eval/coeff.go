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
			63, 90, 64, 98, 83, 42, -39, -68,
			27, 32, 66, 65, 73, 109, 88, 41,
			0, 11, 21, 26, 48, 43, 35, 18,
			-8, -2, 10, 28, 29, 24, 14, 5,
			-9, -4, 5, 7, 22, 15, 30, 8,
			-8, -5, -4, -4, 9, 28, 38, 1,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			0, 0, 0, 0, 0, 0, 0, 0,
			123, 109, 108, 46, 50, 76, 121, 136,
			28, 29, -14, -55, -59, -35, 4, 9,
			12, 1, -14, -33, -35, -26, -12, -11,
			-6, -7, -19, -27, -28, -23, -18, -22,
			-11, -11, -18, -16, -18, -19, -25, -25,
			-10, -11, -14, -11, -7, -19, -28, -27,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			-163, -116, -70, -39, -2, -80, -136, -123,
			-26, -5, 33, 49, 8, 70, -14, 8,
			-9, 29, 36, 35, 86, 61, 44, 11,
			2, 2, 15, 42, 24, 45, 13, 32,
			-10, -2, 12, 15, 25, 24, 25, 5,
			-33, -13, -3, 7, 25, 6, 11, -11,
			-39, -25, -18, 4, 3, 3, -7, -9,
			-72, -36, -34, -16, -15, -7, -30, -39,
		},
		{
			-46, -4, 6, -1, -5, -22, -3, -72,
			3, 9, -1, -3, -8, -19, 4, -22,
			4, 4, 21, 20, -3, -10, -15, -14,
			15, 20, 31, 25, 22, 20, 15, 0,
			24, 21, 39, 31, 36, 26, 14, 17,
			7, 12, 18, 34, 28, 11, 4, 10,
			9, 11, 13, 11, 11, 7, 5, 16,
			10, -5, 12, 15, 16, 2, -4, 5,
		},
		{
			-41, -61, -43, -92, -83, -94, -53, -66,
			-31, -3, -12, -30, 2, -20, -18, -54,
			-9, 9, 11, 26, 7, 52, 25, 16,
			-20, -8, 9, 26, 22, 10, -7, -23,
			-17, -14, -15, 16, 12, -11, -13, -3,
			-12, -3, -2, -8, -5, 0, 4, 4,
			-4, 0, 4, -14, -2, 10, 18, 4,
			-19, 4, -15, -19, -12, -17, 7, -4,
		},
		{
			16, 19, 8, 21, 16, 12, 10, 6,
			3, 6, 9, 10, 1, 8, 12, 3,
			17, 11, 12, 3, 9, 10, 6, 9,
			12, 18, 14, 27, 17, 18, 14, 14,
			7, 16, 24, 21, 20, 19, 12, -9,
			6, 18, 19, 21, 24, 17, 8, -2,
			12, 1, -2, 13, 8, 3, 7, -8,
			-2, 11, 10, 9, 11, 15, -8, -14,
		},
		{
			26, 16, 24, 22, 33, 30, 22, 57,
			6, 0, 26, 43, 27, 44, 19, 50,
			-11, 17, 12, 14, 43, 40, 73, 45,
			-16, -7, -7, 5, 7, 7, 14, 15,
			-34, -32, -19, -7, -8, -27, -1, -15,
			-31, -27, -17, -15, -6, -10, 19, -2,
			-35, -23, -7, -3, 0, 1, 19, -23,
			-20, -14, -5, 4, 8, 0, 7, -21,
		},
		{
			24, 32, 36, 30, 26, 28, 26, 18,
			24, 38, 37, 27, 29, 22, 22, 7,
			24, 23, 23, 20, 11, 6, 0, -3,
			28, 25, 31, 23, 11, 9, 7, 5,
			24, 24, 23, 19, 16, 18, 2, 5,
			16, 14, 13, 14, 7, 2, -20, -14,
			15, 14, 13, 11, 5, -1, -12, 4,
			12, 12, 15, 7, 1, 4, -2, 0,
		},
		{
			-49, -28, -9, 16, 16, 26, 16, -27,
			-11, -32, -16, -20, -29, 13, -9, 28,
			0, -2, -3, 11, 6, 57, 47, 59,
			-16, -6, -5, -8, 0, 11, 13, 8,
			-9, -15, -9, 1, 4, 1, 5, 8,
			-13, -3, 0, -2, 4, 5, 17, 6,
			-6, -1, 7, 15, 12, 20, 26, 36,
			-11, -13, -2, 5, 5, -13, 9, -13,
		},
		{
			46, 36, 54, 49, 51, 42, 8, 55,
			3, 32, 60, 70, 102, 62, 23, 25,
			11, 29, 56, 51, 68, 36, -1, -17,
			30, 37, 45, 59, 63, 44, 41, 31,
			7, 38, 41, 56, 50, 41, 24, 15,
			-2, 8, 30, 30, 35, 30, -3, -10,
			-19, -9, -9, -5, -1, -24, -62, -94,
			-10, -10, -5, -9, -14, -23, -45, -24,
		},
		{
			94, 140, 100, 45, -18, -16, -60, 131,
			-49, 38, 15, 120, 60, 23, 18, -56,
			-28, 65, 29, 10, 46, 102, 26, -6,
			-3, -12, -27, -36, -36, -34, -61, -109,
			-58, -48, -44, -70, -78, -68, -106, -137,
			-46, -36, -62, -62, -49, -74, -50, -72,
			38, -9, -18, -43, -49, -35, 4, 15,
			23, 54, 25, -60, -14, -42, 30, 30,
		},
		{
			-110, -70, -43, -8, -5, -5, -10, -124,
			-9, 16, 26, 6, 28, 43, 40, 10,
			-2, 21, 35, 48, 52, 44, 48, 11,
			-11, 24, 44, 53, 56, 53, 45, 19,
			-15, 15, 35, 52, 51, 39, 30, 13,
			-20, 5, 24, 34, 31, 25, 8, -7,
			-40, -9, 5, 13, 17, 7, -14, -38,
			-72, -58, -35, -7, -27, -14, -50, -83,
		},
	},
	PieceValues: [2][7]Score{
		{0, 79, 392, 404, 518, 1059, 0},
		{0, 127, 328, 347, 650, 1288, 0},
	},
	TempoBonus: [2]Score{25, 23},
	KingAttackPieces: [2][4]Score{
		{7, 8, 8, 15},
		{-49, 12, 8, -81},
	},
	SafeChecks: [2][4]Score{
		{9, 9, 9, 5},
		{19, 17, -10, 8},
	},
	KingShelter: [2]Score{7, -6},
	MobilityKnight: [2][9]Score{
		{-57, -36, -25, -18, -12, -6, 2, 8, 10},
		{-40, -1, 20, 30, 38, 47, 45, 41, 30},
	},
	MobilityBishop: [2][14]Score{
		{-46, -35, -26, -22, -15, -8, -1, 2, 3, 5, 7, 9, 5, 19},
		{-19, -3, 1, 12, 25, 37, 39, 46, 52, 49, 47, 45, 51, 37},
	},
	MobilityRook: [2][11]Score{
		{-26, -19, -16, -12, -7, -1, 7, 14, 14, 17, 23},
		{-6, 2, 11, 17, 24, 28, 29, 32, 40, 45, 39},
	},
	KnightOutpost: [2][40]Score{
		{
			-29, 52, 30, 20, 56, 99, 120, -30,
			63, -8, -24, -55, 3, -39, -4, 0,
			0, -4, 9, 33, 14, 48, 17, 12,
			-1, 25, 37, 35, 48, 50, 68, 17,
			0, 0, 0, 31, 48, 0, 0, 0,
		},
		{
			-92, 139, -14, 26, -40, 72, -47, 43,
			-54, 24, 30, 35, 10, 37, 17, 81,
			25, 13, 28, 26, 36, 34, 29, 36,
			21, 13, 18, 27, 29, 15, 0, 23,
			0, 0, 0, 21, 18, 0, 0, 0,
		},
	},
	ConnectedRooks:  [2]Score{1, 7},
	BishopPair:      [9]Score{-2, 43, 60, 55, 46, 38, 34, 36, 38},
	ProtectedPasser: [2]Score{22, 5},
	PasserKingDist:  [2]Score{8, 8},
	PasserRank: [2][6]Score{
		{-13, -25, -20, 4, 5, 39},
		{11, 14, 41, 68, 139, 95},
	},
	DoubledPawns:  [2]Score{-9, -16},
	IsolatedPawns: [2]Score{-16, -11},
}
