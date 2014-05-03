package index

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	"github.com/balzaczyy/golucene/core/codec/compressing"
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"github.com/balzaczyy/golucene/core/util/packed"
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
	name:             "Lucene42Codec",
	fieldsFormat:     newLucene41StoredFieldsFormat(),
	vectorsFormat:    newLucene42TermVectorsFormat(),
	fieldInfosFormat: newLucene42FieldInfosFormat(),
	infosFormat:      newLucene40SegmentInfoFormat(),
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
	normsFormat: newReadonlyLucene42NormsFormat(),
}

type readonlyLucene42NormsFormat struct {
	*Lucene42NormsFormat
}

func (f *readonlyLucene42NormsFormat) NormsConsumer(state SegmentWriteState) (w DocValuesConsumer, err error) {
	panic("this codec can only be used for reading")
}

func newReadonlyLucene42NormsFormat() *readonlyLucene42NormsFormat {
	return &readonlyLucene42NormsFormat{
		newLucene42NormsFormat(),
	}
}

// lucene42/Lucene42FieldInfosFormat.java

/*
Lucene 4.2 Field Infos format.

Field names are stored in the field info file, with suffix .fnm.

FieldInfos (.fnm) --> Header, HeaderCount, <FieldName, FieldNumber,
                      FieldBits, DocValuesBits, Attribute>^FieldsCount

Data types:
- Header --> CodecHeader
- FieldsCount --> VInt
- FieldName --> string
- FieldBits, DocValuesBit --> byte
- FieldNumber --> VInt
- Attributes --> map[string]string

Field Description:
- FieldsCount: the number of fields in this file.
- FieldName: name of the field as a UTF-8 string.
- FieldNumber: the field's number. NOte that unlike previous versions
  of Lucene, the fields are not numbered implicitly by their order in
  the file, instead explicitly.
- FieldBits: a byte containing field options.
  - The low-order bit is one for indexed fields, and zero for non-indexed
    fields.
  - The second lowest-order bit is one for fields that have term vectors
    stored, and zero for fields without term vectors.
  - If the third lowest order-bit is set (0x4), offsets are stored into
    the postings list in addition to positions.
  - Fourth bit is unsed.
  - If the fifth lowest-order bit is set (0x10), norms are omitted for
    the indexed field.
  - If the sixth lowest-order bit is set (0x20), payloads are stored
    for the indexed field.
  - If the seventh lowest-order bit is set (0x40), term frequencies a
    and ositions omitted for the indexed field.
  - If the eighth lowest-order bit is set (0x80), positions are omitted
    for the indexed field.
- DocValuesBits: a byte containing per-document value types. The type
  recorded as two four-bit intergers, with the high-order bits
  representing norms options, and low-order bits representing DocVlaues
  options. Each four-bit interger can be decoded as such:
  - 0: no DocValues for this field.
  - 1: NumericDocValues.
  - 2: BinaryDocvalues.
  - 3: SortedDocValues.
- Attributes: a key-value map of codec-private attributes.
*/
type Lucene42FieldInfosFormat struct {
	reader FieldInfosReader
	writer FieldInfosWriter
}

func newLucene42FieldInfosFormat() *Lucene42FieldInfosFormat {
	return &Lucene42FieldInfosFormat{
		reader: Lucene42FieldInfosReader,
		writer: Lucene42FieldInfosWriter,
	}
}

func (f *Lucene42FieldInfosFormat) FieldInfosReader() FieldInfosReader {
	return f.reader
}

