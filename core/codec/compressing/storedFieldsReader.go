package compressing

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	"github.com/balzaczyy/golucene/core/codec/lucene40"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"github.com/balzaczyy/golucene/core/util/packed"
)

// codec/compressing/CompressingStoredFieldsReader.java

// Do not reuse the decompression buffer when there is more than 32kb to decompress
const BUFFER_REUSE_THRESHOLD = 1 << 15

// StoredFieldsReader impl for CompressingStoredFieldsFormat
type CompressingStoredFieldsReader struct {
	version           int
	fieldInfos        model.FieldInfos
	indexReader       *CompressingStoredFieldsIndexReader
	maxPointer        int64
	fieldsStream      store.IndexInput
	chunkSize         int
	packedIntsVersion int
	compressionMode   CompressionMode
	decompressor      Decompressor
	bytes             []byte
	numDocs           int
	closed            bool
}

// used by clone
func newCompressingStoredFieldsReaderFrom(reader *CompressingStoredFieldsReader) *CompressingStoredFieldsReader {
	return &CompressingStoredFieldsReader{
		version:           reader.version,
		fieldInfos:        reader.fieldInfos,
		fieldsStream:      reader.fieldsStream.Clone(),
		indexReader:       reader.indexReader.Clone(),
		maxPointer:        reader.maxPointer,
		chunkSize:         reader.chunkSize,
		packedIntsVersion: reader.packedIntsVersion,
		compressionMode:   reader.compressionMode,
		decompressor:      reader.compressionMode.NewDecompressor(),
		numDocs:           reader.numDocs,
		bytes:             make([]byte, len(reader.bytes)),
		closed:            false,
	}
}

// Sole constructor
func newCompressingStoredFieldsReader(d store.Directory,
	si *model.SegmentInfo, segmentSuffix string,
	fn model.FieldInfos, ctx store.IOContext, formatName string,
	compressionMode CompressionMode) (r *CompressingStoredFieldsReader, err error) {

	r = &CompressingStoredFieldsReader{}
	r.compressionMode = compressionMode
	segment := si.Name
	r.fieldInfos = fn
	r.numDocs = si.DocCount()

	var indexStream store.ChecksumIndexInput
	success := false
	defer func() {
		if !success {
			util.CloseWhileSuppressingError(r, indexStream)
		}
	}()

	indexStreamFN := util.SegmentFileName(segment, segmentSuffix, lucene40.FIELDS_INDEX_EXTENSION)
	fieldsStreamFN := util.SegmentFileName(segment, segmentSuffix, lucene40.FIELDS_EXTENSION)
	// Load the index into memory
	if indexStream, err = d.OpenChecksumInput(indexStreamFN, ctx); err != nil {
		return nil, err
	}
	codecNameIdx := formatName + CODEC_SFX_IDX
	if r.version, err = int32AsInt(codec.CheckHeader(indexStream, codecNameIdx,
		VERSION_START, VERSION_CURRENT)); err != nil {
		return nil, err
	}
	assert(int64(codec.HeaderLength(codecNameIdx)) == indexStream.FilePointer())
	if r.indexReader, err = newCompressingStoredFieldsIndexReader(indexStream, si); err != nil {
		return nil, err
	}

	var maxPointer int64 = -1

	if r.version >= VERSION_CHECKSUM {
		if maxPointer, err = indexStream.ReadVLong(); err != nil {
			return nil, err
		}
		if _, err = codec.CheckFooter(indexStream); err != nil {
			return nil, err
		}
	} else {
		if err = codec.CheckEOF(indexStream); err != nil {
			return nil, err
		}
	}

	if err = indexStream.Close(); err != nil {
		return nil, err
	}
	indexStream = nil

	// Open the data file and read metadata
	if r.fieldsStream, err = d.OpenInput(fieldsStreamFN, ctx); err != nil {
		return nil, err
	}
	if r.version >= VERSION_CHECKSUM {
		if maxPointer+codec.FOOTER_LENGTH != r.fieldsStream.Length() {
			return nil, errors.New(fmt.Sprintf(
				"Invalid fieldsStream maxPointer (file truncated?): maxPointer=%v, length=%v",
				maxPointer, r.fieldsStream.Length()))
		}
	} else {
		maxPointer = r.fieldsStream.Length()
	}
	r.maxPointer = maxPointer
	codecNameDat := formatName + CODEC_SFX_DAT
	var fieldsVersion int
	if fieldsVersion, err = int32AsInt(codec.CheckHeader(r.fieldsStream,
		codecNameDat, VERSION_START, VERSION_CURRENT)); err != nil {
		return nil, err
	}
	assert2(r.version == fieldsVersion,
		"Version mismatch between stored fields index and data: %v != %v",
		r.version, fieldsVersion)
	assert(int64(codec.HeaderLength(codecNameDat)) == r.fieldsStream.FilePointer())

	r.chunkSize = -1
	if r.version >= VERSION_BIG_CHUNKS {
		if r.chunkSize, err = int32AsInt(r.fieldsStream.ReadVInt()); err != nil {
			return nil, err
		}
	}

	if r.packedIntsVersion, err = int32AsInt(r.fieldsStream.ReadVInt()); err != nil {
		return nil, err
	}
	r.decompressor = compressionMode.NewDecompressor()
	r.bytes = make([]byte, 0)

	if r.version >= VERSION_CHECKSUM {
		// NOTE: data file is too costly to verify checksum against all the
		// bytes on open, but fo rnow we at least verify proper structure
		// of the checksum footer: which looks for FOOTER_MATIC +
		// algorithmID. This is cheap and can detect some forms of
		// corruption such as file trucation.
		if _, err = codec.RetrieveChecksum(r.fieldsStream); err != nil {
			return nil, err
		}
	}

	success = true
	return r, nil
}

