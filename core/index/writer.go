package index

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/analysis"
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"log"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// index/IndexCommit.java

/*
Expert: represents a single commit into an index as seen by the
IndexDeletionPolicy or IndexReader.

Changes to the content of an index are made visible only after the
writer who made that change commits by writing a new segments file
(segments_N). This point in time, when the action of writing of a new
segments file to the directory is completed, is an index commit.

Each index commit oint has a unique segments file associated with it.
The segments file associated with a later index commit point would
have a larger N.
*/
type IndexCommit interface {
	// Get the segments file (segments_N) associated with the commit point.
	SegmentsFileName() string
	// Returns all index files referenced by this commit point.
	FileNames() []string
	// Returns the Directory for the index.
	Directory() store.Directory
	/*
		Delete this commit point. This only applies when using the commit
		point in the context of IndexWriter's IndexDeletionPolicy.

		Upon calling this, the writer is notified that this commit point
		should be deleted.

		Decision that a commit-point should be deleted is taken by the
		IndexDeletionPolicy in effect and therefore this should only be
		called by its onInit() or onCommit() methods.
	*/
	Delete()
	// Returns true if this commit should be deleted; this is only used
	// by IndexWriter after invoking the IndexDeletionPolicy.
	IsDeleted() bool
	// returns number of segments referenced by this commit.
	SegmentCount() int
	// Returns the generation (the _N in segments_N) for this IndexCommit
	Generation() int64
	// Returns userData, previously passed to SetCommitData(map) for this commit.
	UserData() map[string]string
}

type IndexCommits []IndexCommit

func (s IndexCommits) Len() int      { return len(s) }
func (s IndexCommits) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s IndexCommits) Less(i, j int) bool {
	if s[i].Directory() != s[j].Directory() {
		panic("cannot compare IndexCommits from different Directory instances")
	}
	return s[i].Generation() < s[j].Generation()
}

// Used by search package to assign a default similarity
var DefaultSimilarity func() Similarity

// index/IndexWriter.java

// Use a seprate goroutine to protect closing control
type ClosingControl struct {
	_closed   bool // volatile
	_closing  bool // volatile
	isRunning bool
	closer    chan func() (bool, error)
	done      chan error
}

func newClosingControl() *ClosingControl {
	ans := &ClosingControl{
		isRunning: true,
		closer:    make(chan func() (bool, error)),
		done:      make(chan error),
	}
	go ans.daemon()
	return ans
}

func (cc *ClosingControl) daemon() {
	var err error
	for cc.isRunning {
		err = nil
		select {
		case f := <-cc.closer:
			log.Println("...closing...")
			if !cc._closed {
				cc._closing = true
				cc._closed, err = f()
				cc._closing = false
			}
			cc.done <- err
		}
	}
	log.Println("IW CC daemon is stopped.")
}

// Used internally to throw an AlreadyClosedError if this IndexWriter
// has been closed or is in the process of closing.
func (cc *ClosingControl) ensureOpen(failIfClosing bool) {
	assert2(!cc._closed && (!failIfClosing || !cc._closing), "this IndexWriter is closed")
}

func (cc *ClosingControl) close(f func() (ok bool, err error)) error {
	if cc._closed {
		return nil // already closed
	}
	cc.closer <- f
	log.Println("Closing IW...")
	return <-cc.done
}

const UNBOUNDED_MAX_MERGE_SEGMENTS = -1

/* Name of the write lock in the index. */
const WRITE_LOCK_NAME = "write.lock"

/* Source of a segment which results from a flush. */
const SOURCE_FLUSH = "flush"

/*
Absolute hard maximum length for a term, in bytes once encoded as
UTF8. If a term arrives from the analyzer longer than this length,
it is skipped and a message is printed to infoStream, if set.
*/
const MAX_TERM_LENGTH = MAX_TERM_LENGTH_UTF8

/*
An IndexWriter creates and maintains an index.

The OpenMode option on IndexWriterConfig.SetOpenMode() determines
whether a new index is created, or whether an existing index is
opened. Note that you can open an index with OPEN_MODE_CREATE even
while readers are using the index. The old readers will continue to
search the "point in time" snapshot they had opened, and won't see
the newly created index until they re-open. If OPEN_MODE_CREATE_OR_APPEND
is used, IndexWriter will create a new index if there is not already
an index at the provided path and otherwise open th existing index.

In either case, documents are added with AddDocument() and removed
with DeleteDocumentsByQuery(). A document can be updated with
UpdateDocuments() (which just deletes and then adds the entire
document). When finished adding, deleting and updating documents,
Close() should be called.

...
*/
type IndexWriter struct {
	sync.Locker
	*ClosingControl
	*MergeControl

	hitOOM bool // volatile

	directory store.Directory   // where this index resides
	analyzer  analysis.Analyzer // how to analyze text

	changeCount           int64 // volatile, increments every time a change is completed
	lastCommitChangeCount int64 // volatile, last changeCount that was committed

	rollbackSegments []*SegmentInfoPerCommit // list of segmentInfo we will fallback to if the commit fails

	pendingCommit            *SegmentInfos // set when a commit is pending (after prepareCommit() & before commit())
	pendingCommitChangeCount int64         // volatile

	filesToCommit []string

	segmentInfos         *SegmentInfos // the segments
	globalFieldNumberMap *model.FieldNumbers

	docWriter  *DocumentsWriter
	eventQueue *list.List
	deleter    *IndexFileDeleter

	// used by forceMerge to note those needing merging
	segmentsToMerge map[*SegmentInfoPerCommit]bool

	writeLock store.Lock

	mergePolicy     MergePolicy
	mergeScheduler  MergeScheduler
	mergeExceptions []*OneMerge

	flushCount        int32 // atomic
	flushDeletesCount int32 // atomic

	readerPool            *ReaderPool
	bufferedDeletesStream *BufferedDeletesStream

	bufferedDeletesStreamLock sync.Locker

	// This is a "write once" variable (like the organic dye on a DVD-R
	// that may or may not be heated by a laser and then cooled to
	// permanently record the event): it's false, until Reader() is
	// called for the first time, at which point it's switched to true
	// and never changes back to false. Once this is true, we hold open
	// and reuse SegmentReader instances internally for applying
	// deletes, doing merges, and reopening near real-time readers.
	poolReaders bool

	// The instance that we passed to the constructor. It is saved only
	// in order to allow users to query an IndexWriter settings.
	config *LiveIndexWriterConfig

	codec Codec // for writing new segments

	// If non-nil, information about merges will be printed to this.
	infoStream util.InfoStream

	// A hook for extending classes to execute operations after pending
	// and deleted documents have been flushed ot the Directory but
	// before the change is committed (new segments_N file written).
	doAfterFlush func() error
	// A hook for extending classes to execute operations before
	// pending added and deleted documents are flushed to the Directory.
	doBeforeFlush func() error

	// Used only by commit and prepareCommit, below; lock order is
	// commitLock -> IW
	commitLock sync.Locker

	// Ensures only one flush() is actually flushing segments at a time:
	fullFlushLock sync.Locker

	keepFullyDeletedSegments bool // test only
}

