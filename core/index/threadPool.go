package index

import (
	"container/list"
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
	id   int // used by pool
	dwpt *DocumentsWriterPerThread
	// TODO this should really be part of DocumentsWriterFlushControl
	// write access guarded by DocumentsWriterFlushControl
	flushPending bool // volatile
	// TODO this should really be part of DocumentsWriterFlushControl
	// write access guarded by DocumentsWriterFlushControl
	bytesUsed int64
	isActive  bool
}

func newThreadState(id int) *ThreadState {
	return &ThreadState{id: id, isActive: true}
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
	threadStates  []*ThreadState
	listeners     []*list.List
	freeList      *list.List
	lockedList    *list.List
	hasMoreStates *sync.Cond
}

func NewDocumentsWriterPerThreadPool(maxNumThreadStates int) *DocumentsWriterPerThreadPool {
	assert2(maxNumThreadStates >= 1, fmt.Sprintf("maxNumThreadStates must be >= 1 but was: %v", maxNumThreadStates))
	return &DocumentsWriterPerThreadPool{
		Locker:        &sync.Mutex{},
		threadStates:  make([]*ThreadState, 0, maxNumThreadStates),
		listeners:     make([]*list.List, maxNumThreadStates),
		freeList:      list.New(),
		lockedList:    list.New(),
		hasMoreStates: sync.NewCond(&sync.Mutex{}),
	}
}

func (tp *DocumentsWriterPerThreadPool) numActiveThreadState() int {
	return len(tp.threadStates)
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
strategy is LIFO.
*/
func (tp *DocumentsWriterPerThreadPool) lockAny() (res *ThreadState) {
	for res == nil {
		if res = tp.findNextAvailableThreadState(); res == nil {
			// ThreadState is already locked before return by this method:
			if res = tp.newThreadState(); res == nil {
				// Wait until a thread state freez up:
				func() {
					tp.hasMoreStates.L.Lock()
					defer tp.hasMoreStates.L.Unlock()

					tp.hasMoreStates.Wait()
				}()
			}
		}
	}
	return
}

func (tp *DocumentsWriterPerThreadPool) lock(id int, wait bool) *ThreadState {
	tp.Lock()
	defer tp.Unlock()

	for e := tp.freeList.Front(); e != nil; e = e.Next() {
		if tid := e.Value.(int); tid == id {
			tp.freeList.Remove(e)
			tp.lockedList.PushBack(id)
			return tp.threadStates[tid]
		}
	}

	if !wait {
		return nil
	}
	waitingList := tp.listeners[id]
	if waitingList == nil {
		waitingList = list.New()
		tp.listeners[id] = waitingList
	}
	ch := make(chan *ThreadState)
	waitingList.PushBack(ch)
	return <-ch // block until reserved thread state is released
}

func (tp *DocumentsWriterPerThreadPool) findNextAvailableThreadState() *ThreadState {
	tp.Lock()
	defer tp.Unlock()

	if tp.freeList.Len() > 0 {
		// Important that we are LIFO here! This way if number of
		// concurrent indexing threads was once high, but has now
		// reduced, we only use a limited number of thread states:
		e := tp.freeList.Back()
		id := e.Value.(int)
		tp.freeList.Remove(e)

		if tp.threadStates[id].dwpt == nil {
			// This thread-state is not initialized, e.g. it was just
			// flushed. See if we can instead find another free thread
			// state that already has docs indexed. This way if incoming
			// thread concurrentcy has decreased, we don't leave docs
			// indefinitely buffered, tying up RAM. This will instead get
			// thread states flushed, freein up RAM for larger segment
			// flushes:
			for e = tp.freeList.Front(); e != nil; e = e.Next() {
				if id2 := e.Value.(int); tp.threadStates[id2].dwpt != nil {
					// Use this one instead, and swap it with the
					// un-initialized one:
					tp.freeList.PushFront(id)
					id = id2
					tp.freeList.Remove(e)
					break
				}
			}
		}
		return tp.threadStates[id]
	}
	return nil
}

/*
Returns a new ThreadState iff any new state is available otherwise
nil.

NOTE: the returned ThreadState is already locked iff non-nil.
*/
func (tp *DocumentsWriterPerThreadPool) newThreadState() (ts *ThreadState) {
	tp.Lock()
	defer tp.Unlock()

	// Create a new empty thread state if possible
	if len(tp.threadStates) < cap(tp.threadStates) {
		ts = newThreadState(len(tp.threadStates))
		tp.threadStates = append(tp.threadStates, ts)
		tp.lockedList.PushBack(ts.id)
	}
	return
}

func (tp *DocumentsWriterPerThreadPool) foreach(f func(state *ThreadState)) {
	for i, limit := 0, len(tp.threadStates); i < limit; i++ {
		ts := tp.lock(i, true)
		assert(ts != nil)
		f(ts)
		tp.release(ts)
	}
}

func (tp *DocumentsWriterPerThreadPool) find(f func(state *ThreadState) interface{}) interface{} {
	for i, limit := 0, len(tp.threadStates); i < limit; i++ {
		if ts := tp.lock(i, false); ts != nil {
			res := f(ts)
			tp.release(ts)
			if res != nil {
				return res
			}
		}
	}
	return nil
}

/*
Release the ThreadState back to the pool. Equals to
ThreadState.Unlock() in Lucene Java.
*/
func (tp *DocumentsWriterPerThreadPool) release(ts *ThreadState) {
	tp.Lock()
	defer tp.Unlock()

	if waitingList := tp.listeners[ts.id]; waitingList != nil && waitingList.Len() > 0 {
		// this thread state is reserved
		e := waitingList.Front()
		waitingList.Remove(e)
		// re-allocate to external handler
		e.Value.(chan *ThreadState) <- ts
		return
	}

	// push the thread state back to
	tp.freeList.PushBack(ts.id)
	tp.hasMoreStates.Signal()
}
