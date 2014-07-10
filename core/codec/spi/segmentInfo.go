package spi

import (
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
)

// codecs/SegmentInfoFormat.java

// Expert: Control the format of SegmentInfo (segment metadata file).
type SegmentInfoFormat interface {
	// Returns the SegmentInfoReader for reading SegmentInfo instances.
	SegmentInfoReader() SegmentInfoReader
	// Returns the SegmentInfoWriter for writing SegmentInfo instances.
	SegmentInfoWriter() SegmentInfoWriter
}

// codecs/SegmentInfoReader.java

// Read SegmentInfo data from a directory.

type SegmentInfoReader interface {
	Read(store.Directory, string, store.IOContext) (*SegmentInfo, error)
}

// codecs/SegmentInfoWriter.java

// Write SegmentInfo data.
type SegmentInfoWriter interface {
	Write(store.Directory, *SegmentInfo, FieldInfos, store.IOContext) error
}
