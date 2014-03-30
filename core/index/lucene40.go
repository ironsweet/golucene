package index

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	. "github.com/balzaczyy/golucene/core/codec/lucene40"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
)

// lucene40/Lucene40SegmentInfoFormat.java

/*
Lucene 4.0 Segment info format.

Files:
- .si: Header, SegVersion, SegSize, IsCompoundFile, Diagnostics, Attributes, Files

Data types:
- Header --> CodecHeader
- SegSize --> int32
- SegVersion --> string
- Files --> set[string]
- Diagnostics, Attributes --> map[string]string
- IsCompoundFile --> byte

Field Descriptions:
- SegVersion is the code version that created the segment.
- SegSize is the number of documents contained in the segment index.
- IsCompoundFile records whether the segment is written as a compound
  file or not. If this is -1, the segment is not a compound file. If
  it is 1, the segment is a compound file.
- Checksum contains the CRC32 checksum of all bytes in the segments_N
  file up until the checksum. This is used to verify integrity of the
  file on opening the index.
- The Diagnostics Map is privately written by IndexWriter, as a debugging
  for each segment it creates. It includes metadata like the current
  Lucene version, OS, Java version, why the segment was created (merge,
  flush, addIndexes), etc.
- Attributes: a key-value map of codec-pivate attributes.
- Files is a list of files referred to by this segment.
*/
type Lucene40SegmentInfoFormat struct {
	reader SegmentInfoReader
	writer SegmentInfoWriter
}

func newLucene40SegmentInfoFormat() *Lucene40SegmentInfoFormat {
	return &Lucene40SegmentInfoFormat{
		reader: Lucene40SegmentInfoReader,
		writer: Lucene40SegmentInfoWriter,
	}
}

func (f *Lucene40SegmentInfoFormat) SegmentInfoReader() SegmentInfoReader {
	return f.reader
}

func (f *Lucene40SegmentInfoFormat) SegmentInfoWriter() SegmentInfoWriter {
	return f.writer
}

const (
	LUCENE40_SI_EXTENSION    = "si"
	LUCENE40_CODEC_NAME      = "Lucene40SegmentInfo"
	LUCENE40_VERSION_START   = 0
	LUCENE40_VERSION_CURRENT = LUCENE40_VERSION_START

	SEGMENT_INFO_YES = 1
)

// lucene40/Lucene40SegmentInfoReader.java

var Lucene40SegmentInfoReader = func(dir store.Directory, segment string, context store.IOContext) (si *SegmentInfo, err error) {
	si = &SegmentInfo{}
	fileName := util.SegmentFileName(segment, "", LUCENE40_SI_EXTENSION)
	input, err := dir.OpenInput(fileName, context)
	if err != nil {
		return si, err
	}

	success := false
	defer func() {
		if !success {
			util.CloseWhileSuppressingError(input)
		} else {
			input.Close()
		}
	}()

	_, err = codec.CheckHeader(input, LUCENE40_CODEC_NAME, LUCENE40_VERSION_START, LUCENE40_VERSION_CURRENT)
	if err != nil {
		return si, err
	}
	version, err := input.ReadString()
	if err != nil {
		return si, err
	}
	docCount, err := input.ReadInt()
	if err != nil {
		return si, err
	}
	if docCount < 0 {
		return si, errors.New(fmt.Sprintf("invalid docCount: %v (resource=%v)", docCount, input))
	}
	sicf, err := input.ReadByte()
	if err != nil {
		return si, err
	}
	isCompoundFile := (sicf == SEGMENT_INFO_YES)
	diagnostics, err := input.ReadStringStringMap()
	if err != nil {
		return si, err
	}
	attributes, err := input.ReadStringStringMap()
	if err != nil {
		return si, err
	}
	files, err := input.ReadStringSet()
	if err != nil {
		return si, err
	}

	if input.FilePointer() != input.Length() {
		return si, errors.New(fmt.Sprintf(
			"did not read all bytes from file '%v': read %v vs size %v (resource: %v)",
			fileName, input.FilePointer(), input.Length(), input))
	}

	si = newSegmentInfo(dir, version, segment, int(docCount), isCompoundFile,
		nil, diagnostics, attributes)
	si.setFiles(files)

	success = true
	return si, nil
}

// lucene40/Lucne40SegmentInfoWriter.java

// Lucene 4.0 implementation of SegmentInfoWriter
var Lucene40SegmentInfoWriter = func(dir store.Directory, si *SegmentInfo, fis FieldInfos, ctx store.IOContext) error {
	panic("not implemented yet")
}

// Lucene40StoredFieldsWriter.java
const (
	LUCENE40_SF_FIELDS_EXTENSION       = "fdt"
	LUCENE40_SF_FIELDS_INDEX_EXTENSION = "fdx"
)

// codecs/lucene40/Lucene40LiveDocsFormat.java

/*
Lucene 4.0 Live Documents Format.

The .del file is optional, and only exists when a segment contains
deletions.

Although per-segment, this file is maintained exterior to compound
segment files.

Deletions (.del) --> Format,Heaer,ByteCount,BitCount, Bits | DGaps
  (depending on Format)
	Format,ByteSize,BitCount --> uint32
	Bits --> <byte>^ByteCount
	DGaps --> <DGap,NonOnesByte>^NonzeroBytesCount
	DGap --> vint
	NonOnesByte --> byte
	Header --> CodecHeader

Format is 1: indicates cleard DGaps.

ByteCount indicates the number of bytes in Bits. It is typically
(SegSize/8)+1.

BitCount indicates the number of bits that are currently set in Bits.

Bits contains one bit for each document indexed. When the bit
corresponding to a document number is cleared, that document is
marked as deleted. Bit ordering is from least to most significant.
Thus, if Bits contains two bytes, 0x00 and 0x02, then document 9 is
marked as alive (not deleted).

DGaps represents sparse bit-vectors more efficiently than Bits. It is
makde of DGaps on indexes of nonOnes bytes in Bits, and the nonOnes
bytes themselves. The number of nonOnes byte in Bits
(NonOnesBytesCount) is not stored.

For example, if there are 8000 bits and only bits 10,12,32 are
cleared, DGaps would be used:

(vint) 1, (byte) 20, (vint) 3, (byte) 1
*/
type Lucene40LiveDocsFormat struct {
}

func (format *Lucene40LiveDocsFormat) NewLiveDocs(size int) util.MutableBits {
	ans := NewBitVector(size)
	ans.InvertAll()
	return ans
}

func (format *Lucene40LiveDocsFormat) WriteLiveDocs(bits util.MutableBits,
	dir store.Directory, info *SegmentInfoPerCommit, newDelCount int,
	ctx store.IOContext) error {

	panic("not implemented yet")
}
