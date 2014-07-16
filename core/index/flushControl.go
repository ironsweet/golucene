package index

import (
	"container/list"
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
	"math"
	"sync"
	"sync/atomic"
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
	_activeBytes        int64
	_flushBytes         int64
	numPending          int // volatile
	numDocsSinceStalled int
	flushDeletes        int32 // atomic bool
	fullFlush           bool
	flushQueue          *list.List
	// only for safety reasons if a DWPT is close to the RAM limit
	blockedFlushes  *list.List
	flushingWriters map[*DocumentsWriterPerThread]int64

	maxConfiguredRamBuffer float64
	peakActiveBytes        int64 // assert only
	peakFlushBytes         int64 // assert only
	peakNetBytes           int64 // assert only
	peakDelta              int64 // assert only
	flushByRAMWasDisabled  bool  // assert only

	*DocumentsWriterStallControl // mixin

	perThreadPool *DocumentsWriterPerThreadPool
	flushPolicy   FlushPolicy
	closed        bool

	documentsWriter       *DocumentsWriter
	config                LiveIndexWriterConfig
	bufferedUpdatesStream *BufferedUpdatesStream
	infoStream            util.InfoStream

	fullFlushBuffer []*DocumentsWriterPerThread
}

func newDocumentsWriterFlushControl(documentsWriter *DocumentsWriter,
	config LiveIndexWriterConfig,
	bufferedUpdatesStream *BufferedUpdatesStream) *DocumentsWriterFlushControl {

	objLock := &sync.Mutex{}
	ans := &DocumentsWriterFlushControl{
		Locker:                objLock,
		condFlushWait:         sync.NewCond(objLock),
		flushQueue:            list.New(),
		blockedFlushes:        list.New(),
		flushingWriters:       make(map[*DocumentsWriterPerThread]int64),
		infoStream:            config.InfoStream(),
		perThreadPool:         documentsWriter.perThreadPool,
		flushPolicy:           documentsWriter.flushPolicy,
		config:                config,
		hardMaxBytesPerDWPT:   int64(config.RAMPerThreadHardLimitMB()) * 1024 * 1024,
		documentsWriter:       documentsWriter,
		bufferedUpdatesStream: bufferedUpdatesStream,
	}
	ans.DocumentsWriterStallControl = newDocumentsWriterStallControl()
	return ans
}

func (fc *DocumentsWriterFlushControl) activeBytes() int64 {
	fc.Lock()
	defer fc.Unlock()
	return fc._activeBytes
}

func (fc *DocumentsWriterFlushControl) flushBytes() int64 {
	fc.Lock()
	defer fc.Unlock()
	return fc._flushBytes
}

func (fc *DocumentsWriterFlushControl) netBytes() int64 {
	fc.Lock()
	defer fc.Unlock()
	return fc._netBytes()
}

func (fc *DocumentsWriterFlushControl) _netBytes() int64 {
	return fc._activeBytes + fc._flushBytes
}

func (fc *DocumentsWriterFlushControl) assertMemory() {
	maxRamMB := fc.config.RAMBufferSizeMB()
	if fc.flushByRAMWasDisabled || maxRamMB == DISABLE_AUTO_FLUSH {
		fc.flushByRAMWasDisabled = true
		return
	}
	// for this assert we must be tolerant to ram buffer changes!
	if maxRamMB > fc.maxConfiguredRamBuffer {
		fc.maxConfiguredRamBuffer = maxRamMB
	}
	ram := fc._flushBytes + fc._activeBytes
	ramBufferBytes := int64(fc.maxConfiguredRamBuffer) * 1024 * 1024
	// take peakDelta into account - worst case is that all flushing,
	// pending and blocked DWPT had maxMem and the last doc had the
	// peakDelta

	/*
		2 * ramBufferBytes
		 	-> before we stall we need to across the 2xRAM Buffer border
		 	this is still a valid limit
		(numPending + numFlusingDWPT() + numBlockedFlushes()) * peakDelta
		 	-> those are the total number of DWPT that are not active but
		 	not yet fully flushed. all of them could theoretically be taken
		 	out of the loop once they crossed the RAM buffer and the last
		 	document was the peak delta
		(numDocsSinceStalled * peakDelta)
		 	-> at any given time, there could be n threads in flight that
		 	crossed the stall control before we reached the limit and each
		 	of them could hold a peak document
	*/
	expected := 2*ramBufferBytes + int64(fc.numPending+len(fc.flushingWriters)+
		fc.blockedFlushes.Len())*fc.peakDelta +
		int64(fc.numDocsSinceStalled)*fc.peakDelta

	// the expected ram consumption is an upper bound at this point and
	// not really the expected consumption
	if fc.peakDelta < (ramBufferBytes >> 1) {
		/*
			If we are indexing with very low maxRamBuffer, like 0.1MB.
			Memory can easily overflow if we check out some DWPT based on
			docCount and have several DWPT in flight indexing large
			documents (compared to the ram buffer). This means that those
			DWPT and their threads will not hit the stall control before
			asserting the memory which would in turn fail. To prevent this
			we only assert if the largest document seen is smaller tan the
			1/2 of the maxRamBufferMB
		*/
		assertn(ram <= expected,
			"actual mem: %v, expected mem: %v, flush mem: %v, active mem: %v, "+
				"pending DWPT: %v, flushing DWPT: %v, blocked DWPT: %v, "+
				"peakDelta mem: %v bytes, ramBufferBytes=%v, maxConfiguredRamBuffer=%v",
			ram, expected, fc._flushBytes, fc._activeBytes, fc.numPending,
			len(fc.flushingWriters), fc.blockedFlushes.Len(), fc.peakDelta,
			ramBufferBytes, fc.maxConfiguredRamBuffer)
	}
}

