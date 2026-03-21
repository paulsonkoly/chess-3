package heur_test

import (
	"testing"

	. "github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/heur"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/stretchr/testify/assert"
)

func TestKiller(t *testing.T) {
	k := heur.NewKiller()

	assert.Equal(t, move.Move(0), k.LookUp(0, 0), "<*0*, 0>")

	k.Add(0, move.From(E1)|move.To(G1))

	assert.Equal(t, move.From(E1)|move.To(G1), k.LookUp(0, 0), "< *e1g1*, 0>")
	assert.Equal(t, move.Move(0), k.LookUp(0, 1), "<e1g1, *0* >")

	k.Add(0, move.From(A2)|move.To(A3))

	assert.Equal(t, move.From(A2)|move.To(A3), k.LookUp(0, 0), "< *a2a3*, e1g1>")
	assert.Equal(t, move.From(E1)|move.To(G1), k.LookUp(0, 1), "< a2a3, *e1g1*>")

	k.Add(0, move.From(E1)|move.To(G1))

	assert.Equal(t, move.From(E1)|move.To(G1), k.LookUp(0, 0), "< *e1g1*, a2a3>")
	assert.Equal(t, move.From(A2)|move.To(A3), k.LookUp(0, 1), "< e1g1, *a2a3*>")

	k.Add(0, move.From(E1)|move.To(G1))

	assert.Equal(t, move.From(E1)|move.To(G1), k.LookUp(0, 0), "< *e1g1*, a2a3>")
	assert.Equal(t, move.From(A2)|move.To(A3), k.LookUp(0, 1), "< e1g1, *a2a3*>")
}
