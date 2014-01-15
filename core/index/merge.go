package index

import (
	"fmt"
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
Default maxThreadCount. We default to 1: tests on spinning-magnet
drives showed slower indexing performance if more than one merge
routine runs at once (though on an SSD it was faster)
*/
const DEFAULT_MAX_ROUTINE_COUNT = 1

// Default maxMergeCount.
const DEFAULT_MAX_MERGE_COUNT = 2

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

	// Max number of merge routines allowed to be running at once. When
	// there are more merges then this, we forcefully pause the larger
	// ones, letting the smaller ones run, up until maxMergeCount
	// merges at which point we forcefully pause incoming routines
	// (that presumably are the ones causing so much merging).
	maxRoutineCount int

	// Max number of merges we accept before forcefully throttling the
	// incoming routines
	maxMergeCount int
}

func NewConcurrentMergeScheduler() *ConcurrentMergeScheduler {
	return &ConcurrentMergeScheduler{
		Locker:          &sync.Mutex{},
		maxRoutineCount: DEFAULT_MAX_ROUTINE_COUNT,
		maxMergeCount:   DEFAULT_MAX_MERGE_COUNT,
	}
}

// Sets the maximum number of merge goroutines and simultaneous
// merges allowed.
func (cms *ConcurrentMergeScheduler) SetMaxMergesAndRoutines(maxMergeCount, maxRoutineCount int) {
	assert2(maxRoutineCount >= 1, "maxRoutineCount should be at least 1")
	assert2(maxMergeCount >= 1, "maxMergeCount should be at least 1")
	assert2(maxRoutineCount <= maxMergeCount, fmt.Sprintf(
		"maxRoutineCount should be <= maxMergeCount (= %v)", maxMergeCount))
	cms.maxRoutineCount = maxRoutineCount
	cms.maxMergeCount = maxMergeCount
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
	SetIndexWriter(writer *IndexWriter)
	SetNoCFSRatio(noCFSRatio float64)
	SetMaxCFSSegmentSizeMB(v float64)
}

type MergePolicyImpl struct {
	// Return the byte size of the provided SegmentInfoPerCommit,
	// pro-rated by percentage of non-deleted documents if
	// SetCalibrateSizeByDeletes() is set.
	size func(info *SegmentInfoPerCommit) (n int64, err error)
	// IndexWriter that contains this instance.
	writer *util.SetOnce
	// If the size of te merge segment exceeds this ratio of the total
	// index size then it will remain in non-compound format.
	noCFSRatio float64
	// If the size of the merged segment exceeds this value then it
	// will not use compound file format.
	maxCFSSegmentSize float64
}

type MergeSpecifier interface {
	// Determine what set of merge operations are now necessary on the
	// index. IndexWriter calls this whenever there is a change to the
	// segments. This call is always synchronized on the IndexWriter
	// instance so only one thread at a time will call this method.
	// FindMerges(mergeTrigger MergeTrigger, segmentInfos *SegmentInfos) (spec MergeSpecification, err error)
	// Determine what set of merge operations is necessary in order to
	// merge to <= the specified segment count. IndexWriter calls this
	// when its forceMerge() method is called. This call is always
	// synchronized on the IndexWriter instance so only one thread at a
	// time will call this method.
	// FindForcedMerges(segmentInfos *SegmentInfos, maxSegmentCount int,
	// segmentsToMerge map[SegmentInfoPerCommit]bool) (spec MergeSpecification, err error)
	// Determine what set of merge operations is necessary in order to
	// expunge all deletes from the index.
	// FindForcedDeletesMerges(segmentinfos *SegmentInfos) (spec MergeSpecification, err error)
}

func (mp *MergePolicyImpl) Clone() MergePolicy {
	clone := *mp
	clone.writer = util.NewSetOnce()
	return &clone
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
	ans := &MergePolicyImpl{
		writer:            util.NewSetOnce(),
		noCFSRatio:        defaultNoCFSRatio,
		maxCFSSegmentSize: defaultMaxCFSSegmentSize,
	}
	ans.size = func(info *SegmentInfoPerCommit) (n int64, err error) {
		byteSize, err := info.SizeInBytes()
		if err != nil {
			return 0, err
		}
		docCount := info.info.docCount
		if docCount <= 0 {
			return byteSize, nil
		}

		delCount := ans.writer.Get().(*IndexWriter).numDeletedDocs(info)
		delRatio := float32(delCount) / float32(docCount)
		assert(delRatio <= 1)
		return int64(float32(byteSize) * (1 - delRatio)), nil
	}
	return ans
}

