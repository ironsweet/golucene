package lucene410

import (
	"github.com/balzaczyy/golucene/core/codec/lucene40"
	"github.com/balzaczyy/golucene/core/codec/lucene41"
	"github.com/balzaczyy/golucene/core/codec/lucene42"
	"github.com/balzaczyy/golucene/core/codec/lucene46"
	"github.com/balzaczyy/golucene/core/codec/lucene49"
	"github.com/balzaczyy/golucene/core/codec/perfield"
	. "github.com/balzaczyy/golucene/core/codec/spi"
)

// codec/lucene410/Lucene410Codec.java

func init() {
	RegisterCodec(newLucene410Codec())
}

/*
Implements the Lucene 4.10 index format, with configuration per-field
postings and docvalues format.

If you want to reuse functionality of this codec in another codec,
extend FilterCodec (or embeds the Codec in Go style).
*/
type Lucene410Codec struct {
	*CodecImpl
}

func newLucene410Codec() *Lucene410Codec {
	return &Lucene410Codec{NewCodec("Lucene410",
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
		new(lucene49.Lucene49NormsFormat),
	)}
}
