package index

import (
	"container/list"
	"github.com/balzaczyy/golucene/core/analysis"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"log"
	"sync"
	"sync/atomic"
)

// index/DocConsumer.java

type DocConsumer interface {
	processDocuments(fieldInfos *FieldInfosBuilder) error
	finishDocument() error
	flush(state *SegmentWriteState) error
	abort()
}

// index/DocumentsWriter.java

/*
This class accepts multiple added documents and directly writes
segment files.

Each added document is passed to the DocConsumer, which in turn
processes the document and interacts with other consumers in the
indexing chain. Certain consumers, like StoredFieldsConsumer and
TermVectorsConsumer, digest a document and immediately write bytes to
the "doc store" files (i.e., they do not consume RAM per document,
except while they are processing the document).

Other consumers e.g. FreqProxTermsWriter and NormsConsumer, buffer
bytes in RAM and flush only when a new segment is produced.

Once we have used our allowed RAM buffer, or the number of aded docs
is large enough (in the case we are flushing by doc count instead of
RAM usage), we create a real segment and flush it to the Directory.

Goroutines:

Multiple Goroutines are allowed into AddDocument at once. There is an
initial synchronized call to ThreadState() which allocates a
TheadState for this goroutine. The same goroutine will get the same
ThreadState over time (goroutine affinity) so that if there are
consistent patterns (for example each goroutine is indexing a
different content source) then we make better use of RAM. Then
processDocument() is called on tha tThreadState without
synchronization (most of the "heavy lifting" is in this call).
Finally the synchronized "finishDocument" is called to flush changes
to the directory.

When flush is called by IndexWriter we forcefully idle all goroutines
and flush only once they are all idle. This means you can call flush
with a given goroutine even while other goroutines are actively
adding/deleting documents.

Exceptions:

Because this class directly updates in-memory posting lists, and
flushes stored fields and term vectors directly to files in the
directory, there are certain limited times when an error can corrupt
this state. For example, a disk full while flushing stored fields
leaves this file in a corrupt state. Or, a memory issue while
appending to the in-memory posting lists can corrupt that posting
list. We call such errors "aborting errors". In these cases we must
call abort() to discard all docs added since the last flush.

All other errors ("non-aborting errors") can still partially update
the index structures. These updates are consistent, but, they
represent only a part of the document seen up until the error was hit.
When this happens, we immediately mark the document as deleted so
that the document is always atomically ("all or none") added to the
index.
*/
type DocumentsWriter struct {
	sync.Locker

	directory    store.Directory
	closed       bool // volatile
	infoStream   util.InfoStream
	config       *LiveIndexWriterConfig
	numDocsInRAM int32 // atomic

	// TODO: cut over to BytesRefHash in BufferedDeletes
	deleteQueue *DocumentsWriterDeleteQueue // volatile
	ticketQueue *DocumentsWriterFlushQueue
	// we preserve changes during a full flush since IW might not
	// checkout before we release all changes. NRT Readers otherwise
	// suddenly return true from isCurrent() while there are actually
	// changes currently committed. See also anyChanges() &
	// flushAllThreads()
	pendingChangesInCurrentFullFlush bool // volatile

	perThreadPool *DocumentsWriterPerThreadPool
	flushPolicy   FlushPolicy
	flushControl  *DocumentsWriterFlushControl
	writer        *IndexWriter
	eventsLock    *sync.RWMutex
	events        *list.List // synchronized
	// for asserts
	currentFullFlushDelQueue *DocumentsWriterDeleteQueue
}

func newDocumentsWriter(writer *IndexWriter, config *LiveIndexWriterConfig, directory store.Directory) *DocumentsWriter {
	ans := &DocumentsWriter{
		Locker:        &sync.Mutex{},
		deleteQueue:   newDocumentsWriterDeleteQueue(),
		ticketQueue:   newDocumentsWriterFlushQueue(),
		directory:     directory,
		config:        config,
		infoStream:    config.infoStream,
		perThreadPool: config.indexerThreadPool,
		flushPolicy:   config.flushPolicy,
		writer:        writer,
		eventsLock:    &sync.RWMutex{},
		events:        list.New(),
	}
	ans.flushControl = newDocumentsWriterFlushControl(ans, config, writer.bufferedDeletesStream)
	return ans
}

func (dw *DocumentsWriter) ensureOpen() {
	assert2(!dw.closed, "this IndexWriter is closed")
}

