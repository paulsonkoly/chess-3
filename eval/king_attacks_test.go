package eval

import (
	"testing"

	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/stretchr/testify/assert"
)

func TestSigmoidal(t *testing.T) {
	const (
		rangeStart = Score(-50)
		rangeEnd   = Score(49)
	)
	for i := rangeStart; i < rangeEnd; i++ {
		iVal := sigmoidal(i)
		fVal := sigmoidal(float64(i))

		assert.InDelta(t, float64(iVal), fVal, 0.5)
	}
}
