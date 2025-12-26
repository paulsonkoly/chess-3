package tuning

import (
	"fmt"
	"io"
	"iter"
	"math"
	"path"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/eval"
)

// Vector represents the set of float64 values in a multi dimensional vector.
type Vector struct {
	data []float64
}

func (v *Vector) ModifyElem(i int, mod func(float64) float64) { v.data[i] = mod(v.data[i]) }

// VectorFromSlice constructs a new Vector from data. Array length is not checked.
func VectorFromSlice(data []float64) Vector { return Vector{data: data} }

// NullVector returns a Vector where all elements are set to 0.
func NullVector(targets []string) Vector { return EngineRep{}.ToVector(targets) }

// VectorToSlice returns the elements of a vector. It returns the underlying
// slice, so modifying the slice elements will be reflected in the vector.
func (v Vector) VectorToSlice() []float64 { return v.data }

// Add adds other to v.
func (v *Vector) Add(other Vector) { v.Combine(other, func(v, o float64) float64 { return v + o }) }

// Sub subtracts other from v.
func (v *Vector) Sub(other Vector) { v.Combine(other, func(v, o float64) float64 { return v - o }) }

// DivConst divides v by the constant c element wise.
func (v *Vector) DivConst(c float64) { v.Modify(func(v float64) float64 { return v / c }) }

// Map maps mod element wise on v and returns a new vector with the mapped elements.
func (v Vector) Map(mod func(float64) float64) Vector {
	new := Vector{data: slices.Clone(v.data)}
	new.Modify(mod)
	return new
}

// Modify applies the function mod on v element wise.
func (v *Vector) Modify(mod func(float64) float64) {
	for i := range v.data {
		v.data[i] = mod(v.data[i])
	}
}

// Combine modifies v by zipping it with other and replacing the elements with
// the result of comb(vE, otherE).
func (v *Vector) Combine(other Vector, comb func(float64, float64) float64) {
	for i, e := range v.data {
		v.data[i] = comb(e, other.data[i])
	}
}

// EngineRep is the engine eval structure representation of all known
// coefficients (including the ones that are not tuned).
type EngineRep eval.CoeffSet[float64]

// Eval returns the evaluation function result.
func (e *EngineRep) Eval(b *board.Board) float64 {
	score := eval.Eval(b, (*eval.CoeffSet[float64])(e))

	if b.STM == Black {
		score = -score // convert to side relative
	}
	return score
}

// EngineCoeffs loads and converts the engine stored coeffs. The engine in16
// representation is converted to float64.
func EngineCoeffs() EngineRep {
	result := EngineRep{}
	engineSet := eval.Coefficients

	t := reflect.TypeOf(result)
	dstV := reflect.ValueOf(&result).Elem()
	srcV := reflect.ValueOf(engineSet)

	if dstV.NumField() != srcV.NumField() {
		return result
	}

	for i := range t.NumField() {
		convert(dstV.Field(i), srcV.Field(i))
	}

	return result
}

func convert(dst, src reflect.Value) {
	switch {
	case src.Kind() == reflect.Array && dst.Kind() == reflect.Array:
		if src.Len() != dst.Len() {
			panic(fmt.Sprintf("array length mismatch %d != %d", src.Len(), dst.Len()))
		}

		for i := range src.Len() {
			convert(dst.Index(i), src.Index(i))
		}

	case src.Kind() == reflect.Int16 && dst.Kind() == reflect.Float64:
		dst.Set(reflect.ValueOf(float64(src.Int())))

	default:
		panic(fmt.Sprintf("invalid kind src: %v, dst: %v", src.Kind(), dst.Kind()))
	}
}

// ToVector converts e to a vector extracting the tuned coefficients pointed by targets.
func (e EngineRep) ToVector(targets []string) Vector {
	result := Vector{data: make([]float64, 0)}

	unWrap := eval.CoeffSet[float64](e)
	structV := reflect.ValueOf(unWrap)
	structT := reflect.TypeOf(unWrap)

	for i := range structT.NumField() {
		if slices.Contains(targets, structT.Field(i).Name) {
			floats := getFieldFloats(structV.Field(i))

			result.data = append(result.data, floats...)
		}
	}

	return result
}

func getFieldFloats(v reflect.Value) []float64 {
	switch v.Kind() {

	case reflect.Float64:
		return []float64{v.Float()}

	case reflect.Array:
		floats := make([]float64, 0)
		for i := range v.Len() {
			sub := getFieldFloats(v.Index(i))
			floats = append(floats, sub...)
		}

		return floats

	default:
		panic(fmt.Sprintf("invalid kind %v", v.Kind()))
	}
}

// SetVector injects the values from v into e based on the tuned coefficients
// pointed by targets.
func (e *EngineRep) SetVector(v Vector, targets []string) {
	unWrap := (*eval.CoeffSet[float64])(e)
	structV := reflect.ValueOf(unWrap).Elem()
	structT := reflect.TypeOf(unWrap).Elem()
	floats := v.data

	for i := range structT.NumField() {
		if slices.Contains(targets, structT.Field(i).Name) {
			numUsed := setFieldFloats(structV.Field(i), floats)

			floats = floats[numUsed:]
		}
	}
}

