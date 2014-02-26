package index

import (
	"container/list"
	"fmt"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"sync"
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
	directory    store.Directory
	closed       bool // volatile
	infoStream   util.InfoStream
	config       *LiveIndexWriterConfig
	numDocsInRAM int // atomic

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
	events        *list.List // synchronized
}

func newDocumentsWriter(writer *IndexWriter, config *LiveIndexWriterConfig, directory store.Directory) *DocumentsWriter {
	ans := &DocumentsWriter{
		deleteQueue:   newDocumentsWriterDeleteQueue(),
		ticketQueue:   newDocumentsWriterFlushQueue(),
		directory:     directory,
		config:        config,
		infoStream:    config.infoStream,
		perThreadPool: config.indexerThreadPool,
		flushPolicy:   config.flushPolicy,
		writer:        writer,
		events:        list.New(),
	}
	ans.flushControl = newDocumentsWriterFlushControl(ans, config, writer.bufferedDeletesStream)
	return ans
}

// index/DocumentsWriterPerThread.java

// Returns the DocConsumer that the DocumentsWriter calls to
// process the documents.
type IndexingChain func(documentsWriterPerThread *DocumentsWrtierPerThread) DocConsumer

var defaultIndexingChain = func(documentsWriterPerThread *DocumentsWrtierPerThread) DocConsumer {
	panic("not implemented yet")
}

type DocumentsWrtierPerThread struct {
}

// L600
// if you increase this, you must fix field cache impl for
// Terms/TermsIndex requires <= 32768
const MAX_TERM_LENGTH_UTF8 = util.BYTE_BLOCK_SIZE - 2

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
type ThreadState struct{}

func newThreadState() *ThreadState {
	return &ThreadState{}
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
affinity, I will use channel and concurrent running goroutines to
hold individual DocumentsWriterPerThead instances and states.
*/
type DocumentsWriterPerThreadPool struct {
	allocator             ThreadStateAllocator
	threadStates          []*ThreadState
	numThreadStatesActive int // volatile
}

type ThreadStateAllocator interface {
	getAndLock(id int, documentsWriter *DocumentsWriter) *ThreadState
}

func newDocumentsWriterPerThreadPool(maxNumThreadStates int) *DocumentsWriterPerThreadPool {
	assert2(maxNumThreadStates >= 1, fmt.Sprintf("maxNumThreadStates must be >= 1 but was: %v", maxNumThreadStates))
	threadStates := make([]*ThreadState, maxNumThreadStates)
	for i, _ := range threadStates {
		threadStates[i] = newThreadState()
	}
	return &DocumentsWriterPerThreadPool{
		threadStates:          threadStates,
		numThreadStatesActive: 0,
	}
}

// Returns the max number of ThreadState instances available in this
// DocumentsWriterPerThreadPool
func (tp *DocumentsWriterPerThreadPool) maxThreadStates() int {
	return len(tp.threadStates)
}

// index/ThreadAffinityDocumentsWrierThreadPool.java

/*
A DocumentsWriterPerThreadPool implementation that tries to assign an
indexing goroutine to the same ThreadState each time the goroutine
tries to obtain a TheadState. Once a new ThreadState is created it is
associated with the creating goroutine. Subsequently, if the
goroutines associated ThreadState is not in use it will be associated
with the requesting goroutine. Otherwise, if the ThreadState is used
by another goroutine, ThreadAffinityDocumentsWriterPerThreadPool
tries to find the curently minimal contended ThreadState.
*/
type ThreadAffinityAllocator struct {
	sync.Locker
	bindings map[int]*ThreadState // synchronized
}

// Creates a new ThreadAffinityDocumentsWriterThreadPool with a given
// maximum of ThreadStates.
func newThreadAffinityDocumentsWriterPerThreadPool(maxNumPerThreads int) *DocumentsWriterPerThreadPool {
	ans := newDocumentsWriterPerThreadPool(maxNumPerThreads)
	ans.allocator = &ThreadAffinityAllocator{
		&sync.Mutex{},
		make(map[int]*ThreadState),
	}
	assert(ans.maxThreadStates() >= 1)
	return ans
}

func (alloc *ThreadAffinityAllocator) getAndLock(id int, documentsWriter *DocumentsWriter) *ThreadState {
	// alloc.Lock()
	// threadState := alloc.bindings[id]
	// alloc.Unlock()
	// if threadState != nil && threadState.tryLock() {
	// 	return threadState
	// }
	// var minThreadState *ThreadState

	// GoLucene use channel to implement the pool so it doesn't need to
	// worry about contention at all.
	// Find the state that has minimum number of goroutines waiting
	panic("not implemented yet")
}

// index/FrozenBufferedDeletes.java

/*
Holds buffered deletes by term or query, once pushed. Pushed delets
are write-once, so we shift to more memory efficient data structure
to hold them. We don't hold docIDs because these are applied on flush.
*/
type FrozenBufferedDeletes struct {
}
