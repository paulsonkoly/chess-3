package types

const (
	Short = 0
	Long  = 1
)

type Castle byte

const (
	NoCastle   = 0
	ShortWhite = 1
	LongWhite  = 2
	ShortBlack = 3
	LongBlack  = 4
)

func C(c Color, typ int) Castle {
	return Castle(int(c)*2 + typ + 1)
}

type CastlingRights byte

func CRights(castles ...Castle) CastlingRights {
	result := CastlingRights(0)
	for _, c := range castles {
		result |= 1 << int(c-1)
	}
	return result
}
