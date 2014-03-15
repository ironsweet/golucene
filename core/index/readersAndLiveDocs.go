package index

import (
	"github.com/balzaczyy/golucene/core/store"
	"sync"
)

// index/ReadersAndLiveDocs.java

/*
Used by IndexWriter to hold open SegmentReaders (for searching or
merging), plus pending deletes, for a given segment.
*/
type ReadersAndLiveDocs struct {
	sync.Locker

	info *SegmentInfoPerCommit

	// How many further deletions we;ve doen against
	// liveDocs vs when we loaded it or last wrote it:
	_pendingDeleteCount int
}

func (rld *ReadersAndLiveDocs) pendingDeleteCount() int {
	rld.Lock()
	defer rld.Unlock()
	return rld._pendingDeleteCount
}

/*
Get reader for searching/deleting
*/
func (rld *ReadersAndLiveDocs) reader(ctx store.IOContext) (*SegmentReader, error) {
	panic("not implemented yet")
}

func (rld *ReadersAndLiveDocs) release(sr *SegmentReader) error {
	panic("not implemented yet")
}

func (rld *ReadersAndLiveDocs) String() string {
	panic("not implemented yet")
}
