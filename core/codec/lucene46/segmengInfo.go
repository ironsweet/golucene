package lucene46

import (
	. "github.com/balzaczyy/golucene/core/codec/spi"
)

type Lucene46SegmentInfoFormat struct {
}

func NewLucene46SegmentInfoFormat() *Lucene46SegmentInfoFormat {
	return &Lucene46SegmentInfoFormat{}
}

func (f *Lucene46SegmentInfoFormat) SegmentInfoReader() SegmentInfoReader {
	panic("not implemented yet")
}

func (f *Lucene46SegmentInfoFormat) SegmentInfoWriter() SegmentInfoWriter {
	panic("not implemented yet")
}
