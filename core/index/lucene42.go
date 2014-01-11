package index

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"log"
)

// lucene42/Lucene42Codec.java

/*
Implements the Lucene 4.2 index format, with configurable per-field
postings and docvalues formats.

If you want to reuse functionality of this codec, in another codec,
extend FilterCodec.
*/
var Lucene42Codec = &CodecImpl{
	fieldsFormat:  newLucene41StoredFieldsFormat(),
	vectorsFormat: newLucene42TermVectorsFormat(),
	// fieldInfosFormat: newLucene42FieldInfosFormat(),
	infosFormat: newLucene40SegmentInfoFormat(),
	// liveDocsFormat: newLucene40LiveDocsFormat(),
	// Returns the postings format that should be used for writing new
	// segments of field.
	//
	// The default implemnetation always returns "Lucene41"
	postingsFormat: newPerFieldPostingsFormat(func(field string) PostingsFormat {
		panic("not implemented yet")
	}),
	// Returns the decvalues format that should be used for writing new
	// segments of field.
	//
	// The default implementation always returns "Lucene42"
	docValuesFormat: newPerFieldDocValuesFormat(func(field string) DocValuesFormat {
		panic("not implemented yet")
	}),
	// return Codec{ReadSegmentInfo: Lucene40SegmentInfoReader,
	// 	ReadFieldInfos: Lucene42FieldInfosReader,
	// 	GetFieldsProducer: func(readState SegmentReadState) (fp FieldsProducer, err error) {
	// 		return newPerFieldPostingsReader(readState)
	// 	},
	// 	GetDocValuesProducer: func(s SegmentReadState) (dvp DocValuesProducer, err error) {
	// 		return newPerFieldDocValuesReader(s)
	// 	},
	// 	GetNormsDocValuesProducer: func(s SegmentReadState) (dvp DocValuesProducer, err error) {
	// 		return newLucene42DocValuesProducer(s, "Lucene41NormsData", "nvd", "Lucene41NormsMetadata", "nvm")
	// 	},
	// 	GetStoredFieldsReader: func(d store.Directory, si SegmentInfo, fn FieldInfos, ctx store.IOContext) (r StoredFieldsReader, err error) {
	// 		return newLucene41StoredFieldsReader(d, si, fn, ctx)
	// 	},
	// 	GetTermVectorsReader: func(d store.Directory, si SegmentInfo, fn FieldInfos, ctx store.IOContext) (r TermVectorsReader, err error) {
	// 		return newLucene42TermVectorsReader(d, si, fn, ctx)
	// 	},
	// }
}

const (
	// Extension of field infos
	LUCENE42_FI_EXTENSION = "fnm"

	// Codec header
	LUCENE42_FI_CODEC_NAME     = "Lucene42FieldInfos"
	LUCENE42_FI_FORMAT_START   = 0
	LUCENE42_FI_FORMAT_CURRENT = LUCENE42_FI_FORMAT_START

	// Field flags
	LUCENE42_FI_IS_INDEXED                   = 0x1
	LUCENE42_FI_STORE_TERMVECTOR             = 0x2
	LUCENE42_FI_STORE_OFFSETS_IN_POSTINGS    = 0x4
	LUCENE42_FI_OMIT_NORMS                   = 0x10
	LUCENE42_FI_STORE_PAYLOADS               = 0x20
	LUCENE42_FI_OMIT_TERM_FREQ_AND_POSITIONS = 0x40
	LUCENE42_FI_OMIT_POSITIONS               = 0x80
)

