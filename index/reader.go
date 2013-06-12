package index

import (
	"fmt"
	"lucene/store"
	"lucene/util"
	"math"
	"sync"
)

type IndexReader struct {
	self              interface{} // to infer embedders
	closedByChild     bool
	refCount          uint32 // synchronized
	parentReaders     map[*IndexReader]bool
	parentReadersLock sync.RWMutex
	Context           func() IndexReaderContext
	MaxDoc            func() int
	NumDocs           func() int
}

func newIndexReader(self interface{}) *IndexReader {
	return &IndexReader{self: self, refCount: 1}
}

func (r *IndexReader) ensureOpen() {
	if r.refCount <= 0 {
		panic("this IndexReader is closed")
	}
	// the happens before rule on reading the refCount, which must be after the fake write,
	// ensures that we see the value:
	if r.closedByChild {
		panic("this IndexReader cannot be used anymore as one of its child readers was closed")
	}
}

func (r *IndexReader) registerParentReader(reader *IndexReader) {
	r.ensureOpen()
	r.parentReadersLock.Lock()
	r.parentReaders[reader] = true
	r.parentReadersLock.Unlock()
}

type IndexReaderContext struct {
	self            interface{} // to infer embedders
	parent          *CompositeReaderContext
	isTopLevel      bool
	docBaseInParent int
	ordInParent     int
	Reader          func() *IndexReader
	Leaves          func() []AtomicReaderContext
	Children        func() []IndexReaderContext
}

func newIndexReaderContext(self interface{}, parent *CompositeReaderContext, ordInParent, docBaseInParent int) *IndexReaderContext {
	return &IndexReaderContext{
		parent:          parent,
		isTopLevel:      parent == nil,
		docBaseInParent: docBaseInParent,
		ordInParent:     ordInParent}
}

type AtomicReader struct {
	*IndexReader  // inherit IndexReader
	readerContext *AtomicReaderContext
	Fields        func() Fields
	LiveDocs      func() util.Bits
}

func newAtomicReader() *AtomicReader {
	ans := &AtomicReader{}
	ans.IndexReader = newIndexReader(ans)
	ans.readerContext = newAtomicReaderContextFromReader(ans)
	return ans
}

func (r *AtomicReader) Context() AtomicReaderContext {
	r.IndexReader.ensureOpen()
	return *(r.readerContext)
}

func (r *AtomicReader) Terms(field string) Terms {
	fields := r.Fields()
	if fields == nil {
		return nil
	}
	return fields.Terms(field)
}

type AtomicReaderContext struct {
	*IndexReaderContext // inherit IndexReaderContext
	Ord, DocBase        int
	reader              *AtomicReader
	leaves              []AtomicReaderContext
}

func newAtomicReaderContextFromReader(r *AtomicReader) *AtomicReaderContext {
	return newAtomicReaderContext(nil, r, 0, 0, 0, 0)
}

func newAtomicReaderContext(parent *CompositeReaderContext, reader *AtomicReader, ord, docBase, leafOrd, leafDocBase int) *AtomicReaderContext {
	ans := &AtomicReaderContext{}
	super := newIndexReaderContext(ans, parent, ord, docBase)
	super.Leaves = func() []AtomicReaderContext {
		if !ans.IndexReaderContext.isTopLevel {
			panic("This is not a top-level context.")
		}
		// assert leaves != null
		return ans.leaves
	}
	super.Children = func() []IndexReaderContext {
		return nil
	}
	ans.IndexReaderContext = super
	ans.Ord = leafOrd
	ans.DocBase = leafDocBase
	ans.reader = reader
	if super.isTopLevel {
		ans.leaves = []AtomicReaderContext{*ans}
	}
	return ans
}

func (ctx *AtomicReaderContext) Reader() *AtomicReader {
	return ctx.reader
}

type CompositeReader struct {
	*IndexReader                                    // inherit IndexReader
	readerContext           *CompositeReaderContext // lazy load
	getSequentialSubReaders func() []*IndexReader
}

func newCompositeReader() *CompositeReader {
	ans := &CompositeReader{}
	ans.IndexReader = newIndexReader(ans)
	return ans
}

func (r *CompositeReader) Context() CompositeReaderContext {
	r.IndexReader.ensureOpen()
	// lazy init without thread safety for perf reasons: Building the readerContext twice does not hurt!
	if r.readerContext == nil {
		// assert getSequentialSubReaders() != null;
		r.readerContext = newCompositeReaderContext(r)
	}
	return *(r.readerContext)
}

type CompositeReaderContext struct {
	*IndexReaderContext // inherit IndexReaderContext
	children            []IndexReaderContext
	leaves              []AtomicReaderContext
	reader              *CompositeReader
}

func newCompositeReaderContext(r *CompositeReader) *CompositeReaderContext {
	return newCompositeReaderContextBuilder(r).build()
}

func newCompositeReaderContext3(reader *CompositeReader,
	children []IndexReaderContext, leaves []AtomicReaderContext) *CompositeReaderContext {
	return newCompositeReaderContext6(nil, reader, 0, 0, children, leaves)
}

func newCompositeReaderContext5(parent *CompositeReaderContext, reader *CompositeReader,
	ordInParent, docBaseInParent int, children []IndexReaderContext) *CompositeReaderContext {
	return newCompositeReaderContext6(parent, reader, ordInParent, docBaseInParent, children, nil)
}

