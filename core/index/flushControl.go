package index

import (
	"container/list"
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
	"math"
	"sync"
)

// index/DocumentsWriterFlushControl.java

/*
This class controls DocumentsWriterPerThread (DWPT) flushing during
indexing. It tracks the memory consumption per DWPT and uses a
configured FlushPolicy to decide if a DWPT must flush.

In addition to the FlushPolicy the flush control might set certain
DWPT as flush pending iff a DWPT exceeds the RAMPerThreadHardLimitMB()
to prevent address space exhaustion.
*/
type DocumentsWriterFlushControl struct {
	sync.Locker
	condFlushWait *sync.Cond

	hardMaxBytesPerDWPT int64
	activeBytes         int64
	flushBytes          int64
	numPending          int // volatile
	fullFlush           bool
	flushQueue          *list.List
	// only for safety reasons if a DWPT is close to the RAM limit
	blockedFlushes  *list.List
	flushingWriters map[*DocumentsWriterPerThread]int64

	stallControl  *DocumentsWriterStallControl
	perThreadPool *DocumentsWriterPerThreadPool
	flushPolicy   FlushPolicy
	closed        bool

	documentsWriter       *DocumentsWriter
	config                *LiveIndexWriterConfig
	bufferedDeletesStream *BufferedDeletesStream
	infoStream            util.InfoStream

	fullFlushBuffer []*DocumentsWriterPerThread
}

func newDocumentsWriterFlushControl(documentsWriter *DocumentsWriter,
	config *LiveIndexWriterConfig, bufferedDeletesStream *BufferedDeletesStream) *DocumentsWriterFlushControl {
	objLock := &sync.Mutex{}
	return &DocumentsWriterFlushControl{
		Locker:                objLock,
		condFlushWait:         sync.NewCond(objLock),
		flushQueue:            list.New(),
		blockedFlushes:        list.New(),
		flushingWriters:       make(map[*DocumentsWriterPerThread]int64),
		infoStream:            config.infoStream,
		stallControl:          newDocumentsWriterStallControl(),
		perThreadPool:         documentsWriter.perThreadPool,
		flushPolicy:           documentsWriter.flushPolicy,
		config:                config,
		hardMaxBytesPerDWPT:   int64(config.perRoutineHardLimitMB) * 1024 * 1024,
		documentsWriter:       documentsWriter,
		bufferedDeletesStream: bufferedDeletesStream,
	}
}

func (fc *DocumentsWriterFlushControl) assertMemory() bool {
	panic("not implemented yet")
}

func (fc *DocumentsWriterFlushControl) doAfterFlush(dwpt *DocumentsWriterPerThread) {
	fc.Lock()
	defer fc.Unlock()

	bytes, ok := fc.flushingWriters[dwpt]
	assert(ok)

	defer func() {
		defer func() {
			fc.condFlushWait.Signal()
		}()
		fc.updateStallState()
	}()

	delete(fc.flushingWriters, dwpt)
	fc.flushBytes -= bytes
	// fc.perThreadPool.recycle(dwpt)
	fc.assertMemory()
}

func (fc *DocumentsWriterFlushControl) updateStallState() bool {
	var limit int64 = math.MaxInt64
	if maxRamMB := fc.config.ramBufferSizeMB; maxRamMB != DISABLE_AUTO_FLUSH {
		limit = int64(2 * 1024 * 1024 * maxRamMB)
	}
	// We block indexing threads if net byte grows due to slow flushes,
	// yet, for small ram buffers and large documents, we can easily
	// reach the limit without any ongoing flushes. We need ensure that
	// we don't stall/block if an ongoing or pending flush can not free
	// up enough memory to release the stall lock.
	stall := (fc.activeBytes+fc.flushBytes) > limit &&
		fc.activeBytes < limit &&
		!fc.closed
	fc.stallControl.updateStalled(stall)
	return stall
}

func (fc *DocumentsWriterFlushControl) waitForFlush() {
	fc.Lock()
	defer fc.Unlock()

	for len(fc.flushingWriters) > 0 {
		fc.condFlushWait.Wait()
	}
}

/*
Sets flush pending state on the given ThreadState. The ThreadState
must have indexed at least one Document and must not be already
pending.
*/
func (fc *DocumentsWriterFlushControl) _setFlushPending(perThread *ThreadState) {
	assert(!perThread.flushPending)
	if perThread.dwpt.numDocsInRAM > 0 {
		perThread.flushPending = true // write access synced
		bytes := perThread.bytesUsed
		fc.flushBytes += bytes
		fc.activeBytes -= bytes
		fc.numPending++ // write access synced
		fc.assertMemory()
	}
	// don't assert on numDocs since we could hit an abort except while
	// selecting that dwpt for flushing
}

