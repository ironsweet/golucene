package index

import (
	"fmt"
	"sync"
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
has added its  first document since for its segments private deletes
only the deletes after that document are relevant. The global pool
instead starts maintaining the head once this instance is created by
taking the sentinel instance as its initial head.

Since each DeleteSlicemaintains its own head and list is only single
linked the garbage collector takes care of pruning the list for us.
All nodes in the list that are still relevant should be either
directly or indirectly referenced by one of the DWPT's private
DeleteSlice or by the global BufferedDeletes slice.

Each DWPT as well as the global delete pool maintain their private
DeleteSlice instance. In the DWPT case updating a slice is equivalent
to atomically finishing the document. The slice update guarantees a
"happens before" relationship to all other updates in the same
indexing session. When a DWPT updates a document it:

1. consumes a document and finishes its processing
2. updates its private DeleteSlice either by calling updateSlice() or
   addTermToDeleteSlice() (if the document has a delTerm)
3. applies all deletes in the slice to its private BufferedDeletes
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
	globalBufferedDeletes *BufferedDeletes
	globalBufferLock      sync.Locker

	generation int64
}

func newDocumentsWriterDeleteQueue() *DocumentsWriterDeleteQueue {
	return newDocumentsWriterDeleteQueueWithGeneration(0)
}

func newDocumentsWriterDeleteQueueWithGeneration(generation int64) *DocumentsWriterDeleteQueue {
	return newDocumentsWriterDeleteQueueWith(newBufferedDeletes(), generation)
}

func newDocumentsWriterDeleteQueueWith(globalBufferedDeletes *BufferedDeletes, generation int64) *DocumentsWriterDeleteQueue {
	tail := newNode(nil)
	return &DocumentsWriterDeleteQueue{
		globalBufferedDeletes: globalBufferedDeletes,
		globalBufferLock:      &sync.Mutex{},
		generation:            generation,
		// we use a sentinel instance as our initial tail. No slice will
		// ever try to apply this tail since the head is always omitted.
		tail:        tail, // sentinel
		globalSlice: newDeleteSlice(tail),
	}
}

func (dq *DocumentsWriterDeleteQueue) anyChanges() bool {
	panic("not implemented yet")
}

func (dq *DocumentsWriterDeleteQueue) clear() {
	dq.globalBufferLock.Lock()
	defer dq.globalBufferLock.Unlock()

	currentTail := dq.tail
	dq.globalSlice.head, dq.globalSlice.tail = currentTail, currentTail
	dq.globalBufferedDeletes.clear()
}

func (dq *DocumentsWriterDeleteQueue) String() string {
	return fmt.Sprintf("DWDQ: [ generation: %v ]", dq.generation)
}

type DeleteSlice struct {
	// No need to be volatile, slices are thread captive
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

type Node struct {
	next *Node // volatile
	item interface{}
}

func newNode(item interface{}) *Node {
	return &Node{item: item}
}
