package lucene40

import (
	. "github.com/balzaczyy/golucene/core/codec/spi"
	// . "github.com/balzaczyy/golucene/core/index/model"
	// "github.com/balzaczyy/golucene/core/store"
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

func NewLucene40SegmentInfoFormat() *Lucene40SegmentInfoFormat {
	return &Lucene40SegmentInfoFormat{
		reader: new(Lucene40SegmentInfoReader),
		writer: new(Lucene40SegmentInfoWriter),
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
)
