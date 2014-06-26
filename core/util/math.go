package util

import (
	"math"
)

// util/MathUtil.java

/* Returns x <= 0 ? 0 : floor(log(x) ? log(base)) */
func Log(x int64, base int) int {
	assert2(base > 1, "base must be > 1")
	ret := 0
	for x >= int64(base) {
		x /= int64(base)
		ret++
	}
	return ret
}

/*
Return the greatest common divisor of a and b, consistently with
big.GCD(a, b).

NOTE: A greatest common divisor must be positive, but 2^64 cannot be
expressed as an int64 although it is the GCD of math.MinInt64 and 0
and the GCD of math.MinInt64 and math.MinInt64. So in these 2 cases,
and only them, this method will return math.MinInt64.
*/
func Gcd(a, b int64) int64 {
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}
	if a == 0 {
		return b
	} else if b == 0 {
		return a
	}
	commonTrailingZeros := NumberOfTrailingZeros(a | b)
	a = int64(uint64(a) >> NumberOfTrailingZeros(a))
	for {
		b = int64(uint64(b) >> NumberOfTrailingZeros(b))
		if a == b {
			break
		}
		if a > b || a == math.MinInt64 { // math.MinInt64 is treated as 2^64
			a, b = b, a
		}
		if a == 1 {
			break
		}
		b -= a
	}
	return a << commonTrailingZeros
}

func NumberOfTrailingZeros(n int64) uint {
	if n == 0 {
		return 64
	}
	ans := 0
	for {
		if n&1 != 0 {
			break
		}
		n >>= 1
		ans++
	}
	return uint(ans)
}