var Lucene42FieldInfosReader = func(dir store.Directory, segment string, context store.IOContext) (fi FieldInfos, err error) {
	log.Printf("Reading FieldInfos from %v...", dir)
	fi = FieldInfos{}
	fileName := util.SegmentFileName(segment, "", LUCENE42_FI_EXTENSION)
	log.Printf("Segment: %v", fileName)
	input, err := dir.OpenInput(fileName, context)
	if err != nil {
		return fi, err
	}
	log.Printf("Reading %v", input)

	success := false
	defer func() {
		if success {
			input.Close()
		} else {
			util.CloseWhileHandlingError(err, input)
		}
	}()

	_, err = codec.CheckHeader(input,
		LUCENE42_FI_CODEC_NAME,
		LUCENE42_FI_FORMAT_START,
		LUCENE42_FI_FORMAT_CURRENT)
	if err != nil {
		return fi, err
	}

	size, err := input.ReadVInt() //read in the size
	if err != nil {
		return fi, err
	}
	log.Printf("Found %v FieldInfos.", size)

	infos := make([]FieldInfo, size)
	for i, _ := range infos {
		name, err := input.ReadString()
		if err != nil {
			return fi, err
		}
		fieldNumber, err := input.ReadVInt()
		if err != nil {
			return fi, err
		}
		bits, err := input.ReadByte()
		if err != nil {
			return fi, err
		}
		isIndexed := (bits & LUCENE42_FI_IS_INDEXED) != 0
		storeTermVector := (bits & LUCENE42_FI_STORE_TERMVECTOR) != 0
		omitNorms := (bits & LUCENE42_FI_OMIT_NORMS) != 0
		storePayloads := (bits & LUCENE42_FI_STORE_PAYLOADS) != 0
		var indexOptions IndexOptions
		switch {
		case !isIndexed:
			indexOptions = IndexOptions(0)
		case (bits & LUCENE42_FI_OMIT_TERM_FREQ_AND_POSITIONS) != 0:
			indexOptions = INDEX_OPT_DOCS_ONLY
		case (bits & LUCENE42_FI_OMIT_POSITIONS) != 0:
			indexOptions = INDEX_OPT_DOCS_AND_FREQS
		case (bits & LUCENE42_FI_STORE_OFFSETS_IN_POSTINGS) != 0:
			indexOptions = INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
		default:
			indexOptions = INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS
		}

		// DV Types are packed in one byte
		val, err := input.ReadByte()
		if err != nil {
			return fi, err
		}
		docValuesType, err := getDocValuesType(input, (byte)(val&0x0F))
		if err != nil {
			return fi, err
		}
		normsType, err := getDocValuesType(input, (byte)((uint8(val)>>4)&0x0F))
		if err != nil {
			return fi, err
		}
		attributes, err := input.ReadStringStringMap()
		if err != nil {
			return fi, err
		}
		infos[i] = NewFieldInfo(name, isIndexed, fieldNumber, storeTermVector,
			omitNorms, storePayloads, indexOptions, docValuesType, normsType, attributes)
	}

	if input.FilePointer() != input.Length() {
		return fi, errors.New(fmt.Sprintf(
			"did not read all bytes from file '%v': read %v vs size %v (resource: %v)",
			fileName, input.FilePointer(), input.Length(), input))
	}
	fi = NewFieldInfos(infos)
	success = true
	return fi, nil
}

func getDocValuesType(input store.IndexInput, b byte) (t DocValuesType, err error) {
	switch b {
	case 0:
		return DocValuesType(0), nil
	case 1:
		return DOC_VALUES_TYPE_NUMERIC, nil
	case 2:
		return DOC_VALUES_TYPE_BINARY, nil
	case 3:
		return DOC_VALUES_TYPE_SORTED, nil
	case 4:
		return DOC_VALUES_TYPE_SORTED_SET, nil
	default:
		return DocValuesType(0), errors.New(
			fmt.Sprintf("invalid docvalues byte: %v (resource=%v)", b, input))
	}
}

// lucene42/Lucene42TermVectorsFormat.java

