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
	init(indexWriterConfig LiveIndexWriterConfig)
}

type FlushPolicyImplSPI interface {
	onInsert(*DocumentsWriterFlushControl, *ThreadState)
	onDelete(*DocumentsWriterFlushControl, *ThreadState)
}

type FlushPolicyImpl struct {
	sync.Locker
	spi               FlushPolicyImplSPI
	indexWriterConfig LiveIndexWriterConfig
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

func (fp *FlushPolicyImpl) init(indexWriterConfig LiveIndexWriterConfig) {
	fp.Lock() // synchronized
	defer fp.Unlock()
	fp.indexWriterConfig = indexWriterConfig
	fp.infoStream = indexWriterConfig.InfoStream()
}

/*
Returns the current most RAM consuming non-pending ThreadState with
at least one indexed document.

This method will never return nil
*/
func (p *FlushPolicyImpl) findLargestNonPendingWriter(control *DocumentsWriterFlushControl,
	perThreadState *ThreadState) *ThreadState {
	assert(perThreadState.dwpt.numDocsInRAM > 0)
	maxRamSoFar := perThreadState.bytesUsed
	// the dwpt which needs to be flushed eventually
	maxRamUsingThreadState := perThreadState
	assert2(!perThreadState.flushPending, "DWPT should have flushed")
	count := 0
	control.perThreadPool.foreach(func(next *ThreadState) {
		if !next.flushPending {
			if nextRam := next.bytesUsed; nextRam > 0 && next.dwpt.numDocsInRAM > 0 {
				if p.infoStream.IsEnabled("FP") {
					p.infoStream.Message("FP", "thread state has %v bytes; docInRAM=%v",
						nextRam, next.dwpt.numDocsInRAM)
				}
				count++
				if nextRam > maxRamSoFar {
					maxRamSoFar = nextRam
					maxRamUsingThreadState = next
				}
			}
		}
	})
	if p.infoStream.IsEnabled("FP") {
		p.infoStream.Message("FP", "%v in-use non-flusing threads states", count)
	}
	p.message("set largest ram consuming thread pending on lower watermark")
	return maxRamUsingThreadState
}

func (p *FlushPolicyImpl) message(s string) {
	if p.infoStream.IsEnabled("FP") {
		p.infoStream.Message("FP", s)
	}
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
	if p.flushOnDocCount() && state.dwpt.numDocsInRAM >= p.indexWriterConfig.MaxBufferedDocs() {
		// flush this state by num docs
		control.setFlushPending(state)
	} else if p.flushOnRAM() { // flush by RAM
		limit := int64(p.indexWriterConfig.RAMBufferSizeMB() * 1024 * 1024)
		totalRam := control._activeBytes + control.deleteBytesUsed() // safe w/o sync
		if totalRam >= limit {
			if p.infoStream.IsEnabled("FP") {
				p.infoStream.Message("FP",
					"trigger flush: activeBytes=%v deleteBytes=%v vs limit=%v",
					control._activeBytes, control.deleteBytesUsed(), limit)
			}
			p.markLargestWriterPending(control, state, totalRam)
		}
	}
}

/* Marks the mos tram consuming active DWPT flush pending */
func (p *FlushByRamOrCountsPolicy) markLargestWriterPending(control *DocumentsWriterFlushControl,
	perThreadState *ThreadState, currentBytesPerThread int64) {
	control.setFlushPending(p.findLargestNonPendingWriter(control, perThreadState))
}

/* Returns true if this FLushPolicy flushes on IndexWriterConfig.MaxBufferedDocs(), otherwise false */
func (p *FlushByRamOrCountsPolicy) flushOnDocCount() bool {
	return p.indexWriterConfig.MaxBufferedDocs() != DISABLE_AUTO_FLUSH
}

/* Returns true if this FlushPolicy flushes on IndexWriterConfig.RAMBufferSizeMB(), otherwise false */
func (p *FlushByRamOrCountsPolicy) flushOnRAM() bool {
	return p.indexWriterConfig.RAMBufferSizeMB() != DISABLE_AUTO_FLUSH
}
