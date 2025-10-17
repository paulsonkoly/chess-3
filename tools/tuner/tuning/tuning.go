package tuning

import (
	"errors"
	"math"
	"reflect"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/eval"
	. "github.com/paulsonkoly/chess-3/types"
)

const (
	ExitFailure = 1
	// NumLinesInBatch determines how the epd file is split into batches. A batch
	// completion implies the coefficients update.
	NumLinesInBatch = 100_000

	// NumChunksInBatch determines how a batch is split into chunks. A chunk is a
	// unique work iterm handed over to clients.
	NumChunksInBatch = 16
)

type Coeffs eval.CoeffSet[float64]

func (c *Coeffs) Add(other Coeffs) {
	// for i := range *c {
	// 	(*c)[i] += other[i]
	// }
}

func Sigmoid(v, k float64) float64 {
	return 1 / (1 + math.Exp(-k*v/400))
}

func (c Coeffs)Eval(b *board.Board) float64 {
	score := eval.Eval(b, (*eval.CoeffSet[float64])(&c))

	if b.STM == Black {
		score = -score
	}
	return score
}

// TuningTargets is the list of structure member names in eval.Coeffs subject to tuning.
type TuningTargets []string

var DefaultTargets = []string{
	// "PSqT",
	// "PieceValues",
	// "TempoBonus",
	// "MobilityKnight", "MobilityBishop", "MobilityRook",
	// "KingAttackPieces", "SafeChecks", "KingShelter",
	// "ProtectedPasser", "PasserKingDist", "PasserRank", "DoubledPawns", "IsolatedPawns",
	// "KnightOutpost", "ConnectedRooks", "BishopPair",
	"BishopPair",
}

// EngineCoeffs is the value set saved in the engine (int16) and
// converted to a float64 CoeffSet.
func EngineCoeffs() (Coeffs, error) {
	result := Coeffs{}
	engineSet := eval.Coefficients

	t := reflect.TypeOf(result)
	dstV := reflect.ValueOf(&result).Elem()
	srcV := reflect.ValueOf(engineSet)

	if dstV.NumField() != srcV.NumField() {
		return result, errors.New("field count mismatch")
	}

	for i := range t.NumField() {
		convert(dstV.Field(i), srcV.Field(i))
	}

	return result, nil
}

func convert(dst, src reflect.Value) error {
	switch {
	case src.Kind() == reflect.Array && dst.Kind() == reflect.Array:
		if src.Len() != dst.Len() {
			return errors.New("mismatched array length")
		}

		for i := range src.Len() {
			err := convert(dst.Index(i), src.Index(i))
			if err != nil {
				return err
			}
		}

	case src.Kind() == reflect.Int16 && dst.Kind() == reflect.Float64:
		dst.Set(reflect.ValueOf(float64(src.Int())))

	default:
		return errors.New("invlid structure field kind")
	}

	return nil
}

// func getFieldFloats(v reflect.Value) ([]float64, error) {
// 	switch v.Kind() {
//
// 	case reflect.Int16:
// 		return []float64{float64(v.Int())}, nil
//
// 	case reflect.Array:
// 		floats := make([]float64, 0)
// 		for i := range v.Len() {
// 			sub, err := getFieldFloats(v.Index(i))
// 			if err != nil {
// 				return nil, err
// 			}
// 			floats = append(floats, sub...)
// 		}
//
// 		return floats, nil
//
// 	default:
// 		return nil, ErrInvalidField{kind: v.Kind()}
// 	}
// }
//
// // ToCoeffSet converts c to the engine representation.
// func (c Coeffs) ToCoeffSet(targets TuningTargets) eval.CoeffSet[float64] {
// 	result := eval.Co
//
// }