/*
Sets the IndexWriter to use by this merge policy. This method is
allowed to be called only once, and is usually set by IndexWriter. If
it is called more thanonce, panic is thrown.
*/
func (mp *MergePolicyImpl) SetIndexWriter(writer *IndexWriter) {
	mp.writer.Set(writer)
}

/*
Returns true if this single info is already fully merged (has no
pending deletes, is in the same dir as the writer, and matches the
current compound file setting)
*/
func (mp *MergePolicyImpl) isMerged(info *SegmentInfoPerCommit) bool {
	w := mp.writer.Get().(*IndexWriter)
	assert(w != nil)
	hasDeletions := w.numDeletedDocs(info) > 0
	return !hasDeletions &&
		!info.info.hasSeparateNorms() &&
		info.info.dir == w.directory &&
		(mp.noCFSRatio > 0 && mp.noCFSRatio < 1 || mp.maxCFSSegmentSize < math.MaxInt64)
}

/*
If a merged segment will be more than this percentage of the total
size of the index, leave the segment as non-compound file even if
compound file is enabled. Set to 1.0 to always use CFS regardless or
merge size.
*/
func (mp *MergePolicyImpl) SetNoCFSRatio(noCFSRatio float64) {
	assert2(noCFSRatio >= 0 && noCFSRatio <= 1, fmt.Sprintf(
		"noCFSRatio must be 0.0 to 1.0 inclusive; got %v", noCFSRatio))
	mp.noCFSRatio = noCFSRatio
}

/*
If a merged segment will be more than this value, leave the segment
as non-compound file even if compound file is enabled. Set this to
math.Inf(1) (default) and noCFSRatio to 1.0 to always use CFS
regardless of merge size.
*/
func (mp *MergePolicyImpl) SetMaxCFSSegmentSizeMB(v float64) {
	assert2(v >= 0, fmt.Sprintf("maxCFSSegmentSizeMB must be >=0 (got %v)", v))
	v *= 1024 * 1024
	if v > float64(math.MaxInt64) {
		mp.maxCFSSegmentSize = math.MaxInt64
	} else {
		mp.maxCFSSegmentSize = v
	}
}

// Passed to MergePolicy.FindMerges(MergeTrigger, SegmentInfos) to
// indicate the event that triggered the merge
type MergeTrigger int

const (
	// Merge was triggered by a segment flush.
	MERGE_TRIGGER_SEGMENT_FLUSH = MergeTrigger(1)
	// Merge was triggered by a full flush. Full flushes can be caused
	// by a commit, NRT reader reopen or close call on the index writer
	MERGE_TRIGGER_FULL_FLUSH = MergeTrigger(2)
)

/*
OneMerge provides the information necessary to perform an individual
primitive merge operation, resulting in a single new segment. The
merge spec includes the subset of segments to be merged as well as
whether the new segment should use the compound file format.
*/
type OneMerge struct {
}

/*
A MergeSpecification instance provides the information necessary to
perform multiple merges. It simply contains a list of OneMerge
instances.
*/
type MergeSpecification []OneMerge

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

