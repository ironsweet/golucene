package spi

import (
	. "github.com/balzaczyy/golucene/core/index/model"
	"io"
)

// codecs/DocValuesFormat.java

/*
Encodes/decodes per-document values.

Note, when extending this class, the name Name() may be written into
the index in certain configurations. In order for the segment to be
read, the name must resolve to your implemetation via LoadXYZ().
Since Go doesn't have Java's SPI locate mechanism, this method use
manual mappings to resolve format names.

If you implement your own format, make sure that it is manually
included.
*/
type DocValuesFormat interface {
	Name() string
	// Returns a DocValuesConsumer to write docvalues to the index.
	FieldsConsumer(state *SegmentWriteState) (w DocValuesConsumer, err error)
	// Returns a DocValuesProducer to read docvalues from the index.
	//
	// NOTE: by the time this call returns, it must
	// hold open any files it will need to use; else, those files may
	// be deleted. Additionally, required fiels may be deleted during
	// the execution of this call before there is a chance to open them.
	// Under these circumstances an IO error should be returned by the
	// implementation. IO errors are expected and will automatically
	// cause a retry of the segment opening logic with the newly
	// revised segments.
	FieldsProducer(state SegmentReadState) (r DocValuesProducer, err error)
}

var allDocValuesFormats = map[string]DocValuesFormat{}

// workaround Lucene Java's SPI mechanism
func RegisterDocValuesFormat(formats ...DocValuesFormat) {
	for _, format := range formats {
		allDocValuesFormats[format.Name()] = format
	}
}

func LoadDocValuesProducer(name string, state SegmentReadState) (fp DocValuesProducer, err error) {
	panic("not implemented yet")
	// switch name {
	// case "Lucene42":
	// 	return newLucene42DocValuesProducer(state, LUCENE42_DV_DATA_CODEC, LUCENE42_DV_DATA_EXTENSION,
	// 		LUCENE42_DV_METADATA_CODEC, LUCENE42_DV_METADATA_EXTENSION)
	// case "Lucene45":
	// 	return newLucene45DocValuesProducer(state, LUCENE45_DV_DATA_CODEC, LUCENE45_DV_DATA_EXTENSION,
	// 		LUCENE45_DV_META_CODEC, LUCENE45_DV_META_EXTENSION)
	// }
	// panic(fmt.Sprintf("Service '%v' not found.", name))
}

// codecs/DocValuesConsumer.java
/*
Abstract API that consumes numeric, binary and sorted docvalues.
Concret implementations of this actually do "something" with the
docvalues (write it into the index in a specific format).

The lifecycle is:

1. DocValuesConsumer is created by DocValuesFormat.FieldsConsumer()
or NormsFormat.NormsConsumer().
2. AddNumericField, AddBinaryField, or addSortedField are called for
each Numeric, Binary, or Sorted docvalues field. The API is a "pull"
rather than "push", and the implementation is free to iterate over
the values multiple times.
3. After all fields are added, the consumer is closed.
*/
type DocValuesConsumer interface {
	io.Closer
	// Writes numeric docvalues for a field.
	AddNumericField(*FieldInfo, func() func() (interface{}, bool)) error
}

// codecs/DocvaluesProducer.java

// Abstract API that produces numeric, binary and sorted docvalues.
type DocValuesProducer interface {
	io.Closer
	Numeric(field *FieldInfo) (v NumericDocValues, err error)
	Binary(field *FieldInfo) (v BinaryDocValues, err error)
	Sorted(field *FieldInfo) (v SortedDocValues, err error)
	SortedSet(field *FieldInfo) (v SortedSetDocValues, err error)
}

// type NumericDocValues interface {
// 	Value(docID int) int64
// }
type NumericDocValues func(docID int) int64

/* A per-document []byte */
type BinaryDocValues interface {
	// Lookup the value for document. The returned BytesRef may be
	// re-used across calls to get() so make sure to copy it if you
	// want to keep it around.
	Get(docId int) []byte
}

type SortedDocValues interface {
	BinaryDocValues
	Ord(docID int) int
	LookupOrd(int) []byte
	ValueCount() int
}

type SortedSetDocValues interface {
	NextOrd() int64
	SetDocument(docID int)
	LookupOrd(int64) []byte
	ValueCount() int64
}