/*
Called if we hit an error at a bad time (when updating the index
files) and must discard all currently buffered docs. This resets our
state, discarding any docs added since last flush.
*/
func (dw *DocumentsWriter) abort(writer *IndexWriter) {
	dw.Lock()
	defer dw.Unlock()

	var success = false
	var newFilesSet = make(map[string]bool)
	defer func() {
		if dw.infoStream.IsEnabled("DW") {
			dw.infoStream.Message("DW", "done abort; abortedFiles=%v success=%v", newFilesSet, success)
		}
	}()

	dw.deleteQueue.clear()
	if dw.infoStream.IsEnabled("DW") {
		dw.infoStream.Message("DW", "abort")
	}
	dw.perThreadPool.foreach(func(perThread *ThreadState) {
		dw.abortThreadState(perThread, newFilesSet)
	})
	dw.flushControl.abortPendingFlushes(newFilesSet)
	dw.putEvent(newDeleteNewFilesEvent(newFilesSet))
	dw.flushControl.waitForFlush()
	success = true
}

func (dw *DocumentsWriter) abortThreadState(perThread *ThreadState, newFiles map[string]bool) {
	if perThread.isActive { // we might be closed
		if perThread.dwpt != nil {
			defer func() {
				perThread.dwpt.checkAndResetHasAborted()
				dw.flushControl.doOnAbort(perThread)
			}()
			dw.subtractFlushedNumDocs(perThread.dwpt.numDocsInRAM)
			perThread.dwpt.abort(newFiles)
		} else {
			dw.flushControl.doOnAbort(perThread)
		}
	} else {
		assert(dw.closed)
	}
}

func (dw *DocumentsWriter) anyChanges() bool {
	if dw.infoStream.IsEnabled("DW") {
		dw.infoStream.Message("DW",
			"anyChanges? numDocsInRAM=%v deletes=%v, hasTickets=%v pendingChangesInFullFlush=%v",
			atomic.LoadInt32(&dw.numDocsInRAM), dw.deleteQueue.anyChanges(),
			dw.ticketQueue.hasTickets(), dw.pendingChangesInCurrentFullFlush)
	}
	// Changes are either in a DWPT or in the deleteQueue.
	// Yet if we currently flush deletes and/or dwpt, there
	// could be a window where all changes are in the ticket queue
	// before they are published to the IW, ie, we need to check if the
	// ticket queue has any tickets.
	return atomic.LoadInt32(&dw.numDocsInRAM) != 0 || dw.deleteQueue.anyChanges() ||
		dw.ticketQueue.hasTickets() || dw.pendingChangesInCurrentFullFlush
}

func (dw *DocumentsWriter) close() {
	dw.closed = true
	dw.flushControl.close()
}

func (dw *DocumentsWriter) preUpdate() (bool, error) {
	dw.ensureOpen()
	var hasEvents = false
	if dw.flushControl.anyStalledThreads() || dw.flushControl.numQueuedFlushes() > 0 {
		// Help out flushing any queued DWPTs so we can un-stall:
		if dw.infoStream.IsEnabled("DW") {
			dw.infoStream.Message("DW", "DocumentsWriter has queued dwpt; will hijack this thread to flush pending segment(s)")
		}
		for {
			// Try pick up pending threads here if possible
			for flushingDWPT := dw.flushControl.nextPendingFlush(); flushingDWPT != nil; {
				// Don't push the delete here since the update could fail!
				ok, err := dw.doFlush(flushingDWPT)
				if err != nil {
					return false, err
				}
				if ok {
					hasEvents = true
				}
			}

			if dw.infoStream.IsEnabled("DW") {
				if dw.flushControl.anyStalledThreads() {
					dw.infoStream.Message("DW", "WARNING DocumentsWriter has stalled threads; waiting")
				}
			}

			dw.flushControl.waitIfStalled() // block if stalled
			if dw.flushControl.numQueuedFlushes() == 0 {
				break
			}
		}
	}
	return hasEvents, nil
}

func (dw *DocumentsWriter) postUpdate(flusingDWPT *DocumentsWriterPerThread, hasEvents bool) (bool, error) {
	panic("not implemented yet")
}

func (dw *DocumentsWriter) ensureInitialized(state *ThreadState) {
	panic("not implemented yet")
}

