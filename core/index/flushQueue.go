package index

import (
	"container/list"
	"fmt"
	"sync"
	"sync/atomic"
)

// index/DocumentsWriterFlushQueue.java

type DocumentsWriterFlushQueue struct {
	sync.Locker
	queue *list.List
	// we track tickets separately since count must be present even
	// before the ticket is constructed, ie. queue.size would not
	// reflect it.
	ticketCount int32 // aomitc
	purgeLock   sync.Locker
}

func newDocumentsWriterFlushQueue() *DocumentsWriterFlushQueue {
	return &DocumentsWriterFlushQueue{
		Locker:    &sync.Mutex{},
		queue:     list.New(),
		purgeLock: &sync.Mutex{},
	}
}

func (fq *DocumentsWriterFlushQueue) addDeletes(deleteQueue *DocumentsWriterDeleteQueue) error {
	panic("not implemented yet")
}

func (fq *DocumentsWriterFlushQueue) hasTickets() bool {
	n := atomic.LoadInt32(&fq.ticketCount)
	assertn(n >= 0, "ticketCount should be >= 0 but was: ", n)
	return n != 0
}

func assertn(ok bool, msg string, args ...interface{}) {
	if !ok {
		panic(fmt.Sprintf(msg, args...))
	}
}

func (fq *DocumentsWriterFlushQueue) _purge(writer *IndexWriter) (numPurged int, err error) {
	for {
		if head, canPublish := func() (FlushTicket, bool) {
			fq.Lock()
			defer fq.Unlock()
			if fq.queue.Len() > 0 {
				head := fq.queue.Front().Value.(FlushTicket)
				return head, head.canPublish()
			}
			return nil, false
		}(); canPublish {
			numPurged++
			if err = func() error {
				defer func() {
					fq.Lock()
					defer fq.Unlock()
					// remove the published ticket from the queue
					e := fq.queue.Front()
					fq.queue.Remove(e)
					atomic.AddInt32(&fq.ticketCount, -1)
					assert(e.Value.(FlushTicket) == head)
				}()
				// if we block on publish -> lock IW -> lock BufferedDeletes,
				// we don't block concurrent segment flushes just because
				// they want to append to the queue. The down-side is that we
				// need to force a purge on fullFlush since there could be a
				// ticket still in the queue.
				return head.publish(writer)
			}(); err != nil {
				return
			}
		} else {
			break
		}
	}
	return
}

func (fq *DocumentsWriterFlushQueue) forcePurge(writer *IndexWriter) (int, error) {
	fq.purgeLock.Lock()
	defer fq.purgeLock.Unlock()
	return fq._purge(writer)
}

type FlushTicket interface {
	canPublish() bool
	publish(writer *IndexWriter) error
}

type FlushTicketImpl struct {
	frozenDeletes *FrozenBufferedDeletes
	published     bool
}