/*
Lucene 4.2 term vectors format.

Very similarly to Lucene41StoredFieldsFormat, this format is based on
compressed chunks of data, with document-level granularity so that a
document can never span across distinct chunks. Moreover, data is
made as compact as possible:

- textual data is compressedusing the very light LZ4 compression
algorithm,
- binary data is written using fixed-size blocks of packed ints.

Term vectors are stored using two files

- a data file where terms, frequencies, positions, offsets and
payloads are stored,
- an index file, loaded into memory, used to locate specific
documents in the data file.

Looking up term vectors for any document requires at most 1 disk seek.

File formats

1. vector_data

A vector data file (extension .tvd). This file stores terms,
frequencies, positions, offsets and payloads for every document. Upon
writing a new segment, it accumulates data into memory until the
buffer used to store terms and payloads grows beyond 4KB. Then it
flushes all metadata, terms and positions to disk using LZ4
compression for terms and payloads and blocks of packed ints for
positions

Here is more detailed description of the field data file format:

- VectorData (.tvd) --> <Header>, PackedIntsVersion, ChunkSize, <Chunk>^ChunkCount
- Header --> CodecHeader
- PackedIntsVersion --> PackedInts.CURRENT_VERSION as a VInt
- ChunkSize is the number of bytes of terms to accumulate before
  flusing, as a VInt
- ChunkCount is not known in advance and is the number of chunks
  necessary to store all document of the segment
- Chunk --> DocBase, ChunkDocs, <NumFields>, <FieldNums>, <FieldNumOffs>, <Flags>,
  <NumTerms>, <TermLengths>, <TermFreqs>, <Position>, <StartOffsets>, <Lengths>,
  <PayloadLengths>, <TermAndPayloads>
- DocBase is the ID of the first doc of the chunk as a VInt
- ChunkDocs is the number of documents in the chunk
- NumFields --> DocNumFields^ChunkDocs
- DocNUmFields is the number of fields for each doc, written as a
  VInt if ChunkDocs==1 and as a PackedInts array otherwise
- FieldNums --> FieldNumDelta^TotalFields, as a PackedInts array
- FieldNumOff is the offset of the field number in FieldNums
- TotalFields is the total number of fields (sum of the values of NumFields)
- Flags --> Bit <FieldFlags>
- Bit is a single bit which when true means that fields have the same
  options for every document in the chunk
- FieldFlags --> if Bit==1: Flag^TotalDistinctFields else Flag^TotalFields
- Flag: a 3-bits int where:
  - the first bit means that the field has positions
  - the second bit means that the field has offsets
  - the third bitmeans that the field has payloads
- NumTerms --> FieldNumTerms^TotalFields
- FieldNumTerms: the numer of terms for each field, using blocks of 64 packed ints
- TermLengths --> PrefixLength^TotalTerms SuffixLength^TotalTerms
- TotalTerms: total number of terms (sum of NumTerms)
- SuffixLength: length of the term of a field, the common prefix with
  the previous term otherwise using blocks of 64 packed ints
- TermFreqs --> TermFreqMinus1^TotalTerms
- TermFreqMinus1: (frequency - 1) for each term using blocks of 64 packed ints
- Positions --> PositionDelta^TotalPositions
- TotalPositions is the sum of frequencies of terms of all fields that have positions
- PositionDelta: the absolute position fo rthe first position of a
  term, and the difference with the previous positions for following
  positions using blocks of 64 packed ints
- StartOffsets --> (AvgCharsPerTerm^TotalDistinctFields) StartOffsetDelta^TotalOffsets
- TotalOffsets is the sum of frequencies of terms of all fields tha thave offsets
- AvgCharsPerTerm: average number of chars per term, encoded as a
  float32 on 4 bytes. They are not present if no field has both
  positions and offsets enabled.
- StartOffsetDelta: (startOffset - previousStartOffset - AvgCharsPerTerm
  * PositionDelta). previousStartOffset is 0 for the first offset and
  AvgCharsPerTerm is 0 if the field has no ositions using blocks of
  64 pakced ints
- Lengths --> LengthMinusTermLength^TotalOffsets
- LengthMinusTermLength: (endOffset - startOffset - termLenght) using blocks of 64 packed ints
- PayloadLengths --> PayloadLength^TotalPayloads
- TotalPayloads is the sum of frequencies of terms of all fields that have payloads
- PayloadLength is the payload length encoded using blocks of 64 packed ints
- TermAndPayloads --> LZ4-compressed representation of <FieldTermsAndPayLoads>^TotalFields
- FieldTermsAndPayLoads --> Terms (Payloads)
- Terms: term bytes
- Payloads: payload bytes (if the field has payloads)

2. vector_index

An index file (extension .tvx).

- VectorIndex (.tvx) --> <Header>, <ChunkIndex>
- Header --> CodecHeader
- ChunkIndex: See CompressingStoredFieldsIndexWriter
*/
type Lucene42TermVectorsFormat struct {
	*CompressingTermVectorsFormat
}

func newLucene42TermVectorsFormat() *Lucene42TermVectorsFormat {
	panic("not implemented yet")
}

type Lucene42TermVectorsReader struct {
	*CompressingTermVectorsReader
}

func newLucene42TermVectorsReader(d store.Directory, si SegmentInfo, fn FieldInfos, ctx store.IOContext) (r TermVectorsReader, err error) {
	formatName := "Lucene41StoredFields"
	compressionMode := codec.COMPRESSION_MODE_FAST
	// chunkSize := 1 << 12
	p, err := newCompressingTermVectorsReader(d, si, "", fn, ctx, formatName, compressionMode)
	if err == nil {
		r = &Lucene42TermVectorsReader{p}
	}
	return r, nil
}

type CompressingTermVectorsReader struct {
	vectorsStream store.IndexInput
	closed        bool
}

func newCompressingTermVectorsReader(d store.Directory, si SegmentInfo, segmentSuffix string, fn FieldInfos,
	ctx store.IOContext, formatName string, compressionMode codec.CompressionMode) (r *CompressingTermVectorsReader, err error) {
	panic("not implemented yet")
	return nil, nil
}

func (r *CompressingTermVectorsReader) Close() (err error) {
	if !r.closed {
		err = util.Close(r.vectorsStream)
		r.closed = true
	}
	return err
}

func (r *CompressingTermVectorsReader) get(doc int) Fields {
	panic("not implemented yet")
}

func (r *CompressingTermVectorsReader) clone() TermVectorsReader {
	panic("not implemented yet")
}