func int32AsInt(n int32, err error) (int, error) {
	return int(n), err
}

func (r *CompressingStoredFieldsReader) ensureOpen() {
	assert2(!r.closed, "this FieldsReader is closed")
}

// Close the underlying IndexInputs
func (r *CompressingStoredFieldsReader) Close() (err error) {
	if !r.closed {
		if err = util.Close(r.fieldsStream); err == nil {
			r.closed = true
		}
	}
	return
}

func (r *CompressingStoredFieldsReader) readField(in util.DataInput,
	visitor StoredFieldVisitor, info *model.FieldInfo, bits int) (err error) {
	switch bits & TYPE_MASK {
	case BYTE_ARR:
		panic("not implemented yet")
	case STRING:
		var length int
		if length, err = int32AsInt(in.ReadVInt()); err != nil {
			return err
		}
		data := make([]byte, length)
		if err = in.ReadBytes(data); err != nil {
			return err
		}
		visitor.StringField(info, string(data))
	case NUMERIC_INT:
		panic("not implemented yet")
	case NUMERIC_FLOAT:
		panic("not implemented yet")
	case NUMERIC_LONG:
		panic("not implemented yet")
	case NUMERIC_DOUBLE:
		panic("not implemented yet")
	default:
		panic(fmt.Sprintf("Unknown type flag: %x", bits))
	}
	return nil
}