func NewTieredMergePolicy() *TieredMergePolicy {
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

/*
Maximum number of segments to be merged at a time during "normal"
merging. For explicit merging (e.g., forceMerge or forceMergeDeletes
was called), see SetMaxMergeAtonceExplicit(). Default is 10.
*/
func (tmp *TieredMergePolicy) SetMaxMergeAtOnce(v int) *TieredMergePolicy {
	assert2(v >= 2, fmt.Sprintf("maxMergeAtonce must be > 1 (got %v)", v))
	tmp.maxMergeAtOnce = v
	return tmp
}

/*
Maximum number of segments to be merged at a time, during forceMerge
or forceMergeDeletes. Default is 30.
*/
func (tmp *TieredMergePolicy) SetMaxMergeAtOnceExplicit(v int) *TieredMergePolicy {
	assert2(v >= 2, fmt.Sprintf("maxMergeAtonceExplicit must be > 1 (got %v)", v))
	tmp.maxMergeAtOnceExplicit = v
	return tmp
}

/*
Maximum sized segment to produce during normal merging. This setting
is approximate: the estimate of the merged segment size is made by
summing sizes of to-be-merged segments(compensating for percent
deleted docs). Default is 5 GB.
*/
func (tmp *TieredMergePolicy) SetMaxMergedSegmentMB(v float64) *TieredMergePolicy {
	assert2(v >= 0, fmt.Sprintf("maxMergedSegmentMB must be >= 0 (got %v)", v))
	v *= 1024 * 1024
	tmp.maxMergedSegmentBytes = math.MaxInt64
	if v < math.MaxInt64 {
		tmp.maxMergedSegmentBytes = int64(v)
	}
	return tmp
}

/*
Controls how aggressively merges that reclaim more deletions are
favored. Higher values favor selecting merges that reclaim deletions.
A value of 0 means deletions don't impact merge selection.
*/
func (tmp *TieredMergePolicy) SetReclaimDeletesWeight(v float64) *TieredMergePolicy {
	assert2(v >= 0, fmt.Sprintf("reclaimDeletesWeight must be >= 0 (got %v)", v))
	tmp.reclaimDeletesWeight = v
	return tmp
}

/*
Segments smaller than this are "rounded up" to this size, ie treated
as equal (floor) size for merge selection. This is to prevent
frequent flushing of tiny segments from allowing a long tail in the
index. Default is 2 MB.
*/
func (tmp *TieredMergePolicy) SetFloorSegmentMB(v float64) *TieredMergePolicy {
	assert2(v > 0, fmt.Sprintf("floorSegmentMB must be > 0 (got %v)", v))
	v *= 1024 * 1024
	tmp.floorSegmentBytes = math.MaxInt64
	if v < math.MaxInt64 {
		tmp.floorSegmentBytes = int64(v)
	}
	return tmp
}

/*
When forceMergeDeletes is called, we only merge away a segment if its
delete percentage is over this threshold. Default is 10%.
*/
func (tmp *TieredMergePolicy) SetForceMergeDeletesPctAllowed(v float64) *TieredMergePolicy {
	assert2(v >= 0 && v <= 100, fmt.Sprintf("forceMergeDeletesPctAllowed must be between 0 and 100 inclusive (got %v)", v))
	tmp.forceMergeDeletesPctAllowed = v
	return tmp
}

/*
Sets the allowed number of segments per tier. Smaller values mean
more merging but fewer segments.

NOTE: this value should be >= the SetMaxMergeAtOnce otherwise you'll
force too much merging to occur.
*/
func (tmp *TieredMergePolicy) SetSegmentsPerTier(v float64) *TieredMergePolicy {
	assert2(v >= 2, fmt.Sprintf("segmentsPerTier must be >= 2 (got %v)", v))
	tmp.segsPerTier = v
	return tmp
}

// index/LogMergePolicy.java

/*
Defines the allowed range of log(size) for each level. A level is
computed by taking the max segment log size, minus LEVEL_LOG_SPAN,
and finding all segments falling within that range.
*/
const LEVEL_LOG_SPAN = 0.75

// Default merge factor, which is how many segments are merged at a time
const DEFAULT_MERGE_FACTOR = 10

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
	*MergePolicyImpl

	// How many segments to merge at a time.
	mergeFactor int
	// Any segments whose size is smaller than this value will be
	// rounded up to this value. This ensures that tiny segments are
	// aggressively merged.
	minMergeSize int64
	// If the size of a segment exceeds this value then it will never
	// be merged.
	maxMergeSize int64
	// Although the core MPs set it explicitly, we must default in case
	// someone out there wrote this own LMP ...
	// If the size of a segment exceeds this value then it will never
	// be merged during ForceMerge()
	maxMergeSizeForForcedMerge int64
	// If true, we pro-rate a segment's size by the percentage of
	// non-deleted documents.
	calibrateSizeByDeletes bool
}

