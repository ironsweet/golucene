package index

import (
	// "github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/codec/lucene40"
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
	PostingsFormatForField func(string) PostingsFormat
}

var Lucene45CodecImpl = func() *Lucene45Codec {
	f := func(string) PostingsFormat {
		return newLucene41PostingsFormat()
	}
	codec := &CodecImpl{
		name:             "Lucene45",
		fieldsFormat:     newLucene41StoredFieldsFormat(),
		vectorsFormat:    newLucene42TermVectorsFormat(),
		fieldInfosFormat: newLucene42FieldInfosFormat(),
		infosFormat:      lucene40.NewLucene40SegmentInfoFormat(),
		liveDocsFormat:   new(lucene40.Lucene40LiveDocsFormat),
		postingsFormat: newPerFieldPostingsFormat(func(field string) PostingsFormat {
			return f(field)
		}),
		docValuesFormat: newPerFieldDocValuesFormat(func(field string) DocValuesFormat {
			panic("not implemented yet")
		}),
		normsFormat: newLucene42NormsFormat(),
	}
	return &Lucene45Codec{codec, f}
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
	dvp *Lucene42DocValuesProducer, err error) {
	panic("not implemented yet")
}
