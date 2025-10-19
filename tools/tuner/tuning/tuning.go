package tuning

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"slices"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/eval"
	. "github.com/paulsonkoly/chess-3/types"
)

const (
	ExitFailure = 1
	// NumLinesInBatch determines how the epd file is split into batches. A batch
	// completion implies the coefficients update.
	NumLinesInBatch = 100 //= 100_000

	// NumChunksInBatch determines how a batch is split into chunks. A chunk is a
	// unique work iterm handed over to clients.
	NumChunksInBatch = 16

	Epsilon    = 0.001
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

func (c Coeffs) Eval(b *board.Board) float64 {
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
	// "KingAttackPieces",
	"BishopPair",
}

var ErrBadStruct = errors.New("bad struct")

type ErrInvalidField struct{ kind reflect.Kind }

func (e ErrInvalidField) Error() string { return fmt.Sprintf("invalid kind %s", e.kind) }

type ErrInvalidField2 struct{ a, b reflect.Kind }

func (e ErrInvalidField2) Error() string { return fmt.Sprintf("invalid kind %s %s", e.a, e.b) }

type ErrLengthMismatch struct{ a, b int }

func (e ErrLengthMismatch) Error() string { return fmt.Sprintf("length mismatch %d %d", e.a, e.b) }

// EngineCoeffs is the value set saved in the engine (int16) and
// converted to a float64 CoeffSet.
func EngineCoeffs() (Coeffs, error) {
	result := Coeffs{}
	engineSet := eval.Coefficients

	t := reflect.TypeOf(result)
	dstV := reflect.ValueOf(&result).Elem()
	srcV := reflect.ValueOf(engineSet)

	if dstV.NumField() != srcV.NumField() {
		return result, ErrBadStruct
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
			return ErrLengthMismatch{src.Len(), dst.Len()}
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
		return ErrInvalidField2{src.Kind(), dst.Kind()}
	}

	return nil
}

// Floats extract the target related floats from c.
func (c Coeffs) Floats(target TuningTargets) ([]float64, error) {
	result := make([]float64, 0)

	unWrap := eval.CoeffSet[float64](c)
	structV := reflect.ValueOf(unWrap)
	structT := reflect.TypeOf(unWrap)

	for i := range structT.NumField() {
		if slices.Contains(target, structT.Field(i).Name) {
			floats, err := getFieldFloats(structV.Field(i))

			if err != nil {
				return nil, err
			}

			result = append(result, floats...)
		}
	}

	return result, nil
}

func getFieldFloats(v reflect.Value) ([]float64, error) {
	switch v.Kind() {

	case reflect.Float64:
		return []float64{v.Float()}, nil

	case reflect.Array:
		floats := make([]float64, 0)
		for i := range v.Len() {
			sub, err := getFieldFloats(v.Index(i))
			if err != nil {
				return nil, err
			}
			floats = append(floats, sub...)
		}

		return floats, nil

	default:
		return nil, ErrInvalidField{kind: v.Kind()}
	}
}

// SetFloats sets the target values in c to floats.
func (c *Coeffs) SetFloats(target TuningTargets, floats []float64) error {

	unWrap := (*eval.CoeffSet[float64])(c)
	structV := reflect.ValueOf(unWrap).Elem()
	structT := reflect.TypeOf(unWrap).Elem()

	for i := range structT.NumField() {
		if slices.Contains(target, structT.Field(i).Name) {
			numUsed, err := setFieldFloats(structV.Field(i), floats)
			if err != nil {
				return err
			}

			floats = floats[numUsed:]
		}
	}

	return nil
}

func setFieldFloats(dst reflect.Value, floats []float64) (int, error) {
	switch dst.Kind() {
	case reflect.Array:
		if dst.Len() > len(floats) {
			return 0, ErrLengthMismatch{dst.Len(), len(floats)}
		}

		numUsed := 0
		for i := range dst.Len() {
			rec, err := setFieldFloats(dst.Index(i), floats)
			if err != nil {
				return 0, err
			}

			floats = floats[rec:]
			numUsed += rec
		}
		return numUsed, nil

	case reflect.Float64:
		if len(floats) < 1 {
			return 0, ErrLengthMismatch{1, len(floats)}
		}
		dst.Set(reflect.ValueOf(floats[0]))

		return 1, nil

	default:
		return 0, ErrInvalidField{dst.Kind()}
	}
}