/*
Used internally to throw an AlreadyClosedError if this IndexWriter
has been closed or is in the process of closing.

Calls ensureOpen(true).
*/
func (w *IndexWriter) ensureOpen() {
	w.ClosingControl.ensureOpen(true)
}

/*
Constructs a new IndexWriter per the settings given in conf. If you want to
make "live" changes to this writer instance, use Config().

NOTE: after this writer is created, the given configuration instance cannot be
passed to another writer. If you intend to do so, you should clone it
beforehand.
*/
func NewIndexWriter(d store.Directory, conf *IndexWriterConfig) (w *IndexWriter, err error) {
	ans := &IndexWriter{
		Locker:         &sync.Mutex{},
		ClosingControl: newClosingControl(),

		segmentsToMerge: make(map[*SegmentInfoPerCommit]bool),
		mergeExceptions: make([]*OneMerge, 0),
		doAfterFlush:    func() error { return nil },
		doBeforeFlush:   func() error { return nil },
		commitLock:      &sync.Mutex{},
		fullFlushLock:   &sync.Mutex{},

		config:         newLiveIndexWriterConfigFrom(conf),
		directory:      d,
		analyzer:       conf.analyzer,
		infoStream:     conf.infoStream,
		mergePolicy:    conf.mergePolicy,
		mergeScheduler: conf.mergeScheduler,
		codec:          conf.codec,

		bufferedDeletesStream: newBufferedDeletesStream(conf.infoStream),
		poolReaders:           conf.readerPooling,

		bufferedDeletesStreamLock: &sync.Mutex{},

		writeLock: d.MakeLock(WRITE_LOCK_NAME),
	}
	ans.readerPool = newReaderPool(ans)
	ans.MergeControl = newMergeControl(conf.infoStream, ans.readerPool)

	conf.setIndexWriter(ans)
	ans.mergePolicy.SetIndexWriter(ans)

	// obtain write lock
	if ok, err := ans.writeLock.ObtainWithin(conf.writeLockTimeout); !ok || err != nil {
		if err != nil {
			return nil, err
		}
		return nil, errors.New(fmt.Sprintf("Index locked for write: %v", ans.writeLock))
	}

	var success bool = false
	defer func() {
		if !success {
			if ans.infoStream.IsEnabled("IW") {
				ans.infoStream.Message("IW", "init: hit exception on init; releasing write lock")
			}
			ans.writeLock.Release() // don't mask the original exception
			ans.writeLock = nil
		}
	}()

	var create bool
	switch conf.openMode {
	case OPEN_MODE_CREATE:
		create = true
	case OPEN_MODE_APPEND:
		create = false
	default:
		// CREATE_OR_APPEND - create only if an index does not exist
		ok, err := IsIndexExists(d)
		if err != nil {
			return nil, err
		}
		create = !ok
	}

	// If index is too old, reading the segments will return
	// IndexFormatTooOldError
	ans.segmentInfos = &SegmentInfos{}

	var initialIndexExists bool = true

	if create {
		// Try to read first. This is to allow create against an index
		// that's currently open for searching. In this case we write the
		// next segments_N file with no segments:
		err = ans.segmentInfos.ReadAll(d)
		if err == nil {
			ans.segmentInfos.Clear()
		} else {
			// Likely this means it's a fresh directory
			initialIndexExists = false
			err = nil
		}

		// Record that we have a change (zero out all segments) pending:
		ans.changed()
	} else {
		err = ans.segmentInfos.ReadAll(d)
		if err != nil {
			return
		}

		if commit := conf.commit; commit != nil {
			// Swap out all segments, but, keep metadta in SegmentInfos,
			// like version & generation, to preserve write-once. This is
			// important if readers are open against the future commit
			// points.
			assert2(commit.Directory() == d,
				"IndexCommit's directory doesn't match my directory")
			oldInfos := &SegmentInfos{}
			ans.segmentInfos.replace(oldInfos)
			ans.changed()
			ans.infoStream.Message("IW", "init: loaded commit '%v'",
				commit.SegmentsFileName())
		}
	}

	ans.rollbackSegments = ans.segmentInfos.createBackupSegmentInfos()

	// start with previous field numbers, but new FieldInfos
	ans.globalFieldNumberMap, err = ans.fieldNumberMap()
	if err != nil {
		return
	}
	ans.config.flushPolicy.init(ans.config)
	ans.docWriter = newDocumentsWriter(ans, ans.config, d)
	ans.eventQueue = ans.docWriter.events

	// Default deleter (for backwards compatibility) is
	// KeepOnlyLastCommitDeleter:
	ans.deleter, err = newIndexFileDeleter(d, conf.delPolicy,
		ans.segmentInfos, ans.infoStream, ans, initialIndexExists)
	if err != nil {
		return
	}

	if ans.deleter.startingCommitDeleted {
		// Deletion policy deleted the "head" commit point. We have to
		// mark outsef as changed so that if we are closed w/o any
		// further changes we write a new segments_N file.
		ans.changed()
	}

	if ans.infoStream.IsEnabled("IW") {
		ans.infoStream.Message("IW", "init: create=%v", create)
		ans.messageState()
	}

	success = true
	return ans, nil
}

func (w *IndexWriter) fieldInfos(info *model.SegmentInfo) (infos model.FieldInfos, err error) {
	var cfsDir store.Directory
	defer func() {
		if info.IsCompoundFile() && cfsDir != nil {
			err = mergeError(err, cfsDir.Close())
		}
	}()

	if info.IsCompoundFile() {
		cfsDir, err = store.NewCompoundFileDirectory(
			info.Dir,
			util.SegmentFileName(info.Name, "", store.COMPOUND_FILE_EXTENSION),
			store.IO_CONTEXT_READONCE,
			false,
		)
		if err != nil {
			return
		}
	} else {
		cfsDir = info.Dir
	}
	return info.Codec().(Codec).FieldInfosFormat().FieldInfosReader()(
		cfsDir, info.Name, store.IO_CONTEXT_READONCE)
}

/*
Loads or returns the alread loaded the global field number map for
this SegmentInfos. If this SegmentInfos has no global field number
map the returned instance is empty.
*/
func (w *IndexWriter) fieldNumberMap() (m *model.FieldNumbers, err error) {
	m = model.NewFieldNumbers()
	for _, info := range w.segmentInfos.Segments {
		fis, err := w.fieldInfos(info.info)
		if err != nil {
			return nil, err
		}
		for _, fi := range fis.Values {
			m.AddOrGet(fi)
		}
	}
	return m, nil
}