func (fc *DocumentsWriterFlushControl) doOnAbort(state *ThreadState) {
	fc.Lock()
	defer fc.Unlock()
	defer fc.updateStallState()
	if state.flushPending {
		fc.flushBytes -= state.bytesUsed
	} else {
		fc.activeBytes -= state.bytesUsed
	}
	fc.assertMemory()
	// Take it out of the loop this DWPT is stale
	fc.perThreadPool.reset(state, fc.closed)
}

func (fc *DocumentsWriterFlushControl) tryCheckoutForFlush(perThread *ThreadState) *DocumentsWriterPerThread {
	fc.Lock()
	defer fc.Unlock()

	assert(perThread.flushPending)
	return fc._tryCheckOutForFlush(perThread)
}

func (fc *DocumentsWriterFlushControl) _tryCheckOutForFlush(perThread *ThreadState) *DocumentsWriterPerThread {
	assert(perThread.flushPending)
	defer fc.updateStallState()
	// We are pending so all memory is already moved to flushBytes
	if perThread.isActive && perThread.dwpt != nil {
		bytes := perThread.bytesUsed // do that before replace
		dwpt := fc.perThreadPool.reset(perThread, fc.closed)
		_, ok := fc.flushingWriters[dwpt]
		assert2(!ok, "DWPT is already flushing")
		// Record the flushing DWPT to reduce flushBytes in doAfterFlush
		fc.flushingWriters[dwpt] = bytes
		fc.numPending-- // write access synced
		return dwpt
	}
	return nil
}

func (fc *DocumentsWriterFlushControl) String() string {
	return fmt.Sprintf("DocumentsWriterFlushControl [activeBytes=%v, flushBytes=%v]",
		fc.activeBytes, fc.flushBytes)
}

func (fc *DocumentsWriterFlushControl) close() {
	fc.Lock()
	defer fc.Unlock()
	// set by DW to signal that we should not release new DWPT after close
	if !fc.closed {
		fc.closed = true
	}
}

func (fc *DocumentsWriterFlushControl) markForFullFlush() {
	flushingQueue := func() *DocumentsWriterDeleteQueue {
		fc.Lock()
		defer fc.Unlock()

		assert2(!fc.fullFlush, "called DWFC#markForFullFlush() while full flush is still running")
		assertn(len(fc.fullFlushBuffer) == 0, "full flush buffer should be empty: ", fc.fullFlushBuffer)

		fc.fullFlush = true
		res := fc.documentsWriter.deleteQueue
		// Set a new delete queue - all subsequent DWPT will use this
		// queue untiil we do another full flush
		fc.documentsWriter.deleteQueue = newDocumentsWriterDeleteQueueWithGeneration(res.generation + 1)
		return res
	}()

	fc.perThreadPool.foreach(func(next *ThreadState) {
		if !next.isActive || next.dwpt == nil {
			if fc.closed && next.isActive {
				next.deactivate()
			}
			return
		}
		assertn(next.dwpt.deleteQueue == flushingQueue ||
			next.dwpt.deleteQueue == fc.documentsWriter.deleteQueue,
			" flushingQueue: %v currentQueue: %v perThread queue: %v numDocsInRAM: %v",
			flushingQueue, fc.documentsWriter.deleteQueue, next.dwpt.deleteQueue,
			next.dwpt.numDocsInRAM)
		if next.dwpt.deleteQueue != flushingQueue {
			// this one is already a new DWPT
			return
		}
		fc.addFlushableState(next)
	})

	func() {
		fc.Lock()
		defer fc.Unlock()

		// make sure we move all DWPT that are where concurrently marked
		// as pending and moved to blocked are moved over to the
		// flushQueue. There is a chance that this happens since we
		// marking DWPT for full flush without blocking indexing.
		fc.pruneBlockedQueue(flushingQueue)
		fc.assertBlockedFlushes(fc.documentsWriter.deleteQueue)
		for _, dwpt := range fc.fullFlushBuffer {
			fc.flushQueue.PushBack(dwpt)
		}
		fc.fullFlushBuffer = nil
		fc.updateStallState()
	}()
	fc.assertActiveDeleteQueue(fc.documentsWriter.deleteQueue)
}

func (fc *DocumentsWriterFlushControl) assertActiveDeleteQueue(queue *DocumentsWriterDeleteQueue) {
	fc.perThreadPool.foreach(func(next *ThreadState) {
		n := 0
		if next.dwpt != nil {
			n = next.dwpt.numDocsInRAM
		}
		assertn(!next.isActive || next.dwpt == nil || next.dwpt.deleteQueue == queue,
			"isInitialized: %v numDocs: %v", next.isActive && next.dwpt != nil, n)
	})
}

