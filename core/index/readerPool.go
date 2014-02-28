package index

import (
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
