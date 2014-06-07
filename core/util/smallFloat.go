package util

import (
	"math"
)

// util/SmallFloat.java

/*
floatToByte(b, mantissaBits=3, zeroExponent=15)
smallest non-zero value = 5.820766E-10
largest value = 7.5161928E9
epsilon = 0.125
*/
func FloatToByte315(f float32) int8 {
	bits := math.Float32bits(f)
	smallfloat := bits >> (24 - 3)
	if smallfloat <= ((63 - 15) << 3) {
		if bits <= 0 {
			return 0
		}
		return 1
	}
	if smallfloat >= ((63-15)<<3)+0x100 {
		return -1
	}
	return int8(smallfloat - ((63 - 15) << 3))
}
