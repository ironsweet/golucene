package index

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// index/DocumentsWriterDeleteQueue.java

/*
DocumentsWriterDeleteQueue is a non-blocking linked pending deletes
queue. In contrast to other queue implementation we only maintain the
tail of the queue. A delete queue is always used in a context of a
set of DWPTs and a global delete pool. Each of the DWPT and the
global pool need to maintain their 'own' head of the queue (as a
DeleteSlice instance per DWPT). The difference between the DWPT and
the global pool is that the DWPT starts maintaining a head once it
has added its first document since for its segments private deletes
only the deletes after that document are relevant. The global pool
instead starts maintaining the head once this instance is created by
taking the sentinel instance as its initial head.

Since each DeleteSlice maintains its own head and list is only single
linked, the garbage collector takes care of pruning the list for us.
All nodes in the list that are still relevant should be either
directly or indirectly referenced by one of the DWPT's private
DeleteSlice or by the global BufferedUpdates slice.

Each DWPT as well as the global delete pool maintain their private
DeleteSlice instance. In the DWPT case, updating a slice is equivalent
to atomically finishing the document. The slice update guarantees a
"happens before" relationship to all other updates in the same
indexing session. When a DWPT updates a document it:

1. consumes a document and finishes its processing
2. updates its private DeleteSlice either by calling updateSlice() or
   addTermToDeleteSlice() (if the document has a delTerm)
3. applies all deletes in the slice to its private BufferedUpdates
   and resets it
4. increments its internal document id

The DWPT also doesn't apply its current docments delete term until it
has updated its delete slice which ensures the consistency of the
update. If the update fails before the DeleteSlice could have been
updated the deleteTerm will also not be added to its private deletes
neither to the global deletes.
*/
type DocumentsWriterDeleteQueue struct {
	tail                  *Node // volatile
	globalSlice           *DeleteSlice
	globalBufferedUpdates *BufferedUpdates
	globalBufferLock      sync.Locker

	generation int64
}

func newDocumentsWriterDeleteQueue() *DocumentsWriterDeleteQueue {
	return newDocumentsWriterDeleteQueueWithGeneration(0)
}

func newDocumentsWriterDeleteQueueWithGeneration(generation int64) *DocumentsWriterDeleteQueue {
	return newDocumentsWriterDeleteQueueWith(newBufferedUpdates(), generation)
}

func newDocumentsWriterDeleteQueueWith(globalBufferedUpdates *BufferedUpdates, generation int64) *DocumentsWriterDeleteQueue {
	tail := newNode(nil)
	return &DocumentsWriterDeleteQueue{
		globalBufferedUpdates: globalBufferedUpdates,
		globalBufferLock:      &sync.Mutex{},
		generation:            generation,
		// we use a sentinel instance as our initial tail. No slice will
		// ever try to apply this tail since the head is always omitted.
		tail:        tail, // sentinel
		globalSlice: newDeleteSlice(tail),
	}
}

/* Invariant for document update */
func (q *DocumentsWriterDeleteQueue) add(term *Term, slice *DeleteSlice) {
	panic("not implemented yet")
}

func (dq *DocumentsWriterDeleteQueue) freezeGlobalBuffer(callerSlice *DeleteSlice) *FrozenBufferedUpdates {
	dq.globalBufferLock.Lock()
	defer dq.globalBufferLock.Unlock()

	// Here we freeze the global buffer so we need to lock it, apply
	// all deletes in the queue and reset the global slice to let the
	// GC prune the queue.
	currentTail := dq.tail
	// take the current tail and make this local. Any changes after
	// this call are applied later and not relevant here
	if callerSlice != nil {
		// update the callers slices so we are on the same page
		callerSlice.tail = currentTail
	}
	if dq.globalSlice.tail != currentTail {
		dq.globalSlice.tail = currentTail
		dq.globalSlice.apply(dq.globalBufferedUpdates, MAX_INT)
	}

	packet := freezeBufferedUpdates(dq.globalBufferedUpdates, false)
	dq.globalBufferedUpdates.clear()
	return packet
}

func (dq *DocumentsWriterDeleteQueue) anyChanges() bool {
	dq.globalBufferLock.Lock()
	defer dq.globalBufferLock.Unlock()
	// check if all items in the global slice were applied
	// and if the global slice is up-to-date
	// and if globalBufferedUpdates has changes
	return dq.globalBufferedUpdates.any() ||
		!dq.globalSlice.isEmpty() ||
		dq.globalSlice.tail != dq.tail ||
		dq.tail.next != nil
}

func (dq *DocumentsWriterDeleteQueue) newSlice() *DeleteSlice {
	return newDeleteSlice(dq.tail)
}

func (q *DocumentsWriterDeleteQueue) updateSlice(slice *DeleteSlice) bool {
	if slice.tail != q.tail { // if we are the same just
		slice.tail = q.tail
		return true
	}
	return false
}

func (dq *DocumentsWriterDeleteQueue) clear() {
	dq.globalBufferLock.Lock()
	defer dq.globalBufferLock.Unlock()

	currentTail := dq.tail
	dq.globalSlice.head, dq.globalSlice.tail = currentTail, currentTail
	dq.globalBufferedUpdates.clear()
}

func (q *DocumentsWriterDeleteQueue) RamBytesUsed() int64 {
	return atomic.LoadInt64(&q.globalBufferedUpdates.bytesUsed)
}

func (dq *DocumentsWriterDeleteQueue) String() string {
	return fmt.Sprintf("DWDQ: [ generation: %v ]", dq.generation)
}

type DeleteSlice struct {
	head *Node // we don't apply this one
	tail *Node
}

func newDeleteSlice(currentTail *Node) *DeleteSlice {
	assert(currentTail != nil)
	// Initially this is a 0 length slice pointing to the 'current'
	// tail of the queue. Once we update the slice we only need to
	// assig the tail and have a new slice
	return &DeleteSlice{head: currentTail, tail: currentTail}
}

func (ds *DeleteSlice) apply(del *BufferedUpdates, docIDUpto int) {
	if ds.head == ds.tail {
		// 0 length slice
		return
	}
	// When we apply a slice we take the head and get its next as our
	// first item to apply and continue until we applied the tail. If
	// the head and tail in this slice are not equal then there will be
	// at least one more non-nil node in the slice!
	for current := ds.head; current != ds.tail; {
		current = current.next
		assert2(current != nil,
			"slice property violated between the head on the tail must not be a null node")
		current.apply(del, docIDUpto)
	}
	ds.reset()
}

func (ds *DeleteSlice) reset() {
	// reset to 0 length slice
	ds.head = ds.tail
}

/*
Returns true iff the given item is identical to the item held by the
slice's tail, otherwise false.
*/
func (ds *DeleteSlice) isTailItem(item interface{}) bool {
	return ds.tail.item == item
}

func (ds *DeleteSlice) isEmpty() bool {
	return ds.head == ds.tail
}

type Node struct {
	next *Node // volatile
	item interface{}
}

func newNode(item interface{}) *Node {
	return &Node{item: item}
}

func (node *Node) apply(BufferedUpdates *BufferedUpdates, docIDUpto int) {
	panic("sentinel item must never be applied")
}
