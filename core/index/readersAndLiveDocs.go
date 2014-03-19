package index

import (
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"log"
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
