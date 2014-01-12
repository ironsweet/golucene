package index

import (
	"github.com/balzaczyy/golucene/core/util"
	"io"
	"math"
	"sync"
)

// index/MergeScheduler.java

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

// index/MergePolicy.java

// Default max segment size in order to use compound file system.
// Set to maxInt64.
const DEFAULT_MAX_CFS_SEGMENT_SIZE = math.MaxInt64

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
type MergePolicy interface {
}

type MergePolicyImpl struct {
	// IndexWriter that contains this instance.
	writer *util.SetOnce
	// If the size of te merge segment exceeds this ratio of the total
	// index size then it will remain in non-compound format.
	noCFSRatio float64
	// If the size of the merged segment exceeds this value then it
	// will not use compound file format.
	maxCFSSegmentSize float64
}

/*
Creates a new merge policy instance. Note that if you intend to use
it without passing it to IndexWriter, you should call SetIndexWriter()
*/
func NewDefaultMergePolicyImpl() *MergePolicyImpl {
	return newMergePolicyImpl(DEFAULT_NO_CFS_RATIO, DEFAULT_MAX_CFS_SEGMENT_SIZE)
}

/*
Create a new merge policy instance with default settings for noCFSRatio
and maxCFSSegmentSize. This ctor should be used by subclasses using
different defaults than the MergePolicy.
*/
func newMergePolicyImpl(defaultNoCFSRatio, defaultMaxCFSSegmentSize float64) *MergePolicyImpl {
	return &MergePolicyImpl{
		util.NewSetOnce(),
		defaultNoCFSRatio,
		defaultMaxCFSSegmentSize,
	}
}

/*
OneMerge provides the information necessary to perform an individual
primitive merge operation, resulting in a single new segment. The
merge spec includes the subset of segments to be merged as well as
whether the new segment should use the compound file format.
*/
type OneMerge struct {
}

// index/TieredMergePolicy.java

// Default noCFSRatio. If a merge's size is >= 10% of the index, then
// we disable compound file for it.
const DEFAULT_NO_CFS_RATIO = 0.1

/*
Merges segments of approximately equal size, subject to an allowed
number of segments per tier. This is similar to LogByteSizeMergePolicy,
except this merge policy is able to merge non-adjacent segment, and
separates how many segments are merged at once (SetMaxMergeAtOnce())
from how many segments are allowed per tier (SetSegmentsPerTier()).
This merge policy also does not over-merge (i.e. cascade merges).

For normal merging, this policy first computes a "budget" of how many
segments are allowed to be in the index. If the index is over-budget,
then the policy sorts segments by decreasing size (pro-rating by
percent deletes), and then finds the least-cost merge. Merge cost is
measured by a combination of the "skew" of the merge (size of largest
segments divided by smallest segment), total merge size and percent
deletes reclaimed, so tha tmerges with lower skew, smaller size and
those reclaiming more deletes, are flavored.

If a merge wil produce a segment that's larger than SetMaxMergedSegmentMB(),
then the policy will merge fewer segments (down to 1 at once, if that
one has deletions) to keep the segment size under budget.

NOTE: this policy freely merges non-adjacent segments; if this is a
problem, use LogMergePolicy.

NOTE: This policy always merges by byte size of the segments, always
pro-rates by percent deletes, and does not apply any maximum segment
size duirng forceMerge (unlike LogByteSizeMergePolicy).
*/
type TieredMergePolicy struct {
	*MergePolicyImpl

	maxMergeAtOnce         int
	maxMergedSegmentBytes  int64
	maxMergeAtOnceExplicit int

	floorSegmentBytes           int64
	segsPerTier                 float64
	forceMergeDeletesPctAllowed float64
	reclaimDeletesWeight        float64
}

func newTieredMergePolicy() *TieredMergePolicy {
	return &TieredMergePolicy{
		MergePolicyImpl:             newMergePolicyImpl(DEFAULT_NO_CFS_RATIO, DEFAULT_MAX_CFS_SEGMENT_SIZE),
		maxMergeAtOnce:              10,
		maxMergedSegmentBytes:       5 * 1024 * 1024 * 1024,
		maxMergeAtOnceExplicit:      30,
		floorSegmentBytes:           2 * 1024 * 1024,
		segsPerTier:                 10,
		forceMergeDeletesPctAllowed: 10,
		reclaimDeletesWeight:        2,
	}
}

// index/LogMergePolicy.java

/*
This class implements a MergePolicy that tries to merge segments into
levels of exponentially increasing size, where each level has fewer
segments than the value of the merge factor. Whenver extra segments
(beyond the merge factor upper bound) are encountered, all segments
within the level are merged. You can get or set the merge factor
using MergeFactor() and SetMergeFactor() repectively.

This class is abstract and required a subclass to define the Size()
method  which specifies how a segment's size is determined.
LogDocMergePolicy is one subclass that measures size by document
count in the segment. LogByteSizeMergePolicy is another subclass that
measures size as the total byte size of the file(s) for the segment.
*/
type LogMergePolicy struct {
}