func (w *IndexWriter) messageState() {
	if w.infoStream.IsEnabled("IW") {
		w.infoStream.Message("IW", "\ndir=%v\nindex=%v\nversion=%v\n%v",
			w.directory, w.segString(), util.LUCENE_VERSION, w.config)
	}
}

/*
Commits all changes to an index, wait for pending merges to complete,
and closes all associate files.

This is a "slow graceful shutdown" which may take a long time
especially if a big merge is pending. If you only want to close
resources, use rollback(). If you only want to commit pending changes
and close resources, see closeAndWait().

Note that this may be a costly operation, so, try to re-use a single
writer instead of closing and opening a new one. See commit() for
caveats about write caching done by some IO devices.

If an error is hit during close, e.g., due to disk full or some other
reason, then both the on-disk index and the internal state of the
IndexWriter instance will be consistent. However, the close will not
be complete even though part of it (flushing buffered documents) may
have succeeded, so the write lock will still be held.

If you can correct the underlying cause (e.g., free up some disk
space) then you can call close() again. Failing that, if you want to
force the write lock to be released (dangerous, because ou may then
lose buffered docs in the IndexWriter instance) then you can do
something like this:

	defer func() {
		if IsDirectoryLocked(directory) {
			UnlockDIrectory(directory)
		}
	}
	err = writer.Close()

after which, you must be certain not ot use the writer instance
anymore.

NOTE: if this method hits a memory issue, you should immediately
close the writer, again. See above for details. But it's probably
impossible for GoLucene.
*/
func (w *IndexWriter) Close() error {
	return w.CloseAndWait(true)
}

/*
Closes the index with or without waiting for currently running merges
to finish. This is only meaningful when using a MergeScheduler that
runs  merges in background threads.

NOTE: if this method hits a memory issue, you should immediately
close the writer, again. See above for details. But it's probably
impossible for GoLucene.

NOTE: it is dangerous to always call closeAndWait(false), especially
when IndexWriter is not open for very long, because this can result
in "merge starvation" whereby long merges will never have a chance to
finish. This will cause too many segments in your index over time.
*/
func (w *IndexWriter) CloseAndWait(waitForMerge bool) error {
	// Ensure that only one goroutine actaully gets to do the closing,
	// and make sure no commit is also in progress:
	w.commitLock.Lock()
	defer w.commitLock.Unlock()
	return w.close(func() (ok bool, err error) {
		// If any methods have hit memory issue, then abort on
		// close, in case the internal state of IndexWriter or
		// DocumentsWriter is corrupt
		if w.hitOOM {
			return w.rollbackInternal()
		}
		return w.closeInternal(waitForMerge, true)
	})
}

/*
Returns true if this goroutine should attempt to close, or false if
IndexWriter is now closed; else, waits until another thread finishes
closing.
*/
func (w *IndexWriter) closeInternal(waitForMerges bool, doFlush bool) (ok bool, err error) {
	defer func() {
		if !ok {
			if w.infoStream.IsEnabled("IW") {
				w.infoStream.Message("IW", "hit error while closing")
			}
		}
	}()

	assert2(w.pendingCommit == nil,
		"cannot close: prepareCommit was already called with no corresponding call to commit")

	if w.infoStream.IsEnabled("IW") {
		w.infoStream.Message("IW", "now flush at close waitForMerges=%v", waitForMerges)
	}

	w.docWriter.close()

	err = w.closeInternalFlush(waitForMerges, doFlush)
	if err != nil {
		return false, err
	}

	if w.infoStream.IsEnabled("IW") {
		w.infoStream.Message("IW", "now call final commit()")
	}

	if doFlush {
		err = w.commitInternal()
		if err != nil {
			return false, err
		}
	}

	// commitInternal calls ReaderPool.commit, which writes any pending
	// liveDocs from ReaderPool, so it's safe to drop all readers now:
	err = w.readerPool.dropAll(true)
	if err != nil {
		return false, err
	}
	w.deleter.Close() // no error

	// used by assert below
	oldWriter := w.docWriter
	w.docWriter = nil

	if w.infoStream.IsEnabled("IW") {
		w.infoStream.Message("IW", "at close: %v", w.segString())
	}

	if w.writeLock != nil {
		err = w.writeLock.Release()
		if err != nil {
			return false, err
		}
		w.writeLock = nil
	}

	ok = true

	oldWriter.perThreadPool.foreach(func(state *ThreadState) {
		assert(!state.isActive)
	})

	return ok, nil
}

func (w *IndexWriter) closeInternalFlush(waitForMerges, doFlush bool) (err error) {
	defer func() {
		err2 := w.closeInternalCleanup(waitForMerges)
		if err != nil && err2 != nil {
			log.Printf("Flush failed and error hidden: %v", err)
		}
		err = err2
	}()

	// Only allow a new merge to be triggered if we are going to wait
	// for merges:
	if doFlush {
		err = w.flush(waitForMerges, true)
	} else {
		w.docWriter.abort(w) // already closed -- never sync on IW
	}
	return
}

func (w *IndexWriter) closeInternalCleanup(waitForMerges bool) error {
	defer func() {
		// shutdown policy, scheduler and all threads (this call is not
		// interruptible):
		util.CloseWhileSuppressingError(w.mergePolicy, w.mergeScheduler)
	}()

	// clean up merge scheduler in all cases, although flushing may have failed:
	if waitForMerges {
		err := w.mergeScheduler.Merge(w)
		if err != nil {
			return err
		}
		w.waitForMerges()
	} else {
		w.abortAllMerges()
	}
	w.stopMerges = true
	return nil
}

// Retuns the Directory used by this index.
func (w *IndexWriter) Directory() store.Directory {
	return w.directory
}

// L1201
/*
Adds a document to this index.

Note that if an Error is hit (for example disk full) then the index
will be consistent, but this document may not have been added.
Furthermore, it's possible the index will have one segment in
non-compound format even when using compound files (when a merge has
partially succeeded).

This method periodically flushes pending documents to the Directory
(see flush()), and also periodically triggers segment merges in the
index according to the MergePolicy in use.

Merges temporarily consume space in the directory. The amount of
space required is up to 1X the size of all segments being merged,
when no readers/searchers are open against the index, and up to 2X
the size of all segments being merged when readers/searchers are open
against the index (see forceMerge() for details). The sequence of
primitive merge operations performed is governed by the merge policy.

Note that each term in the document can be no longer than 16383
characters, otherwise error will be returned.

Note that it's possible to creat an invalid Unicode string in Java if
a UTF16 surrogate pair is malformed. In this case, the invalid
characters are silently replaced with the Unicode replacement
character U+FFFD.

NOTE: if this method hits a memory issue, you should immediately
close the writer. See above for details.
*/
func (w *IndexWriter) AddDocument(doc []model.IndexableField) error {
	return w.AddDocumentWithAnalyzer(doc, w.analyzer)
}

