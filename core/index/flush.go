package index

import (
	"github.com/balzaczyy/golucene/core/util"
	"sync"
)

/*
FlushPlicy controls when segments are flushed from a RAM resident
internal data-structure to the IndexWriter's Directory.

Segments are traditionally flushed by:
1. RAM consumption - configured via IndexWriterConfig.SetRAMBufferSizeMB()
2. Number of RAM resident documents - configured via IndexWriterConfig.SetMaxBufferedDocs()

The policy also applies pending delete operations (by term and/or
query), given the threshold set in IndexcWriterConfig.SetMaxBufferedDeleteTerms().

IndexWriter consults the provided FlushPolicy to control the flushing
process. The policy is informed for each added or updated document as
well as for each delete term. Based on the FlushPolicy, the
information provided via ThreadState and DocumentsWriterFlushControl,
the FlushPolicy decides if a DocumentsWriterPerThread needs flushing
and mark it as flush-pending via DocumentsWriterFlushControl.SetFLushingPending(),
or if deletes need to be applied.
*/
type FlushPolicy interface {
	// Called for each delete term. If this is a delte triggered due to
	// an update the given ThreadState is non-nil.
	//
	// Note: this method is called synchronized on the given
	// DocumentsWriterFlushControl and it is guaranteed that the
	// calling goroutine holds the lock on the given ThreadState
	onDelete(*DocumentsWriterFlushControl, *ThreadState)
	// Called for each document update on the given ThreadState's DWPT
	//
	// Note: this method is called synchronized on the given DWFC and
	// it is guaranteed that the calling thread holds the lock on the
	// given ThreadState
	onUpdate(*DocumentsWriterFlushControl, *ThreadState)
	// Called for each document addition on the given ThreadState's DWPT.
	//
	// Note: this method is synchronized by the given DWFC and it is
	// guaranteed that the calling thread holds the lock on the given
	// ThreadState
	onInsert(*DocumentsWriterFlushControl, *ThreadState)
	// Called by DocumentsWriter to initialize the FlushPolicy
	init(indexWriterConfig *LiveIndexWriterConfig)
}

type FlushPolicyImplSPI interface {
	onInsert(*DocumentsWriterFlushControl, *ThreadState)
	onDelete(*DocumentsWriterFlushControl, *ThreadState)
}

type FlushPolicyImpl struct {
	sync.Locker
	spi               FlushPolicyImplSPI
	indexWriterConfig *LiveIndexWriterConfig
	infoStream        util.InfoStream
}

func newFlushPolicyImpl(spi FlushPolicyImplSPI) *FlushPolicyImpl {
	return &FlushPolicyImpl{
		Locker: &sync.Mutex{},
		spi:    spi,
	}
}

func (fp *FlushPolicyImpl) onUpdate(control *DocumentsWriterFlushControl, state *ThreadState) {
	fp.spi.onInsert(control, state)
	fp.spi.onDelete(control, state)
}

func (fp *FlushPolicyImpl) init(indexWriterConfig *LiveIndexWriterConfig) {
	fp.Lock() // synchronized
	defer fp.Unlock()
	fp.indexWriterConfig = indexWriterConfig
	fp.infoStream = indexWriterConfig.infoStream
}

// index/FlushByRamOrCountsPolicy.java

/*
Default FlushPolicy implementation that flushes new segments based on
RAM used and document count depending on the IndexWriter's
IndexWriterConfig. It also applies pending deletes based on the
number of buffered delete terms.

1. onDelete() - applies pending delete operations based on the global
number of buffered delete terms iff MaxBufferedDeleteTerms() is
enabled
2. onInsert() - flushes either on the number of documents per
DocumentsWriterPerThread (NumDocsInRAM()) or on the global active
memory consumption in the current indexing session iff
MaxBufferedDocs() or RAMBufferSizeMB() is enabled respectively
3. onUpdate() - calls onInsert() and onDelete() in order

All IndexWriterConfig settings are used to mark DocumentsWriterPerThread
as flush pending during indexing with respect to their live updates.

If SetRAMBufferSizeMB() is enabled, the largest ram consuming
DocumentsWriterPerThread will be marked as pending iff the global
active RAM consumption is >= the configured max RAM buffer.
*/
type FlushByRamOrCountsPolicy struct {
	*FlushPolicyImpl
}

func newFlushByRamOrCountsPolicy() *FlushByRamOrCountsPolicy {
	ans := new(FlushByRamOrCountsPolicy)
	ans.FlushPolicyImpl = newFlushPolicyImpl(ans)
	return ans
}

func (p *FlushByRamOrCountsPolicy) onDelete(control *DocumentsWriterFlushControl, state *ThreadState) {
	panic("not implemented yet")
}

func (p *FlushByRamOrCountsPolicy) onInsert(control *DocumentsWriterFlushControl, state *ThreadState) {
	panic("not implemented yet")
}
