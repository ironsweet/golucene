package lucene42

import (
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/util/packed"
)

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

func (f *Lucene42DocValuesFormat) FieldsConsumer(state *SegmentWriteState) (w DocValuesConsumer, err error) {
	panic("this codec can only be used for reading")
}

func (f *Lucene42DocValuesFormat) FieldsProducer(state SegmentReadState) (r DocValuesProducer, err error) {
	return newLucene42DocValuesProducer(state,
		LUCENE42_DV_DATA_CODEC, LUCENE42_DV_DATA_EXTENSION,
		LUCENE42_DV_METADATA_CODEC, LUCENE42_DV_METADATA_EXTENSION)
}
