//go:build !spsa

// Package params provides tunable engine parameter functions.
package params

// This file is not included in spsa builds. It must contain
// exported constants for all the tunable engine parameters. In an
// spsa build look into params/spsa.go instead. The symmetry
// between the two files must be maintained.

const (
	NMPDiffFactor    = 51
	NMPDepthLimit    = 1
	NMPInit          = 4
	RFPDepthLimit    = 8
	WindowSize       = 50
	LMRStart         = 2
	StandPatDelta    = 110
	HistBonusMul     = 20
	HistBonusLin     = 15
	HistAdjRange     = 8
	HistAdjReduction = 7
)

func UCIOptions() string { return "" }

func OpenbenchInfo() string { return "" }