/*
Adds a document to this index, using the provided analyzer instead of
the value of Analyzer().

See AddDocument() for details on index and IndexWriter state after an
error, and flushing/merging temporary free space requirements.

NOTE: if this method hits a memory issue, you hsould immediately
close the writer. See above for details.
*/
func (w *IndexWriter) AddDocumentWithAnalyzer(doc []model.IndexableField, analyzer analysis.Analyzer) error {
	return w.UpdateDocument(nil, doc, analyzer)
}

// L1545
/*
Updates a document by first deleting the document(s) containing term
and then adding the new document. The delete and then add are atomic
as seen by a reader on the same index (flush may happen only after
the add).

NOTE: if this method hits a memory issue, you should immediately
close he write. See above for details.
*/
func (w *IndexWriter) UpdateDocument(term *Term, doc []model.IndexableField, analyzer analysis.Analyzer) error {
	w.ensureOpen()
	var success = false
	defer func() {
		if !success {
			if w.infoStream.IsEnabled("IW") {
				w.infoStream.Message("IW", "hit error updating document")
			}
		}
	}()

	ok, err := w.docWriter.updateDocument(doc, analyzer, term)
	if err != nil {
		return err
	}
	if ok {
		_, err = w.docWriter.processEvents(w, true, false)
		if err != nil {
			return err
		}
	}
	success = true
	return nil
}

func (w *IndexWriter) newSegmentName() string {
	// Cannot synchronize on IndexWriter because that causes deadlook
	// Ian: but why?
	w.Lock()
	defer w.Unlock()
	// Important to increment changeCount so that the segmentInfos is
	// written on close. Otherwise we could close, re-open and
	// re-return the same segment name that was previously returned
	// which can cause problems at least with ConcurrentMergeScheculer.
	w.changeCount++
	w.segmentInfos.changed()
	defer func() { w.segmentInfos.counter++ }()
	return fmt.Sprintf("_%v", strconv.FormatInt(int64(w.segmentInfos.counter), 36))
}

/*
Forces merge policy to merge segments until there are <=
maxNumSegments. The actual merge to be executed are determined by the
MergePolicy.

This is a horribly costly operation, especially when you pass a small
maxNumSegments; usually you should only call this if the index is
static (will no longer be changed).

Note that this requires up to 2X the index size free space in your
Directory (3X if you're using compound file format). For example, if
your index size is 10 MB, then you need up to 20 MB free for this to
complete (30 MB if you're using compound file format). Also, it's
best to call commit() afterwards, to allow IndexWriter to free up
disk space.

If some but not all readers re-open while merging is underway, this
will cause > 2X temporary space to be consumed as those new readers
will then hold open the temporary segments at that time. it is best
not to re-open readers while merging is running.

The actual temporary usage could be much less than these figures (it
depends on many factors).

In general, once this completes, the total size of the index will be
less than the size of the starting index. It could be quite a bit
smaller (if there were many pending deletes) or just slightly smaller.

If an error is hit, for example, due to disk full, the index will not
be corrupted and no documents will be list. However, it may have been
partially merged (some segments were merged but not all), and it's
possible that one of the segments in the index will be in
non-compound format even when using compound file format. This will
occur when the error is hit during conversion of the segment into
compound format.

This call will merge those segments present in the index when call
started. If other routines are still adding documents and flushing
segments, those newly created segments will not be merged unless you
call forceMerge again.

NOTE: if this method hits a memory issue, you should immediately
close the writer.

NOTE: if you call CloseAndWait() with false, which aborts all running
merges, then any routine still running this method might hit a
MergeAbortedError.
*/
func (w *IndexWriter) forceMerge(maxNumSegments int) error {
	return w.forceMergeAndWait(maxNumSegments, true)
}

/*
Just like forceMerge(), except you can specify whether the call
should block until all merging completes. This is only meaningful
with  a Mergecheduler that is able to run merges in background
routines.

NOTE: if this method hits a memory issue, you should immediately
close the writer.
*/
func (w *IndexWriter) forceMergeAndWait(maxNumSegments int, doWait bool) error {
	panic("not implemented yet")
}

// Returns true if any merges in pendingMerges or runningMerges
// are maxNumSegments merges.
func (w *IndexWriter) maxSegmentsMergePending() bool {
	w.Lock() // synchronized
	defer w.Unlock()

	panic("not implemented yet")
}

func (w *IndexWriter) maybeMerge(trigger MergeTrigger, maxNumSegments int) (err error) {
	w.ClosingControl.ensureOpen(false)
	if err = w.updatePendingMerges(trigger, maxNumSegments); err == nil {
		err = w.mergeScheduler.Merge(w)
	}
	return
}

