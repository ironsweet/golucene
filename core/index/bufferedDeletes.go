package index

import (
	"bytes"
	"fmt"
	"sync/atomic"
)

// index/BufferedDeletes.java

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
			bd.gen, bd.numTermDeletes, bd.terms, bd.queries, bd.docIDs, bd.bytesUsed)
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
