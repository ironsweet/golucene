package index

import (
	"container/list"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	"github.com/balzaczyy/golucene/core/util"
	"sync"
)

type MergeControl struct {
	sync.Locker
	infoStream util.InfoStream

	readerPool *ReaderPool

	// Holds all SegmentInfo instances currently involved in merges
	mergingSegments map[*SegmentCommitInfo]bool

	pendingMerges *list.List
	runningMerges map[*OneMerge]bool
	mergeSignal   *sync.Cond

	stopMerges bool
}

func newMergeControl(infoStream util.InfoStream, readerPool *ReaderPool) *MergeControl {
	lock := &sync.Mutex{}
	return &MergeControl{
		Locker:          lock,
		infoStream:      infoStream,
		readerPool:      readerPool,
		mergingSegments: make(map[*SegmentCommitInfo]bool),
		pendingMerges:   list.New(),
		runningMerges:   make(map[*OneMerge]bool),
		mergeSignal:     sync.NewCond(lock),
	}
}

// L2272
/*
Aborts runing merges. Be careful when using this method: when you
abort a long-running merge, you lose a lot of work that must later be
redone.
*/
func (mc *MergeControl) abortAllMerges() {
	mc.Lock() // synchronized
	defer mc.Unlock()

	mc.stopMerges = true

	// Abort all pending & running merges:
	for e := mc.pendingMerges.Front(); e != nil; e = e.Next() {
		merge := e.Value.(*OneMerge)
		if mc.infoStream.IsEnabled("IW") {
			mc.infoStream.Message("IW", "now abort pending merge %v",
				mc.readerPool.segmentsToString(merge.segments))
		}
		merge.abort()
		mc.mergeFinish(merge)
	}
	mc.pendingMerges.Init()

	for merge, _ := range mc.runningMerges {
		if mc.infoStream.IsEnabled("IW") {
			mc.infoStream.Message("IW", "now abort running merge %v",
				mc.readerPool.segmentsToString(merge.segments))
		}
		merge.abort()
	}

	// These merges periodically check whether they have
	// been aborted, and stop if so.  We wait here to make
	// sure they all stop.  It should not take very long
	// because the merge threads periodically check if
	// they are aborted.
	for len(mc.runningMerges) > 0 {
		if mc.infoStream.IsEnabled("IW") {
			mc.infoStream.Message("IW", "now wait for %v running merge(s) to abort",
				len(mc.runningMerges))
		}
		mc.mergeSignal.Wait()
	}

	mc.stopMerges = false

	assert(len(mc.mergingSegments) == 0)

	if mc.infoStream.IsEnabled("IW") {
		mc.infoStream.Message("IW", "all running merges have aborted")
	}
}

/*
Wait for any currently outstanding merges to finish.

It is guaranteed that any merges started prior to calling this method
will have completed once this method completes.
*/
func (mc *MergeControl) waitForMerges() {
	mc.Lock() // synchronized
	defer mc.Unlock()
	// ensureOpen(false)

	if mc.infoStream.IsEnabled("IW") {
		mc.infoStream.Message("IW", "waitForMerges")
	}

	for mc.pendingMerges.Len() > 0 || len(mc.runningMerges) > 0 {
		mc.mergeSignal.Wait()
	}

	assert(len(mc.mergingSegments) == 0)

	if mc.infoStream.IsEnabled("IW") {
		mc.infoStream.Message("IW", "waitForMerges done")
	}
}

// L3696
/*
Does finishing for a merge, which is fast but holds the synchronized
lock on MergeControl instance.

Note: it must be externally synchronized or used internally.
*/
func (mc *MergeControl) mergeFinish(merge *OneMerge) {
	// forceMerge, addIndexes or abortAllmerges may be waiting on
	// merges to finish
	// notifyAll()

	// It's possible we are called twice, eg if there was an error
	// inside mergeInit()
	if merge.registerDone {
		for _, info := range merge.segments {
			delete(mc.mergingSegments, info)
		}
		merge.registerDone = false
	}

	delete(mc.runningMerges, merge)
	if len(mc.runningMerges) == 0 {
		mc.mergeSignal.Signal()
	}
}
