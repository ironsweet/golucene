package index

import (
	. "github.com/balzaczyy/golucene/core/codec/spi"
	"math/rand"
)

// Test-only

// index/RandomCodec.java

/*
Codec that assigns per-field random postings format.

The same field/format assignment will happen regardless of order, a
hash is computed up front that determines the mapping. This means
fields can be put into things like HashSets and added to documents
in different orders and the tests will still be deterministic and
reproducible.
*/
type RandomCodec struct {
	*CodecImpl
}

func NewRandomCodec(r *rand.Rand, avoidCodecs map[string]bool) *RandomCodec {
	panic("not implemented yet")
}
