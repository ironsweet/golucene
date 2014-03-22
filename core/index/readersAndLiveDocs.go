package index

import (
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"log"
	"sync"
	"sync/atomic"
)

// index/ReadersAndLiveDocs.java

type refCountMixin struct {
	// Tracks how many consumers are using this instance:
	_refCount int32 // atomic, 1
}

func newRefCountMixin() *refCountMixin {
	return &refCountMixin{1}
}

func (rc *refCountMixin) incRef() {
	assert(atomic.AddInt32(&rc._refCount, 1) > 1)
}

func (rc *refCountMixin) decRef() {
	assert(atomic.AddInt32(&rc._refCount, -1) >= 0)
}

func (rc *refCountMixin) refCount() int {
	n := atomic.LoadInt32(&rc._refCount)
	assert(n >= 0)
	return int(n)
}

/*
Used by IndexWriter to hold open SegmentReaders (for searching or
merging), plus pending deletes, for a given segment.
*/
type ReadersAndLiveDocs struct {
	sync.Locker
	*refCountMixin

	info *SegmentInfoPerCommit

	// Set once (nil, and then maybe set, and never set again)
	_reader *SegmentReader

	// TODO: it's sometimes wasteful that we hold open two separate SRs
	// (one for merging one for reading)... maybe just use a single SR?
	// The gains of not loading the terms index (for merging in the
	// non-NRT case) are far less now... and if the app has any deletes
	// it'll open real readers anyway.
	// Set once (nil, and then maybe set, and never set again)
	mergeReader *SegmentReader

	// Holds the current shared (readable and writeable liveDocs). This
	// is nil when there are no deleted docs, and it's copy-on-write
	// (cloned whenever we need to change it but it's been shared to an
	// external NRT reader).
	_liveDocs util.Bits

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

// NOTE: removes callers ref
func (rld *ReadersAndLiveDocs) dropReaders() error {
	rld.Lock()
	defer rld.Unlock()

	// TODO: can we somehow use IOUtils here...?
	// problem is we are calling .decRef not .close)...
	err := func() (err error) {
		defer func() {
			if rld.mergeReader != nil {
				log.Printf("  pool.drop info=%v merge rc=%v", rld.info, rld.mergeReader.refCount)
				defer func() { rld.mergeReader = nil }()
				err2 := rld.mergeReader.decRef()
				if err == nil {
					err = err2
				} else {
					log.Printf("Escaped error: %v", err2)
				}
			}
		}()

		if rld._reader != nil {
			log.Printf("  pool.drop info=%v merge rc=%v", rld.info, rld._reader.refCount)
			defer func() { rld._reader = nil }()
			return rld._reader.decRef()
		}
		return nil
	}()
	if err != nil {
		return err
	}
	rld.decRef()
	return nil
}

func (rld *ReadersAndLiveDocs) liveDocs() util.Bits {
	rld.Lock()
	defer rld.Unlock()
	return rld._liveDocs
}

/*
Commit live docs to the directory (writes new _X_N.del files);
returns true if it wrote the file and false if there were no new
deletes to write:
*/
func (rld *ReadersAndLiveDocs) writeLiveDocs(dir store.Directory) (bool, error) {
	rld.Lock()
	defer rld.Unlock()

	log.Printf("rld.writeLiveDocs seg=%v pendingDelCount=%v", rld.info, rld._pendingDeleteCount)
	if rld._pendingDeleteCount != 0 {
		// We have new deletes
		assert(rld._liveDocs.Length() == int(rld.info.info.docCount))

		// Do this so we can delete any created files on error; this
		// saves all codecs from having to do it:
		trackingDir := newTrackingDirectoryWrapper(dir)

		// We can write directly to the actual name (vs to a .tmp &
		// renaming it) becaues the file is not live until segments file
		// is written:
		var success = false
		defer func() {
			if !success {
				// Advance only the nextWriteDelGen so that a 2nd attempt to
				// write will write to a new file
				rld.info.advanceNextWriteDelGen()

				// Dleete any prtially created files(s):
				for filename, _ := range trackingDir.createdFiles() {
					dir.DeleteFile(filename) // ignore error
				}
			}
		}()

		err := rld.info.info.codec.LiveDocsFormat().WriteLiveDocs(rld._liveDocs.(util.MutableBits),
			trackingDir, rld.info, rld._pendingDeleteCount, store.IO_CONTEXT_DEFAULT)
		if err != nil {
			return false, err
		}
		success = true

		// If we hit an error in the line above (e.g. disk full) then
		// info's delGen remains pointing to the previous (successfully
		// written) del docs:
		rld.info.advanceDelGen()
		rld.info.delCount += rld._pendingDeleteCount
		assert(rld.info.delCount <= int(rld.info.info.docCount))

		rld._pendingDeleteCount = 0
		return true, nil
	}
	return false, nil
}

func (rld *ReadersAndLiveDocs) String() string {
	panic("not implemented yet")
}
