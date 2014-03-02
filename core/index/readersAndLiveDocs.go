package index

import (
	"sync"
)

// index/ReadersAndLiveDocs.java

/*
Used by IndexWriter to hold open SegmentReaders (for searching or
merging), plus pending deletes, for a given segment.
*/
type ReadersAndLiveDocs struct {
	sync.Locker

	// How many further deletions we;ve doen against
	// liveDocs vs when we loaded it or last wrote it:
	_pendingDeleteCount int
}

func (rld *ReadersAndLiveDocs) pendingDeleteCount() int {
	rld.Lock()
	defer rld.Unlock()
	return rld._pendingDeleteCount
}
