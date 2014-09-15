package compressing

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	"github.com/balzaczyy/golucene/core/codec/lucene40"
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"github.com/balzaczyy/golucene/core/util/packed"
	"math"
)

/* hard limit on the maximum number of documents per chunk */
const MAX_DOCUMENTS_PER_CHUNK = 128

const (
	STRING         = 0x00
	BYTE_ARR       = 0x01
	NUMERIC_INT    = 0x02
	NUMERIC_FLOAT  = 0x03
	NUMERIC_LONG   = 0x04
	NUMERIC_DOUBLE = 0x05
)

var (
	TYPE_BITS = packed.BitsRequired(NUMERIC_DOUBLE)
	TYPE_MASK = int(packed.MaxValue(TYPE_BITS))
)

const (
	CODEC_SFX_IDX      = "Index"
	CODEC_SFX_DAT      = "Data"
	VERSION_START      = 0
	VERSION_BIG_CHUNKS = 1
	VERSION_CHECKSUM   = 2
	VERSION_CURRENT    = VERSION_CHECKSUM
)

/* StoredFieldsWriter impl for CompressingStoredFieldsFormat */
type CompressingStoredFieldsWriter struct {
	directory     store.Directory
	segment       string
	segmentSuffix string
	indexWriter   *StoredFieldsIndexWriter
	fieldsStream  store.IndexOutput

	compressionMode CompressionMode
	compressor      Compressor
	chunkSize       int

	bufferedDocs    *GrowableByteArrayDataOutput
	numStoredFields []int // number of stored fields
	endOffsets      []int // ned offsets in bufferedDocs
	docBase         int   // doc ID at the beginning of the chunk
	numBufferedDocs int   // docBase + numBufferedDocs == current doc ID

	numStoredFieldsInDoc int
}

func NewCompressingStoredFieldsWriter(dir store.Directory, si *model.SegmentInfo,
	segmentSuffix string, ctx store.IOContext, formatName string,
	compressionMode CompressionMode, chunkSize int) (*CompressingStoredFieldsWriter, error) {

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
	assert(indexStream != nil)
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
	err = codec.WriteHeader(indexStream, codecNameIdx, VERSION_CURRENT)
	if err != nil {
		return nil, err
	}
	err = codec.WriteHeader(ans.fieldsStream, codecNameDat, VERSION_CURRENT)
	if err != nil {
		return nil, err
	}
	assert(int64(codec.HeaderLength(codecNameIdx)) == indexStream.FilePointer())
	assert(int64(codec.HeaderLength(codecNameDat)) == ans.fieldsStream.FilePointer())

	ans.indexWriter, err = NewStoredFieldsIndexWriter(indexStream)
	if err != nil {
		return nil, err
	}
	assert(ans.indexWriter != nil)
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
	assert(w != nil)
	defer func() {
		if w != nil {
			w.fieldsStream = nil
			w.indexWriter = nil
		}
	}()
	return util.Close(w.fieldsStream, w.indexWriter)
}

func (w *CompressingStoredFieldsWriter) StartDocument() error { return nil }

func (w *CompressingStoredFieldsWriter) FinishDocument() error {
	if w.numBufferedDocs == len(w.numStoredFields) {
		newLength := util.Oversize(w.numBufferedDocs+1, 4)

		oldArray := w.endOffsets
		w.endOffsets = make([]int, newLength)
		copy(w.endOffsets, oldArray)

		oldArray = w.numStoredFields
		w.numStoredFields = make([]int, newLength)
		copy(w.numStoredFields, oldArray)
	}
	w.numStoredFields[w.numBufferedDocs] = w.numStoredFieldsInDoc
	w.numStoredFieldsInDoc = 0
	w.endOffsets[w.numBufferedDocs] = w.bufferedDocs.length
	w.numBufferedDocs++
	if w.triggerFlush() {
		return w.flush()
	}
	return nil
}