func setFieldFloats(dst reflect.Value, floats []float64) int {
	switch dst.Kind() {
	case reflect.Array:
		if dst.Len() > len(floats) {
			panic(fmt.Sprintf("array length mismatch %d != %d", len(floats), dst.Len()))
		}

		numUsed := 0
		for i := range dst.Len() {
			rec := setFieldFloats(dst.Index(i), floats)

			floats = floats[rec:]
			numUsed += rec
		}
		return numUsed

	case reflect.Float64:
		if len(floats) < 1 {
			panic("array empty")
		}
		dst.SetFloat(floats[0])

		return 1

	default:
		panic(fmt.Sprintf("invalid kind %v", dst.Kind()))
	}
}

// Save saves e in the output stream out, with the header information
// containing fn, epoch and mse.
func (c *EngineRep) Save(out io.Writer, fn string, epoch int, mse float64) error {
	b := strings.Builder{}

	b.WriteString("package eval\n\n")
	b.WriteString("// this file is autogenerated by chess-3/tools/tuner\n")
	b.WriteString(fmt.Sprintf("// epoch: %d, mse: %f %s\n", epoch, mse, path.Base(fn)))
	b.WriteString(fmt.Sprintf("// %v\n\n", time.Now()))

	b.WriteString("import (\n")
	b.WriteString("	. \"github.com/paulsonkoly/chess-3/chess\"\n")
	b.WriteString(")\n\n")

	b.WriteString("var Coefficients = CoeffSet[Score]{\n")
	unWrap := eval.CoeffSet[float64](*c)
	structV := reflect.ValueOf(unWrap)
	structT := reflect.TypeOf(unWrap)

	for i := range structT.NumField() {
		typ := strings.ReplaceAll(structT.Field(i).Type.String(), "float64", "Score")
		b.WriteString(fmt.Sprintf("%s: %s", structT.Field(i).Name, typ))
		writeField(&b, structV.Field(i), 0)
		b.WriteString(",\n")
	}
	b.WriteString("}\n")

	_, err := out.Write([]byte(b.String()))
	return err
}

func writeField(b *strings.Builder, v reflect.Value, indent int) {
	switch v.Kind() {

	case reflect.Array:

		switch {

		case v.Len()%8 == 0: // assume some form of square table (PSqT)
			b.WriteString("{\n")
			newLine := ""
			for i := range v.Len() / 8 {
				fmt.Fprintf(b, "%s%s", newLine, align(indent+1))
				comma := ""
				for j := range 8 {
					e := v.Index(i*8 + j)
					b.WriteString(comma)
					writeField(b, e, indent+1)
					comma = ", "
				}
				newLine = ",\n"
			}
			fmt.Fprintf(b, ",\n%s}", align(indent))

		case v.Index(0).Kind() == reflect.Array: // if the first elem is an array
			b.WriteString("{")
			comma := ""
			fmt.Fprint(b, align(indent))
			for i := range v.Len() {
				fmt.Fprintf(b, "%s\n%s", comma, align(indent+1))
				writeField(b, v.Index(i), indent+1)
				comma = ","
			}
			fmt.Fprintf(b, ",\n%s}", align(indent))

		default: // one line array

			b.WriteString("{ ")
			comma := ""
			for i := range v.Len() {

				fmt.Fprintf(b, "%s", comma)
				writeField(b, v.Index(i), indent+1)
				comma = ", "
			}
			b.WriteString("}")
		}

	case reflect.Float64:
		f := v.Float()
		r := int(math.Round(f))

		fmt.Fprintf(b, "%4d", r)

	default:
		panic("unexpected kind " + reflect.TypeOf(v).Kind().String())
	}
}

func align(indent int) string {
	return strings.Repeat("\t", indent)
}

// TunedParams yields a sequential index and a pointer to each tuned parameter
// in e. Tuned params are selected by targets.
func (e *EngineRep) TunedParams(targets []string) iter.Seq2[int, *float64] {
	unWrap := (*eval.CoeffSet[float64])(e)
	structV := reflect.ValueOf(unWrap).Elem()
	structT := reflect.TypeOf(unWrap).Elem()
	cnt := 0

	return func(yield func(int, *float64) bool) {
		for i := range structT.NumField() {
			if slices.Contains(targets, structT.Field(i).Name) {
				if !yieldFields(yield, &cnt, structV.Field(i)) {
					return
				}
			}
		}
	}
}

func yieldFields(yield func(int, *float64) bool, cnt *int, v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !yieldFields(yield, cnt, v.Index(i)) {
				return false
			}
		}

	case reflect.Float64:
		if !yield(*cnt, v.Addr().Interface().(*float64)) {
			return false
		}
		*cnt++

	default:
		panic("unexpected kind " + v.Kind().String())
	}
	return true
}
