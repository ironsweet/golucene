package index

import (
	"container/list"
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
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

func (fc *DocumentsWriterFlushControl) memoryCheck() bool {
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
	assert(fc.memoryCheck())
}

func (fc *DocumentsWriterFlushControl) updateStallState() {
	panic("not implemented yet")
}

func (fc *DocumentsWriterFlushControl) waitForFlush() {
	fc.Lock()
	defer fc.Unlock()

	for len(fc.flushingWriters) > 0 {
		fc.condFlushWait.Wait()
	}
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
	assert(fc.memoryCheck())
	// Take it out of the loop this DWPT is stale
	fc.perThreadPool.reset(state, fc.closed)
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
		fc.perThreadPool.deactivateUnreleasedStates()
	}
}

func (fc *DocumentsWriterFlushControl) markForFullFlush() {
	panic("not implemented yet")
}

func (fc *DocumentsWriterFlushControl) nextPendingFlush() *DocumentsWriterPerThread {
	panic("not implemented yet")
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
