package fst

import (
	"github.com/balzaczyy/golucene/core/util"
)

// fst/Util.java

/* Just takes unsigned byte values from the BytesRef and converts into an IntsRef. */
func ToIntsRef(input []byte, scratch *util.IntsRef) *util.IntsRef {
	scratch.Grow(len(input))
	for i, v := range input {
		scratch.Ints[i] = int(v)
	}
	scratch.Offset = 0
	scratch.Length = len(input)
	return scratch
}
