package index

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
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
}

func newLucene40SegmentInfoFormat() *Lucene40SegmentInfoFormat {
	panic("not implemented yet")
}

func (f *Lucene40SegmentInfoFormat) SegmentInfoReader() SegmentInfoReader {
	panic("not implemented yet")
}

func (f *Lucene40SegmentInfoFormat) SegmentInfoWriter() SegmentInfoWriter {
	panic("not implemented yet")
}

const (
	LUCENE40_SI_EXTENSION    = "si"
	LUCENE40_CODEC_NAME      = "Lucene40SegmentInfo"
	LUCENE40_VERSION_START   = 0
	LUCENE40_VERSION_CURRENT = LUCENE40_VERSION_START

	SEGMENT_INFO_YES = 1
)

var (
	Lucene40SegmentInfoReader = func(dir store.Directory, segment string, context store.IOContext) (si SegmentInfo, err error) {
		si = SegmentInfo{}
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

		si = SegmentInfo{dir, version, segment, docCount, isCompoundFile, nil, diagnostics, attributes, nil}
		si.CheckFileNames(files)
		si.Files = files

		success = true
		return si, nil
	}
)

// Lucene40StoredFieldsWriter.java
const (
	LUCENE40_SF_FIELDS_EXTENSION       = "fdt"
	LUCENE40_SF_FIELDS_INDEX_EXTENSION = "fdx"
)
