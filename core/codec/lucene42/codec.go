package lucene42

import (
	"github.com/balzaczyy/golucene/core/codec/lucene40"
	"github.com/balzaczyy/golucene/core/codec/lucene41"
	"github.com/balzaczyy/golucene/core/codec/perfield"
	// "github.com/balzaczyy/golucene/core/codec/lucene42"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
)

// lucene42/Lucene42Codec.java

/*
Implements the Lucene 4.2 index format, with configurable per-field
postings and docvalues formats.

If you want to reuse functionality of this codec, in another codec,
extend FilterCodec.
*/
var Lucene42Codec = NewCodec("Lucene42",
	lucene41.NewLucene41StoredFieldsFormat(),
	NewLucene42TermVectorsFormat(),
	NewLucene42FieldInfosFormat(),
	lucene40.NewLucene40SegmentInfoFormat(),
	nil, // liveDocsFormat
	perfield.NewPerFieldPostingsFormat(func(field string) PostingsFormat {
		panic("not implemented yet")
	}),
	perfield.NewPerFieldDocValuesFormat(func(field string) DocValuesFormat {
		panic("not implemented yet")
	}),
	newReadonlyLucene42NormsFormat(),
)

func init() {
	RegisterCodec(Lucene42Codec)
}

type readonlyLucene42NormsFormat struct {
	*Lucene42NormsFormat
}

func (f *readonlyLucene42NormsFormat) NormsConsumer(state *SegmentWriteState) (w DocValuesConsumer, err error) {
	panic("this codec can only be used for reading")
}

func newReadonlyLucene42NormsFormat() *readonlyLucene42NormsFormat {
	return &readonlyLucene42NormsFormat{
		NewLucene42NormsFormat(),
	}
}
