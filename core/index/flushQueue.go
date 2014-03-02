package index

import (
	"fmt"
	"sync/atomic"
)

// index/DocumentsWriterFlushQueue.java

type DocumentsWriterFlushQueue struct {
	ticketCount int32
}

func newDocumentsWriterFlushQueue() *DocumentsWriterFlushQueue {
	return &DocumentsWriterFlushQueue{}
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

func (fq *DocumentsWriterFlushQueue) forcePurge(writer *IndexWriter) (int, error) {
	panic("not implemented yet")
}
