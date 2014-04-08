package compressing

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	"github.com/balzaczyy/golucene/core/codec/lucene40"
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"github.com/balzaczyy/golucene/core/util/packed"
)

/* hard limit on the maximum number of documents per chunk */
const MAX_DOCUMENTS_PER_CHUNK = 128

const CODEC_SFX_IDX = "Index"
const CODEC_SFX_DAT = "Data"
const CP_VERSION_BIG_CHUNKS = 1
const CP_VERSION_CURRENT = CP_VERSION_BIG_CHUNKS

/* StoredFieldsWriter impl for CompressingStoredFieldsFormat */
type CompressingStoredFieldsWriter struct {
	directory     store.Directory
	segment       string
	segmentSuffix string
	indexWriter   *StoredFieldsIndexWriter
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

func NewCompressingStoredFieldsWriter(dir store.Directory, si *model.SegmentInfo,
	segmentSuffix string, ctx store.IOContext, formatName string,
	compressionMode codec.CompressionMode, chunkSize int) (*CompressingStoredFieldsWriter, error) {

	assert(dir != nil)
	ans := &CompressingStoredFieldsWriter{
		directory:       dir,
		segment:         si.Name,
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
	indexStream, err := dir.CreateOutput(util.SegmentFileName(si.Name, segmentSuffix,
		lucene40.FIELDS_INDEX_EXTENSION), ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if !success {
			util.CloseWhileSuppressingError(indexStream)
			ans.Abort()
		}
	}()

	ans.fieldsStream, err = dir.CreateOutput(util.SegmentFileName(si.Name, segmentSuffix,
		lucene40.FIELDS_EXTENSION), ctx)
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

	ans.indexWriter, err = NewStoredFieldsIndexWriter(indexStream)
	if err != nil {
		return nil, err
	}
	indexStream = nil

	err = ans.fieldsStream.WriteVInt(int32(chunkSize))
	if err != nil {
		return nil, err
	}
	err = ans.fieldsStream.WriteVInt(packed.VERSION_CURRENT)
	if err != nil {
		return nil, err
	}

	success = true
	return ans, nil
}

func assert(ok bool) {
	assert2(ok, "assert fail")
}

func assert2(ok bool, msg string, args ...interface{}) {
	if !ok {
		panic(fmt.Sprintf(msg, args...))
	}
}

func (w *CompressingStoredFieldsWriter) Close() error {
	defer func() {
		w.fieldsStream = nil
		w.indexWriter = nil
	}()
	return util.Close(w.fieldsStream, w.indexWriter)
}

func (w *CompressingStoredFieldsWriter) StartDocument(numStoredFields int) error {
	if w.numBufferedDocs == len(w.numStoredFields) {
		newLength := util.Oversize(w.numBufferedDocs+1, 4)
		oldArray := w.endOffsets
		w.numStoredFields = make([]int, newLength)
		w.endOffsets = make([]int, newLength)
		copy(w.numStoredFields, oldArray)
		copy(w.endOffsets, oldArray)
	}
	w.numStoredFields[w.numBufferedDocs] = numStoredFields
	w.numBufferedDocs++
	return nil
}

func (w *CompressingStoredFieldsWriter) FinishDocument() error {
	w.endOffsets[w.numBufferedDocs-1] = w.bufferedDocs.length
	if w.triggerFlush() {
		return w.flush()
	}
	return nil
}

func (w *CompressingStoredFieldsWriter) triggerFlush() bool {
	return w.bufferedDocs.length >= w.chunkSize || // chunks of at least chunkSize bytes
		w.numBufferedDocs >= MAX_DOCUMENTS_PER_CHUNK
}

func (w *CompressingStoredFieldsWriter) flush() error {
	panic("not implemented yet")
}

func (w *CompressingStoredFieldsWriter) Abort() {
	assert(w != nil)
	util.CloseWhileSuppressingError(w)
	util.DeleteFilesIgnoringErrors(w.directory,
		util.SegmentFileName(w.segment, w.segmentSuffix, lucene40.FIELDS_EXTENSION),
		util.SegmentFileName(w.segment, w.segmentSuffix, lucene40.FIELDS_INDEX_EXTENSION))
}

func (w *CompressingStoredFieldsWriter) Finish(fis model.FieldInfos, numDocs int) error {
	if w.numBufferedDocs > 0 {
		err := w.flush()
		if err != nil {
			return err
		}
	} else {
		assert(w.bufferedDocs.length == 0)
	}
	assert2(w.docBase == numDocs,
		"Wrote %v docs, finish called with numDocs=%v", w.docBase, numDocs)
	err := w.indexWriter.finish(numDocs)
	if err != nil {
		return err
	}
	assert(w.bufferedDocs.length == 0)
	return nil
}

// util/GrowableByteArrayDataOutput.java

/* A DataOutput that can be used to build a []byte */
type GrowableByteArrayDataOutput struct {
	bytes  []byte
	length int
}

func newGrowableByteArrayDataOutput(cp int) *GrowableByteArrayDataOutput {
	return &GrowableByteArrayDataOutput{make([]byte, 0, util.Oversize(cp, 1)), 0}
}
