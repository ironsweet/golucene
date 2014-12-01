package index

import (
	"container/list"
	"fmt"
	"github.com/balzaczyy/golucene/core/analysis"
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"sync"
	"sync/atomic"
)

// index/DocumentsWriter.java

/*
This class accepts multiple added documents and directly writes
segment files.

Each added document is passed to the indexing chain, which in turn
processes the document into the different codec formats. Some format
write bytes to files immediately, e.g. stored fields and term vectors,
while others are buffered by the indexing chain and written only on
flush.

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
	config       LiveIndexWriterConfig
	numDocsInRAM int32 // atomic

	// TODO: cut over to BytesRefHash in BufferedUpdates
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

func newDocumentsWriter(writer *IndexWriter, config LiveIndexWriterConfig,
	directory store.Directory) *DocumentsWriter {
	ans := &DocumentsWriter{
		Locker:        &sync.Mutex{},
		deleteQueue:   newDocumentsWriterDeleteQueue(),
		ticketQueue:   newDocumentsWriterFlushQueue(),
		directory:     directory,
		config:        config,
		infoStream:    config.InfoStream(),
		perThreadPool: config.indexerThreadPool(),
		flushPolicy:   config.flushPolicy(),
		writer:        writer,
		eventsLock:    &sync.RWMutex{},
		events:        list.New(),
	}
	ans.flushControl = newDocumentsWriterFlushControl(ans, config, writer.bufferedUpdatesStream)
	return ans
}

func (dw *DocumentsWriter) applyAllDeletes(deleteQueue *DocumentsWriterDeleteQueue) (bool, error) {
	if dw.flushControl.getAndResetApplyAllDeletes() {
		if deleteQueue != nil && !dw.flushControl.fullFlush {
			err := dw.ticketQueue.addDeletes(deleteQueue)
			if err != nil {
				return false, err
			}
		}
		dw.putEvent(applyDeletesEvent) // apply deletes event forces a purge
		return true, nil
	}
	return false, nil
}

func (w *DocumentsWriter) purgeBuffer(writer *IndexWriter, forced bool) (int, error) {
	// forced flag is ignored since Go doesn't encourage tryLock idea
	return w.ticketQueue.forcePurge(writer)
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

func (dw *DocumentsWriter) postUpdate(flushingDWPT *DocumentsWriterPerThread, hasEvents bool) (bool, error) {
	ok, err := dw.applyAllDeletes(dw.deleteQueue)
	if err != nil {
		return false, err
	}
	if ok {
		hasEvents = true
	}
	if flushingDWPT != nil {
		ok, err = dw.doFlush(flushingDWPT)
		if err != nil {
			return false, err
		}
		if ok {
			hasEvents = true
		}
	} else {
		nextPendingFlush := dw.flushControl.nextPendingFlush()
		if nextPendingFlush != nil {
			ok, err = dw.doFlush(nextPendingFlush)
			if err != nil {
				return false, err
			}
			if ok {
				hasEvents = true
			}
		}
	}
	return hasEvents, nil
}

func (dw *DocumentsWriter) ensureInitialized(state *ThreadState) {
	if state.isActive && state.dwpt == nil {
		infos := model.NewFieldInfosBuilder(dw.writer.globalFieldNumberMap)
		state.dwpt = newDocumentsWriterPerThread(dw.writer.newSegmentName(),
			dw.directory, dw.config, dw.infoStream, dw.deleteQueue, infos, &dw.writer.pendingNumDocs)
	}
}

// L428
func (dw *DocumentsWriter) updateDocument(doc []model.IndexableField,
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
				// We don't know whether the document actually counted as
				// being indexed, so we must subtract here to accumulate our
				// separate counter:
				atomic.AddInt32(&dw.numDocsInRAM, int32(dwpt.numDocsInRAM-dwptNuMDocs))
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
	var hasEvents = false
	for flushingDWPT != nil {
		hasEvents = true
		stop, err := func() (bool, error) {
			defer func() {
				dw.flushControl.doAfterFlush(flushingDWPT)
				flushingDWPT.checkAndResetHasAborted()
			}()

			assertn(dw.currentFullFlushDelQueue == nil ||
				flushingDWPT.deleteQueue == dw.currentFullFlushDelQueue,
				"expected: %v but was %v %v",
				dw.currentFullFlushDelQueue,
				flushingDWPT.deleteQueue,
				dw.flushControl.fullFlush)

			/*
				Since, with DWPT, the flush process is concurrent and several
				DWPT could flush at the same time, we must maintain the order
				or the flushes before we can apply the flushed segment and
				the frozen global deletes it is buffering. The reason for
				this is that the global deletes mark a certain point in time
				where we took a DWPT out of rotation and freeze the global
				deletes.

				Example: A flush 'A' starts and freezes the global deletes,
				then flush 'B' starts and freezes all deletes occurred since
				'A' has started. If 'B' finishes before 'A', we need to wait
				until 'A' is done, otherwise the deletes frozen by 'B' are
				not applied to 'A' and we might miss to deletes documents in
				'A'.
			*/

			err := func() error {
				var success = false
				var ticket *SegmentFlushTicket
				defer func() {
					if !success && ticket != nil {
						// In the case of a failure, make sure we are making
						// progress and apply all the deletes since the segment
						// flush failed since the flush ticket could hold global
						// deletes. See FlushTicket.canPublish().
						dw.ticketQueue.markTicketFailed(ticket)
					}
				}()

				// Each flush is assigned a ticket in the order they acquire the ticketQueue lock
				ticket = dw.ticketQueue.addFlushTicket(flushingDWPT)

				flushingDocsInRAM := flushingDWPT.numDocsInRAM
				err := func() error {
					var dwptSuccess = false
					defer func() {
						dw.subtractFlushedNumDocs(flushingDocsInRAM)
						if len(flushingDWPT.filesToDelete) > 0 {
							dw.putEvent(newDeleteNewFilesEvent(flushingDWPT.filesToDelete))
							hasEvents = true
						}
						if !dwptSuccess {
							dw.putEvent(newFlushFailedEvent(flushingDWPT.segmentInfo))
							hasEvents = true
						}
					}()

					// flush concurrently without locking
					newSegment, err := flushingDWPT.flush()
					if err != nil {
						return err
					}
					dw.ticketQueue.addSegment(ticket, newSegment)
					dwptSuccess = true
					return nil
				}()
				if err != nil {
					return err
				}
				// flush was successful once we reached this point - new seg.
				// has been assigned to the ticket!
				success = true
				return nil
			}()
			if err != nil {
				return false, err
			}
			// Now we are done and try to flush the ticket queue if the
			// head of the queue has already finished the flush.
			if dw.ticketQueue.ticketCount() >= dw.perThreadPool.numActiveThreadState() {
				// This means there is a backlog: the one thread in
				// innerPurge can't keep up with all other threads flusing
				// segments. In this case we forcefully stall the producers.
				dw.putEvent(forcedPurgeEvent)
				return true, nil
			}
			return false, nil
		}()
		if err != nil {
			return false, err
		}
		if stop {
			break
		}

		flushingDWPT = dw.flushControl.nextPendingFlush()
	}
	if hasEvents {
		dw.putEvent(mergePendingEvent)
	}
	// If deletes alone are consuming > 1/2 our RAM buffer, force them
	// all to apply now. This is to prevent too-frequent flushing of a
	// long tail tiny segments:
	if ramBufferSizeMB := dw.config.RAMBufferSizeMB(); ramBufferSizeMB != DISABLE_AUTO_FLUSH &&
		dw.flushControl.deleteBytesUsed() > int64(1024*1024*ramBufferSizeMB/2) {

		if dw.infoStream.IsEnabled("DW") {
			dw.infoStream.Message("DW", "force apply deletes bytesUsed=%v vs ramBuffer=%v",
				dw.flushControl.deleteBytesUsed(), 1024*1024*ramBufferSizeMB)
		}
		hasEvents = true
		ok, err := dw.applyAllDeletes(dw.deleteQueue)
		if err != nil {
			return false, err
		}
		if !ok {
			dw.putEvent(applyDeletesEvent)
		}
	}

	return hasEvents, nil
}

