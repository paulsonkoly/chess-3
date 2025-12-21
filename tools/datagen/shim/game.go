package shim

import (
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/types"
)

type Game struct {
	WDL       WDL
	Positions []Position
}

type WDL byte

const (
	Draw = WDL(iota)
	WhiteWins
	BlackWins
)

type Position struct {
	FEN   string
	BM    move.Move
	Score types.Score
}
