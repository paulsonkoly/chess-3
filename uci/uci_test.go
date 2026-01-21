package uci_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/search"
	"github.com/paulsonkoly/chess-3/transp"
	"github.com/paulsonkoly/chess-3/uci"
	"github.com/stretchr/testify/assert"
)

type MockSearch struct {
	Cleared bool
	TTSize  int
	Options search.Options
	move    move.Move
	score   Score
}

func (ms *MockSearch) Clear() {
	ms.Cleared = true
}

func (ms *MockSearch) ResizeTT(size int) {
	ms.TTSize = size
}

func (ms *MockSearch) Go(_ *board.Board, opts ...search.Option) (Score, move.Move) {
	for _, opt := range opts {
		opt(&ms.Options)
	}

	return ms.score, ms.move
}

func (ms *MockSearch) MockScore(score Score) {
	ms.score = score
}

func (ms *MockSearch) MockMove(move move.Move) {
	ms.move = move
}

func TestUCI(t *testing.T) {
	inputs := `uci

quit
`

	outputs := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	d := uci.NewDriver(uci.WithInput(strings.NewReader(inputs)),
		uci.WithOutput(outputs),
		uci.WithError(errors),
		uci.WithSearch(&MockSearch{}))

	d.Run()

	assert.Empty(t, errors)
	assert.Contains(t, outputs.String(), "id name chess-3")
	assert.Contains(t, outputs.String(), "id author Paul Sonkoly")
}

func TestIsReady(t *testing.T) {
	inputs := `uci

isready
quit
`

	outputs := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	d := uci.NewDriver(uci.WithInput(strings.NewReader(inputs)),
		uci.WithOutput(outputs),
		uci.WithError(errors),
		uci.WithSearch(&MockSearch{}))

	d.Run()

	assert.Empty(t, errors)
	assert.Contains(t, outputs.String(), "readyok")
}

func TestUCINewGame(t *testing.T) {
	inputs := `uci

ucinewgame
quit
`

	outputs := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	search := &MockSearch{}

	d := uci.NewDriver(uci.WithInput(strings.NewReader(inputs)),
		uci.WithOutput(outputs),
		uci.WithError(errors),
		uci.WithSearch(search))

	d.Run()

	assert.Empty(t, errors)
	assert.True(t, search.Cleared)
}

func TestHashSettings(t *testing.T) {
	inputs := `uci

setoption name Hash value 16
quit
`

	outputs := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	search := &MockSearch{}

	d := uci.NewDriver(uci.WithInput(strings.NewReader(inputs)),
		uci.WithOutput(outputs),
		uci.WithError(errors),
		uci.WithSearch(search))

	d.Run()

	assert.Empty(t, errors)
	assert.Equal(t, 16 * transp.MegaBytes, search.TTSize)
}

func TestInitialFen(t *testing.T) {
	inputs := `uci

fen
quit
`

	outputs := &bytes.Buffer{}
	errors := &bytes.Buffer{}

	d := uci.NewDriver(uci.WithInput(strings.NewReader(inputs)),
		uci.WithOutput(outputs),
		uci.WithError(errors),
		uci.WithSearch(&MockSearch{}))

	d.Run()

	assert.Empty(t, errors)
	assert.Contains(t, outputs.String(), "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
}

func TestSettingFen(t *testing.T) {
	inputs := `uci

position fen 4k3/8/2K5/8/8/8/8/8 w - - 0 1 moves c6c5
fen
quit
`

	outputs := &bytes.Buffer{}
	errors := &bytes.Buffer{}

	d := uci.NewDriver(uci.WithInput(strings.NewReader(inputs)),
		uci.WithOutput(outputs),
		uci.WithError(errors),
		uci.WithSearch(&MockSearch{}))

	d.Run()

	assert.Empty(t, errors)
	assert.Contains(t, outputs.String(), "4k3/8/8/2K5/8/8/8/8 b - - 1 1")
}

func TestStartPos(t *testing.T) {
	inputs := `uci

position fen 4k3/8/2K5/8/8/8/8/8 w - - 0 1 moves c6c5
position startpos
fen
quit
`

	outputs := &bytes.Buffer{}
	errors := &bytes.Buffer{}

	d := uci.NewDriver(uci.WithInput(strings.NewReader(inputs)),
		uci.WithOutput(outputs),
		uci.WithError(errors),
		uci.WithSearch(&MockSearch{}))

	d.Run()

	assert.Empty(t, errors)
	assert.Contains(t, outputs.String(), "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
}

func TestGo(t *testing.T) {
	inputs := `uci
go depth 5
quit
`
	outputs := &bytes.Buffer{}
	errors := &bytes.Buffer{}

	search := &MockSearch{}

	search.MockMove(move.From(E2) | move.To(E4))
	search.MockScore(-123)

	d := uci.NewDriver(
		uci.WithInput(strings.NewReader(inputs)),
		uci.WithOutput(outputs),
		uci.WithError(errors),
		uci.WithSearch(search),
	)

	d.Run()

	assert.Empty(t, errors)
	assert.Contains(t, outputs.String(), "bestmove e2e4")
}

func TestGoDepth(t *testing.T) {
	inputs := `uci
go depth 5
quit
`
	outputs := &bytes.Buffer{}
	errors := &bytes.Buffer{}

	search := &MockSearch{}

	d := uci.NewDriver(
		uci.WithInput(strings.NewReader(inputs)),
		uci.WithOutput(outputs),
		uci.WithError(errors),
		uci.WithSearch(search),
	)

	d.Run()

	assert.Empty(t, errors)
	assert.Equal(t, Depth(5), search.Options.Depth)
}
