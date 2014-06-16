package util

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
