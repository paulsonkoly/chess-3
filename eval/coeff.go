package eval

import (
	. "github.com/paulsonkoly/chess-3/chess"
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
			101, 112, 90, 111, 109, 66, -66, -64,
			8, 29, 50, 45, 60, 112, 74, 23,
			3, 16, 17, 30, 48, 48, 36, 12,
			-2, 1, 10, 19, 25, 23, 20, 7,
			-4, -5, 1, 4, 16, 17, 27, 11,
			-5, -6, -9, -5, -2, 28, 38, 6,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			0, 0, 0, 0, 0, 0, 0, 0,
			116, 107, 100, 51, 49, 69, 125, 129,
			38, 31, 1, -33, -39, -20, 3, 10,
			10, -1, -13, -36, -33, -28, -14, -13,
			-8, -9, -24, -28, -28, -23, -23, -24,
			-14, -16, -20, -22, -17, -18, -28, -30,
			-10, -14, -12, -13, -3, -12, -28, -36,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{
			-185, -133, -93, -25, 25, -108, -111, -142,
			-14, -9, 33, 57, 36, 73, -14, 14,
			-18, 19, 34, 39, 86, 73, 49, 18,
			0, 5, 21, 33, 21, 46, 21, 34,
			-9, 5, 13, 15, 25, 21, 35, 8,
			-25, -8, -1, 13, 19, 8, 7, -6,
			-38, -31, -16, 1, -1, 2, -8, -9,
			-86, -28, -44, -20, -17, -10, -23, -59,
		},
		{
			-36, 11, 21, 7, 4, 28, 9, -65,
			-14, 4, -1, 15, 10, -20, 4, -23,
			-5, 2, 32, 29, 10, 13, -10, -13,
			1, 14, 34, 40, 39, 30, 24, 4,
			10, 19, 45, 40, 45, 44, 22, 18,
			-18, 7, 16, 36, 32, 13, 7, -3,
			-2, 11, 0, 12, 15, -5, 3, 6,
			-14, -27, -2, 9, 9, -3, -14, -16,
		},
		{
			-60, -59, -97, -98, -90, -126, -33, -56,
			-46, -9, -12, -30, -7, -8, -23, -52,
			-9, 6, 26, 22, 28, 52, 37, 10,
			-20, 5, 7, 35, 20, 18, 2, -13,
			-13, -5, -3, 12, 14, -9, -4, -5,
			-13, 2, -3, -3, -4, 0, 1, 5,
			-1, 1, 5, -12, -5, 1, 15, 5,
			-2, 2, -19, -28, -23, -14, -9, -10,
		},
		{
			32, 29, 30, 36, 31, 32, 16, 18,
			16, 21, 21, 20, 15, 18, 23, 14,
			18, 23, 19, 13, 14, 24, 19, 15,
			13, 22, 20, 27, 31, 21, 28, 18,
			5, 14, 28, 29, 27, 25, 14, 2,
			6, 16, 20, 22, 26, 15, 9, 3,
			5, -4, -2, 11, 10, 3, 2, -20,
			-9, 4, 7, 6, 10, 10, -2, -10,
		},
		{
			49, 54, 31, 46, 48, 58, 61, 79,
			13, 2, 32, 48, 41, 62, 30, 55,
			-5, 31, 21, 40, 64, 66, 90, 36,
			-14, -3, 6, 25, 18, 24, 28, 18,
			-32, -27, -22, -11, -14, -23, -1, -15,
			-34, -24, -24, -17, -15, -18, 8, -14,
			-49, -24, -14, -9, -10, -3, 10, -35,
			-21, -15, -8, 0, 0, -4, 5, -20,
		},
		{
			34, 38, 49, 41, 45, 44, 41, 36,
			38, 47, 44, 41, 42, 23, 32, 20,
			35, 25, 35, 25, 16, 19, 3, 19,
			31, 32, 35, 29, 26, 22, 14, 20,
			21, 32, 33, 28, 27, 29, 17, 14,
			6, 17, 16, 14, 14, 10, 0, -4,
			10, 6, 11, 8, 6, -1, -9, 4,
			9, 10, 14, 8, 6, 11, 2, 0,
		},
		{
			-30, 1, 9, 20, 29, 56, 39, 15,
			-19, -44, -8, -30, -25, 30, -29, 27,
			-9, 1, -1, 6, 16, 66, 62, 40,
			-4, -4, -6, -5, 2, 10, 23, 16,
			-1, 0, -2, 1, 2, 8, 10, 18,
			-5, 7, 9, 3, 8, 8, 19, 11,
			0, 9, 12, 15, 14, 24, 28, 20,
			3, 1, 7, 13, 14, -10, 3, -9,
		},
		{
			64, 55, 74, 86, 99, 91, 72, 90,
			35, 63, 62, 102, 139, 104, 86, 58,
			28, 29, 62, 68, 95, 84, 56, 59,
			16, 52, 51, 76, 92, 98, 86, 64,
			2, 39, 44, 71, 63, 60, 41, 37,
			-11, 8, 27, 24, 26, 32, 3, -11,
			-23, -16, -22, -7, -8, -46, -67, -57,
			-21, -33, -24, -15, -31, -31, -51, -9,
		},
		{
			117, 191, 156, 97, 31, 56, 24, 151,
			-26, 93, 87, 157, 90, 90, 70, -57,
			27, 130, 119, 66, 99, 138, 91, -11,
			25, 60, 56, 25, 35, 49, 31, -46,
			-27, 16, 33, -15, 3, 2, -2, -85,
			-37, -20, -31, -31, -18, -35, -17, -50,
			18, -11, -29, -60, -54, -48, -1, 10,
			8, 39, 8, -70, -27, -66, 18, 23,
		},
		{
			-156, -92, -64, -24, -24, -28, -23, -153,
			-19, 11, 12, -8, 13, 19, 35, -9,
			-3, 19, 25, 35, 29, 34, 37, 7,
			-6, 22, 39, 46, 43, 36, 30, 6,
			-17, 14, 31, 48, 41, 32, 16, 2,
			-14, 7, 23, 33, 30, 22, 1, -10,
			-22, -4, 10, 18, 17, 14, -11, -36,
			-68, -43, -23, -13, -34, -6, -39, -89,
		},
	},
	PieceValues: [2][7]Score{
		{0, 76, 392, 402, 514, 1115, 0},
		{0, 139, 355, 376, 689, 1332, 0},
	},
	TempoBonus: [2]Score{14, 14},
	KingAttackPieces: [2][4]Score{
		{8, 8, 10, 16},
		{-54, 16, 3, -81},
	},
	SafeChecks: [2][4]Score{
		{7, 7, 7, 6},
		{12, -3, 4, 7},
	},
	KingShelter: [2]Score{7, -1},
	MobilityKnight: [2][9]Score{
		{-42, -31, -25, -22, -18, -14, -8, 0, 8},
		{-64, -6, 22, 38, 47, 55, 54, 46, 30},
	},
	MobilityBishop: [2][14]Score{
		{-33, -29, -24, -22, -18, -12, -7, -6, -6, -4, -1, 2, -1, 35},
		{-30, -8, 2, 16, 29, 43, 49, 57, 63, 62, 59, 58, 59, 39},
	},
	MobilityRook: [2][11]Score{
		{-20, -14, -13, -10, -7, -1, 5, 12, 14, 17, 35},
		{-1, 6, 18, 25, 36, 41, 45, 47, 53, 57, 46},
	},
	KnightOutpost: [2][40]Score{
		{
			-2, 12, 87, 28, 85, 86, 167, -43,
			63, 21, -10, -17, 26, -31, -1, -8,
			2, -5, 17, 23, 6, 36, 13, 29,
			11, 23, 28, 34, 44, 60, 55, 18,
			0, 0, 0, 33, 39, 0, 0, 0,
		},
		{
			-45, 52, 10, 13, -6, 52, -22, 28,
			-21, 8, 19, 8, 4, 28, 20, 32,
			34, 40, 22, 29, 40, 38, 46, 36,
			20, 17, 23, 29, 33, 17, 11, 19,
			0, 0, 0, 20, 18, 0, 0, 0,
		},
	},
	ConnectedRooks:  [2]Score{2, 14},
	BishopPair:      [9]Score{9, 52, 58, 52, 46, 41, 35, 31, 24},
	ProtectedPasser: [2]Score{20, 5},
	PasserKingDist:  [2]Score{-2, 11},
	PasserRank: [2][6]Score{
		{-13, -21, -16, 11, 36, 68},
		{17, 19, 43, 67, 120, 77},
	},
	DoubledPawns:  [2]Score{-13, -16},
	IsolatedPawns: [2]Score{-13, -14},
}