func saveInts(values []int, out DataOutput) error {
	length := len(values)
	assert(length > 0)
	if length == 1 {
		return out.WriteVInt(int32(values[0]))
	}

	var allEqual = true
	var sentinel = values[0]
	for _, v := range values[1:] {
		if v != sentinel {
			allEqual = false
			break
		}
	}
	if allEqual {
		err := out.WriteVInt(0)
		if err == nil {
			err = out.WriteVInt(int32(values[0]))
		}
		return err
	}

	var max int64 = 0
	for _, v := range values {
		max |= int64(v)
	}
	var bitsRequired = packed.BitsRequired(max)
	err := out.WriteVInt(int32(bitsRequired))
	if err != nil {
		return err
	}

	w := packed.WriterNoHeader(out, packed.PackedFormat(packed.PACKED), length, bitsRequired, 1)
	for _, v := range values {
		if err = w.Add(int64(v)); err != nil {
			return err
		}
	}
	return w.Finish()
}

func (w *CompressingStoredFieldsWriter) writeHeader(docBase,
	numBufferedDocs int, numStoredFields, lengths []int) error {

	// save docBase and numBufferedDocs
	err := w.fieldsStream.WriteVInt(int32(docBase)) // TODO precision loss risk
	if err == nil {
		err = w.fieldsStream.WriteVInt(int32(numBufferedDocs)) // TODO precision loss risk
		if err == nil {
			// save numStoredFields
			err = saveInts(numStoredFields[:numBufferedDocs], w.fieldsStream)
			if err == nil {
				// save lengths
				err = saveInts(lengths[:numBufferedDocs], w.fieldsStream)
			}
		}
	}
	return err
}

func (w *CompressingStoredFieldsWriter) triggerFlush() bool {
	return w.bufferedDocs.length >= w.chunkSize || // chunks of at least chunkSize bytes
		w.numBufferedDocs >= MAX_DOCUMENTS_PER_CHUNK
}

func (w *CompressingStoredFieldsWriter) flush() error {
	err := w.indexWriter.writeIndex(w.numBufferedDocs, w.fieldsStream.FilePointer())
	if err != nil {
		return err
	}

	// transform end offsets into lengths
	lengths := w.endOffsets
	for i := w.numBufferedDocs - 1; i > 0; i-- {
		lengths[i] = w.endOffsets[i] - w.endOffsets[i-1]
		assert(lengths[i] >= 0)
	}
	err = w.writeHeader(w.docBase, w.numBufferedDocs, w.numStoredFields, lengths)
	if err != nil {
		return err
	}

	// compress stored fields to fieldsStream
	if w.bufferedDocs.length >= 2*w.chunkSize {
		// big chunk, slice it
		for compressed := 0; compressed < w.bufferedDocs.length; compressed += w.chunkSize {
			size := w.bufferedDocs.length - compressed
			if w.chunkSize < size {
				size = w.chunkSize
			}
			err = w.compressor(w.bufferedDocs.bytes[compressed:compressed+size], w.fieldsStream)
			if err != nil {
				return err
			}
		}
	} else {
		err = w.compressor(w.bufferedDocs.bytes[:w.bufferedDocs.length], w.fieldsStream)
		if err != nil {
			return err
		}
	}

	// reset
	w.docBase += w.numBufferedDocs
	w.numBufferedDocs = 0
	w.bufferedDocs.length = 0
	return nil
}

