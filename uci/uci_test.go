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
	tests := []struct {
		name   string
		inputs string
	}{
		{"isready", "uci\nisready\nquit\n"},
		{"whitespace at front", "uci\n  \tisready\nquit\n"},
		{"whitespace at the back", "uci\nisready \t \nquit\n"},
		{"whitespaces everywhere", "uci \n \t isready \n \t \n quit\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputs := &bytes.Buffer{}
			errors := &bytes.Buffer{}
			search := &MockSearch{}

			d := uci.NewDriver(
				uci.WithInput(strings.NewReader(tt.inputs)),
				uci.WithOutput(outputs),
				uci.WithError(errors),
				uci.WithSearch(search),
			)

			d.Run()

			assert.Empty(t, errors)
			assert.Contains(t, outputs.String(), "readyok")
		})
	}
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
	tests := []struct {
		name      string
		inputs    string
		want      int
		wantError string
	}{
		{"set hash to 16 Mb", "setoption name Hash value 16", 16 * transp.MegaBytes, ""},
		{"dangling tokens", "setoption name Hash value 17 extra", 17 * transp.MegaBytes, ""},
		{"arguments missing", "setoption name Hash", 0, "argument missing"},
		{"hash at minimum", "setoption name Hash value 1", 1 * transp.MegaBytes, ""},
		{"hash at maximum", "setoption name Hash value 1024", 1024 * transp.MegaBytes, ""},
		{"hash below minimum", "setoption name Hash value 0", 0, ""},
		{"hash above maximum", "setoption name Hash value 2048", 0, ""},
		{"invalid hash value", "nsetoption name Hash value invalid", 0, ""},
		{"malformed setoption - missing value", "setoption name Hash", 0, "argument missing"},
		{"malformed setoption - wrong structure", "setoption whops Hash value 16", 0, ""},
		{"unknown option name", "setoption name Unknown value 100", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputs := &bytes.Buffer{}
			errors := &bytes.Buffer{}
			search := &MockSearch{}

			prelude := "uci\n"
			prolog := "\nquit\n"

			inputs := prelude + tt.inputs + prolog

			d := uci.NewDriver(
				uci.WithInput(strings.NewReader(inputs)),
				uci.WithOutput(outputs),
				uci.WithError(errors),
				uci.WithSearch(search),
			)

			d.Run()

			if tt.wantError != "" {
				assert.Contains(t, errors.String(), tt.wantError)
			} else {
				assert.Empty(t, errors)
				assert.Equal(t, tt.want, search.TTSize)
			}
		})
	}
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

func TestGoTime(t *testing.T) {
	tests := []struct {
		name      string
		inputs    string
		wantError string
	}{
		{"go wtime/winc/btime/binc", "go wtime 1000 winc 300 btime 900 binc 200", ""},
		{"go wtime time missing", "go wtime", "argument missing"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputs := &bytes.Buffer{}
			errors := &bytes.Buffer{}
			search := &MockSearch{}

			search.MockMove(move.From(E2) | move.To(E4))
			search.MockScore(-123)

			prelude := "uci\n"
			prolog := "\nquit\n"

			inputs := prelude + tt.inputs + prolog

			d := uci.NewDriver(
				uci.WithInput(strings.NewReader(inputs)),
				uci.WithOutput(outputs),
				uci.WithError(errors),
				uci.WithSearch(search),
			)

			d.Run()

			if tt.wantError != "" {
				assert.NotContains(t, outputs.String(), "bestmove")
				assert.Contains(t, errors.String(), tt.wantError)
			} else {
				assert.Empty(t, errors)
				assert.Contains(t, outputs.String(), "bestmove e2e4")
			}
		})
	}
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

func TestDebug(t *testing.T) {
	tests := []struct {
		name      string
		inputs    string
		want      bool
		wantError string
	}{
		{"debug with no arguments", "debug", false, "on/off missing"},
		{"debug with extra arguments after", "debug on extra", true, ""},
		{"debug with extra arguments in between", "debug extra on", false, ""},
		{"debug with invalid argument", "debug invalid", false, ""},
		{"debug on and off sequence", "debug on\ndebug off", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputs := &bytes.Buffer{}
			errors := &bytes.Buffer{}
			search := &MockSearch{}

			prelude := "uci\n"
			prolog := "\ngo\nquit\n"

			inputs := prelude + tt.inputs + prolog

			d := uci.NewDriver(
				uci.WithInput(strings.NewReader(inputs)),
				uci.WithOutput(outputs),
				uci.WithError(errors),
				uci.WithSearch(search),
			)

			d.Run()

			assert.NotEmpty(t, outputs.String())
			assert.Equal(t, tt.want, search.Options.Debug)
			assert.Contains(t, errors.String(), tt.wantError)
		})
	}
}

