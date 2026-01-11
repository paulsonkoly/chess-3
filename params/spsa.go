//go:build spsa

package params

// This file is only included in spsa builds. It must contain
// exported variables for all the tunable engine parameters. In an
// normal build look into params/params.go instead. The symmetry
// between the two files must be maintained.

import (
	"errors"
	"fmt"
	"strings"

	. "github.com/paulsonkoly/chess-3/chess"
)

var (
	NMPDiffFactor    = 50
	NMPDepthLimit    = 1
	NMPInit          = 4
	RFPDepthLimit    = 8
	RFPScoreFactor   = 105
	WindowSize       = 46
	LMRStart         = 2
	StandPatDelta    = 111
	HistBonusMul     = 20
	HistBonusLin     = 16
	HistAdjRange     = 8
	HistAdjReduction = 7
)

var tunables = [...]struct {
	ptr  *int
	name string
	min  int
	max  int
}{
	{&NMPDiffFactor, "NMPDiffFactor", 30, 70},
	{&NMPDepthLimit, "NMPDepthLimit", 0, 5},
	{&NMPInit, "NMPInit", 1, 6},
	{&RFPDepthLimit, "RFPDepthLimit", 5, 10},
	{&RFPScoreFactor, "ScoreFactor", 70, 130},
	{&WindowSize, "WindowSize", 30, 100},
	{&LMRStart, "LMRStart", 0, 10},
	{&StandPatDelta, "StandPatDelta", 80, 130},
	{&HistBonusMul, "HistBonusMul", 15, 25},
	{&HistBonusLin, "HistBonusLin", 0, 20},
	{&HistAdjRange, "HistAdjRange", 4, 10},
	{&HistAdjReduction, "HistAdjReduction", 4, 10},
}

func UCIOptions() string {
	b := strings.Builder{}

	for _, t := range tunables {
		fmt.Fprintf(&b, "option name %s type spin default %d min %d max %d\n", t.name, *t.ptr, t.min, t.max)
	}

	return b.String()
}

func OpenbenchInfo() string {
	b := strings.Builder{}

	for _, t := range tunables {
		lrStep := float64(Abs(t.max-t.min)) / 20

		if lrStep < 1.0 {
			lrStep = 1.0
		}

		fmt.Fprintf(&b, "%s, int, %d.0, %d.0, %d.0, %.3f, 0.002\n",
			t.name, *t.ptr, t.min, t.max, lrStep)
	}

	return b.String()
}

func Set(name string, val int) error {
	for _, t := range tunables {
		if t.name == name {
			if val < t.min || t.max < val {
				return errors.New("out of bounds param")
			}
			*t.ptr = val
			return nil
		}
	}
	return errors.New("no such parameter")
}
