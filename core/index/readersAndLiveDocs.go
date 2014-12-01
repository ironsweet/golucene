package index

import (
	. "github.com/balzaczyy/golucene/core/codec/spi"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"log"
	"sync"
	"sync/atomic"
)

// index/ReadersAndUpdates.java

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
merging), plus pending deletes and updates, for a given segment.
*/
type ReadersAndUpdates struct {
	sync.Locker
	*refCountMixin

	info *SegmentCommitInfo

	writer *IndexWriter

	// Set once (nil, and then maybe set, and never set again)
	_reader *SegmentReader

	// TODO: it's sometimes wasteful that we hold open two separate SRs
	// (one for merging one for reading)... maybe just use a single SR?
	// The gains of not loading the terms index (for merging in the
	// non-NRT case) are far less now... and if the app has any deletes
	// it'll open real readers anyway.
	// Set once (nil, and then maybe set, and never set again)
	mergeReader *SegmentReader

	// Holds the current shared (readable and writeable) liveDocs. This
	// is nil when there are no deleted docs, and it's copy-on-write
	// (cloned whenever we need to change it but it's been shared to an
	// external NRT reader).
	_liveDocs util.Bits

	// How many further deletions we;ve doen against
	// liveDocs vs when we loaded it or last wrote it:
	_pendingDeleteCount int

	// True if the current liveDOcs is reference dby an external NRT reader:
	liveDocsShared bool
}

func newReadersAndUpdates(writer *IndexWriter, info *SegmentCommitInfo) *ReadersAndUpdates {
	return &ReadersAndUpdates{info: info, writer: writer, liveDocsShared: true}
}

func (rld *ReadersAndUpdates) pendingDeleteCount() int {
	rld.Lock()
	defer rld.Unlock()
	return rld._pendingDeleteCount
}

/*
Get reader for searching/deleting
*/
func (rld *ReadersAndUpdates) reader(ctx store.IOContext) (*SegmentReader, error) {
	panic("not implemented yet")
}

func (rld *ReadersAndUpdates) release(sr *SegmentReader) error {
	panic("not implemented yet")
}

// NOTE: removes callers ref
func (rld *ReadersAndUpdates) dropReaders() error {
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

func (rld *ReadersAndUpdates) liveDocs() util.Bits {
	rld.Lock()
	defer rld.Unlock()
	return rld._liveDocs
}

/*
Commit live docs (writes new _X_N.del files) and field update (writes
new _X_N.del files) to the directory; returns true if it wrote any
file and false if there were no new deletes or updates to write:
*/
func (rld *ReadersAndUpdates) writeLiveDocs(dir store.Directory) (bool, error) {
	panic("not implemented yet")
	rld.Lock()
	defer rld.Unlock()

	log.Printf("rld.writeLiveDocs seg=%v pendingDelCount=%v", rld.info, rld._pendingDeleteCount)
	if rld._pendingDeleteCount != 0 {
		// We have new deletes
		assert(rld._liveDocs.Length() == rld.info.Info.DocCount())

		// Do this so we can delete any created files on error; this
		// saves all codecs from having to do it:
		trackingDir := store.NewTrackingDirectoryWrapper(dir)

		// We can write directly to the actual name (vs to a .tmp &
		// renaming it) becaues the file is not live until segments file
		// is written:
		var success = false
		defer func() {
			if !success {
				// Advance only the nextWriteDelGen so that a 2nd attempt to
				// write will write to a new file
				rld.info.AdvanceNextWriteDelGen()

				// Delete any partially created files(s):
				trackingDir.EachCreatedFiles(func(filename string) {
					dir.DeleteFile(filename) // ignore error
				})
			}
		}()

		err := rld.info.Info.Codec().(Codec).LiveDocsFormat().WriteLiveDocs(rld._liveDocs.(util.MutableBits),
			trackingDir, rld.info, rld._pendingDeleteCount, store.IO_CONTEXT_DEFAULT)
		if err != nil {
			return false, err
		}
		success = true

		// If we hit an error in the line above (e.g. disk full) then
		// info's delGen remains pointing to the previous (successfully
		// written) del docs:
		rld.info.AdvanceDelGen()
		rld.info.SetDelCount(rld.info.DelCount() + rld._pendingDeleteCount)
		assert(rld.info.DelCount() <= rld.info.Info.DocCount())

		rld._pendingDeleteCount = 0
		return true, nil
	}
	return false, nil
}

/* Writes field updates (new _X_N updates files) to the directory */
func (r *ReadersAndUpdates) writeFieldUpdates(dir store.Directory,
	dvUpdates *DocValuesFieldUpdatesContainer) error {
	panic("not implemented yet")
}

func (rld *ReadersAndUpdates) String() string {
	panic("not implemented yet")
}