func newLogMergePolicy() *LogMergePolicy {
	return &LogMergePolicy{
		MergePolicyImpl:            newMergePolicyImpl(DEFAULT_NO_CFS_RATIO, DEFAULT_MAX_CFS_SEGMENT_SIZE),
		mergeFactor:                DEFAULT_MERGE_FACTOR,
		maxMergeSizeForForcedMerge: math.MaxInt64,
		calibrateSizeByDeletes:     true,
	}
}

// Returns true if LMP is enabled in IndexWriter's InfoStream.
func (mp *LogMergePolicy) verbose() bool {
	w := mp.writer.Get().(*IndexWriter)
	return w != nil && w.infoStream.IsEnabled("LMP")
}

// Print a debug message to IndexWriter's infoStream.
func (mp *LogMergePolicy) message(message string) {
	if mp.verbose() {
		mp.writer.Get().(*IndexWriter).infoStream.Message("LMP", message)
	}
}

/*
Determines how often segment indices are merged by AdDocument(). With
smaller values, less RAM is used while indexing, and searches are
faster, but indexing speed is slower. With larger values, more RAM is
used during indexing, and while searches is slower, indexing is
faster. Thus larger values (> 10) are best for batch index creation,
and smaller values (< 10) for indces that are interactively
maintained.
*/
func (mp *LogMergePolicy) SetMergeFactor(mergeFactor int) {
	assert2(mergeFactor >= 2, "mergeFactor cannot be less than 2")
	mp.mergeFactor = mergeFactor
}

// Sets whether the segment size should be calibrated by the number
// of delets when choosing segments to merge
func (mp *LogMergePolicy) SetCalbrateSizeByDeletes(calibrateSizeByDeletes bool) {
	mp.calibrateSizeByDeletes = calibrateSizeByDeletes
}

func (mp *LogMergePolicy) Close() error {
	return nil
}

/*
Return the byte size of the provided SegmentInfoPerCommit, pro-rated
by percentage of non-deleted documents if SetCalibratedSizeByDeletes()
is set.
*/
func (mp *LogMergePolicy) sizeBytes(info *SegmentInfoPerCommit) (n int64, err error) {
	if mp.calibrateSizeByDeletes {
		return mp.MergePolicyImpl.size(info)
	}
	return info.SizeInBytes()
}

/*
Returns true if the number of segments eligible for merging is less
than or equal to the specified maxNumSegments.
*/
func (mp *LogMergePolicy) isMergedBy(infos *SegmentInfos, maxNumSegments int, segmentsToMerge map[*SegmentInfoPerCommit]bool) bool {
	panic("not implemented yet")
}

type SegmentInfoAndLevel struct {
	info  *SegmentInfoPerCommit
	level float32
	index int
}

type SegmentInfoAndLevels []SegmentInfoAndLevel

func (ss SegmentInfoAndLevels) Len() int           { return len(ss) }
func (ss SegmentInfoAndLevels) Swap(i, j int)      { ss[i], ss[j] = ss[j], ss[i] }
func (ss SegmentInfoAndLevels) Less(i, j int) bool { return ss[i].level < ss[j].level }

