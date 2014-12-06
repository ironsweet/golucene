package index

import (
	. "github.com/balzaczyy/golucene/core/codec/spi"
	"strings"
	"sync"
)

type ReaderPool struct {
	owner *IndexWriter
	sync.Locker
	readerMap map[*SegmentCommitInfo]*ReadersAndUpdates
}

func newReaderPool(owner *IndexWriter) *ReaderPool {
	return &ReaderPool{
		owner:     owner,
		Locker:    &sync.Mutex{},
		readerMap: make(map[*SegmentCommitInfo]*ReadersAndUpdates),
	}
}

func (pool *ReaderPool) infoIsLive(info *SegmentCommitInfo) bool {
	panic("not implemented yet")
}

func (pool *ReaderPool) drop(info *SegmentCommitInfo) error {
	pool.Lock()
	defer pool.Unlock()
	panic("not implemented yet")
}

func (pool *ReaderPool) release(rld *ReadersAndUpdates) error {
	panic("not implemented yet")
}

func (pool *ReaderPool) Close() error {
	return pool.dropAll(false)
}

// Remove all our references to readers, and commits any pending changes.
func (pool *ReaderPool) dropAll(doSave bool) error {
	pool.Lock() // synchronized
	defer pool.Unlock()

	var priorE error
	for len(pool.readerMap) > 0 {
		for k, rld := range pool.readerMap {
			if doSave {
				ok, err := rld.writeLiveDocs(pool.owner.directory)
				if err != nil {
					return err
				}
				if ok {
					// Make sure we only write del docs and field updates for a live segment:
					assert(pool.infoIsLive(rld.info))
					// Must checkpoint because we just
					// created new _X_N.del and field updates files;
					// don't call IW.checkpoint because that also
					// increments SIS.version, which we do not want to
					// do here: it was done previously (after we
					// invoked BDS.applyDeletes), whereas here all we
					// did was move the state to disk:
					err = pool.owner.checkpointNoSIS()
					if err != nil {
						return err
					}
				}
			}

			// Important to remove as-we-go, not with .clear()
			// in the end, in case we hit an exception;
			// otherwise we could over-decref if close() is
			// called again:
			delete(pool.readerMap, k)

			// NOTE: it is allowed that these decRefs do not
			// actually close the SRs; this happens when a
			// near real-time reader is kept open after the
			// IndexWriter instance is closed:
			err := rld.dropReaders()
			if err != nil {
				if doSave {
					return err
				}
				if priorE == nil {
					priorE = err
				}
			}
		}
	}
	assert(len(pool.readerMap) == 0)
	return priorE
}

/* Commit live docs changes for the segment readers for the provided infos. */
func (pool *ReaderPool) commit(infos *SegmentInfos) error {
	pool.Lock() // synchronized
	defer pool.Unlock()

	for _, info := range infos.Segments {
		if rld, ok := pool.readerMap[info]; ok {
			assert(rld.info == info)
			ok, err := rld.writeLiveDocs(pool.owner.directory)
			if err != nil {
				return err
			}
			if ok {
				// Make sure we only write del docs for a live segment:
				assert(pool.infoIsLive(info))
				// Must checkpoint because we just created new _X_N.del and
				// field updates files; don't call IW.checkpoint because that
				// also increments SIS.version, which we do not want to do
				// here: it was doen previously (after we invoked
				// BDS.applyDeletes), whereas here all we did was move the
				// stats to disk:
				err = pool.owner.checkpointNoSIS()
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

/*
Obtain a ReadersAndUpdates instance from the ReaderPool. If
create is true, you must later call release().
*/
func (pool *ReaderPool) get(info *SegmentCommitInfo, create bool) *ReadersAndUpdates {
	pool.Lock() // synchronized
	defer pool.Unlock()

	assertn(info.Info.Dir == pool.owner.directory, "info.dir=%v vs %v", info.Info.Dir, pool.owner.directory)

	rld, ok := pool.readerMap[info]
	if !ok {
		if !create {
			return nil
		}
		rld = newReadersAndUpdates(pool.owner, info)
		// Steal initial reference:
		pool.readerMap[info] = rld
	} else {
		assertn(rld.info == info, "rld.info=%v info=%v isLive?= %v vs %v",
			rld.info, info, pool.infoIsLive(rld.info), pool.infoIsLive(info))
	}

	if create {
		// Return ref to caller:
		rld.incRef()
	}

	assert(pool.noDups())

	return rld
}

/* Make sure that every segment appears only once in the pool: */
func (pool *ReaderPool) noDups() bool {
	seen := make(map[string]bool)
	for info, _ := range pool.readerMap {
		_, ok := seen[info.Info.Name]
		assert(!ok)
		seen[info.Info.Name] = true
	}
	return true
}

/*
Obtain the number of deleted docs for a pooled reader. If the reader
isn't being pooled, the segmentInfo's delCount is returned.
*/
func (pool *ReaderPool) numDeletedDocs(info *SegmentCommitInfo) int {
	// ensureOpen(false)
	delCount := info.DelCount()
	if rld := pool.get(info, false); rld != nil {
		delCount += rld.pendingDeleteCount()
	}
	return delCount
}

/*
returns a string description of the specified segments, for debugging.
*/
func (pool *ReaderPool) segmentsToString(infos []*SegmentCommitInfo) string {
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
func (pool *ReaderPool) segmentToString(info *SegmentCommitInfo) string {
	// TODO synchronized
	return info.StringOf(info.Info.Dir, pool.numDeletedDocs(info)-info.DelCount())
}
