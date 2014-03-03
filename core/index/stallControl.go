package index

import (
	"sync"
)

/*
Controls the health status of a DocumentsWriter sessions. This class
used to block incoming index threads if flushing significantly slower
than indexing to ensure the DocumentsWriter's healthiness. If
flushing is significantly slower than indexing the net memory used
within an IndexWriter session can increase very quickly and easily
exceed the JVM's available memory.

To prevent OOM errors and ensure IndexWriter's stability this class
blocks incoming threads from indexing once 2 x number of available
ThreadState(s) in DocumentsWriterPerThreadPool is exceeded. Once
flushing catches up and number of flushing DWPT is equal of lower
than the number of active ThreadState(s) threads are released and can
continue indexing.
*/
type DocumentsWriterStallControl struct {
	sync.Locker
	*sync.Cond

	stalled    bool // volatile
	wasStalled bool // assert only
}

func newDocumentsWriterStallControl() *DocumentsWriterStallControl {
	lock := &sync.Mutex{}
	return &DocumentsWriterStallControl{
		Locker: lock,
		Cond:   sync.NewCond(lock),
	}
}

/*
Update the stalled flag status. This method will set the stalled flag
to true iff the number of flushing DWPT is greater than the number of
active DWPT. Otherwise it will reset the DWSC to healthy and release
all threads waiting on waitIfStalled()
*/
func (sc *DocumentsWriterStallControl) updateStalled(stalled bool) {
	sc.Lock()
	defer sc.Unlock()
	sc.stalled = stalled
	if stalled {
		sc.wasStalled = true
	}
	sc.Signal()
}
