package index

import (
	"bytes"
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
	"math"
	"sync/atomic"
)

// index/BufferedDeletes.java

/* Go slice consumes two int for an extra doc ID, assuming 50% pre-allocation. */
const BYTES_PER_DEL_DOCID = 2 * util.NUM_BYTES_INT

/* Go map (amd64) consumes about 40 bytes for an extra entry. */
const BYTES_PER_DEL_QUERY = 40 + util.NUM_BYTES_OBJECT_REF + util.NUM_BYTES_INT

const MAX_INT = int(math.MaxInt32)

const VERBOSE = false

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
		return fmt.Sprintf(
			"BufferedDeletes[gen=%v, numTerms=%v, terms=%v, queries=%v, docIDs=%v, bytesUsed=%v]",
			bd.gen, atomic.LoadInt32(&bd.numTermDeletes), bd.terms, bd.queries, bd.docIDs, bd.bytesUsed)
	} else {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "BufferedDeletes[gen=%v", bd.gen)
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
		buf.WriteRune(']')
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
	// Terms, in sorted order:
	terms     *PrefixCodedTerms
	termCount int // just for debugging

	// Parallel array of deleted query, and the docIDUpto for each
	_queries       []Query
	queryLimits    []int
	bytesUsed      int64
	numTermDeletes int
	gen            int64 // -1, assigned by BufferedDeletesStream once pushed
	// true iff this frozen packet represents a segment private deletes
	// in that case it should only have queries
	isSegmentPrivate bool
}

func freezeBufferedDeletes(deletes *BufferedDeletes, isPrivate bool) *FrozenBufferedDeletes {
	assert2(!isPrivate || len(deletes.terms) == 0,
		"segment private package should only have del queries")
	var termsArray []*Term
	for k, _ := range deletes.terms {
		termsArray = append(termsArray, k)
	}
	util.TimSort(TermSorter(termsArray))
	builder := newPrefixCodedTermsBuilder()
	for _, term := range termsArray {
		builder.add(term)
	}
	terms := builder.finish()

	queries := make([]Query, len(deletes.queries))
	queryLimits := make([]int, len(deletes.queries))
	var upto = 0
	for k, v := range deletes.queries {
		queries[upto] = k
		queryLimits[upto] = v
		upto++
	}

	return &FrozenBufferedDeletes{
		gen:              -1,
		isSegmentPrivate: isPrivate,
		termCount:        len(termsArray),
		terms:            terms,
		_queries:         queries,
		queryLimits:      queryLimits,
		bytesUsed:        terms.sizeInBytes() + int64(len(queries))*BYTES_PER_DEL_QUERY,
		numTermDeletes:   int(atomic.LoadInt32(&deletes.numTermDeletes)),
	}
}

func (bd *FrozenBufferedDeletes) queries() []*QueryAndLimit {
	panic("not implemented yet")
}

func (bd *FrozenBufferedDeletes) String() string {
	panic("not implemented yet")
}

func (d *FrozenBufferedDeletes) any() bool {
	return d.termCount > 0 || len(d._queries) > 0
}