func (fc *DocumentsWriterFlushControl) commitPerThreadBytes(perThread *ThreadState) {
	delta := perThread.dwpt.bytesUsed() - perThread.bytesUsed
	perThread.bytesUsed += delta
	// We need to differentiate here if we are pending since
	// setFlushPending moves the perThread memory to the flushBytes and
	// we could be set to pending during a delete
	if perThread.flushPending {
		fc._flushBytes += delta
	} else {
		fc._activeBytes += delta
	}
	fc.assertUpdatePeaks(delta)
}

func (fc *DocumentsWriterFlushControl) assertUpdatePeaks(delta int64) {
	if fc.peakActiveBytes < fc._activeBytes {
		fc.peakActiveBytes = fc._activeBytes
	}
	if fc.peakFlushBytes < fc._flushBytes {
		fc.peakFlushBytes = fc._flushBytes
	}
	if n := fc._netBytes(); fc.peakNetBytes < n {
		fc.peakNetBytes = n
	}
	if fc.peakDelta < delta {
		fc.peakDelta = delta
	}
}

func (fc *DocumentsWriterFlushControl) doAfterDocument(perThread *ThreadState, isUpdate bool) *DocumentsWriterPerThread {
	fc.Lock()
	defer fc.Unlock()

	defer func() {
		stalled := fc.updateStallState()
		fc.assertNumDocsSinceStalled(stalled)
		fc.assertMemory()
	}()

	fc.commitPerThreadBytes(perThread)
	if !perThread.flushPending {
		if isUpdate {
			fc.flushPolicy.onUpdate(fc, perThread)
		} else {
			fc.flushPolicy.onInsert(fc, perThread)
		}
		if !perThread.flushPending && perThread.bytesUsed > fc.hardMaxBytesPerDWPT {
			// Safety check to prevent a single DWPT exceeding its RAM
			// limit. This is super important since we can not address more
			// than 2048 MB per DWPT
			fc._setFlushPending(perThread)
		}
	}
	var flushingDWPT *DocumentsWriterPerThread
	if fc.fullFlush {
		if perThread.flushPending {
			fc.checkoutAndBlock(perThread)
			flushingDWPT = fc.nextPendingFlush()
		}
	} else {
		flushingDWPT = fc._tryCheckoutForFlush(perThread)
	}
	return flushingDWPT
}

/*
updates the number of documents "finished" while we are in a stalled
state. this is important for asserting memory upper bounds since it
corresponds to the number of threads that are in-flight and crossed
the stall control check before we actually stalled.
*/
func (fc *DocumentsWriterFlushControl) assertNumDocsSinceStalled(stalled bool) {
	if fc.stalled {
		fc.numDocsSinceStalled++
	} else {
		fc.numDocsSinceStalled = 0
	}
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
	fc._flushBytes -= bytes
	// fc.perThreadPool.recycle(dwpt)
	fc.assertMemory()
}