func (w *IndexWriter) updatePendingMerges(trigger MergeTrigger, maxNumSegments int) error {
	w.Lock() // synchronized
	defer w.Unlock()
	assert(maxNumSegments == -1 || maxNumSegments > 0)
	if w.stopMerges {
		return nil
	}

	var err error
	var spec MergeSpecification
	if maxNumSegments != UNBOUNDED_MAX_MERGE_SEGMENTS {
		assertn(trigger == MERGE_TRIGGER_EXPLICIT || trigger == MERGE_FINISHED,
			"Expected EXPLIT or MEGE_FINISHED as trigger even with maxNumSegments set but was: %v",
			MergeTriggerName(trigger))
		spec, err = w.mergePolicy.FindForcedMerges(
			w.segmentInfos,
			maxNumSegments,
			w.segmentsToMerge)
		if err != nil {
			return err
		}
		if spec != nil {
			for _, merge := range spec {
				merge.maxNumSegments = maxNumSegments
			}
		}
	} else {
		spec, err = w.mergePolicy.FindMerges(trigger, w.segmentInfos)
		if err != nil {
			return err
		}
	}

	if spec != nil {
		for _, merge := range spec {
			_, err = w.registerMerge(merge)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

/*
Experts: to be used by a MergePolicy to avoid selecting merges for
segments already being merged. The returned collection is not cloned,
and thus is only safe to access if you hold IndexWriter's lock (which
you do when IndexWriter invokes the MergePolicy).
*/
func (w *IndexWriter) MergingSegments() map[*SegmentInfoPerCommit]bool {
	// no need to synchronized but should be
	return w.mergingSegments
}

/*
Expert: the MergeScheduler calls this method to retrieve the next
merge requested by the MergePolicy.
*/
func (w *IndexWriter) nextMerge() *OneMerge {
	w.Lock() // synchronized
	defer w.Unlock()

	if w.pendingMerges.Len() == 0 {
		return nil
	}
	// Advance the merge from pending to running
	merge := w.pendingMerges.Front().Value.(*OneMerge)
	w.pendingMerges.Remove(w.pendingMerges.Front())
	w.runningMerges[merge] = true
	return merge
}

// Expert: returns true if there are merges waiting to be scheduled.
func (w *IndexWriter) hasPendingMerges() bool {
	return w.pendingMerges.Len() > 0
}

/*
Close the IndexWriter without committing any changes that have
occurred since the last commit (or since it was opened, if commit
hasn't been called). This removes any temporary files that had been
created, after which the state of the index will be the same as it
was when commit() was last called or when this writer was first
opened. This also clears a previous call to prepareCommit()
*/
func (w *IndexWriter) Rollback() error {
	w.ensureOpen()
	return w.close(w.rollbackInternal)
}

func (w *IndexWriter) rollbackInternal() (ok bool, err error) {
	if w.infoStream.IsEnabled("IW") {
		w.infoStream.Message("IW", "rollback")
	}

	err = func() error {
		var success = false
		defer func() {
			if w.infoStream.IsEnabled("IW") {
				w.infoStream.Message("IW", "hit error during rollback")
			}
		}()

		func() {
			w.Lock()
			defer w.Unlock()

			w.finishMerges(false)
			w.stopMerges = true
		}()

		if w.infoStream.IsEnabled("IW") {
			w.infoStream.Message("IW", "rollback: done finish merges")
		}

		// Must pre-close these two, in case they increment changeCount
		// so that we can then set it to false before calling closeInternal
		w.mergePolicy.Close()
		err = w.mergeScheduler.Close()
		if err != nil {
			return err
		}

		w.bufferedDeletesStream.clear()
		_, err = w.docWriter.processEvents(w, false, true)
		if err != nil {
			return err
		}
		w.docWriter.close()  // mark it as closed first to prevent subsequent indexing actions/flushes
		w.docWriter.abort(w) // don't sync on IW here

		err = func() error {
			w.Lock()
			defer w.Unlock()

			if w.pendingCommit != nil {
				w.pendingCommit.rollbackCommit(w.directory)
				w.deleter.decRefInfos(w.pendingCommit)
				w.pendingCommit = nil
			}

			// Don't bother saving any changes in our segmentInfos
			err = w.readerPool.dropAll(false)
			if err != nil {
				return err
			}

			// Keep the same segmentInfos instance but replace all of its
			// SegmentInfo instances. This is so the next attempt to commit
			// using this instance of IndexWriter will always write to a
			// new generation ("write once").
			w.segmentInfos.rollbackSegmentInfos(w.rollbackSegments)
			if w.infoStream.IsEnabled("IW") {
				w.infoStream.Message("IW", "rollback: infos=%v", w.readerPool.segmentsToString(w.segmentInfos.Segments))
			}

			w.testPoint("rollback before checkpoint")

			// Ask deleter to locate unreferenced files & remove them:
			err = w.deleter.checkpoint(w.segmentInfos, false)
			if err == nil {
				err = w.deleter.refreshList()
			}

			success = err != nil
			return err
		}()
		if err != nil {
			return err
		}

		success = true
		return nil
	}()

	if err == nil {
		ok, err = w.closeInternal(false, false)
	}
	return
}

/*
Called whenever the SegmentInfos has been updatd and the index files
referenced exist (correctly) in the index directory.
*/
func (w *IndexWriter) checkpoint() error {
	w.Lock() // synchronized
	defer w.Unlock()
	return w._checkpoint()
}

func (w *IndexWriter) _checkpoint() error {
	w.changeCount++
	w.segmentInfos.changed()
	return w.deleter.checkpoint(w.segmentInfos, false)
}

/*
Checkpoints with IndexFileDeleter, so it's aware of new files, and
increments changeCount, so on close/commit we will write a new
segments file, but does NOT bump segmentInfos.version.
*/
func (w *IndexWriter) checkpointNoSIS() (err error) {
	w.Lock() // synchronized
	defer w.Unlock()
	w.changeCount++
	return w.deleter.checkpoint(w.segmentInfos, false)
}

/* Called internally if any index state has changed. */
func (w *IndexWriter) changed() {
	w.Lock()
	defer w.Unlock()
	w.changeCount++
	w.segmentInfos.changed()
}

func (w *IndexWriter) publishFrozenDeletes(packet *FrozenBufferedDeletes) {
	w.Lock()
	defer w.Unlock()
	assert(packet != nil && packet.any())
	w.bufferedDeletesStreamLock.Lock()
	defer w.bufferedDeletesStreamLock.Unlock()
	w.bufferedDeletesStream.push(packet)
}

/*
Atomically adds the segment private delete packet and publishes the
flushed segments SegmentInfo to the index writer.
*/
func (w *IndexWriter) publishFlushedSegment(newSegment *SegmentInfoPerCommit,
	packet *FrozenBufferedDeletes, globalPacket *FrozenBufferedDeletes) (err error) {
	defer func() {
		atomic.AddInt32(&w.flushCount, 1)
		err = mergeError(err, w.doAfterFlush())
	}()

	// Lock order IW -> BDS
	w.Lock()
	defer w.Unlock()
	w.bufferedDeletesStreamLock.Lock()
	defer w.bufferedDeletesStreamLock.Unlock()

	if w.infoStream.IsEnabled("IW") {
		w.infoStream.Message("IW", "publishFlushedSegment")
	}

	if globalPacket != nil && globalPacket.any() {
		w.bufferedDeletesStream.push(globalPacket)
	}
	// Publishing the segment must be synched on IW -> BDS to make sure
	// that no merge prunes away the seg. private delete packet
	var nextGen int64
	if packet != nil && packet.any() {
		nextGen = w.bufferedDeletesStream.push(packet)
	} else {
		// Since we don't have a delete packet to apply we can get a new
		// generation right away
		nextGen = w.bufferedDeletesStream.nextGen
	}
	if w.infoStream.IsEnabled("IW") {
		w.infoStream.Message("IW", "publish sets newSegment delGen=%v seg=%v", nextGen, w.readerPool.segmentToString(newSegment))
	}
	newSegment.setBufferedDeletesGen(nextGen)
	w.segmentInfos.Segments = append(w.segmentInfos.Segments, newSegment)
	return w._checkpoint()
}

func (w *IndexWriter) resetMergeExceptions() {
	w.Lock() // synchronized
	defer w.Unlock()
	panic("not implemented yet")
}

/*
Requires commitLock
*/
func (w *IndexWriter) prepareCommitInternal() error {
	w.ClosingControl.ensureOpen(false)
	if w.infoStream.IsEnabled("IW") {
		w.infoStream.Message("IW", "prepareCommit: flush")
		w.infoStream.Message("IW", "  index before flush %v", w.segString())
	}

	assert2(!w.hitOOM, "this writer hit an OOM; cannot commit")
	assert2(w.pendingCommit == nil, "prepareCommit was already called with no corresponding call to commit")

	err := w.doBeforeFlush()
	if err != nil {
		return err
	}
	w.testPoint("startDoFlush")

	// This is copied from doFLush, except it's modified to clone &
	// incRef the flushed SegmentInfos inside the sync block:

	toCommit, anySegmentsFlushed, err := func() (toCommit *SegmentInfos, anySegmentsFlushed bool, err error) {
		w.fullFlushLock.Lock()
		defer w.fullFlushLock.Unlock()

		var flushSuccess = false
		var success = false
		defer func() {
			if !success {
				if w.infoStream.IsEnabled("IW") {
					w.infoStream.Message("IW", "hit error during prepareCommit")
				}
			}
			// Done: finish the full flush!
			w.docWriter.finishFullFlush(flushSuccess)
			err2 := w.doAfterFlush()
			if err2 != nil {
				log.Printf("Error in doAfterFlush: %v", err2)
			}
		}()

		anySegmentsFlushed, err = w.docWriter.flushAllThreads(w)
		if err != nil {
			return
		}
		if !anySegmentsFlushed {
			// prevent double increment since docWriter.doFlush increments
			// the flushCount if we flushed anything.
			atomic.AddInt32(&w.flushCount, -1)
		}
		w.docWriter.processEvents(w, false, true)
		flushSuccess = true

		err = func() (err error) {
			w.Lock()
			defer w.Unlock()

			err = w._maybeApplyDeletes(true)
			if err != nil {
				return
			}

			err = w.readerPool.commit(w.segmentInfos)
			if err != nil {
				return
			}

			// Must clone the segmentInfos while we still
			// hold fullFlushLock and while sync'd so that
			// no partial changes (eg a delete w/o
			// corresponding add from an updateDocument) can
			// sneak into the commit point:
			toCommit = w.segmentInfos.Clone()

			w.pendingCommitChangeCount = w.changeCount

			// This protects the segmentInfos we are now going
			// to commit.  This is important in case, eg, while
			// we are trying to sync all referenced files, a
			// merge completes which would otherwise have
			// removed the files we are now syncing.
			w.filesToCommit = toCommit.files(w.directory, false)
			w.deleter.incRefFiles(w.filesToCommit)
			return
		}()
		if err != nil {
			return
		}
		success = true
		return
	}()

	var success = false
	defer func() {
		if !success {
			func() {
				w.Lock()
				defer w.Unlock()
				w.deleter.decRefFiles(w.filesToCommit)
				w.filesToCommit = nil
			}()
		}
	}()
	if anySegmentsFlushed {
		err := w.maybeMerge(MERGE_TRIGGER_FULL_FLUSH, UNBOUNDED_MAX_MERGE_SEGMENTS)
		if err != nil {
			return err
		}
	}
	success = true

	return w.startCommit(toCommit)
}

/*
Commits all pending changes (added & deleted documents, segment
merges, added indexes, etc.) to the index, and syncs all referenced
index files, such that a reader will see the changes and the index
updates will survive an OS or machine crash or power loss. Note that
this does not wait for any running background merges to finish. This
may be a costly operation, so you should test the cost in your
application and do it only when really necessary.

Note that this operation calls Directory.sync on the index files.
That call  should not return until the file contents & metadata are
on stable storage. For FSDirectory, this calls the OS's fsync. But,
beware: some hardware devices may in fact cache writes even during
fsync, and return before the bits are actually on stable storage, to
give the appearance of faster performance. If you have such a device,
and it does not hav a battery backup (for example) then on power loss
it may still lose data. Lucene cannot guarantee consistency on such
devices.

NOTE: if this method hits a memory issue, you should immediately
close the writer.
*/
func (w *IndexWriter) Commit() error {
	w.ensureOpen()
	w.commitLock.Lock()
	defer w.commitLock.Unlock()
	return w.commitInternal()
}

/*
Assume commitLock is locked.
*/
func (w *IndexWriter) commitInternal() error {
	if w.infoStream.IsEnabled("IW") {
		w.infoStream.Message("IW", "commit: start")
	}

	w.ClosingControl.ensureOpen(false)

	if w.infoStream.IsEnabled("IW") {
		w.infoStream.Message("IW", "commit: enter lock")
	}

	if w.pendingCommit == nil {
		if w.infoStream.IsEnabled("IW") {
			w.infoStream.Message("IW", "commit: now prepare")
		}
		err := w.prepareCommitInternal()
		if err != nil {
			return err
		}
	} else {
		if w.infoStream.IsEnabled("IW") {
			w.infoStream.Message("IW", "commit: already prepared")
		}
	}
	return w.finishCommit()
}

func (w *IndexWriter) finishCommit() error {
	w.Lock() // synchronized
	defer w.Unlock()

	if w.pendingCommit == nil {
		if w.infoStream.IsEnabled("IW") {
			w.infoStream.Message("IW", "commit: pendingCommit == nil; skip")
			w.infoStream.Message("IW", "commit: done")
		}
		return nil
	}

	defer func() {
		// Matches the incRef done in prepareCommit:
		w.deleter.decRefFiles(w.filesToCommit)
		w.filesToCommit = nil
		w.pendingCommit = nil
		// TODO check if any wait()
	}()

	if w.infoStream.IsEnabled("IW") {
		w.infoStream.Message("IW", "commit: pendingCommit != nil")
	}
	err := w.pendingCommit.finishCommit(w.directory)
	if err != nil {
		return err
	}
	if w.infoStream.IsEnabled("IW") {
		w.infoStream.Message("IW", "commit: wrote segments file '%v'", w.pendingCommit.SegmentsFileName())
	}
	w.segmentInfos.updateGeneration(w.pendingCommit)
	w.lastCommitChangeCount = w.pendingCommitChangeCount
	w.rollbackSegments = w.pendingCommit.createBackupSegmentInfos()
	// NOTE: don't use this.checkpoint() here, because
	// we do not want to increment changeCount:
	return w.deleter.checkpoint(w.pendingCommit, true)
}

/*
Flush all in-memory buffered updates (adds and deletes) to the
Directory.
*/
func (w *IndexWriter) flush(triggerMerge bool, applyAllDeletes bool) error {
	// NOTE: this method cannot be sync'd because
	// maybeMerge() in turn calls mergeScheduler.merge which
	// in turn can take a long time to run and we don't want
	// to hold the lock for that.  In the case of
	// ConcurrentMergeScheduler this can lead to deadlock
	// when it stalls due to too many running merges.

	// We can be called during close, when closing==true, so we must pass false to ensureOpen:
	w.ClosingControl.ensureOpen(false)
	ok, err := w.doFlush(applyAllDeletes)
	if err != nil {
		return err
	}
	if ok && triggerMerge {
		return w.maybeMerge(MERGE_TRIGGER_FULL_FLUSH, UNBOUNDED_MAX_MERGE_SEGMENTS)
	}
	return nil
}

func (w *IndexWriter) doFlush(applyAllDeletes bool) (bool, error) {
	assert2(!w.hitOOM, "this writer hit an OutOfMemoryError; cannot flush")

	err := w.doBeforeFlush()
	if err != nil {
		return false, err
	}
	if w.infoStream.IsEnabled("TP") {
		w.infoStream.Message("TP", "startDoFlush")
	}

	success := false
	defer func() {
		if !success && w.infoStream.IsEnabled("IW") {
			w.infoStream.Message("IW", "hit error during flush")
		}
	}()

	if w.infoStream.IsEnabled("IW") {
		w.infoStream.Message("IW", "  start flush: applyAllDeletes=%v", applyAllDeletes)
		w.infoStream.Message("IW", "  index before flush %v", w.segString())
	}

	anySegmentFlushed, err := func() (ok bool, err error) {
		w.fullFlushLock.Lock()
		defer w.fullFlushLock.Unlock()

		flushSuccess := false
		defer func() {
			w.docWriter.finishFullFlush(flushSuccess)
			w.docWriter.processEvents(w, false, true)
		}()

		if ok, err = w.docWriter.flushAllThreads(w); err == nil {
			flushSuccess = true
		}
		return
	}()
	if err != nil {
		return false, err
	}

	err = func() error {
		w.Lock()
		defer w.Unlock()
		err := w._maybeApplyDeletes(applyAllDeletes)
		if err != nil {
			return err
		}
		err = w.doAfterFlush()
		if err != nil {
			return err
		}
		if !anySegmentFlushed {
			//flushCount is incremented in flushAllThreads
			atomic.AddInt32(&w.flushCount, 1)
		}
		return nil
	}()
	if err != nil {
		return false, err
	}

	success = true
	return anySegmentFlushed, nil
}

func (w *IndexWriter) _maybeApplyDeletes(applyAllDeletes bool) error {
	if applyAllDeletes {
		if w.infoStream.IsEnabled("IW") {
			w.infoStream.Message("IW", "apply all deletes during flush")
		}
		return w._applyAllDeletes()
	} else if w.infoStream.IsEnabled("IW") {
		w.infoStream.Message("IW", "don't apply deletes now delTermCount=%v bytesUsed=%v",
			atomic.LoadInt32(&w.bufferedDeletesStream.numTerms),
			atomic.LoadInt64(&w.bufferedDeletesStream.bytesUsed))
	}
	return nil
}

func (w *IndexWriter) applyAllDeletes() error {
	w.Lock() // synchronized
	defer w.Unlock()
	return w._applyAllDeletes()
}

func (w *IndexWriter) _applyAllDeletes() error {
	atomic.AddInt32(&w.flushDeletesCount, 1)
	result, err := w.bufferedDeletesStream.applyDeletes(w.readerPool, w.segmentInfos.Segments)
	if err != nil {
		return err
	}
	if result.anyDeletes {
		err = w.checkpoint()
		if err != nil {
			return err
		}
	}
	if !w.keepFullyDeletedSegments && result.allDeleted != nil {
		if w.infoStream.IsEnabled("IW") {
			w.infoStream.Message("IW", "drop 100%% deleted segments: %v",
				w.readerPool.segmentsToString(result.allDeleted))
		}
		for _, info := range result.allDeleted {
			// If a merge has already registered for this segment, we leave
			// it in the readerPool; the merge will skip merging it and
			// will then drop it once it's done:
			if _, ok := w.mergingSegments[info]; !ok {
				w.segmentInfos.remove(info)
				err = w.readerPool.drop(info)
				if err != nil {
					return err
				}
			}
		}
		err = w.checkpoint()
		if err != nil {
			return err
		}
	}
	w.bufferedDeletesStream.prune(w.segmentInfos)
	return nil
}

// L3440
/*
Merges the indicated segments, replacing them in the stack with a
single segment.
*/
func (w *IndexWriter) merge(merge *OneMerge) error {
	panic("not implemented yet")
}

/*
Checks whether this merge involves any segments already participating
in a merge. If not, this merge is "registered", meaning we record
that its semgents are now participating in a merge, and true is
returned. Else (the merge conflicts) false is returned.
*/
func (w *IndexWriter) registerMerge(merge *OneMerge) (bool, error) {
	panic("not implemented yet")
}

func setDiagnostics(info *model.SegmentInfo, source string) {
	setDiagnosticsAndDetails(info, source, nil)
}

func setDiagnosticsAndDetails(info *model.SegmentInfo, source string, details map[string]string) {
	ans := map[string]string{
		"source":         source,
		"lucene.version": util.LUCENE_VERSION,
		"os":             runtime.GOOS,
		"os.arch":        runtime.GOARCH,
		"go.version":     runtime.Version(),
		"timestamp":      fmt.Sprintf("%v", time.Now().Unix()),
	}
	if details != nil {
		for k, v := range details {
			ans[k] = v
		}
	}
	info.SetDiagnostics(ans)
}

// Returns a string description of all segments, for debugging.
func (w *IndexWriter) segString() string {
	// TODO synchronized
	return w.readerPool.segmentsToString(w.segmentInfos.Segments)
}

// called only from assert
func (w *IndexWriter) assertFilesExist(toSync *SegmentInfos) {
	files := toSync.files(w.directory, false)
	for _, filename := range files {
		assertn(w.directory.FileExists(filename), "file %v does not exist", filename)
		// If this trips it means we are missing a call to checkpoint
		// somewhere, because by the time we are called, deleter should
		// know about every file referenced by the current head
		// segmentInfos:
		assertn(w.deleter.exists(filename), "IndexFileDeleter doesn't know about file %v", filename)
	}
}

/* For infoStream output */
func (w *IndexWriter) toLiveInfos(sis *SegmentInfos) *SegmentInfos {
	w.Lock() // synchronized
	defer w.Unlock()
	return w._toLiveInfos(sis)
}

func (w *IndexWriter) _toLiveInfos(sis *SegmentInfos) *SegmentInfos {
	newSIS := new(SegmentInfos)
	// liveSIS := make(map[*SegmentInfoPerCommit]bool)
	// for _, info := range w.segmentInfos.Segments {
	// 	liveSIS[info] = true
	// }
	for _, info := range sis.Segments {
		// if _, ok :=  liveSIS[info] ; ok {
		newSIS.Segments = append(newSIS.Segments, info)
		// }
	}
	return newSIS
}

/*
Walk through all files referenced by the current segmentInfos and ask
the  Directory to sync each file, if it wans't already. If that
succeeds, then we prepare a new segments_N file but do not fully
commit it.
*/
func (w *IndexWriter) startCommit(toSync *SegmentInfos) error {
	w.testPoint("startStartCommit")
	assert(w.pendingCommit == nil)
	assert2(!w.hitOOM, "this writer hit an OutOfMemoryError; cannot commit")

	if w.infoStream.IsEnabled("IW") {
		w.infoStream.Message("IW", "startCommit(): start")
	}

	func() {
		w.Lock()
		defer w.Unlock()

		assertn(w.lastCommitChangeCount <= w.changeCount,
			"lastCommitChangeCount=%v changeCount=%v", w.lastCommitChangeCount, w.changeCount)
		if w.pendingCommitChangeCount == w.lastCommitChangeCount {
			if w.infoStream.IsEnabled("IW") {
				w.infoStream.Message("IW", "  skip startCommit(): no changes pending")
			}
			w.deleter.decRefFiles(w.filesToCommit)
			w.filesToCommit = nil
			return
		}

		if w.infoStream.IsEnabled("IW") {
			w.infoStream.Message("IW", "startCommit index=%v changeCount=%v",
				w.readerPool.segmentsToString(toSync.Segments), w.changeCount)
		}

		w.assertFilesExist(toSync)
	}()

	w.testPoint("midStartCommit")

	var pendingCommitSet = false
	defer func() {
		w.Lock()
		defer w.Unlock()

		// Have out master segmentInfos record the generations we just
		// prepared. We do this on error or success so we don't
		// double-write a segments_N file.
		w.segmentInfos.updateGeneration(toSync)

		if !pendingCommitSet {
			if w.infoStream.IsEnabled("IW") {
				w.infoStream.Message("IW", "hit error committing segments file")
			}

			// Hit error
			w.deleter.decRefFiles(w.filesToCommit)
			w.filesToCommit = nil
		}
	}()

	w.testPoint("midStartCommit2")
	err := func() (err error) {
		w.Lock()
		defer w.Unlock()

		assert(w.pendingCommit == nil)
		assert(w.segmentInfos.generation == toSync.generation)

		// Eror here means nothing is prepared (this method unwinds
		// everything it did on an error)
		err = toSync.prepareCommit(w.directory)
		if err != nil {
			return err
		}
		log.Print("DONE prepareCommit")

		pendingCommitSet = true
		w.pendingCommit = toSync
		return nil
	}()
	if err != nil {
		return err
	}

	// This call can take a long time -- 10s of seconds or more. We do
	// it without syncing on this:
	var success = false
	var filesToSync []string
	defer func() {
		if !success {
			pendingCommitSet = false
			w.pendingCommit = nil
			toSync.rollbackCommit(w.directory)
		}
	}()

	filesToSync = toSync.files(w.directory, false)
	err = w.directory.Sync(filesToSync)
	if err != nil {
		return err
	}
	success = true

	if w.infoStream.IsEnabled("IW") {
		w.infoStream.Message("IW", "done all syncs: %v", filesToSync)
	}

	w.testPoint("midStartCommitSuccess")
	w.testPoint("finishStartCommit")
	return nil
}

/*
Used only  by assert for testing. Current points:
- startDoFlush
- startCommitMerge
- startStartCommit
- midStartCommit
- midStartCommit2
- midStartCommitSuccess
- finishStartCommit
- startCommitMergeDeletes
- startMergeInit
- DocumentsWriter.ThreadState.init start
*/
func (w *IndexWriter) testPoint(message string) {
	if w.infoStream.IsEnabled("TP") {
		w.infoStream.Message("TP", message)
	}
}

// L4356

/* Called by DirectoryReader.doClose() */
func (w *IndexWriter) deletePendingFiles() {
	w.deleter.deletePendingFiles()
}

/*
NOTE: this method creates a compound file for all files returned by
info.files(). While, generally, this may include separate norms and
deleteion files, this SegmentInfos must not reference such files when
this method is called, because they are not allowed within a compound
file.
*/
func createCompoundFile(infoStream util.InfoStream,
	directory store.Directory,
	checkAbort CheckAbort,
	info *model.SegmentInfo,
	context store.IOContext) (names []string, err error) {

	filename := util.SegmentFileName(info.Name, "", store.COMPOUND_FILE_EXTENSION)
	if infoStream.IsEnabled("IW") {
		infoStream.Message("IW", "create compound file %v", filename)
	}
	// Now merge all added files
	files := info.Files()
	var cfsDir *store.CompoundFileDirectory
	cfsDir, err = store.NewCompoundFileDirectory(directory, filename, context, true)
	if err != nil {
		return
	}
	func() {
		defer func() {
			var success = false
			defer func() {
				if !success {
					directory.DeleteFile(filename) // ignore error
					directory.DeleteFile(util.SegmentFileName(info.Name, "", store.COMPOUND_FILE_EXTENSION))
				}
			}()

			err = util.CloseWhileHandlingError(err, cfsDir)
			success = err == nil
		}()

		var length int64
		for file, _ := range files {
			err = directory.Copy(cfsDir, file, file, context)
			if err != nil {
				return
			}
			length, err = directory.FileLength(file)
			if err != nil {
				return
			}
			err = checkAbort.work(float64(length))
			if err != nil {
				return
			}
		}
	}()
	if err != nil {
		return
	}

	// Replace all previous files with the CFS/CFE files:
	siFiles := make(map[string]bool)
	siFiles[filename] = true
	siFiles[util.SegmentFileName(info.Name, "", store.COMPOUND_FILE_EXTENSION)] = true
	info.SetFiles(siFiles)

	for file, _ := range files {
		names = append(names, file)
	}
	return
}

// Tries to delete the given files if unreferenced.
func (w *IndexWriter) deleteNewFiles(files []string) error {
	w.Lock() // synchronized
	defer w.Unlock()
	panic("not implemented yet")
}

/* Cleans up residuals from a segment that could not be entirely flushed due to an error */
func (w *IndexWriter) flushFailed(info *model.SegmentInfo) error {
	w.Lock()
	defer w.Unlock()
	return w.deleter.refresh(info.Name)
}

func (w *IndexWriter) purge(forced bool) (n int, err error) {
	return w.docWriter.purgeBuffer(w, forced)
}

func (w *IndexWriter) doAfterSegmentFlushed(triggerMerge bool, forcePurge bool) (err error) {
	defer func() {
		if triggerMerge {
			err = mergeError(err, w.maybeMerge(MERGE_TRIGGER_SEGMENT_FLUSH, UNBOUNDED_MAX_MERGE_SEGMENTS))
		}
	}()
	_, err = w.purge(forcePurge)
	return err
}

/*
If openDirectoryReader() has been called (ie, this writer is in near
real-time mode), then after a merge comletes, this class can be
invoked to warm the reader on the newly merged segment, before the
merge commits. This is not required for near real-time search, but
will reduce search latency on opening a new near real-time reader
after a merge completes.

NOTE: warm is called before any deletes have been carried over to the
merged segment.
*/
type IndexReaderWarmer interface {
	// Invoked on the AtomicReader for the newly merged segment, before
	// that segment is made visible to near-real-time readers.
	warm(reader AtomicReader) error
}