func (f *Lucene42FieldInfosFormat) FieldInfosWriter() FieldInfosWriter {
	return f.writer
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

var Lucene42FieldInfosReader = func(dir store.Directory,
	segment string, context store.IOContext) (fi model.FieldInfos, err error) {

	log.Printf("Reading FieldInfos from %v...", dir)
	fi = model.FieldInfos{}
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

	infos := make([]model.FieldInfo, size)
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
		var indexOptions model.IndexOptions
		switch {
		case !isIndexed:
			indexOptions = model.IndexOptions(0)
		case (bits & LUCENE42_FI_OMIT_TERM_FREQ_AND_POSITIONS) != 0:
			indexOptions = model.INDEX_OPT_DOCS_ONLY
		case (bits & LUCENE42_FI_OMIT_POSITIONS) != 0:
			indexOptions = model.INDEX_OPT_DOCS_AND_FREQS
		case (bits & LUCENE42_FI_STORE_OFFSETS_IN_POSTINGS) != 0:
			indexOptions = model.INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
		default:
			indexOptions = model.INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS
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
		infos[i] = model.NewFieldInfo(name, isIndexed, fieldNumber, storeTermVector,
			omitNorms, storePayloads, indexOptions, docValuesType, normsType, attributes)
	}

	if input.FilePointer() != input.Length() {
		return fi, errors.New(fmt.Sprintf(
			"did not read all bytes from file '%v': read %v vs size %v (resource: %v)",
			fileName, input.FilePointer(), input.Length(), input))
	}
	fi = model.NewFieldInfos(infos)
	success = true
	return fi, nil
}

func getDocValuesType(input store.IndexInput, b byte) (t model.DocValuesType, err error) {
	switch b {
	case 0:
		return model.DocValuesType(0), nil
	case 1:
		return model.DOC_VALUES_TYPE_NUMERIC, nil
	case 2:
		return model.DOC_VALUES_TYPE_BINARY, nil
	case 3:
		return model.DOC_VALUES_TYPE_SORTED, nil
	case 4:
		return model.DOC_VALUES_TYPE_SORTED_SET, nil
	default:
		return model.DocValuesType(0), errors.New(
			fmt.Sprintf("invalid docvalues byte: %v (resource=%v)", b, input))
	}
}

// lucene42/Lucene42FieldInfosWriter.java
var Lucene42FieldInfosWriter = func(dir store.Directory,
	segName string, infos model.FieldInfos, ctx store.IOContext) error {
	panic("not implemented yet")
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
	return &Lucene42TermVectorsFormat{
		newCompressingTermVectorsFormat("Lucene41StoredFields", "", compressing.COMPRESSION_MODE_FAST, 1<<12),
	}
}

type Lucene42TermVectorsReader struct {
	*CompressingTermVectorsReader
}

func newLucene42TermVectorsReader(d store.Directory,
	si *model.SegmentInfo, fn model.FieldInfos,
	ctx store.IOContext) (r TermVectorsReader, err error) {

	formatName := "Lucene41StoredFields"
	compressionMode := compressing.COMPRESSION_MODE_FAST
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

func newCompressingTermVectorsReader(d store.Directory,
	si *model.SegmentInfo, segmentSuffix string,
	fn model.FieldInfos, ctx store.IOContext, formatName string,
	compressionMode compressing.CompressionMode) (r *CompressingTermVectorsReader, err error) {

	panic("not implemented yet")
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

// lucene42/Lucene42NormsFormat.java

/*
Lucene 4.2 score normalization format.

NOTE: this uses the same format as Lucene42DocValuesFormat Numeric
DocValues, but with different fiel extensions, and passing FASTEST
for uncompressed encoding: trading off space for performance.

Fiels:
- .nvd: Docvalues data
- .nvm: DocValues metadata
*/
type Lucene42NormsFormat struct {
	acceptableOverheadRatio float32
}

func newLucene42NormsFormat() *Lucene42NormsFormat {
	// note: we choose FASTEST here (otherwise our norms are half as big
	// but 15% slower than previous lucene)
	return newLucene42NormsFormatWithOverhead(packed.PackedInts.FASTEST)
}

func newLucene42NormsFormatWithOverhead(acceptableOverheadRatio float32) *Lucene42NormsFormat {
	return &Lucene42NormsFormat{acceptableOverheadRatio}
}

func (f *Lucene42NormsFormat) NormsConsumer(state SegmentWriteState) (w DocValuesConsumer, err error) {
	panic("not implemented yet")
}

func (f *Lucene42NormsFormat) NormsProducer(state SegmentReadState) (r DocValuesProducer, err error) {
	return newLucene42DocValuesProducer(state, "Lucene41NormsData", "nvd", "Lucene41NormsMetadata", "nvm")
}

// lucene42/Lucene42DocValuesFormat.java

/*
Lucene 4.2 DocValues format.

Encodes the four per-document value types (Numeric, Binary, Sorted,
SortedSet) with seven basic strategies.

- Delta-compressed Numerics: per-document integers written in blocks
  of 4096. For each block the minimum value is encoded, and each entry
  is a delta from that minimum value.
- Table-compressed Numerics: when the number of unique values is very
  small, a lookup table is written instead. Each per-document entry
  is instead the ordinal to this table.
- Uncompressed Numerics: when all values would fit into a single byte,
  and the  acceptableOverheadRatio would pack values into 8 bits per
  value anyway, they are written as absolute values (with no indirection
  or packing) for performance.
- GCD-compressed Numerics: when all numbers share a common divisor,
  such as dates, the greatest common denominator (GCD) is computed,
  and quotients are stored using Delta-compressed Numerics.
- Fixed-width Binary: one large concatenated []byte is written, along
  with the fixed length. Each document's value can be addressed by
  maxDoc*length.
- Variable-width Binary: one large concatenated []byte is written,
  along  with end addresses for each document. The addresses are written
  in blocks of 4096, with the current absolute start for the block,
  and the average (expected) delta per entry. For each document the
  deviation from the delta (actual - expected) is written.
- Sorted: an FST mapping deduplicated terms to ordinals is written,
  along with the per-document ordinals written using one of the numeric
  strategies above.
- SortedSet: an FST mapping deduplicated terms to ordinals is written,
  along with the per-document ordinal list written using one of the
  binary strategies above.

Files:

1. .dvd: DocValues data
2. .dvm: DocValues metadata

###### 1. dvm

The DocValues metadata or .dvm files.

The DocValues field, this stores metadata, such as the offset into the
DocValues data (.dvd)

DocValues metadata (.dvm) --> Header, <FieldNumber, EntryType, Entry>^NumFields

- Entry --> NumericEntry | BinaryEntry | SortedEntry
- NumericEntry --> DataOffset, CompressionType, packedVersion
- BinaryEntry --> DataOffset, DataLength, MinLength, MaxLength, packedVersion, BlockSize?
- SortedEntry --> DataOffset, ValueCount
- FieldNumber, PackedVersion, MinLength, MaxLength, BlockSize, ValudCount --> VInt
- DataOffset, DataLength --> int64
- EntryType, CompressionType --> byte
- Header --> CodecHeader

Sorted fields have two entries: a Sortedentry with the FST metadata,
and an ordinary NumericEntry for the document-to-ord metadata.

SortedSet fields have two entries: a SortedEntry with the FST metadata,
and an ordinary BinaryEntry for the document-to-ord-list meatadata.

FieldNumber of -1 indicates the end of metadata.

EntryType is a 0 (NumericEntry), 1 (BinaryEntry), or 2 (SortedEntry)

DataOffset is the pointer to the start of the dta in the DocValues
data (.dvd)

CompressionType indicates how Numeric values will be compressed:

- 0 --> delta-compressed. For each block of 4096 integers, every integer
  is delta-encoded from the minimum value within the block.
- 1 --> table-compressed. When the number of unique numeric values is
  small and it would save space, a lookup table of unique values is
  written, followed by the ordinal for each document.
- 2 --> uncompressed. When the acceptableOverHeadratio parameter would
  upgrade the number of bits required to 8, and all values fit in a
  byte, these are written as absolute binary values for performance.
- 3 --> gcd-compressed. When all integers share a common divisor, only
  quotients are stored using blocks of delta-encoded ints.


MinLength and MaxLength represent the min and max []byte value lengths
for Binary values. If they are equal, then all values are of a fixed
size, and can be addressed as DataOffset + (docID * length). Otherwise,
the binary values are of variable size, and packed integer metadata (
PackedVersion, BlockSize) is written for addresses.

###### 2. dvd

The DocVlaues data or .dvd file.

For Docvalues field, this stores the actual per-document data (the
heavy-lifting)

DocValues data (.dvd) --> Header, <NumericData | BinaryData | SortedData>^NumFields

- NumericData --> DeltaCOmpressedNumerics | TableCompressedNumerics |
  UncompressedNumerics | GCDCOmpressedNumerics
- BinaryData --> byte^DataLength, Addresses
- Sorteddata --> FST<int64>
- DeltaCompressedNumerics --> BlockPackedInts(blockSize=4096)
- TableCompressedNumerics --> TableSize, int64^TableSize, PackedInts
- UncompressedNumerics --> byte^maxdoc
- Addresses --> MonotonicBlockpackedInts(blockSize=4096)


SortedSet entries store the list of ordinals in their BinaryData as a
sequences of increasing VLongs, delta-encoded.

Limitations:
- Binary doc values can be at most MAX_BINARY_FIELD_LENGTH in length.
*/
type Lucene42DocValuesFormat struct {
	AcceptableOverheadRatio float32
}

func NewLucene42DocValuesFormat() *Lucene42DocValuesFormat {
	return &Lucene42DocValuesFormat{packed.PackedInts.DEFAULT}
}

func (f *Lucene42DocValuesFormat) Name() string {
	return "Lucene42"
}

func (f *Lucene42DocValuesFormat) FieldsConsumer(state SegmentWriteState) (w DocValuesConsumer, err error) {
	panic("this codec can only be used for reading")
}

func (f *Lucene42DocValuesFormat) FieldsProducer(state SegmentReadState) (r DocValuesProducer, err error) {
	return newLucene42DocValuesProducer(state,
		LUCENE42_DV_DATA_CODEC, LUCENE42_DV_DATA_EXTENSION,
		LUCENE42_DV_METADATA_CODEC, LUCENE42_DV_METADATA_EXTENSION)
}
