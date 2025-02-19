package tuning

import (
	"fmt"
	"iter"
	"math"
	"reflect"
	"slices"
	"strings"

	"github.com/paulsonkoly/chess-3/eval"
)

// Targets controls which fields in eval.CoeffSet are going to be tuned.
var Targets = [...]string{
	"PSqT",
	"PieceValues",
	"TempoBonus",
	"MobilityKnight", "MobilityBishop", "MobilityRook",
	"KingAttackPieces", "KingAttackCount",
	"ProtectedPasser", "PasserKingDist", "PasserRank",
	 "KnightOutpost", "ConnectedRooks", "BishopPair",
}

// Anchor is a known value in tuning, anchoring down the tuning process. We
// skip tuning the value with this index, for example a good setting would be
// setting a Pawn's value to 100 so we measure evalution in centipawns.
var Anchor = Index {s : []int{1, 0, 1} }

type Coeffs eval.CoeffSet[float64]

// InitialCoeffs creates the initial coefficents by converting eval.Coefficients from Score to Tunable.
func InitialCoeffs() *Coeffs {
	result := Coeffs{}

	t := reflect.TypeOf(result)
	dst := reflect.ValueOf(&result).Elem()
	src := reflect.ValueOf(&eval.Coefficients).Elem()

	for ix := 0; ix < t.NumField(); ix++ {
		convert(dst.Field(ix), src.Field(ix))
	}

	return &result
}

func (t *Coeffs) ToEvalType() *eval.CoeffSet[float64] {
	return (*eval.CoeffSet[float64])(t)
}

func convert(dst, src reflect.Value) {
	switch dst.Kind() {

	case reflect.Array:
		if src.Kind() != reflect.Array {
			panic("array expected")
		}
		if src.Len() != dst.Len() {
			panic("len mismatch")
		}
		for i := 0; i < dst.Len(); i++ {
			convert(dst.Index(i), src.Index(i))
		}

	case reflect.Float64:
		if src.Kind() != reflect.Int16 {
			panic("int16 expected")
		}
		dst.Set(reflect.ValueOf(float64(src.Int())))

	default:
		panic("unexpected kind " + dst.Kind().String())
	}
}

func (t *Coeffs) Print() {
	typ := reflect.TypeOf(*t)
	for ix := 0; ix < typ.NumField(); ix++ {
		f := typ.Field(ix)
    if slices.Contains(Targets[:], f.Name) {
			v := reflect.ValueOf(*t).Field(ix)

			fmt.Printf("%s: ", f.Name)

			printField(v, 0)
			fmt.Printf(",\n")
		}
	}
}

func indent(i int) string {
	return strings.Repeat("\t", i)
}

func printField(v reflect.Value, in int) {
	switch v.Kind() {

	case reflect.Array:

		switch {

		case v.Len()%8 == 0: // assume some form of square table (PSqT)
			fmt.Print("{\n")
			newLine := ""
			for i := range v.Len() / 8 {
				fmt.Printf("%s%s", newLine, indent(in+1))
				comma := ""
				for j := range 8 {
					e := v.Index(i*8 + j)
					fmt.Print(comma)
					printField(e, in+1)
					comma = ", "
				}
				newLine = ",\n"
			}
			fmt.Printf(",\n%s}", indent(in))

		case v.Index(0).Kind() == reflect.Array: // if the first elem is an array
			fmt.Printf("{")
			comma := ""
			fmt.Print(indent(in))
			for i := 0; i < v.Len(); i++ {
				fmt.Printf("%s\n%s", comma, indent(in+1))
				printField(v.Index(i), in+1)
				comma = ","
			}
			fmt.Printf(",\n%s}", indent(in))

		default: // one line array

			fmt.Printf("{ ")
			comma := ""
			for i := 0; i < v.Len(); i++ {

				fmt.Printf("%s", comma)
				printField(v.Index(i), in+1)
				comma = ", "
			}
			fmt.Printf("}")
		}

	case reflect.Float64:
		f := v.Float()
		r := int(math.Round(f))

		fmt.Printf("%4d", r)

	default:
		panic("unexpected kind " + reflect.TypeOf(v).Kind().String())
	}
}

type Index struct {
	s []int
}

func (t *Coeffs) Loop() iter.Seq[Index] {
	return func(yield func(Index) bool) {
		typ := reflect.TypeOf(*t.ToEvalType())
		v := reflect.ValueOf(*t.ToEvalType())

		for ix := 0; ix < typ.NumField(); ix++ {
			if !slices.Contains(Targets[:], typ.Field(ix).Name) {
				continue
			}

			e := v.Field(ix)
			if e.Kind() != reflect.Array {
				panic("expected array " + e.Kind().String())
			}

			if !recurse(yield, Index{[]int{ix}}, e) {
				return
			}
		}
	}
}

func recurse(yield func(Index) bool, ix Index, v reflect.Value) bool {

	switch v.Kind() {

	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !recurse(yield, Index{append(ix.s, i)}, v.Index(i)) {
				return false
			}
		}

	case reflect.Float64:
    if slices.Equal(ix.s[:], Anchor.s[:]) {
      return true
    }
		return yield(ix)

	}

	return true
}
func (t *Coeffs) At(ix Index) *float64 {
	v := reflect.ValueOf(t)
	ixs := ix.s

	return recurseAt(ixs[1:], v.Elem().Field(ixs[0]))
}

func recurseAt(ixs []int, v reflect.Value) *float64 {
	switch v.Kind() {

	case reflect.Array:
		return recurseAt(ixs[1:], v.Index(ixs[0]))

	case reflect.Float64:
		return v.Addr().Interface().(*float64)
	}
	panic("unreachable")
}
