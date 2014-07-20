package lucene42

import (
	"github.com/balzaczyy/golucene/core/codec/compressing"
)

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
	*compressing.CompressingTermVectorsFormat
}

func NewLucene42TermVectorsFormat() *Lucene42TermVectorsFormat {
	return &Lucene42TermVectorsFormat{
		compressing.NewCompressingTermVectorsFormat("Lucene41StoredFields", "", compressing.COMPRESSION_MODE_FAST, 1<<12),
	}
}