func (fc *DocumentsWriterFlushControl) updateStallState() bool {
	var limit int64 = math.MaxInt64
	if maxRamMB := fc.config.RAMBufferSizeMB(); maxRamMB != DISABLE_AUTO_FLUSH {
		limit = int64(2 * 1024 * 1024 * maxRamMB)
	}
	// We block indexing threads if net byte grows due to slow flushes,
	// yet, for small ram buffers and large documents, we can easily
	// reach the limit without any ongoing flushes. We need ensure that
	// we don't stall/block if an ongoing or pending flush can not free
	// up enough memory to release the stall lock.
	stall := (fc._activeBytes+fc._flushBytes) > limit &&
		fc._activeBytes < limit &&
		!fc.closed
	fc.updateStalled(stall)
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
func (fc *DocumentsWriterFlushControl) setFlushPending(perThread *ThreadState) {
	fc.Lock()
	defer fc.Unlock()
	fc._setFlushPending(perThread)
}

func (fc *DocumentsWriterFlushControl) _setFlushPending(perThread *ThreadState) {
	assert(!perThread.flushPending)
	if perThread.dwpt.numDocsInRAM > 0 {
		perThread.flushPending = true // write access synced
		bytes := perThread.bytesUsed
		fc._flushBytes += bytes
		fc._activeBytes -= bytes
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
		fc._flushBytes -= state.bytesUsed
	} else {
		fc._activeBytes -= state.bytesUsed
	}
	fc.assertMemory()
	// Take it out of the loop this DWPT is stale
	fc.perThreadPool.reset(state, fc.closed)
}

func (fc *DocumentsWriterFlushControl) tryCheckoutForFlush(perThread *ThreadState) *DocumentsWriterPerThread {
	fc.Lock()
	defer fc.Unlock()
	return fc._tryCheckoutForFlush(perThread)
}

func (fc *DocumentsWriterFlushControl) _tryCheckoutForFlush(perThread *ThreadState) *DocumentsWriterPerThread {
	if perThread.flushPending {
		return fc.internalTryCheckOutForFlush(perThread)
	}
	return nil
}

func (fc *DocumentsWriterFlushControl) checkoutAndBlock(perThread *ThreadState) {
	// perThread is already locked
	assert2(perThread.flushPending, "can not block non-pending threadstate")
	assert2(fc.fullFlush, "can not block if fullFlush == false")
	bytes := perThread.bytesUsed
	dwpt := fc.perThreadPool.reset(perThread, fc.closed)
	fc.numPending--
	fc.blockedFlushes.PushBack(&BlockedFlush{dwpt, bytes})
}

func (fc *DocumentsWriterFlushControl) internalTryCheckOutForFlush(perThread *ThreadState) *DocumentsWriterPerThread {
	// perThread is already locked
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

/* Various statistics */

func (fc *DocumentsWriterFlushControl) deleteBytesUsed() int64 {
	return fc.documentsWriter.deleteQueue.RamBytesUsed() + fc.bufferedUpdatesStream.RamBytesUsed()
}

// L444

func (fc *DocumentsWriterFlushControl) obtainAndLock() *ThreadState {
	perThread := fc.perThreadPool.lockAny()
	var success = false
	defer func() {
		if !success {
			fc.perThreadPool.release(perThread)
		}
	}()

	if perThread.isActive &&
		perThread.dwpt != nil &&
		perThread.dwpt.deleteQueue != fc.documentsWriter.deleteQueue {

		// Threre is a flush-all in process and this DWPT is now stale --
		// enroll it for flush and try for another DWPT:
		fc.addFlushableState(perThread)
	}
	success = true
	// simply return the ThreadState even in a flush all case since we
	// already hold the lock
	return perThread
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
			fc.flushQueue.Remove(e)
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
			flushingDWPT := fc.internalTryCheckOutForFlush(perThread)
			assert2(flushingDWPT != nil, "DWPT must never be null here since we hold the lock and it holds documents")
			assert2(dwpt == flushingDWPT, "flushControl returned different DWPT")
			fc.fullFlushBuffer = append(fc.fullFlushBuffer, flushingDWPT)
		}()
	} else {
		fc.perThreadPool.reset(perThread, fc.closed) // make this state inactive
	}
}

func (fc *DocumentsWriterFlushControl) getAndResetApplyAllDeletes() bool {
	return atomic.SwapInt32(&fc.flushDeletes, 0) == 1
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

func (fc *DocumentsWriterFlushControl) numQueuedFlushes() int {
	fc.Lock()
	defer fc.Unlock()
	return fc.flushQueue.Len()
}

type BlockedFlush struct {
	dwpt  *DocumentsWriterPerThread
	bytes int64
}

/*
This mehtod will block if too many DWPT are currently flushing and no
checked out DWPT are available.
*/
func (fc *DocumentsWriterFlushControl) waitIfStalled() {
	if fc.infoStream.IsEnabled("DWFC") {
		fc.infoStream.Message(
			"DWFC", "waitIfStalled: numFlushesPending: %v netBytes: %v flushingBytes: %v fullFlush: %v",
			fc.flushQueue.Len(), fc.netBytes(), fc.flushBytes(), fc.fullFlush)
	}
	fc.DocumentsWriterStallControl.waitIfStalled()
}
