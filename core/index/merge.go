package index

// index/MergeScheduler.java

import (
	"io"
	"sync"
)

// index/MergePolicy.java

/*
Expert: a MergePolicy determines the sequence of primitive merge
operations.

Whenever the segments in an index have been altered by IndexWriter,
either the addition of a newly flushed segment, addition of many
segments from addIndexes* calls, or a previous merge that may now
seed to cascade, IndexWriter invokes findMerges() to give the
MergePolicy a chance to pick merges that are now required. This
method returns a MergeSpecification instance describing the set of
merges that should be done, or nil if no merges are necessary. When
IndexWriter.forceMerge() is called, it calls findForcedMerges() and
the MergePolicy should then return the necessary merges.

Note that the policy can return more than one merge at a time. In
this case, if the writer is using SerialMergeScheduler, the merges
will be run sequentially but if it is using ConcurrentMergeScheduler
they will be run concurrently.

The default MergePolicy is TieredMergePolicy.
*/
type MergePolicy struct {
}

/*
OneMerge provides the information necessary to perform an individual
primitive merge operation, resulting in a single new segment. The
merge spec includes the subset of segments to be merged as well as
whether the new segment should use the compound file format.
*/
type OneMerge struct {
}

/*
Expert: IndexWriter uses an instance implementing this interface to
execute the merges selected by a MergePolicy. The default
MergeScheduler is ConcurrentMergeScheduler.

Implementers of sub-classes shold make sure that Clone() returns an
independent instance able to work with any IndexWriter instance.
*/
type MergeScheduler interface {
	io.Closer
	Merge(writer *IndexWriter) error
	Clone() MergeScheduler
}

// Passed to MergePolicy.FindMerges(MergeTrigger, SegmentInfos) to
// indicate the event that triggered the merge
type MergeTrigger int

const (
	// Merge was triggered by a segment flush.
	MERGE_TRIGGER_SEGMENT_FLUSH = 1
	// Merge was triggered by a full flush. Full flushes can be caused
	// by a commit, NRT reader reopen or close call on the index writer
	MERGE_TRIGGER_FULL_FLUSH = 2
)

// index/MergeState.java

// Recording units of work when merging segments.
type CheckAbort struct {
}

// index/SerialMergeScheduler.java

// A MergeScheduler that simply does each merge sequentially, using
// the current thread.
type SerialMergeScheduler struct {
	sync.Locker
}

func NewSerialMergeScheduler() *SerialMergeScheduler {
	return &SerialMergeScheduler{&sync.Mutex{}}
}

func (ms *SerialMergeScheduler) Merge(writer *IndexWriter) (err error) {
	ms.Lock() // synchronized
	defer ms.Unlock()

	for merge := writer.nextMerge(); merge != nil && err == nil; merge = writer.nextMerge() {
		err = writer.merge(merge)
	}
	return
}

func (ms *SerialMergeScheduler) Clone() MergeScheduler {
	return NewSerialMergeScheduler()
}

func (ms *SerialMergeScheduler) Close() error {
	return nil
}

// index/ConcurrentMergeScheduler.java

/*
A MergeScheduler that runs each merge using a separate goroutine.

Specify the max number of goroutines that may run at once, and the
maximum number of simultaneous merges with SetMaxMergesAndRoutines().

If the number of merges exceeds the max number of threads then the
largest merges are paused until one of the smaller merges completes.

If more than MaxMergeCount() merges are requested then this class
will forcefully throttle the incoming goroutines by pausing until one
or more merges complete.
*/
type ConcurrentMergeScheduler struct {
	sync.Locker
}

func NewConcurrentMergeScheduler() *ConcurrentMergeScheduler {
	return &ConcurrentMergeScheduler{&sync.Mutex{}}
}

// Sets the maximum number of merge goroutines and simultaneous
// merges allowed.
func (cms *ConcurrentMergeScheduler) SetMaxMergesAndRoutines(maxMergeCount, maxRoutineCount int) {
	panic("not implemented yet")
}

func (cms *ConcurrentMergeScheduler) Close() error {
	cms.sync()
	return nil
}

// Wait for any running merge threads to finish. This call is not
// Interruptible as used by Close()
func (cms *ConcurrentMergeScheduler) sync() {
	panic("not implemented yet")
}

func (cms *ConcurrentMergeScheduler) Merge(writer *IndexWriter) error {
	cms.Lock() // synchronized
	defer cms.Unlock()
	panic("not implemented yet")
}

func (cms *ConcurrentMergeScheduler) String() string {
	panic("not implemented yet")
}

func (cms *ConcurrentMergeScheduler) Clone() MergeScheduler {
	panic("not implemented yet")
}