func (w *CompressingStoredFieldsWriter) WriteField(info *model.FieldInfo, field model.IndexableField) error {
	w.numStoredFieldsInDoc++

	bits := 0
	var bytes []byte
	var str string

	number := field.NumericValue()
	if number != nil {
		switch t := number.(type) {
		case int32:
			bits = NUMERIC_INT
		case int64:
			bits = NUMERIC_LONG
		case float32:
			bits = NUMERIC_FLOAT
		case float64:
			bits = NUMERIC_DOUBLE
		default:
			panic(fmt.Sprintf("cannot store numeric value %v of type %v", number, t))
		}
	} else {
		bytes = field.BinaryValue()
		if bytes != nil {
			bits = BYTE_ARR
		} else {
			bits = STRING
			str = field.StringValue()
			assert2(str != "",
				"field %v is stored but does not have binaryValue, stringValue nor numericValue",
				field.Name())
		}
	}

	infoAndBits := (int64(info.Number) << uint(TYPE_BITS)) | int64(bits)
	err := w.bufferedDocs.WriteVLong(infoAndBits)
	if err != nil {
		return err
	}

	switch {
	case bytes != nil:
		err = w.bufferedDocs.WriteVInt(int32(len(bytes)))
		if err == nil {
			err = w.bufferedDocs.WriteBytes(bytes)
		}
	case str != "":
		err = w.bufferedDocs.WriteString(str)
	case bits == NUMERIC_INT:
		err = w.bufferedDocs.WriteInt(number.(int32))
	case bits == NUMERIC_LONG:
		err = w.bufferedDocs.WriteLong(number.(int64))
	case bits == NUMERIC_FLOAT:
		err = w.bufferedDocs.WriteInt(int32(math.Float32bits(number.(float32))))
	case bits == NUMERIC_DOUBLE:
		err = w.bufferedDocs.WriteLong(int64(math.Float64bits(number.(float64))))
	default:
		panic("Cannot get here")
	}
	return err
}

func (w *CompressingStoredFieldsWriter) Abort() {
	if w == nil { // tolerate early released pointer
		return
	}
	util.CloseWhileSuppressingError(w)
	util.DeleteFilesIgnoringErrors(w.directory,
		util.SegmentFileName(w.segment, w.segmentSuffix, lucene40.FIELDS_EXTENSION),
		util.SegmentFileName(w.segment, w.segmentSuffix, lucene40.FIELDS_INDEX_EXTENSION))
}

func (w *CompressingStoredFieldsWriter) Finish(fis model.FieldInfos, numDocs int) (err error) {
	if w == nil {
		return errors.New("Nil class pointer encountered.")
	}
	assert2(w.indexWriter != nil, "already closed?")
	if w.numBufferedDocs > 0 {
		if err = w.flush(); err != nil {
			return err
		}
	} else {
		assert(w.bufferedDocs.length == 0)
	}
	assert2(w.docBase == numDocs,
		"Wrote %v docs, finish called with numDocs=%v", w.docBase, numDocs)
	if err = w.indexWriter.finish(numDocs, w.fieldsStream.FilePointer()); err != nil {
		return err
	}
	if err = codec.WriteFooter(w.fieldsStream); err != nil {
		return err
	}
	assert(w.bufferedDocs.length == 0)
	return nil
}

// util/GrowableByteArrayDataOutput.java

/* A DataOutput that can be used to build a []byte */
type GrowableByteArrayDataOutput struct {
	*util.DataOutputImpl
	bytes  []byte
	length int
}

func newGrowableByteArrayDataOutput(cp int) *GrowableByteArrayDataOutput {
	ans := &GrowableByteArrayDataOutput{bytes: make([]byte, 0, util.Oversize(cp, 1))}
	ans.DataOutputImpl = util.NewDataOutput(ans)
	return ans
}

func (out *GrowableByteArrayDataOutput) WriteByte(b byte) error {
	assert(out.length <= len(out.bytes))
	if out.length < len(out.bytes) {
		out.bytes[out.length] = b
	} else {
		out.bytes = append(out.bytes, b)
	}
	out.length++
	return nil
}

func (out *GrowableByteArrayDataOutput) WriteBytes(b []byte) error {
	assert(out.length <= len(out.bytes))
	remaining := len(out.bytes) - out.length
	if remaining > len(b) {
		copy(out.bytes[out.length:], b)
	} else if remaining == 0 {
		out.bytes = append(out.bytes, b...)
	} else {
		copy(out.bytes[out.length:], b[:remaining])
		out.bytes = append(out.bytes, b[remaining:]...)
	}
	out.length += len(b)
	return nil
}
