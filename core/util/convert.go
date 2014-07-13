package util

import (
	"math"
	"strconv"
)

// util/SmallFloat.java
// Floating point numbers smaller than 32 bits.

/** byteToFloat(b, mantissaBits=3, zeroExponent=15) */
func Byte315ToFloat(b byte) float32 {
	// on Java1.5 & 1.6 JVMs, prebuilding a decoding array and doing a lookup
	// is only a little bit faster (anywhere from 0% to 7%)
	// However, is it faster for Go? TODO need benchmark
	if b == 0 {
		return 0
	}
	bits := uint32(b) << (24 - 3)
	bits += (63 - 15) << 24
	return math.Float32frombits(bits)
}

func ItoHex(i int64) string {
	return strconv.FormatInt(i, 16)
}
