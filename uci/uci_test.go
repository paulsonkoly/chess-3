package uci_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/paulsonkoly/chess-3/transp"
	"github.com/paulsonkoly/chess-3/uci"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUCI(t *testing.T) {
	inputs := `uci
quit
`
	outputs := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	d := uci.NewDriver(uci.WithInput(strings.NewReader(inputs)),
		uci.WithOutput(outputs),
		uci.WithError(errors),
		uci.WithSearch(uci.NewMockSearch(t)))

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
		uci.WithSearch(uci.NewMockSearch(t)))

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
	search := uci.NewMockSearch(t)

	search.EXPECT().Clear().Once()

	d := uci.NewDriver(uci.WithInput(strings.NewReader(inputs)),
		uci.WithOutput(outputs),
		uci.WithError(errors),
		uci.WithSearch(search))

	d.Run()

	assert.Empty(t, errors)
}


func TestHashSettings(t *testing.T) {
	inputs := `uci
setoption name Hash value 16
quit
`
	outputs := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	search := uci.NewMockSearch(t)

	search.On("ResizeTT", mock.Anything).Return()

	d := uci.NewDriver(uci.WithInput(strings.NewReader(inputs)),
		uci.WithOutput(outputs),
		uci.WithError(errors),
		uci.WithSearch(search))

	d.Run()

	assert.Empty(t, errors)
	search.AssertCalled(t, "ResizeTT", 16*transp.MegaBytes)
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
		uci.WithSearch(uci.NewMockSearch(t)))

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
		uci.WithSearch(uci.NewMockSearch(t)))

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
		uci.WithSearch(uci.NewMockSearch(t)))

	d.Run()

	assert.Empty(t, errors)
	assert.Contains(t, outputs.String(), "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
}
