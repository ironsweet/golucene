package index

import (
	"github.com/balzaczyy/golucene/core/util"
)

// index/SimpleMergedSegmentWarmer.java

/*
A very simple meged segment warmer that just ensures data structures
are initialized.
*/
type SimpleMergedSegmentWarmer struct {
	infoStream util.InfoStream
}

// Creates a new SimpleMergedSegmentWarmer
func NewSimpleMergedSegmentWarmer(infoStream util.InfoStream) *SimpleMergedSegmentWarmer {
	return &SimpleMergedSegmentWarmer{infoStream}
}

func (warmer *SimpleMergedSegmentWarmer) warm(reader AtomicReader) error {
	panic("not implemented yet")
}
