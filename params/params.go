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
	RFPScoreFactor   = 105
	WindowSize       = 50
	LMRStart         = 2
	StandPatDelta    = 110
	HistBonusMul     = 20
	HistBonusLin     = 15
	HistAdjRange     = 8
	HistAdjReduction = 7
)

// UCIOptions returns the uci options string for tunable parameters in an spsa
// build. For non-spsa builds it returns empty.
func UCIOptions() string { return "" }

// OpenbenchInfo returns the openbench spsa input in an spsa build. For
// non-spsa builds it returns empty.
func OpenbenchInfo() string { return "" }

// Set sets the named parameter to value val in spsa build. In non-spsa
// build it does nothing.
func Set(name string, val int) error { return nil }
