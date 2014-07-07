package index

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// index/ConcurrentMergeScheduler.java

type MergeJob struct {
	start  time.Time
	writer *IndexWriter
	merge  *OneMerge
}

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

	// IndexWriter that owns this instance.
	writer *IndexWriter

	// How many merges have kicked off (this is used to name them).
	mergeThreadCount int32 // atomic

	suppressErrors bool

	chRequest            chan *MergeJob
	chSync               chan *sync.WaitGroup
	concurrentMergeCount int32 // atomic
	numMergeRoutines     int32 // atomic
}

func NewConcurrentMergeScheduler() *ConcurrentMergeScheduler {
	cms := &ConcurrentMergeScheduler{
		Locker:    &sync.Mutex{},
		chRequest: make(chan *MergeJob),
		chSync:    make(chan *sync.WaitGroup),
	}
	cms.SetMaxMergesAndRoutines(DEFAULT_MAX_MERGE_COUNT, DEFAULT_MAX_ROUTINE_COUNT)
	return cms
}

/*
Daemon worker that accepts and processes merge job.

GoLucene assumes each merge is pre-sorted according to its merge size
before acquired from IndexWriter. It makes use of pre-allocated
go routines, instead of MergeThread to do the real merge work,
witout explicit synchronizations and waitings.

Note, however, change of merge count won't pause/resume workers.
*/
func (cms *ConcurrentMergeScheduler) worker(id int) {
	atomic.AddInt32(&cms.numMergeRoutines, 1)
	fmt.Printf("CMS Worker %v is started.\n", id)
	var isRunning = true
	var wg *sync.WaitGroup
	for isRunning && id < cms.maxRoutineCount {
		select {
		case job := <-cms.chRequest:
			cms.process(job)
		case wg = <-cms.chSync:
			isRunning = false
			defer wg.Done()
		}
	}
	fmt.Printf("CMS Worker %v is stopped.\n", id)
	atomic.AddInt32(&cms.numMergeRoutines, -1)
}

func (cms *ConcurrentMergeScheduler) process(job *MergeJob) {
	atomic.AddInt32(&cms.concurrentMergeCount, 1)
	defer func() {
		atomic.AddInt32(&cms.concurrentMergeCount, -1)
	}()

	if cms.verbose() {
		elapsed := time.Now().Sub(job.start)
		cms.message("  stalled for %v", elapsed)
		cms.message("  consider merge %v", job.writer.readerPool.segmentsToString(job.merge.segments))
		// OK to spawn a new merge routine to handle this merge
		cms.message("    launch new thread [%v]", atomic.AddInt32(&cms.mergeThreadCount, 1))
		cms.message("  merge thread: start")
	}

	err := job.writer.merge(job.merge)
	if err != nil {
		// Ignore the error if it was due to abort:
		if _, ok := err.(MergeAbortedError); !ok && !cms.suppressErrors {
			// suppressErrors is normally only set during testing.
			cms.handleMergeError(err)
		}
	}
}

// Sets the maximum number of merge goroutines and simultaneous
// merges allowed.
func (cms *ConcurrentMergeScheduler) SetMaxMergesAndRoutines(maxMergeCount, maxRoutineCount int) {
	assert2(maxRoutineCount >= 1, "maxRoutineCount should be at least 1")
	assert2(maxMergeCount >= 1, "maxMergeCount should be at least 1")
	assert2(maxRoutineCount <= maxMergeCount, fmt.Sprintf(
		"maxRoutineCount should be <= maxMergeCount (= %v)", maxMergeCount))

	oldCount := cms.maxRoutineCount
	cms.maxRoutineCount = maxRoutineCount
	cms.maxMergeCount = maxMergeCount

	cms.Lock()
	defer cms.Unlock()
	for i := oldCount; i < maxRoutineCount; i++ {
		go cms.worker(i)
	}
}

/*
Returns true if verbosing is enabled. This method is usually used in
conjunction with message(), like that:

	if cms.verbose() {
		cms.message("your message")
	}
*/
func (cms *ConcurrentMergeScheduler) verbose() bool {
	return cms.writer != nil && cms.writer.infoStream.IsEnabled("CMS")
}

/*
Outputs the given message - this method assumes verbose() was called
and returned true.
*/
func (cms *ConcurrentMergeScheduler) message(format string, args ...interface{}) {
	cms.writer.infoStream.Message("CMS", format, args...)
}

func (cms *ConcurrentMergeScheduler) Close() error {
	cms.sync()
	return nil
}

/*
Wait for any running merge threads to finish. This call is not
Interruptible as used by Close()
*/
func (cms *ConcurrentMergeScheduler) sync() {
	cms.Lock()
	defer cms.Unlock()

	wg := new(sync.WaitGroup)
	// no need to synchronize on numMergeRoutines
	for i, limit := 0, int(cms.numMergeRoutines); i < limit; i++ {
		wg.Add(1)
		cms.chSync <- wg
	}
	wg.Wait()
}

func (cms *ConcurrentMergeScheduler) Merge(writer *IndexWriter,
	trigger MergeTrigger, newMergesFound bool) error {
	cms.Lock() // synchronized
	defer cms.Unlock()

	// assert !Thread.holdsLock(writer)
	cms.writer = writer

	// First, quickly run through the newly proposed merges
	// and add any orthogonal merges (ie a merge not
	// involving segments already pending to be merged) to
	// the queue.  If we are way behind on merging, many of
	// these newly proposed merges will likely already be
	// registered.
	if cms.verbose() {
		cms.message("now merge")
		cms.message("  index: %v", writer.segString())
	}

	// Iterate, pulling from the IndexWriter's queue of
	// pending merges, until it's empty:
	for merge := writer.nextMerge(); merge != nil; merge = writer.nextMerge() {
		if atomic.LoadInt32(&cms.concurrentMergeCount) >= int32(cms.maxMergeCount) {
			// This means merging has fallen too far behind: we
			// have already created maxMergeCount threads, and
			// now there's at least one more merge pending.
			// Note that only maxThreadCount of
			// those created merge threads will actually be
			// running; the rest will be paused (see
			// updateMergeThreads).  We stall this producer
			// thread to prevent creation of new segments,
			// until merging has caught up:
			if cms.verbose() {
				cms.message("    too many merges; stalling...")
			}
		}
		cms.chRequest <- &MergeJob{time.Now(), writer, merge}
	}
	if cms.verbose() {
		cms.message("  no more merges pending; now return")
	}
	return nil
}

/*
Called when an error is hit in a background merge thread
*/
func (cms *ConcurrentMergeScheduler) handleMergeError(err error) {
	// When an exception is hit during merge, IndexWriter
	// removes any partial files and then allows another
	// merge to run.  If whatever caused the error is not
	// transient then the exception will keep happening,
	// so, we sleep here to avoid saturating CPU in such
	// cases:
	time.Sleep(250 * time.Millisecond)
	// Lucene Java throw Unchecked exception in a separate thread.
	// GoLucene just dump the error in console.
	log.Printf("Merge error: %v", err)
}

func (cms *ConcurrentMergeScheduler) String() string {
	panic("not implemented yet")
}
