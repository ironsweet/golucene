package index

import (
	"bytes"
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
	"sync/atomic"
)

// index/BufferedDeletes.java

const BYTES_PER_DEL_DOCID = util.NUM_BYTES_INT

const VERBOSE = true

/*
Holds buffered deletes, by docID, term or query for a single segment.
This is used to hold buffered pending deletes against the
to-be-flushed segment. Once the deletes are pushed (on flush in DW),
these deletes are converted to a FronzenDeletes instance.

NOTE: instances of this class are accessed either via a private
instance on DocumentsWriterPerThread, or via sync'd code by
DocumentsWriterDeleteQueue
*/
type BufferedDeletes struct {
	numTermDeletes int32 // atomic
	terms          map[*Term]int
	queries        map[interface{}]int
	docIDs         []int

	bytesUsed int64 // atomic

	gen int64
}

func newBufferedDeletes() *BufferedDeletes {
	return &BufferedDeletes{
		terms:   make(map[*Term]int),
		queries: make(map[interface{}]int),
	}
}

func (bd *BufferedDeletes) String() string {
	if VERBOSE {
		return fmt.Sprintf("gen=%v, numTerms=%v, terms=%v, queries=%v, docIDs=%v, bytesUsed=%v",
			bd.gen, atomic.LoadInt32(&bd.numTermDeletes), bd.terms, bd.queries, bd.docIDs, bd.bytesUsed)
	} else {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "gen=%v", bd.gen)
		if n := atomic.LoadInt32(&bd.numTermDeletes); n != 0 {
			fmt.Fprintf(&buf, " %v deleted terms (unique count=%v)", n, len(bd.terms))
		}
		if len(bd.queries) > 0 {
			fmt.Fprintf(&buf, " %v deleted queries", len(bd.queries))
		}
		if len(bd.docIDs) > 0 {
			fmt.Fprintf(&buf, " %v deleted docIDs", len(bd.docIDs))
		}
		if n := atomic.LoadInt64(&bd.bytesUsed); n != 0 {
			fmt.Fprintf(&buf, " bytesUsed=%v", n)
		}
		return buf.String()
	}
}

func (bd *BufferedDeletes) addDocID(docID int) {
	bd.docIDs = append(bd.docIDs, docID)
	atomic.AddInt64(&bd.bytesUsed, BYTES_PER_DEL_DOCID)
}

func (bd *BufferedDeletes) clear() {
	bd.terms = make(map[*Term]int)
	bd.queries = make(map[interface{}]int)
	bd.docIDs = nil
	atomic.StoreInt32(&bd.numTermDeletes, 0)
	atomic.StoreInt64(&bd.bytesUsed, 0)
}

func (bd *BufferedDeletes) any() bool {
	return len(bd.terms) > 0 || len(bd.docIDs) > 0 || len(bd.queries) > 0
}

// index/FrozenBufferedDeletes.java

/*
Holds buffered deletes by term or query, once pushed. Pushed delets
are write-once, so we shift to more memory efficient data structure
to hold them. We don't hold docIDs because these are applied on flush.
*/
type FrozenBufferedDeletes struct {
	bytesUsed      int
	numTermDeletes int
	gen            int64 // -1, assigned by BufferedDeletesStream once pushed
	// true iff this frozen packet represents a segment private deletes
	// in that case it should only have queries
	isSegmentPrivate bool
}

func newFrozenBufferedDeletes() *FrozenBufferedDeletes {
	return &FrozenBufferedDeletes{gen: -1}
}

func (bd *FrozenBufferedDeletes) queries() []*QueryAndLimit {
	panic("not implemented yet")
}

func (bd *FrozenBufferedDeletes) String() string {
	panic("not implemented yet")
}