func (r *CompressingStoredFieldsReader) VisitDocument(docID int, visitor StoredFieldVisitor) error {
	err := r.fieldsStream.Seek(r.indexReader.startPointer(docID))
	if err != nil {
		return err
	}

	docBase, err := int32AsInt(r.fieldsStream.ReadVInt())
	if err != nil {
		return err
	}
	chunkDocs, err := int32AsInt(r.fieldsStream.ReadVInt())
	if err != nil {
		return err
	}
	if docID < docBase ||
		docID >= docBase+chunkDocs ||
		docBase+chunkDocs > r.numDocs {
		return errors.New(fmt.Sprintf(
			"Corrupted: docID=%v, docBase=%v, chunkDocs=%v, numDocs=%v (resource=%v)",
			docID, docBase, chunkDocs, r.numDocs, r.fieldsStream))
	}

	var numStoredFields, offset, length, totalLength int
	if chunkDocs == 1 {
		if numStoredFields, err = int32AsInt(r.fieldsStream.ReadVInt()); err != nil {
			return err
		}
		offset = 0
		if length, err = int32AsInt(r.fieldsStream.ReadVInt()); err != nil {
			return err
		}
		totalLength = length
	} else {
		bitsPerStoredFields, err := int32AsInt(r.fieldsStream.ReadVInt())
		if err != nil {
			return err
		}
		if bitsPerStoredFields == 0 {
			numStoredFields, err = int32AsInt(r.fieldsStream.ReadVInt())
			if err != nil {
				return err
			}
		} else if bitsPerStoredFields > 31 {
			return errors.New(fmt.Sprintf("bitsPerStoredFields=%v (resource=%v)",
				bitsPerStoredFields, r.fieldsStream))
		} else {
			panic("not implemented yet")
		}

		bitsPerLength, err := int32AsInt(r.fieldsStream.ReadVInt())
		if err != nil {
			return err
		}
		if bitsPerLength == 0 {
			if length, err = int32AsInt(r.fieldsStream.ReadVInt()); err != nil {
				return err
			}
			offset = (docID - docBase) * length
			totalLength = chunkDocs * length
		} else if bitsPerLength > 31 {
			return errors.New(fmt.Sprintf("bitsPerLength=%v (resource=%v)",
				bitsPerLength, r.fieldsStream))
		} else {
			it := packed.ReaderIteratorNoHeader(
				r.fieldsStream, packed.PackedFormat(packed.PACKED), r.packedIntsVersion,
				chunkDocs, bitsPerLength, 1)
			var n int64
			off := 0
			for i := 0; i < docID-docBase; i++ {
				if n, err = it.Next(); err != nil {
					return err
				}
				off += int(n)
			}
			offset = off
			if n, err = it.Next(); err != nil {
				return err
			}
			length = int(n)
			off += length
			for i := docID - docBase + 1; i < chunkDocs; i++ {
				if n, err = it.Next(); err != nil {
					return err
				}
				off += int(n)
			}
			totalLength = off
		}
	}

	if (length == 0) != (numStoredFields == 0) {
		return errors.New(fmt.Sprintf(
			"length=%v, numStoredFields=%v (resource=%v)",
			length, numStoredFields, r.fieldsStream))
	}
	if numStoredFields == 0 {
		// nothing to do
		return nil
	}

	var documentInput util.DataInput
	if r.version >= VERSION_BIG_CHUNKS && totalLength >= 2*r.chunkSize {
		panic("not implemented yet")
	} else {
		var bytes []byte
		if totalLength <= BUFFER_REUSE_THRESHOLD {
			bytes = r.bytes
		} else {
			bytes = make([]byte, 0)
		}
		bytes, err = r.decompressor(r.fieldsStream, totalLength, offset, length, bytes)
		if err != nil {
			return err
		}
		assert(len(bytes) == length)
		documentInput = store.NewByteArrayDataInput(bytes)
	}

	for fieldIDX := 0; fieldIDX < numStoredFields; fieldIDX++ {
		infoAndBits, err := documentInput.ReadVLong()
		if err != nil {
			return err
		}
		fieldNumber := int(uint64(infoAndBits) >> uint64(TYPE_BITS))
		fieldInfo := r.fieldInfos.FieldInfoByNumber(fieldNumber)

		bits := int(infoAndBits & int64(TYPE_MASK))
		assertWithMessage(bits <= NUMERIC_DOUBLE, fmt.Sprintf("bits=%x", bits))

		status, err := visitor.NeedsField(fieldInfo)
		if err != nil {
			return err
		}
		switch status {
		case STORED_FIELD_VISITOR_STATUS_YES:
			r.readField(documentInput, visitor, fieldInfo, bits)
		case STORED_FIELD_VISITOR_STATUS_NO:
			panic("not implemented yet")
		case STORED_FIELD_VISITOR_STATUS_STOP:
			return nil
		}
	}

	return nil
}

func assertWithMessage(ok bool, msg string) {
	if !ok {
		panic(msg)
	}
}

func (r *CompressingStoredFieldsReader) Clone() StoredFieldsReader {
	r.ensureOpen()
	return newCompressingStoredFieldsReaderFrom(r)
}