/*
Checks if any merges are now necessary and returns a MergeSpecification
if so. A merge is necessary when there are more than SetMergeFactor()
segments at a given level. When multiple levels have too many
segments, this method will return multiple merges, allowing the
MergeScheduler to use concurrency.
*/
func (mp *LogMergePolicy) FindMerges(mergeTrigger MergeTrigger, infos *SegmentInfos) (spec MergeSpecification, err error) {
	numSegments := len(infos.Segments)
	mp.message(fmt.Sprintf("findMerges: %v segments", numSegments))

	// Compute levels, whic is just log (base mergeFactor) of the size
	// of each segment
	levels := make([]*SegmentInfoAndLevel, 0)
	norm := math.Log(float64(mp.mergeFactor))

	mergingSegments := mp.writer.Get().(*IndexWriter).mergingSegments

	for i, info := range infos.Segments {
		size, err := mp.size(info)
		if err != nil {
			return nil, err
		}

		// Floor tiny segments
		if size < 1 {
			size = 1
		}

		infoLevel := &SegmentInfoAndLevel{info, float32(math.Log(float64(size)) / norm), i}
		levels = append(levels, infoLevel)

		if mp.verbose() {
			segBytes, err := mp.sizeBytes(info)
			if err != nil {
				return nil, err
			}
			var extra string
			if _, ok := mergingSegments[info]; ok {
				extra = " [merging]"
			}
			if size >= mp.maxMergeSize {
				extra = fmt.Sprintf("%v [skip: too large]", extra)
			}
			mp.message(fmt.Sprintf("seg=%v level=%v size=%.3f MB%v",
				mp.writer.Get().(*IndexWriter).SegmentToString(info),
				infoLevel.level,
				segBytes/1024/1024,
				extra))
		}
	}

	var levelFloor float32 = 0
	if mp.minMergeSize > 0 {
		levelFloor = float32(math.Log(float64(mp.minMergeSize)) / float64(norm))
	}

	// Now, we quantize the log values into levfels. The first level is
	// any segment whose log size is within LEVEL_LOG_SPAN of the max
	// size, or, who has such as segment "to the right". Then, we find
	// the max of all other segments and use that to define the next
	// level segment, etc.

	numMergeableSegments := len(levels)

	for start := 0; start < numMergeableSegments; {
		// Find max level of all segments not already quantized.
		maxLevel := levels[start].level
		for i := 1 + start; i < numMergeableSegments; i++ {
			level := levels[i].level
			if level > maxLevel {
				maxLevel = level
			}
		}

		// Now search backwards for the rightmost segment that falls into
		// this level:
		var levelBottom float32
		if maxLevel <= levelFloor {
			// All remaining segments fall into the min level
			levelBottom = -1
		} else {
			levelBottom = float32(float64(maxLevel) - LEVEL_LOG_SPAN)

			// Force a boundary at the level floor
			if levelBottom < levelFloor && maxLevel >= levelFloor {
				levelBottom = levelFloor
			}
		}

		upto := numMergeableSegments - 1
		for upto >= start {
			if levels[upto].level >= levelBottom {
				break
			}
			upto--
		}
		mp.message(fmt.Sprintf("  level %v to %v: %v segments",
			levelBottom, maxLevel, 1+upto-start))

		// Finally, record all merges that are viable at this level:
		end := start + mp.mergeFactor
		for end <= 1+upto {
			panic("not implemented yet")
		}

		start = 1 + upto
	}

	return
}

func (mp LogMergePolicy) String() string {
	panic("not implemented yet")
}

// index/LogDocMergePolicy.java

/*
This is a LogMergePolicy that measures size of a segment as the
number of  documents (not taking deletions into account).
*/
type LogDocMergePolicy struct {
}

func NewLogDocMergePolicy() *LogMergePolicy {
	panic("not implemented yet")
}

// index/LogByteSizeMergePolicy.java

// Default minimum segment size.
var DEFAULT_MIN_MERGE_MB = 1.6

// Default maximum segment size. A segment of this size or larger
// will never be merged.
const DEFAULT_MAX_MERGE_MB = 2048

// Default maximum segment size. A segment of this size or larger
// will never be merged during forceMerge.
var DEFAULT_MAX_MERGE_MB_FOR_FORCED_MERGE = math.MaxInt64

// this is a LogMergePolicy that measures size of a segment as the
// total byte size of the segment's files.
type LogByteSizeMergePolicy struct {
	*LogMergePolicy
}

func NewLogByteSizeMergePolicy() *LogMergePolicy {
	ans := &LogByteSizeMergePolicy{
		LogMergePolicy: newLogMergePolicy(),
	}
	ans.minMergeSize = int64(DEFAULT_MIN_MERGE_MB * 1024 * 1024)
	ans.maxMergeSize = int64(DEFAULT_MAX_MERGE_MB * 1024 * 1024)
	ans.maxMergeSizeForForcedMerge = int64(DEFAULT_MAX_MERGE_MB_FOR_FORCED_MERGE * 1024 * 1024)
	ans.size = func(info *SegmentInfoPerCommit) (n int64, err error) {
		return ans.sizeBytes(info)
	}
	return ans.LogMergePolicy
}
