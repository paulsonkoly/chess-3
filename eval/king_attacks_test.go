package eval

import (
	"testing"

	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/stretchr/testify/assert"
)

func TestSigmoidal(t *testing.T) {
	const (
		rangeEnd = Score(100)
	)
	for i := range rangeEnd {
		iVal := sigmoidal(i)
		fVal := sigmoidal(float64(i))

		assert.InDelta(t, float64(iVal), fVal, 0.5)
	}
}
