package compressing

import (
	"github.com/balzaczyy/golucene/core/codec/spi"
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
)

// compressing/CompressingTermVectorsFormat.java

// A TermVectorsFormat that compresses chunks of documents together
// in order to improve the compression ratio.
type CompressingTermVectorsFormat struct {
	formatName      string
	segmentSuffix   string
	compressionMode CompressionMode
	chunkSize       int
}

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
func NewCompressingTermVectorsFormat(formatName, segmentSuffix string,
	compressionMode CompressionMode, chunkSize int) *CompressingTermVectorsFormat {
	assert2(chunkSize >= 1, "chunkSize must be >= 1")
	return &CompressingTermVectorsFormat{
		formatName:      formatName,
		segmentSuffix:   segmentSuffix,
		compressionMode: compressionMode,
		chunkSize:       chunkSize,
	}
}

func (vf *CompressingTermVectorsFormat) VectorsReader(d store.Directory,
	segmentInfo *model.SegmentInfo, fieldsInfos model.FieldInfos,
	context store.IOContext) (spi.TermVectorsReader, error) {

	panic("not implemented yet")
}

func (vf *CompressingTermVectorsFormat) VectorsWriter(d store.Directory,
	segmentInfo *model.SegmentInfo,
	context store.IOContext) (spi.TermVectorsWriter, error) {

	panic("not implemented yet")
}
