package board

import (
	"fmt"

	"github.com/paulsonkoly/chess-3/types"
	"github.com/pborman/ansi"
)

const PString = " PNBRQK pnbrqk"

func (b Board) Print(w ansi.Writer) {
	if _, err := w.Write([]byte(" abcdefgh\n")); err != nil {
		panic(err)
	}
	for invRank := range 8 {
		if _, err := w.Write([]byte{8 + '0' - byte(invRank)}); err != nil {
			panic(err)
		}
		for file := range 8 {
			sq := (7-invRank)*8 + file
			sqBB := BitBoard(1 << sq)

			var color *ansi.Writer

			if (invRank+file)%2 == 0 {
				color = w.SetBackground(ansi.White)
			} else {
				color = w.SetBackground(ansi.Magenta)
			}

			var cix int
			if b.Colors[types.White]&sqBB != 0 {
				color = color.Red().Bold()
				cix = 0
			} else if b.Colors[types.Black]&sqBB != 0 {
				color = color.Red()
				cix = 1
			}

			p := int(b.SquaresToPiece[sq])

			if _, err := color.Write([]byte{PString[cix*7+p]}); err != nil {
				panic(err)
			}
		}
		if _, err := w.Reset().Write([]byte("\n")); err != nil {
			panic(fmt.Sprintf("write error %s", err))
		}
	}
}
