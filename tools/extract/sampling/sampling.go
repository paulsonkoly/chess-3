package sampling

import (
	"fmt"
	"math"
)

// Discretizer maps some feature of a data point to a discrete numeric scale
// (bin index).
type Discretizer interface {
	Dim() int        // Dim is the dimension of the feature.
	Value(d any) int // Value is the value of the feature, 0 <= Value(d) < Dim()
}

// Feature is a single feature discretizer.
type Feature struct {
	bc    int
	value func(d any) int
}

func NewFeature(bc int, value func(d any) int) Feature {
	return Feature{bc, value}
}

// Dim is the dimension of the feature.
func (f Feature) Dim() int { return f.bc }

// Value maps a data point to a discrete scale of f.
func (f Feature) Value(d any) int {
	ret := f.value(d)
	if ret < 0 || ret >= f.Dim() {
		panic(fmt.Sprintf("ret %d out of range of [0, %d)", ret, f.Dim()))
	}
	return ret
}

// Combined is a combined discretizer of many features.
type Combined struct {
	features []Discretizer
}

// NewCombined is a new combined discretizer.
func NewCombined(features ...Discretizer) Combined {
	return Combined{features}
}

// Dim is the combined dimension of sub-features, the product of sub-feature
// dimensions.
func (c Combined) Dim() int {
	prod := 1
	for _, feature := range c.features {
		prod *= feature.Dim()
	}
	return prod
}

// Value is the combined value of all sub-features.
func (c Combined) Value(d any) int {
	ix := 0
	stride := 1

	for i := len(c.features) - 1; i >= 0; i-- {
		f := c.features[i]
		ix += f.Value(d) * stride
		stride *= f.Dim()
	}

	return ix
}

// Scale is a resize of some Discretizer to a new scale.
type Scale struct {
	orig Discretizer
	size int
}

// NewScale creates a resized version of d to 0..size.
func NewScale(d Discretizer, size int) Scale {
	return Scale{d, size}
}

// Dim is the resized dimension.
func (s Scale) Dim() int { return s.size }

// Value is the re-scaled discrete value.
func (s Scale) Value(d any) int {
	origV := s.orig.Value(d)
	origS := s.orig.Dim()

	ret := int(math.Round(float64(origV) * float64(s.size) / float64(origS)))
	if ret >= s.size {
		ret = s.size - 1
	}
	return ret
}

// Counter counts the occurances of discrete values.
type Counter struct {
	counts []int
	total  int
}

// NewCounter creates a counter of stream of discrete values in the range of
// 0..dim.
func NewCounter(dim int) Counter {
	return Counter{counts: make([]int, dim)}
}

// Add adds one data point.
func (c *Counter) Add(v int) {
	c.counts[v]++
	c.total++
}

// Count is the total number of occurances of v.
func (c Counter) Count(v int) int { return c.counts[v] }

// Total is the number of all data points.
func (c Counter) Total() int { return c.total }

// Dim is the dimensionality of the counter.
func (c Counter) Dim() int { return len(c.counts) }

// Sampler is a uniform sampler of some Counter.
type Sampler struct {
	keepProbs []float64
}

// NewSampler creates a uniform sampler based on data collected in c.
func NewUniformSampler(c Counter) Sampler {
	k := math.MaxFloat64
	for ix := range c.Dim() {
		bucket := c.Count(ix)
		if bucket == 0 {
			continue
		}
		rat := (float64(bucket) / float64(c.Total())) / (1.0 / float64(c.Dim()))
		if rat < k {
			k = rat
		}
	}

	keepProbs := make([]float64, c.Dim())
	for ix := range c.Dim() {
		bucket := c.Count(ix)
		if bucket == 0 {
			continue
		}
		dist := float64(bucket) / float64(c.Total())
		keepProbs[ix] = k * (1.0 / float64(c.Dim())) / dist
	}
	return Sampler{keepProbs}
}

func NewSqrtSampler(c Counter) Sampler {
	keepProbs := make([]float64, c.Dim())

	maxW := 0.0
	for ix := range c.Dim() {
		n := c.Count(ix)
		if n == 0 {
			continue
		}
		w := 1.0 / math.Sqrt(float64(n))
		keepProbs[ix] = w
		if w > maxW {
			maxW = w
		}
	}

	for ix := range c.Dim() {
		keepProbs[ix] /= maxW
	}

	return Sampler{keepProbs}
}

// KeepProb is the probabilty of keeping a data point of value v.
func (s Sampler) KeepProb(v int) float64 {
	return s.keepProbs[v]
}