func newCompositeReaderContext6(parent *CompositeReaderContext,
	reader *CompositeReader,
	ordInParent, docBaseInParent int,
	children []IndexReaderContext,
	leaves []AtomicReaderContext) *CompositeReaderContext {
	ans := &CompositeReaderContext{}
	super := newIndexReaderContext(ans, parent, ordInParent, docBaseInParent)
	super.Leaves = func() []AtomicReaderContext {
		if !ans.IndexReaderContext.isTopLevel {
			panic("This is not a top-level context.")
		}
		// assert leaves != null
		return ans.leaves
	}
	super.Children = func() []IndexReaderContext {
		return ans.children
	}
	ans.children = children
	ans.leaves = leaves
	ans.reader = reader
	return ans
}

func (ctx *CompositeReaderContext) Reader() *CompositeReader {
	return ctx.reader
}

type CompositeReaderContextBuilder struct {
	reader      *CompositeReader
	leaves      []AtomicReaderContext
	leafDocBase int
}

func newCompositeReaderContextBuilder(r *CompositeReader) CompositeReaderContextBuilder {
	return CompositeReaderContextBuilder{reader: r}
}

func (b CompositeReaderContextBuilder) build() *CompositeReaderContext {
	return b.build4(nil, b.reader.IndexReader, 0, 0).self.(*CompositeReaderContext)
}

func (b CompositeReaderContextBuilder) build4(parent *CompositeReaderContext,
	reader *IndexReader, ord, docBase int) *IndexReaderContext {
	if ar, ok := reader.self.(*AtomicReader); ok {
		atomic := newAtomicReaderContext(parent, ar, ord, docBase, len(b.leaves), b.leafDocBase)
		b.leaves = append(b.leaves, *atomic)
		b.leafDocBase += reader.MaxDoc()
		return atomic.IndexReaderContext
	}
	cr := reader.self.(*CompositeReader)
	sequentialSubReaders := cr.getSequentialSubReaders()
	children := make([]IndexReaderContext, len(sequentialSubReaders))
	var newParent *CompositeReaderContext
	if parent == nil {
		newParent = newCompositeReaderContext3(cr, children, b.leaves)
	} else {
		newParent = newCompositeReaderContext5(parent, cr, ord, docBase, children)
	}
	newDocBase := 0
	for i, r := range sequentialSubReaders {
		children[i] = *(b.build4(parent, r, i, newDocBase))
		newDocBase = r.MaxDoc()
	}
	// assert newDocBase == cr.maxDoc()
	return newParent.IndexReaderContext
}

var (
	EMPTY_ARRAY = []ReaderSlice{}
)

type ReaderSlice struct {
	start, length, readerIndex int
}

func (rs ReaderSlice) String() string {
	return fmt.Sprintf("slice start=%v length=%v readerIndex=%v", rs.start, rs.length, rs.readerIndex)
}

type BaseCompositeReader struct {
	*CompositeReader
	subReaders []*IndexReader
	starts     []int
	maxDoc     int
	numDocs    int
}

func newBaseCompositeReader(readers []*IndexReader) *BaseCompositeReader {
	ans := &BaseCompositeReader{}
	ans.subReaders = readers
	ans.starts = make([]int, len(readers)+1) // build starts array
	var maxDoc, numDocs int
	for i, r := range readers {
		ans.starts[i] = maxDoc
		maxDoc += r.MaxDoc() // compute maxDocs
		if maxDoc < 0 {      // overflow
			panic(fmt.Sprintf("Too many documents, composite IndexReaders cannot exceed %v", math.MaxInt32))
		}
		numDocs += r.NumDocs() // compute numDocs
		r.registerParentReader(ans.CompositeReader.IndexReader)
	}
	ans.starts[len(readers)] = maxDoc
	ans.maxDoc = maxDoc
	ans.numDocs = numDocs
	return ans
}

const DEFAULT_TERMS_INDEX_DIVISOR = 1

type DirectoryReader struct {
	*BaseCompositeReader
	directory store.Directory
}

func newDirectoryReader(directory store.Directory, segmentReaders []*AtomicReader) *DirectoryReader {
	readers := make([]*IndexReader, len(segmentReaders))
	for i, v := range segmentReaders {
		readers[i] = v.IndexReader
	}
	return &DirectoryReader{newBaseCompositeReader(readers), directory}
}

func OpenDirectoryReader(directory store.Directory) *DirectoryReader {
	return OpenStandardDirectoryReader(directory, DEFAULT_TERMS_INDEX_DIVISOR)
}

type StandardDirectoryReader struct {
}

// TODO support IndexCommit
func OpenStandardDirectoryReader(directory store.Directory, termInfosIndexDivisor int) *DirectoryReader {
	return NewFindSegmentsFile(directory, func(segmentFileName string) (obj interface{}, err error) {
		sis := NewSegmentInfos()
		sis.Read(directory, segmentFileName)
		readers := make([]*SegmentReader, sis.Size())
		for i := sis.Size() - 1; i >= 0; i-- {
			readers[i], err = NewSegmentReader(sis.info(i), termInfosIndexDivisor, store.IO_CONTEXT_READ)
			if err != nil {
				util.CloseWhileHandlingException(err, readers)
			}
		}
		return newStandardDirectoryReader(directory, readers, nil, sis, termInfosIndexDivisor, false)
	}).run(nil).(*DirectoryReader)
}
