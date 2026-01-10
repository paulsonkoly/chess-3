package chess

import "golang.org/x/exp/constraints"

// Abs is the absolute value of some signed int type.
func Abs[T constraints.Signed](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

// Signum is the sign, -1, 0, +1 for some signed int type.
func Signum[T constraints.Signed](x T) T {
	switch {
	case x < 0:
		return -1
	case x > 0:
		return 1
	}
	return 0
}

// Clamp clamps the value x between a and b.
func Clamp[T constraints.Signed](x, a, b T) T {
	return min(b, max(x, a))
}

// Must panics if err is not nil with err, otherwise it's v.
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
