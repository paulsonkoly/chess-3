package tuning

import (
	"math"
)

const (
	// NumLinesInBatch determines how the epd file is split into batches. A batch
	// completion implies the coefficients update.
	NumLinesInBatch = 100_000

	// NumChunksInBatch determines how a batch is split into chunks. A chunk is a
	// unique work item handed over to clients.
	NumChunksInBatch = 16

	// Perturbation
	Epsilon = 0.001

	Beta1               = 0.9
	Beta2               = 0.999
	InitialLearningRate = 1.0
)

func Sigmoid(v, k float64) float64 {
	return 1 / (1 + math.Exp(-k*v/400))
}

var DefaultTargets = []string{
	"PSqT",
	"PieceValues",
	"TempoBonus",
	"MobilityKnight", "MobilityBishop", "MobilityRook",
	"KingAttackPieces", "SafeChecks", "KingShelter",
	"OpBishopsOutsidePassers", "OpBishopsPawnDelta",
	"ProtectedPasser", "PasserKingDist", "PasserRank", "DoubledPawns", "IsolatedPawns",
	"KnightOutpost", "ConnectedRooks", "BishopPair",
}
