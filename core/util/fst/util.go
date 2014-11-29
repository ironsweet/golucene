package fst

import (
	"github.com/balzaczyy/golucene/core/util"
)

// fst/Util.java

/* Just takes unsigned byte values from the BytesRef and converts into an IntsRef. */
func ToIntsRef(input []byte, scratch *util.IntsRefBuilder) *util.IntsRef {
	scratch.Clear()
	for _, v := range input {
		scratch.Append(int(v))
	}
	return scratch.Get()
}
