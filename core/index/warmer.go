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
}

// Creates a new SimpleMergedSegmentWarmer
func NewSimpleMergedSegmentWarmer(infoStream util.InfoStream) *SimpleMergedSegmentWarmer {
	panic("not implemented yet")
}

func (warmer *SimpleMergedSegmentWarmer) warm(reader AtomicReader) error {
	panic("not implemented yet")
}
