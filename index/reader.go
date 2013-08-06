package index

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/store"
	"github.com/balzaczyy/golucene/util"
	"io"
	"math"
	"sync"
	"sync/atomic"
)

type IndexReader interface {
	io.Closer
	decRef() error
	ensureOpen()
	registerParentReader(r IndexReader)
	NumDocs() int
	MaxDoc() int
	doClose() error
	Context() IndexReaderContext
}

type IndexReaderImpl struct {
	IndexReader
	lock              sync.Mutex
	closed            bool
	closedByChild     bool
	refCount          int32 // synchronized
	parentReaders     map[IndexReader]bool
	parentReadersLock sync.RWMutex
}

func newIndexReader(self IndexReader) *IndexReaderImpl {
	return &IndexReaderImpl{IndexReader: self, refCount: 1}
}

func (r *IndexReaderImpl) decRef() error {
	// only check refcount here (don't call ensureOpen()), so we can
	// still close the reader if it was made invalid by a child:
	if r.refCount <= 0 {
		return errors.New("this IndexReader is closed")
	}

	rc := atomic.AddInt32(&r.refCount, -1)
	if rc == 0 {
		success := false
		defer func() {
			if !success {
				// Put reference back on failure
				atomic.AddInt32(&r.refCount, 1)
			}
		}()
		r.doClose()
		success = true
	} else if rc < 0 {
		panic(fmt.Sprintf("too many decRef calls: refCount is %v after decrement", rc))
	}

	return nil
}

func (r *IndexReaderImpl) ensureOpen() {
	if atomic.LoadInt32(&r.refCount) <= 0 {
		panic("this IndexReader is closed")
	}
	// the happens before rule on reading the refCount, which must be after the fake write,
	// ensures that we see the value:
	if r.closedByChild {
		panic("this IndexReader cannot be used anymore as one of its child readers was closed")
	}
}

func (r *IndexReaderImpl) registerParentReader(reader IndexReader) {
	r.ensureOpen()
	r.parentReadersLock.Lock()
	r.parentReaders[reader] = true
	r.parentReadersLock.Unlock()
}

func (r *IndexReaderImpl) Close() error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if !r.closed {
		r.closed = true
		return r.decRef()
	}
	return nil
}

type IndexReaderContext interface {
	Reader() IndexReader
	Leaves() []AtomicReaderContext
	Children() []IndexReaderContext
}

type IndexReaderContextImpl struct {
	IndexReaderContext
	parent          *CompositeReaderContext
	isTopLevel      bool
	docBaseInParent int
	ordInParent     int
}

func newIndexReaderContext(self IndexReaderContext, parent *CompositeReaderContext, ordInParent, docBaseInParent int) *IndexReaderContextImpl {
	return &IndexReaderContextImpl{
		IndexReaderContext: self,
		parent:             parent,
		isTopLevel:         parent == nil,
		docBaseInParent:    docBaseInParent,
		ordInParent:        ordInParent}
}

type AtomicReader struct {
	*IndexReaderImpl
	readerContext *AtomicReaderContext
	Fields        func() Fields
	LiveDocs      func() util.Bits
}

func newAtomicReader(self IndexReader) *AtomicReader {
	r := &AtomicReader{IndexReaderImpl: newIndexReader(self)}
	r.readerContext = newAtomicReaderContextFromReader(r)
	return r
}

func (r *AtomicReader) Context() IndexReaderContext {
	r.IndexReader.ensureOpen()
	return r.readerContext
}

func (r *AtomicReader) Terms(field string) Terms {
	fields := r.Fields()
	if fields == nil {
		return nil
	}
	return fields.Terms(field)
}

type AtomicReaderContext struct {
	*IndexReaderContextImpl
	Ord, DocBase int
	reader       *AtomicReader
	leaves       []AtomicReaderContext
}

func newAtomicReaderContextFromReader(r *AtomicReader) *AtomicReaderContext {
	return newAtomicReaderContext(nil, r, 0, 0, 0, 0)
}

func newAtomicReaderContext(parent *CompositeReaderContext, reader *AtomicReader, ord, docBase, leafOrd, leafDocBase int) *AtomicReaderContext {
	ans := &AtomicReaderContext{}
	ans.IndexReaderContext = newIndexReaderContext(ans, parent, ord, docBase)
	ans.Ord = leafOrd
	ans.DocBase = leafDocBase
	ans.reader = reader
	if ans.isTopLevel {
		ans.leaves = []AtomicReaderContext{*ans}
	}
	return ans
}

