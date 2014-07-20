package lucene42

import (
	"github.com/balzaczyy/golucene/core/codec/lucene40"
	"github.com/balzaczyy/golucene/core/codec/lucene41"
	"github.com/balzaczyy/golucene/core/codec/perfield"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
)

// lucene42/Lucene42RWCodec.java

var dv = newLucene42RWDocValuesFormat()

// Read-write version of Lucene42Codec for testing.
var Lucene42RWCodec = NewCodec("Lucene42",
	lucene41.NewLucene41StoredFieldsFormat(),
	NewLucene42TermVectorsFormat(),
	NewLucene42FieldInfosFormat(),
	lucene40.NewLucene40SegmentInfoFormat(),
	nil, // liveDocsFormat
	perfield.NewPerFieldPostingsFormat(func(field string) PostingsFormat {
		panic("not implemented yet")
	}),
	perfield.NewPerFieldDocValuesFormat(func(field string) DocValuesFormat {
		return dv
	}),
	NewLucene42NormsFormat(),
)

// lucene42/Lucene42RWDocValuesFormat.java

// Read-write version of Lucene42DocValuesFormat for testing
type Lucene42RWDocValuesFormat struct {
	*Lucene42DocValuesFormat
}

func (f *Lucene42RWDocValuesFormat) FieldsConsumer(state *SegmentWriteState) (w DocValuesConsumer, err error) {
	panic("not implemented yet")
}

func newLucene42RWDocValuesFormat() *Lucene42RWDocValuesFormat {
	return &Lucene42RWDocValuesFormat{
		NewLucene42DocValuesFormat(),
	}
}
