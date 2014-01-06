package index

import (
	"github.com/balzaczyy/golucene/core/util"
)

// index/DocConsumer.java

type DocConsumer interface {
	processDocuments(fieldInfos *FieldInfosBuilder) error
	finishDocument() error
	flush(state *SegmentWriteState) error
	abort()
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
type DocumentsWriterPerThreadPool struct{}

func newDocumentsWriterPerThreadPool(maxNumPerThreads int) *DocumentsWriterPerThreadPool {
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