func (fc *DocumentsWriterFlushControl) nextPendingFlush() *DocumentsWriterPerThread {
	numPending, fullFlush, dwpt := func() (int, bool, *DocumentsWriterPerThread) {
		fc.Lock()
		defer fc.Unlock()

		if e := fc.flushQueue.Front(); e != nil {
			pool := e.Value.(*DocumentsWriterPerThread)
			fc.updateStallState()
			return 0, false, pool
		}
		return fc.numPending, fc.fullFlush, nil
	}()
	if dwpt != nil {
		return dwpt
	}

	if numPending > 0 && !fullFlush {
		// don't check if we are doing a full flush
		dwpt = fc.perThreadPool.find(func(next *ThreadState) interface{} {
			if numPending > 0 && next.flushPending {
				if dwpt := fc.tryCheckoutForFlush(next); dwpt != nil {
					return dwpt
				}
			}
			return nil
		}).(*DocumentsWriterPerThread)
	}
	return dwpt
}

func (fc *DocumentsWriterFlushControl) addFlushableState(perThread *ThreadState) {
	if fc.infoStream.IsEnabled("DWFC") {
		fc.infoStream.Message("DWFC", "addFlushableState %v", perThread.dwpt)
	}
	dwpt := perThread.dwpt
	assert(perThread.isActive && perThread.dwpt != nil)
	assert(fc.fullFlush)
	assert(dwpt.deleteQueue != fc.documentsWriter.deleteQueue)
	if dwpt.numDocsInRAM > 0 {
		func() {
			fc.Lock()
			defer fc.Unlock()
			if !perThread.flushPending {
				fc._setFlushPending(perThread)
			}
			flushingDWPT := fc._tryCheckOutForFlush(perThread)
			assert2(flushingDWPT != nil, "DWPT must never be null here since we hold the lock and it holds documents")
			assert2(dwpt == flushingDWPT, "flushControl returned different DWPT")
			fc.fullFlushBuffer = append(fc.fullFlushBuffer, flushingDWPT)
		}()
	} else {
		fc.perThreadPool.reset(perThread, fc.closed) // make this state inactive
	}
}

/*
Prunes the blockedQueue by removing all DWPT that are associated with
the given flush queue.
*/
func (fc *DocumentsWriterFlushControl) pruneBlockedQueue(flushingQueue *DocumentsWriterDeleteQueue) {
	for e := fc.blockedFlushes.Front(); e != nil; e = e.Next() {
		if blockedFlush := e.Value.(*BlockedFlush); blockedFlush.dwpt.deleteQueue == flushingQueue {
			fc.blockedFlushes.Remove(e)
			_, ok := fc.flushingWriters[blockedFlush.dwpt]
			assert2(!ok, "DWPT is already flushing")
			// Record the flushing DWPT to reduce flushBytes in doAfterFlush
			fc.flushingWriters[blockedFlush.dwpt] = blockedFlush.bytes
			// don't decr pending here - its already done when DWPT is blocked
			fc.flushQueue.PushBack(blockedFlush.dwpt)
		}
	}
}

func (fc *DocumentsWriterFlushControl) finishFullFlush() {
	fc.Lock()
	defer fc.Unlock()

	assert(fc.fullFlush)
	assert(fc.flushQueue.Len() == 0)
	assert(len(fc.flushingWriters) == 0)

	defer func() { fc.fullFlush = false }()

	if fc.blockedFlushes.Len() > 0 {
		fc.assertBlockedFlushes(fc.documentsWriter.deleteQueue)
		fc.pruneBlockedQueue(fc.documentsWriter.deleteQueue)
		assert(fc.blockedFlushes.Len() == 0)
	}
}

func (fc *DocumentsWriterFlushControl) abortFullFlushes(newFiles map[string]bool) {
	fc.Lock()
	defer fc.Unlock()
	defer func() { fc.fullFlush = false }()
}

func (fc *DocumentsWriterFlushControl) assertBlockedFlushes(flushingQueue *DocumentsWriterDeleteQueue) {
	for e := fc.blockedFlushes.Front(); e != nil; e = e.Next() {
		blockedFlush := e.Value.(*BlockedFlush)
		assert(blockedFlush.dwpt.deleteQueue == flushingQueue)
	}
}

func (fc *DocumentsWriterFlushControl) abortPendingFlushes(newFiles map[string]bool) {
	fc.Lock()
	defer fc.Unlock()

	defer func() {
		fc.flushQueue.Init()
		fc.blockedFlushes.Init()
		fc.updateStallState()
	}()

	for e := fc.flushQueue.Front(); e != nil; e = e.Next() {
		dwpt := e.Value.(*DocumentsWriterPerThread)
		fc.documentsWriter.subtractFlushedNumDocs(dwpt.numDocsInRAM)
		fc.doAfterFlush(dwpt)
	}

	for e := fc.blockedFlushes.Front(); e != nil; e = e.Next() {
		blockedFlush := e.Value.(*BlockedFlush)
		fc.flushingWriters[blockedFlush.dwpt] = blockedFlush.bytes
		fc.documentsWriter.subtractFlushedNumDocs(blockedFlush.dwpt.numDocsInRAM)
		blockedFlush.dwpt.abort(newFiles)
		fc.doAfterFlush(blockedFlush.dwpt)
	}
}

type BlockedFlush struct {
	dwpt  *DocumentsWriterPerThread
	bytes int64
}
