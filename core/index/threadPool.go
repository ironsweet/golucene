package index

import (
	"fmt"
	"sync"
)

// index/DocumentsWriterPerThreadPool.java

/*
ThreadState references and guards a DocumentsWriterPerThread instance
that is used during indexing to build a in-memory index segment.
ThreadState also holds all flush related per-thread data controlled
by DocumentsWriterFlushControl.

A ThreadState, its methods and members should only accessed by one
goroutine a time. users must acquire the lock via lock() and release
the lock in a finally block via unlock() before accesing the state.
*/
type ThreadState struct {
	dwpt *DocumentsWriterPerThread
	// TODO this should really be part of DocumentsWriterFlushControl
	// write access guarded by DocumentsWriterFlushControl
	flushPending bool // volatile
	// TODO this should really be part of DocumentsWriterFlushControl
	// write access guarded by DocumentsWriterFlushControl
	bytesUsed int64
	isActive  bool
}

func newThreadState() *ThreadState {
	return &ThreadState{isActive: true}
}

func (ts *ThreadState) deactivate() {
	ts.isActive = false
	ts.reset()
}

func (ts *ThreadState) reset() {
	ts.dwpt = nil
	ts.bytesUsed = 0
	ts.flushPending = false
}

/*
DocumentsWriterPerThreadPool controls ThreadState instances and their
goroutine assignment during indexing. Each TheadState holds a
reference to a DocumentsWriterPerThread that is once a ThreadState is
obtained from the pool exclusively used for indexing a single
document by the obtaining thread. Each indexing thread must obtain
such a ThreadState to make progress. Depending on the DocumentsWriterPerThreadPool
implementation ThreadState assingments might differ from document to
document.

Once a DocumentWriterPerThread is selected for flush the thread pool
is reusing the flushing DocumentsWriterPerthread's ThreadState with a
new DocumentsWriterPerThread instance.

GoRoutine is different from Java's thread. So intead of thread
affinity, I will use channels and concurrent running goroutines to
hold individual DocumentsWriterPerThread instances and states.
*/
type DocumentsWriterPerThreadPool struct {
	sync.Locker
	threadStates          []*ThreadState
	numThreadStatesActive int // volatile
	numThreadStatesLocked int
	hasMoreStates         *sync.Cond
}

func NewDocumentsWriterPerThreadPool(maxNumThreadStates int) *DocumentsWriterPerThreadPool {
	assert2(maxNumThreadStates >= 1, fmt.Sprintf("maxNumThreadStates must be >= 1 but was: %v", maxNumThreadStates))
	return &DocumentsWriterPerThreadPool{
		Locker:        &sync.Mutex{},
		threadStates:  make([]*ThreadState, maxNumThreadStates),
		hasMoreStates: sync.NewCond(&sync.Mutex{}),
	}
}

func (tp *DocumentsWriterPerThreadPool) reset(threadState *ThreadState, closed bool) *DocumentsWriterPerThread {
	dwpt := threadState.dwpt
	if !closed {
		threadState.reset()
	} else {
		threadState.deactivate()
	}
	return dwpt
}

/*
It's unfortunately that Go doesn't support 'Thread Affinity'. Default
strategy is FIFO.
*/
func (tp *DocumentsWriterPerThreadPool) getAndLock(documentsWriter *DocumentsWriter) *ThreadState {
	res := tp.findNextAvailableThreadState()
	if res == nil {
		res = tp.newThreadState()
		if res == nil {
			// Wait for thread state released by others
			tp.hasMoreStates.L.Lock()
			defer tp.hasMoreStates.L.Unlock()

			for res == nil {
				tp.hasMoreStates.Wait()
				res = tp.findNextAvailableThreadState()
			}
		}
	}
	return res
}

func (tp *DocumentsWriterPerThreadPool) findNextAvailableThreadState() *ThreadState {
	tp.Lock()
	defer tp.Unlock()

	if tp.numThreadStatesLocked < tp.numThreadStatesActive {
		res := tp.threadStates[tp.numThreadStatesLocked]
		tp.numThreadStatesLocked++
		return res
	}
	return nil
}

func (tp *DocumentsWriterPerThreadPool) newThreadState() *ThreadState {
	tp.Lock()
	defer tp.Unlock()

	// Create a new empty thread state if possible
	if tp.numThreadStatesActive < len(tp.threadStates) {
		res := newThreadState()
		tp.threadStates[tp.numThreadStatesActive] = res
		tp.numThreadStatesActive++
		return res
	}
	return nil
}

func (tp *DocumentsWriterPerThreadPool) foreach(f func(state *ThreadState)) {
	tp.Lock()
	defer tp.Unlock()

	for _, state := range tp.threadStates {
		if state == nil {
			continue
		}
		f(state)
	}
}

/*
Release the ThreadState back to the pool. Equals to
ThreadState.Unlock() in Lucene Java.
*/
func (tp *DocumentsWriterPerThreadPool) release(ts *ThreadState) {
	tp.Lock()
	defer tp.Unlock()

	// sequential search since n is small
	var pos int = -1
	for i, v := range tp.threadStates {
		if v == ts {
			pos = i
			break
		}
	}

	if pos >= 0 { // found
		tp.numThreadStatesLocked--
		tp.threadStates[pos], tp.threadStates[tp.numThreadStatesLocked] =
			tp.threadStates[tp.numThreadStatesLocked], tp.threadStates[pos]
		tp.hasMoreStates.Signal()
	}
}
