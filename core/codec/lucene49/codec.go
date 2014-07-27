package lucene49

import (
	"github.com/balzaczyy/golucene/core/codec/lucene40"
	"github.com/balzaczyy/golucene/core/codec/lucene41"
	"github.com/balzaczyy/golucene/core/codec/lucene42"
	"github.com/balzaczyy/golucene/core/codec/lucene46"
	"github.com/balzaczyy/golucene/core/codec/perfield"
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
	return &Lucene49Codec{NewCodec("Lucene49",
		lucene41.NewLucene41StoredFieldsFormat(),
		lucene42.NewLucene42TermVectorsFormat(),
		lucene46.NewLucene46FieldInfosFormat(),
		lucene46.NewLucene46SegmentInfoFormat(),
		new(lucene40.Lucene40LiveDocsFormat),
		perfield.NewPerFieldPostingsFormat(func(field string) PostingsFormat {
			return LoadPostingsFormat("Lucene41")
		}),
		perfield.NewPerFieldDocValuesFormat(func(field string) DocValuesFormat {
			panic("not implemented yet")
		}),
		new(Lucene49NormsFormat),
	)}
}