func (ctx *AtomicReaderContext) Leaves() []AtomicReaderContext {
	if !ctx.IndexReaderContextImpl.isTopLevel {
		panic("This is not a top-level context.")
	}
	// assert leaves != null
	return ctx.leaves
}

func (ctx *AtomicReaderContext) Children() []IndexReaderContext {
	return nil
}

func (ctx *AtomicReaderContext) Reader() IndexReader {
	return ctx.reader
}

type CompositeReader struct {
	*IndexReaderImpl
	readerContext           *CompositeReaderContext // lazy load
	getSequentialSubReaders func() []IndexReader
}

func newCompositeReader() *CompositeReader {
	ans := &CompositeReader{}
	ans.IndexReader = newIndexReader(ans)
	return ans
}

func (r *CompositeReader) Context() IndexReaderContext {
	r.IndexReader.ensureOpen()
	// lazy init without thread safety for perf reasons: Building the readerContext twice does not hurt!
	if r.readerContext == nil {
		// assert getSequentialSubReaders() != null;
		r.readerContext = newCompositeReaderContext(r)
	}
	return r.readerContext
}

type CompositeReaderContext struct {
	*IndexReaderContextImpl
	children []IndexReaderContext
	leaves   []AtomicReaderContext
	reader   *CompositeReader
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
	ans.IndexReaderContextImpl = newIndexReaderContext(ans, parent, ordInParent, docBaseInParent)
	ans.children = children
	ans.leaves = leaves
	ans.reader = reader
	return ans
}

func (ctx *CompositeReaderContext) Leaves() []AtomicReaderContext {
	if !ctx.isTopLevel {
		panic("This is not a top-level context.")
	}
	// assert leaves != null
	return ctx.leaves
}

func (ctx *CompositeReaderContext) Children() []IndexReaderContext {
	return ctx.children
}

func (ctx *CompositeReaderContext) Reader() IndexReader {
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
	return b.build4(nil, b.reader.IndexReader, 0, 0).(*CompositeReaderContext)
}

func (b CompositeReaderContextBuilder) build4(parent *CompositeReaderContext,
	reader IndexReader, ord, docBase int) IndexReaderContext {
	if ar, ok := reader.(*AtomicReader); ok {
		atomic := newAtomicReaderContext(parent, ar, ord, docBase, len(b.leaves), b.leafDocBase)
		b.leaves = append(b.leaves, *atomic)
		b.leafDocBase += reader.MaxDoc()
		return atomic.IndexReaderContext
	}
	cr := reader.(*CompositeReader)
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
		children[i] = b.build4(parent, r, i, newDocBase)
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
	subReaders []IndexReader
	starts     []int
	maxDoc     int
	numDocs    int
}

func newBaseCompositeReader(readers []IndexReader) *BaseCompositeReader {
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
	readers := make([]IndexReader, len(segmentReaders))
	for i, v := range segmentReaders {
		readers[i] = v.IndexReader
	}
	return &DirectoryReader{newBaseCompositeReader(readers), directory}
}

func OpenDirectoryReader(directory store.Directory) (r *DirectoryReader, err error) {
	return openStandardDirectoryReader(directory, DEFAULT_TERMS_INDEX_DIVISOR)
}

type StandardDirectoryReader struct {
	*DirectoryReader
}

// TODO support IndexWriter
func newStandardDirectoryReader(directory store.Directory, readers []*AtomicReader,
	sis SegmentInfos, termInfosIndexDivisor int, applyAllDeletes bool) *StandardDirectoryReader {
	return &StandardDirectoryReader{newDirectoryReader(directory, readers)}
}

// TODO support IndexCommit
func openStandardDirectoryReader(directory store.Directory,
	termInfosIndexDivisor int) (r *DirectoryReader, err error) {
	obj, err := NewFindSegmentsFile(directory, func(segmentFileName string) (obj interface{}, err error) {
		sis := &SegmentInfos{}
		sis.Read(directory, segmentFileName)
		readers := make([]*AtomicReader, len(sis.Segments))
		for i := len(sis.Segments) - 1; i >= 0; i-- {
			sr, err := NewSegmentReader(sis.Segments[i], termInfosIndexDivisor, store.IO_CONTEXT_READ)
			readers[i] = sr.AtomicReader
			if err != nil {
				rs := make([]io.Closer, len(readers))
				for i, v := range readers {
					rs[i] = v
				}
				return nil, util.CloseWhileHandlingError(err, rs...)
			}
		}
		return newStandardDirectoryReader(directory, readers, *sis, termInfosIndexDivisor, false), nil
	}).run()
	return obj.(*DirectoryReader), err
}
