package types

type Color byte

const (
  White = Color(iota)
  Black
)

func (c Color) Flip() Color { return c ^ 1 }