func (dw *DocumentsWriter) subtractFlushedNumDocs(numFlushed int) {
	oldValue := atomic.LoadInt32(&dw.numDocsInRAM)
	for !atomic.CompareAndSwapInt32(&dw.numDocsInRAM, oldValue, oldValue-int32(numFlushed)) {
		oldValue = atomic.LoadInt32(&dw.numDocsInRAM)
	}
	assert(atomic.LoadInt32(&dw.numDocsInRAM) >= 0)
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
			next := dw.flushControl.nextPendingFlush()
			assert(next != flushingDWPT)
			flushingDWPT = next
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
		dw.infoStream.Message("DW", "finishFullFlush success=%v", success)
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

func (dw *DocumentsWriter) processEvents(writer *IndexWriter,
	triggerMerge, forcePurge bool) (processed bool, err error) {
	dw.eventsLock.RLock()
	defer dw.eventsLock.RUnlock()

	for e := dw.events.Front(); e != nil; e = e.Next() {
		dw.events.Remove(e)
		processed = true
		if err = e.Value.(Event)(writer, triggerMerge, forcePurge); err != nil {
			break
		}
	}
	return
}

func (dw *DocumentsWriter) assertEventQueueAfterClose() {
	dw.eventsLock.RLock()
	defer dw.eventsLock.RUnlock()

	for e := dw.events.Front(); e != nil; e = e.Next() {
		// TODO find a better way to compare event type
		assert(fmt.Sprintf("%v", e.Value) == fmt.Sprintf("%v", mergePendingEvent))
	}
}
