package transp_test

import (
	"testing"

	"github.com/paulsonkoly/chess-3/board"
	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/transp"
	"github.com/stretchr/testify/assert"
)

const (
	key = 0xdeadbeeff00baa4
)

func TestTransposition(t *testing.T) {
	assert.PanicsWithValue(t, "invalid transposition table size 0", func() { transp.New(0) })

	tt := transp.New(1 * transp.MegaBytes)

	entry, ok := tt.LookUp(key)

	assert.False(t, ok)
	assert.Nil(t, entry)

	tt.Insert(key, 0, 1, 1, move.New(E1, F1), 100, transp.Exact)

	entry, ok = tt.LookUp(key)

	assert.True(t, ok)
	assert.NotNil(t, entry)

	assert.Equal(t, move.New(E1, F1), entry.Move)
	assert.Equal(t, Score(100), entry.Value(1))
	assert.Equal(t, Depth(1), entry.Depth())
	assert.Equal(t, transp.Exact, entry.Type())

	tt.Clear()

	entry, ok = tt.LookUp(key)

	assert.False(t, ok)
	assert.Nil(t, entry)
}

func TestMateScores(t *testing.T) {
	tt := transp.New(1 * transp.MegaBytes)

	tt.Insert(key, 0, 1, 3, move.New(E1, F1), -Inf+5, transp.Exact)

	entry, ok := tt.LookUp(key)

	assert.True(t, ok)
	assert.NotNil(t, entry)

	assert.Equal(t, move.New(E1, F1), entry.Move)
	assert.Equal(t, -Inf+9, entry.Value(7))
	assert.Equal(t, Depth(1), entry.Depth())
	assert.Equal(t, transp.Exact, entry.Type())
}

func TestBucket(t *testing.T) {
	key1 := board.Hash(0x0000deadbeefdead)
	key2 := board.Hash(0x0001deadbeefdead)
	key3 := board.Hash(0x0002deadbeefdead)
	key4 := board.Hash(0x0003deadbeefdead)
	key5 := board.Hash(0x0004deadbeefdead)

	tt := transp.New(1 * transp.MegaBytes)

	// key1 is lowest quality, 0 gen depth 1
	tt.Insert(key1, 0, 1, 3, move.New(E1, F1), 0, transp.UpperBound)
	tt.Insert(key2, 2, 2, 3, move.New(E1, G1), 100, transp.Exact)
	tt.Insert(key3, 2, 5, 2, move.New(E1, H1), -50, transp.LowerBound)
	tt.Insert(key4, 1, 1, 4, move.New(E1, H1), 70, transp.Exact)
	// no matching key, replace lowest quality
	tt.Insert(key5, 2, 5, 4, 0, 70, transp.Exact)

	entry, ok := tt.LookUp(key5)

	assert.True(t, ok)
	assert.NotNil(t, entry)

	// move not kept, not the same position.
	assert.Equal(t, move.Move(0), entry.Move)
	assert.Equal(t, Score(70), entry.Value(5))
	assert.Equal(t, Depth(5), entry.Depth())
	assert.Equal(t, transp.Exact, entry.Type())

	// new gen, lower depth replace keeping move
	tt.Insert(key3, 3, 3, 4, 0, 70, transp.Exact)

	entry, ok = tt.LookUp(key3)

	assert.True(t, ok)
	assert.NotNil(t, entry)

	// move kept, not the same position.
	assert.Equal(t, move.New(E1, H1), entry.Move)
	assert.Equal(t, Score(70), entry.Value(2))
	assert.Equal(t, Depth(3), entry.Depth())
	assert.Equal(t, transp.Exact, entry.Type())

	// low quality insert not performed on matching key
	tt.Insert(key3, 3, 0, 6, 0, 100, transp.LowerBound)

	entry, ok = tt.LookUp(key3)

	assert.True(t, ok)
	assert.NotNil(t, entry)

	assert.Equal(t, move.New(E1, H1), entry.Move)
	assert.Equal(t, Score(70), entry.Value(2))
	assert.Equal(t, Depth(3), entry.Depth())
	assert.Equal(t, transp.Exact, entry.Type())
}