// L428
func (dw *DocumentsWriter) updateDocument(doc []IndexableField,
	analyzer analysis.Analyzer, delTerm *Term) (bool, error) {

	hasEvents, err := dw.preUpdate()
	if err != nil {
		return false, err
	}

	flushingDWPT, err := func() (*DocumentsWriterPerThread, error) {
		perThread := dw.flushControl.obtainAndLock()
		defer dw.flushControl.perThreadPool.release(perThread)

		if !perThread.isActive {
			dw.ensureOpen()
			panic("perThread is not active but we are still open")
		}
		dw.ensureInitialized(perThread)
		assert(perThread.dwpt != nil)
		dwpt := perThread.dwpt
		dwptNuMDocs := dwpt.numDocsInRAM

		err := func() error {
			defer func() {
				if dwpt.checkAndResetHasAborted() {
					if len(dwpt.filesToDelete) > 0 {
						dw.putEvent(newDeleteNewFilesEvent(dwpt.filesToDelete))
					}
					dw.subtractFlushedNumDocs(dwptNuMDocs)
					dw.flushControl.doOnAbort(perThread)
				}
			}()

			err := dwpt.updateDocument(doc, analyzer, delTerm)
			if err != nil {
				return err
			}
			atomic.AddInt32(&dw.numDocsInRAM, 1)
			return nil
		}()
		if err != nil {
			return nil, err
		}

		isUpdate := delTerm != nil
		return dw.flushControl.doAfterDocument(perThread, isUpdate), nil
	}()
	if err != nil {
		return false, err
	}

	return dw.postUpdate(flushingDWPT, hasEvents)
}

func (dw *DocumentsWriter) doFlush(flushingDWPT *DocumentsWriterPerThread) (bool, error) {
	panic("not implemented yet")
}

func (dw *DocumentsWriter) subtractFlushedNumDocs(numFlushed int) {
	oldValue := atomic.LoadInt32(&dw.numDocsInRAM)
	for !atomic.CompareAndSwapInt32(&dw.numDocsInRAM, oldValue, oldValue-int32(numFlushed)) {
		oldValue = atomic.LoadInt32(&dw.numDocsInRAM)
	}
}

/*
FlushAllThreads is synced by IW fullFlushLock. Flushing all threads
is a two stage operation; the caller must ensure (in try/finally)
that finishFLush is called after this method, to release the flush
lock in DWFlushControl
*/
func (dw *DocumentsWriter) flushAllThreads(indexWriter *IndexWriter) (bool, error) {
	if dw.infoStream.IsEnabled("DW") {
		dw.infoStream.Message("DW", "startFullFlush")
	}

	flushingDeleteQueue := func() *DocumentsWriterDeleteQueue {
		dw.Lock()
		defer dw.Unlock()
		dw.pendingChangesInCurrentFullFlush = dw.anyChanges()
		dq := dw.deleteQueue
		// Cut over to a new delete queue. This must be synced on the
		// flush control otherwise a new DWPT could sneak into the loop
		// with an already flushing delete queue
		dw.flushControl.markForFullFlush() // swaps the delQueue synced on FlushControl
		dw.currentFullFlushDelQueue = dq
		return dq
	}()
	assert(dw.currentFullFlushDelQueue != nil)
	assert(dw.currentFullFlushDelQueue != dw.deleteQueue)

	return func() (bool, error) {
		anythingFlushed := false
		defer func() { assert(flushingDeleteQueue == dw.currentFullFlushDelQueue) }()

		flushingDWPT := dw.flushControl.nextPendingFlush()
		for flushingDWPT != nil {
			flushed, err := dw.doFlush(flushingDWPT)
			if err != nil {
				return false, err
			}
			if flushed {
				anythingFlushed = true
			}
			flushingDWPT = dw.flushControl.nextPendingFlush()
		}

		// If a concurrent flush is still in flight wait for it
		dw.flushControl.waitForFlush()
		if !anythingFlushed && flushingDeleteQueue.anyChanges() {
			// apply deletes if we did not flush any document
			if dw.infoStream.IsEnabled("DW") {
				dw.infoStream.Message("DW", "flush naked frozen global deletes")
			}
			err := dw.ticketQueue.addDeletes(flushingDeleteQueue)
			if err != nil {
				return false, err
			}
		}
		_, err := dw.ticketQueue.forcePurge(indexWriter)
		if err != nil {
			return false, err
		}
		assert(!flushingDeleteQueue.anyChanges() && !dw.ticketQueue.hasTickets())
		return anythingFlushed, nil
	}()
}

