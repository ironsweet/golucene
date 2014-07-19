package lucene49

import (
	. "github.com/balzaczyy/golucene/core/codec/spi"
)

// lucene49/lucene49Codec.java

func init() {
	RegisterCodec(newLucene49Codec())
}

/*
Implements the Lucene 4.9 index format, with configurable per-field
postings and docvalues formats.

If you want to reuse functionality of this codec in another codec,
extend FilterCodec.
*/
type Lucene49Codec struct {
	*CodecImpl
}

func newLucene49Codec() *Lucene49Codec {
	panic("not implemented yet")
}
