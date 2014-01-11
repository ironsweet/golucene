package index

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	"github.com/balzaczyy/golucene/core/store"
)

// compressing/CompressingStoredFieldsFormat.java

/*
A StoredFieldsFormat that is very similar to Lucene40StoredFieldsFormat
but compresses documents in chunks in order to improve the
compression ratio.

For a chunk size of chunkSize bytes, this StoredFieldsFormat does not
support documents larger than (2^31 - chunkSize) bytes. In case this
is a problem, you should use another format, such as Lucene40StoredFieldsFormat.

For optimal performance, you should use a MergePolicy that returns
segments that have the biggest byte size first.
*/
type CompressingStoredFieldsFormat struct {
	formatName      string
	segmentSuffix   string
	compressionMode codec.CompressionMode
	chunkSize       int
}

func (format *CompressingStoredFieldsFormat) FieldsReader(d store.Directory, si SegmentInfo,
	fn FieldInfos, context store.IOContext) (r StoredFieldsReader, err error) {
	panic("not implemented yet")
}

func (format *CompressingStoredFieldsFormat) FieldsWriter(d store.Directory, si SegmentInfo,
	context store.IOContext) (w StoredFieldsWriter, err error) {
	panic("not implemented yet")
}

func (format *CompressingStoredFieldsFormat) String() string {
	return fmt.Sprintf("CompressingStoredFieldsFormat(compressionMode=%v, chunkSize=%v)",
		format.compressionMode, format.chunkSize)
}

// compressing/CompressingTermVectorsFormat.java

// A TermVectorsFormat that compresses chunks of documents together
// in order to improve the compression ratio.
type CompressingTermVectorsFormat struct{}

/*
Create a new CompressingTermVectorsFormat

formatName is the name of the format. This name will be used in the
file formats to perform codec header checks.

The compressionMode parameter allows you to choose between compression
algorithms that have various compression and decompression speeds so
that you can pick the one that best fits your indexing and searching
throughput. You should never instantiate two CompressingTermVectorsFormats
that have the same name but different CompressionModes.

chunkSize is the minimum byte size of a chunk of documents. Highter
values of chunkSize should improve the compression ratio but will
require more memory at indexing time and might make document loading
a little slower (depending on the size of your OS cache compared to
the size of your index).
*/
func newCompressingTermVectorsFormat(formatName, segmentSuffix string,
	compressionMode codec.CompressionMode, chunkSize int) *CompressingTermVectorsFormat {
	panic("not implemented yet")
}

func (vf *CompressingTermVectorsFormat) VectorsReader(d store.Directory,
	segmentInfo SegmentInfo, fieldsInfos FieldInfos, context store.IOContext) (r TermVectorsReader, err error) {
	panic("not implemented yet")
}

func (vf *CompressingTermVectorsFormat) VectorsWriter(d store.Directory,
	segmentInfo SegmentInfo, context store.IOContext) (w TermVectorsWriter, err error) {
	panic("not implemented yet")
}