func (dw *DocumentsWriter) finishFullFlush(success bool) {
	defer func() { dw.pendingChangesInCurrentFullFlush = false }()
	if dw.infoStream.IsEnabled("DW") {
		dw.infoStream.Message("DW", "finishFullFlush success %v", success)
	}
	dw.currentFullFlushDelQueue = nil
	if success {
		// Release the flush lock
		dw.flushControl.finishFullFlush()
	} else {
		newFilesSet := make(map[string]bool)
		dw.flushControl.abortFullFlushes(newFilesSet)
		dw.putEvent(newDeleteNewFilesEvent(newFilesSet))
	}
}

func (dw *DocumentsWriter) putEvent(event Event) {
	dw.eventsLock.Lock()
	defer dw.eventsLock.Unlock()
	dw.events.PushBack(event)
}

func (dw *DocumentsWriter) processEvents(writer *IndexWriter, triggerMerge, forcePurge bool) bool {
	dw.eventsLock.RLock()
	defer dw.eventsLock.RUnlock()

	processed := false
	for e := dw.events.Front(); e != nil; e = e.Next() {
		processed = true
		e.Value.(Event)(writer, triggerMerge, forcePurge)
	}
	return processed
}

/*
Interface for internal atomic events. See DocumentsWriter fo details.
Events are executed concurrently and no order is guaranteed. Each
event should only rely on the serializeability within its process
method. All actions that must happen before or after a certain action
must be encoded inside the process() method.
*/
type Event func(writer *IndexWriter, triggerMerge, clearBuffers bool)

func newDeleteNewFilesEvent(files map[string]bool) Event {
	return Event(func(writer *IndexWriter, triggerMerge, forcePurge bool) {
		writer.Lock()
		defer writer.Unlock()
		var fileList []string
		for file, _ := range files {
			fileList = append(fileList, file)
		}
		writer.deleter.deleteNewFiles(fileList)
	})
}

// index/DocumentsWriterPerThread.java

// Returns the DocConsumer that the DocumentsWriter calls to
// process the documents.
type IndexingChain func(documentsWriterPerThread *DocumentsWriterPerThread) DocConsumer

var defaultIndexingChain = func(documentsWriterPerThread *DocumentsWriterPerThread) DocConsumer {
	panic("not implemented yet")
}

type DocumentsWriterPerThread struct {
	directory *TrackingDirectoryWrapper
	consumer  DocConsumer

	// Deletes for our still-in-RAM (to be flushed next) segment
	pendingDeletes *BufferedDeletes
	segmentInfo    *SegmentInfo // Current segment we are working on
	aborting       bool         // True if an abort is pending
	hasAborted     bool         // True if the last exception throws by #updateDocument was aborting

	infoStream   util.InfoStream
	numDocsInRAM int // the number of RAM resident documents
	deleteQueue  *DocumentsWriterDeleteQueue

	filesToDelete map[string]bool
}

/*
Called if we hit an error at a bad time (when updating the index
files) and must discard all currently buffered docs. This resets our
state, discarding any docs added since last flush.
*/
func (dwpt *DocumentsWriterPerThread) abort(createdFiles map[string]bool) {
	log.Printf("now abort seg=%v", dwpt.segmentInfo.name)
	dwpt.hasAborted, dwpt.aborting = true, true
	defer func() {
		dwpt.aborting = false
		if dwpt.infoStream.IsEnabled("DWPT") {
			dwpt.infoStream.Message("DWPT", "done abort")
		}
	}()

	if dwpt.infoStream.IsEnabled("DWPT") {
		dwpt.infoStream.Message("DWPT", "now abort")
	}
	dwpt.consumer.abort()

	dwpt.pendingDeletes.clear()
	for file, _ := range dwpt.directory.createdFiles() {
		createdFiles[file] = true
	}
}

func (dwpt *DocumentsWriterPerThread) checkAndResetHasAborted() (res bool) {
	res, dwpt.hasAborted = dwpt.hasAborted, false
	return
}

func (dwpt *DocumentsWriterPerThread) updateDocument(doc []IndexableField, analyzer analysis.Analyzer, delTerm *Term) error {
	panic("not implemented yet")
}

// L600
// if you increase this, you must fix field cache impl for
// Terms/TermsIndex requires <= 32768
const MAX_TERM_LENGTH_UTF8 = util.BYTE_BLOCK_SIZE - 2
