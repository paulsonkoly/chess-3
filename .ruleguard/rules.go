package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

const chessPkg = "github.com/paulsonkoly/chess-3/chess"

//doc:summary prefers chess helper methods over manual coordinate and bitboard arithmetic
func chessHelpers(m dsl.Matcher) {
	m.Import(chessPkg)

	m.Match(`$s / 8`).
		Where(m.File().Imports(chessPkg) && m["s"].Type.Is("chess.Square")).
		Report(`use $s.Rank()`)

	m.Match(`$s % 8`).
		Where(m.File().Imports(chessPkg) && m["s"].Type.Is("chess.Square")).
		Report(`use $s.File()`)

	m.Match(`$c ^ 1`).
		Where(m.File().Imports(chessPkg) && m["c"].Type.Is("chess.Color")).
		Report(`use $c.Flip()`)
}
