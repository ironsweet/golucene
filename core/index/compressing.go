package index

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	"github.com/balzaczyy/golucene/core/codec/compressing"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"github.com/balzaczyy/golucene/core/util/packed"
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

/*
Create a new CompressingStoredFieldsFormat

formatName is the name of the format. This name will be used in the
file formats to perform CheckHeader().

segmentSuffix is the segment suffix. This suffix is added to the result
 file name only if it's not te empty string.

 The compressionMode parameter allows you to choose between compresison
 algorithms that have various compression and decompression speeds so
 that you can pick the one that best fits your indexing and searching
 throughput. You should never instantiate two CoompressingStoredFieldsFormats
 that have the same name but different CompressionModes.

 chunkSize is the minimum byte size of a chunk of documents. A value
 of 1 can make sense if there is redundancy across fields. In that
 case, both performance and compression ratio should be better than
 with Lucene40StoredFieldsFormat with compressed fields.

 Higher values of chunkSize should improve the compresison ratio but
 will require more memery at indexing time and might make document
 loading a little slower (depending on the size of our OS cache compared
 to the size of your index).
*/
func newCompressingStoredFieldsFormat(formatName, segmentSuffix string,
	compressionMode codec.CompressionMode, chunkSize int) *CompressingStoredFieldsFormat {
	assert2(chunkSize >= 1, "chunkSize must be >= 1")
	return &CompressingStoredFieldsFormat{
		formatName:      formatName,
		segmentSuffix:   segmentSuffix,
		compressionMode: compressionMode,
		chunkSize:       chunkSize,
	}
}

func (format *CompressingStoredFieldsFormat) FieldsReader(d store.Directory, si *SegmentInfo,
	fn FieldInfos, ctx store.IOContext) (r StoredFieldsReader, err error) {
	return newCompressingStoredFieldsReader(d, si, format.segmentSuffix, fn,
		ctx, format.formatName, format.compressionMode)
}

func (format *CompressingStoredFieldsFormat) FieldsWriter(d store.Directory, si *SegmentInfo,
	ctx store.IOContext) (w StoredFieldsWriter, err error) {

	return newCompressingStoredFieldsWriter(d, si, format.segmentSuffix, ctx,
		format.formatName, format.compressionMode, format.chunkSize)
}

func (format *CompressingStoredFieldsFormat) String() string {
	return fmt.Sprintf("CompressingStoredFieldsFormat(compressionMode=%v, chunkSize=%v)",
		format.compressionMode, format.chunkSize)
}

// compressing/CompressingTermVectorsFormat.java

