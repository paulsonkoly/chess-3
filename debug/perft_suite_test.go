package debug_test

import (
	"fmt"
	"testing"

	"github.com/paulsonkoly/chess-3/debug"
	"github.com/stretchr/testify/assert"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

const epd = "standard.epd"

func TestPerftSuite(t *testing.T) {
	inp := Must(debug.NewEPDReader(epd))

	for inp.Scan() {
		entry := inp.Entry()

		t.Run(fmt.Sprintf("%s at depth %d", entry.Fen, entry.D),
			func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, entry.Cnt, debug.Perft(entry.Board, int(entry.D)))
			})
	}
}
