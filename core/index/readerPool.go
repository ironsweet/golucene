package index

import (
	"strings"
	"sync"
)

// L421
type ReaderPool struct {
	owner *IndexWriter
	sync.Locker
	readerMap map[*SegmentInfoPerCommit]*ReadersAndLiveDocs
}

func newReaderPool(owner *IndexWriter) *ReaderPool {
	return &ReaderPool{
		owner:     owner,
		Locker:    &sync.Mutex{},
		readerMap: make(map[*SegmentInfoPerCommit]*ReadersAndLiveDocs),
	}
}

func (pool *ReaderPool) drop(info *SegmentInfoPerCommit) error {
	pool.Lock()
	defer pool.Unlock()
	panic("not implemented yet")
}

// Remove all our references to readers, and commits any pending changes.
func (pool *ReaderPool) dropAll(doSave bool) error {
	pool.Lock() // synchronized
	defer pool.Unlock()
	panic("not implemented yet")
}

// Commit live docs changes for the segment readers for the provided infos.
func (pool *ReaderPool) commit(infos *SegmentInfos) error {
	pool.Lock() // synchronized
	defer pool.Unlock()
	panic("not implemented yet")
}

// Obtain a readersAndLiveDocs instance from the ReaderPool. If
// create is true, you must later call release().
func (pool *ReaderPool) get(info *SegmentInfoPerCommit, create bool) *ReadersAndLiveDocs {
	pool.Lock() // synchronized
	defer pool.Unlock()
	panic("not implemented yet")
}

/*
Obtain the number of deleted docs for a pooled reader. If the reader
isn't being pooled, the segmentInfo's delCount is returned.
*/
func (pool *ReaderPool) numDeletedDocs(info *SegmentInfoPerCommit) int {
	// ensureOpen(false)
	delCount := info.delCount
	if rld := pool.get(info, false); rld != nil {
		delCount += rld.pendingDeleteCount()
	}
	return delCount
}

/*
returns a string description of the specified segments, for debugging.
*/
func (pool *ReaderPool) segmentsToString(infos []*SegmentInfoPerCommit) string {
	// TODO synchronized
	var parts []string
	for _, info := range infos {
		parts = append(parts, pool.segmentToString(info))
	}
	return strings.Join(parts, " ")
}

/*
Returns a string description of the specified segment, for debugging.
*/
func (pool *ReaderPool) segmentToString(info *SegmentInfoPerCommit) string {
	// TODO synchronized
	return info.StringOf(info.info.dir, pool.numDeletedDocs(info)-info.delCount)
}