// A TermVectorsFormat that compresses chunks of documents together
// in order to improve the compression ratio.
type CompressingTermVectorsFormat struct {
	formatName      string
	segmentSuffix   string
	compressionMode codec.CompressionMode
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
func newCompressingTermVectorsFormat(formatName, segmentSuffix string,
	compressionMode codec.CompressionMode, chunkSize int) *CompressingTermVectorsFormat {
	assert2(chunkSize >= 1, "chunkSize must be >= 1")
	return &CompressingTermVectorsFormat{
		formatName:      formatName,
		segmentSuffix:   segmentSuffix,
		compressionMode: compressionMode,
		chunkSize:       chunkSize,
	}
}

func (vf *CompressingTermVectorsFormat) VectorsReader(d store.Directory,
	segmentInfo *SegmentInfo, fieldsInfos FieldInfos, context store.IOContext) (r TermVectorsReader, err error) {
	panic("not implemented yet")
}

func (vf *CompressingTermVectorsFormat) VectorsWriter(d store.Directory,
	segmentInfo *SegmentInfo, context store.IOContext) (w TermVectorsWriter, err error) {
	panic("not implemented yet")
}

// codecs/compressing/CompressingStoredFieldsWriter.java

const CP_VERSION_BIG_CHUNKS = 1
const CP_VERSION_CURRENT = CP_VERSION_BIG_CHUNKS

/* StoredFieldsWriter impl for CompressingStoredFieldsFormat */
type CompressingStoredFieldsWriter struct {
	directory     store.Directory
	segment       string
	segmentSuffix string
	indexWriter   *compressing.StoredFieldsIndexWriter
	fieldsStream  store.IndexOutput

	compressionMode codec.CompressionMode
	compressor      codec.Compressor
	chunkSize       int

	bufferedDocs    *GrowableByteArrayDataOutput
	numStoredFields []int // number of stored fields
	endOffsets      []int // ned offsets in bufferedDocs
	docBase         int   // doc ID at the beginning of the chunk
	numBufferedDocs int   // docBase + numBufferedDocs == current doc ID
}

func newCompressingStoredFieldsWriter(dir store.Directory, si *SegmentInfo,
	segmentSuffix string, ctx store.IOContext, formatName string,
	compressionMode codec.CompressionMode, chunkSize int) (*CompressingStoredFieldsWriter, error) {

	assert(dir != nil)
	ans := &CompressingStoredFieldsWriter{
		directory:       dir,
		segment:         si.name,
		segmentSuffix:   segmentSuffix,
		compressionMode: compressionMode,
		compressor:      compressionMode.NewCompressor(),
		chunkSize:       chunkSize,
		docBase:         0,
		bufferedDocs:    newGrowableByteArrayDataOutput(chunkSize),
		numStoredFields: make([]int, 16),
		endOffsets:      make([]int, 16),
		numBufferedDocs: 0,
	}

	var success = false
	indexStream, err := dir.CreateOutput(util.SegmentFileName(si.name, segmentSuffix,
		LUCENE40_SF_FIELDS_INDEX_EXTENSION), ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if !success {
			util.CloseWhileSuppressingError(indexStream)
			ans.abort()
		}
	}()

	ans.fieldsStream, err = dir.CreateOutput(util.SegmentFileName(si.name, segmentSuffix,
		LUCENE40_SF_FIELDS_EXTENSION), ctx)
	if err != nil {
		return nil, err
	}

	codecNameIdx := formatName + CODEC_SFX_IDX
	codecNameDat := formatName + CODEC_SFX_DAT
	err = codec.WriteHeader(indexStream, codecNameIdx, CP_VERSION_CURRENT)
	if err != nil {
		return nil, err
	}
	err = codec.WriteHeader(ans.fieldsStream, codecNameDat, CP_VERSION_CURRENT)
	if err != nil {
		return nil, err
	}
	assert(int64(codec.HeaderLength(codecNameIdx)) == indexStream.FilePointer())
	assert(int64(codec.HeaderLength(codecNameDat)) == ans.fieldsStream.FilePointer())

	ans.indexWriter, err = compressing.NewStoredFieldsIndexWriter(indexStream)
	if err != nil {
		return nil, err
	}
	indexStream = nil

	err = ans.fieldsStream.WriteVInt(int32(chunkSize))
	if err != nil {
		return nil, err
	}
	err = ans.fieldsStream.WriteVInt(packed.PACKED_VERSION_CURRENT)
	if err != nil {
		return nil, err
	}

	success = true
	return ans, nil
}

func (w *CompressingStoredFieldsWriter) Close() error {
	defer func() {
		w.fieldsStream = nil
		w.indexWriter = nil
	}()
	return util.Close(w.fieldsStream, w.indexWriter)
}

func (w *CompressingStoredFieldsWriter) startDocument(numStoredFields int) error {
	panic("not implemented yet")
}

func (w *CompressingStoredFieldsWriter) finishDocument() error {
	panic("not implemented yet")
}

func (w *CompressingStoredFieldsWriter) abort() {
	// util.CloseWhileSuppressingError(w)
	// util.DeleteFilesIgnoringErrors(w.directory,
	// 	SegmentFileName(w.segment, w.segmentSuffix, FIELDS_EXTENSION),
	// 	SegmentFileName(w.segment, w.segmentSuffix, FIELDS_INDEX_EXTENSION))
	panic("not implemented yet")
}

func (w *CompressingStoredFieldsWriter) finish(fis FieldInfos, numDocs int) error {
	panic("not implemented yet")
}

// util/GrowableByteArrayDataOutput.java

/* A DataOutput that can be used to build a []byte */
type GrowableByteArrayDataOutput struct {
	bytes []byte
}

func newGrowableByteArrayDataOutput(cp int) *GrowableByteArrayDataOutput {
	return &GrowableByteArrayDataOutput{make([]byte, 0, util.Oversize(cp, 1))}
}
