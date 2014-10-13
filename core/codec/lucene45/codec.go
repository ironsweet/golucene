package lucene45

import (
	// "github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/codec/lucene40"
	"github.com/balzaczyy/golucene/core/codec/lucene41"
	"github.com/balzaczyy/golucene/core/codec/lucene42"
	"github.com/balzaczyy/golucene/core/codec/perfield"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
)

// codec/lucene45/Lucene45Codec.java

// NOTE: if we make largish changes in a minor release, easier to
// just make Lucene46Codec or whatever if they are backwards
// compatible or smallish we can probably do the backwards in the
// postingreader (it writes a minor version, etc).
/*
Implements the Lucene 4.5 index format, with configurable per-field
postings and docvalues formats.

If you want to reuse functionality of this codec in another codec,
extend FilterCodec.
*/
type Lucene45Codec struct {
	*CodecImpl
}

func init() {
	RegisterCodec(Lucene45CodecImpl)
}

var Lucene45CodecImpl = func() *Lucene45Codec {
	codec := NewCodec("Lucene45",
		lucene41.NewLucene41StoredFieldsFormat(),
		lucene42.NewLucene42TermVectorsFormat(),
		lucene42.NewLucene42FieldInfosFormat(),
		lucene40.NewLucene40SegmentInfoFormat(),
		new(lucene40.Lucene40LiveDocsFormat),
		perfield.NewPerFieldPostingsFormat(func(field string) PostingsFormat {
			return LoadPostingsFormat("Lucene41")
		}),
		perfield.NewPerFieldDocValuesFormat(func(field string) DocValuesFormat {
			panic("not implemented yet")
		}),
		lucene42.NewLucene42NormsFormat(),
	)
	return &Lucene45Codec{
		CodecImpl: codec,
	}
}()

// codec/lucene45/Lucene45DocValuesFormat.java

const (
	LUCENE45_DV_DATA_CODEC     = "Lucene45DocValuesData"
	LUCENE45_DV_DATA_EXTENSION = "dvd"
	LUCENE45_DV_META_CODEC     = "Lucene45valuesMetadata"
	LUCENE45_DV_META_EXTENSION = "dvm"
)

// codec/lucene45/Lucene45DocValuesProducer.java

type Lucene45DocvaluesProducer struct {
}

// expert: instantiate a new reader
func newLucene45DocValuesProducer(
	state SegmentReadState, dataCodec, dataExtension, metaCodec, metaExtension string) (
	dvp *lucene42.Lucene42DocValuesProducer, err error) {
	panic("not implemented yet")
}
