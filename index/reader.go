package index

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/store"
	"github.com/balzaczyy/golucene/util"
	"io"
	"log"
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
	Leaves() []AtomicReaderContext
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
		r.reportCloseToParentReaders()
		r.notifyReaderClosedListeners()
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
	defer r.parentReadersLock.Unlock()
	r.parentReaders[reader] = true
}

func (r *IndexReaderImpl) notifyReaderClosedListeners() {
	panic("not implemented yet")
}

func (r *IndexReaderImpl) reportCloseToParentReaders() {
	r.parentReadersLock.Lock()
	defer r.parentReadersLock.Unlock()
	for parent, _ := range r.parentReaders {
		p := parent.(*IndexReaderImpl)
		p.closedByChild = true
		// cross memory barrier by a fake write:
		// FIXME do we need it in Go?
		atomic.AddInt32(&p.refCount, 0)
		// recurse:
		p.reportCloseToParentReaders()
	}
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

func (r *IndexReaderImpl) Leaves() []AtomicReaderContext {
	log.Printf("Debug %%v", r.Context())
	return r.Context().Leaves()
}

type IndexReaderContext interface {
	Reader() IndexReader
	Leaves() []AtomicReaderContext
	Children() []IndexReaderContext
}

type IndexReaderContextImpl struct {
	parent          *CompositeReaderContext
	isTopLevel      bool
	docBaseInParent int
	ordInParent     int
}

func newIndexReaderContext(parent *CompositeReaderContext, ordInParent, docBaseInParent int) *IndexReaderContextImpl {
	return &IndexReaderContextImpl{
		parent:          parent,
		isTopLevel:      parent == nil,
		docBaseInParent: docBaseInParent,
		ordInParent:     ordInParent}
}

type ARFieldsReader interface {
	Terms(field string) Terms
	Fields() Fields
	LiveDocs() util.Bits
}

type AtomicReader interface {
	IndexReader
	ARFieldsReader
}

type AtomicReaderImpl struct {
	*IndexReaderImpl
	ARFieldsReader // need child class implementation
	readerContext  *AtomicReaderContext
}

func newAtomicReader(self IndexReader) *AtomicReaderImpl {
	r := &AtomicReaderImpl{IndexReaderImpl: newIndexReader(self)}
	r.readerContext = newAtomicReaderContextFromReader(r)
	return r
}

func (r *AtomicReaderImpl) Context() IndexReaderContext {
	r.ensureOpen()
	return r.readerContext
}

func (r *AtomicReaderImpl) DocFreq(term Term) (n int, err error) {
	panic("not implemented yet")
}

func (r *AtomicReaderImpl) TotalTermFreq(term Term) (n int64, err error) {
	panic("not implemented yet")
}

func (r *AtomicReaderImpl) SumDocFreq(field string) (n int64, err error) {
	panic("not implemented yet")
}

func (r *AtomicReaderImpl) DocCount(field string) (n int, err error) {
	panic("not implemented yet")
}

func (r *AtomicReaderImpl) SumTotalTermFreq(field string) (n int64, err error) {
	panic("not implemented yet")
}

func (r *AtomicReaderImpl) Terms(field string) Terms {
	fields := r.Fields()
	if fields == nil {
		return nil
	}
	return fields.Terms(field)
}

type AtomicReaderContext struct {
	*IndexReaderContextImpl
	Ord, DocBase int
	reader       AtomicReader
	leaves       []AtomicReaderContext
}

func newAtomicReaderContextFromReader(r AtomicReader) *AtomicReaderContext {
	return newAtomicReaderContext(nil, r, 0, 0, 0, 0)
}

func newAtomicReaderContext(parent *CompositeReaderContext, reader AtomicReader, ord, docBase, leafOrd, leafDocBase int) *AtomicReaderContext {
	ans := &AtomicReaderContext{}
	ans.IndexReaderContextImpl = newIndexReaderContext(parent, ord, docBase)
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

func newCompositeReader(self IndexReader) *CompositeReader {
	ans := &CompositeReader{}
	ans.IndexReaderImpl = newIndexReader(self)
	return ans
}

func (r *CompositeReader) Context() IndexReaderContext {
	log.Print("Obtaining context for CompositeReader...")
	r.ensureOpen()
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
	ans.IndexReaderContextImpl = newIndexReaderContext(parent, ordInParent, docBaseInParent)
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
	log.Print("Building CompositeReaderContext...")
	if ar, ok := reader.(AtomicReader); ok {
		log.Print("AtomicReader is detected.")
		atomic := newAtomicReaderContext(parent, ar, ord, docBase, len(b.leaves), b.leafDocBase)
		b.leaves = append(b.leaves, *atomic)
		b.leafDocBase += reader.MaxDoc()
		return atomic
	}
	log.Print("CompositeReader is detected.")
	cr := reader.(*DirectoryReader).CompositeReader
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
	return newParent
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

func newBaseCompositeReader(self IndexReader, readers []IndexReader) *BaseCompositeReader {
	log.Printf("Initializing BaseCompositeReader with %v IndexReaders", len(readers))
	ans := &BaseCompositeReader{}
	ans.CompositeReader = newCompositeReader(self)
	ans.CompositeReader.getSequentialSubReaders = func() []IndexReader {
		log.Printf("Found %v sub readers.", len(ans.subReaders))
		return ans.subReaders
	}
	ans.subReaders = readers
	ans.starts = make([]int, len(readers)+1) // build starts array
	var maxDoc, numDocs int
	for i, r := range readers {
		log.Print("DEBUG ", r)
		ans.starts[i] = maxDoc
		maxDoc += r.MaxDoc() // compute maxDocs
		log.Print("DEBUG2")
		if maxDoc < 0 { // overflow
			panic(fmt.Sprintf("Too many documents, composite IndexReaders cannot exceed %v", math.MaxInt32))
		}
		numDocs += r.NumDocs() // compute numDocs
		log.Printf("Obtained %v docs (max %v)", numDocs, maxDoc)
		r.registerParentReader(ans.CompositeReader.IndexReader)
	}
	ans.starts[len(readers)] = maxDoc
	ans.maxDoc = maxDoc
	ans.numDocs = numDocs
	log.Print("Success")
	return ans
}

const DEFAULT_TERMS_INDEX_DIVISOR = 1

type DirectoryReader struct {
	*BaseCompositeReader
	directory store.Directory
}

func newDirectoryReader(directory store.Directory, segmentReaders []AtomicReader) *DirectoryReader {
	log.Printf("Initializing DirectoryReader with %v segment readers...", len(segmentReaders))
	readers := make([]IndexReader, len(segmentReaders))
	for i, v := range segmentReaders {
		readers[i] = v
	}
	ans := &DirectoryReader{directory: directory}
	ans.BaseCompositeReader = newBaseCompositeReader(ans, readers)
	return ans
}

func OpenDirectoryReader(directory store.Directory) (r *DirectoryReader, err error) {
	return openStandardDirectoryReader(directory, DEFAULT_TERMS_INDEX_DIVISOR)
}

type StandardDirectoryReader struct {
	*DirectoryReader
}

// TODO support IndexWriter
func newStandardDirectoryReader(directory store.Directory, readers []AtomicReader,
	sis SegmentInfos, termInfosIndexDivisor int, applyAllDeletes bool) *StandardDirectoryReader {
	log.Printf("Initializing StandardDirectoryReader with %v sub readers...", len(readers))
	return &StandardDirectoryReader{newDirectoryReader(directory, readers)}
}

// TODO support IndexCommit
func openStandardDirectoryReader(directory store.Directory,
	termInfosIndexDivisor int) (r *DirectoryReader, err error) {
	log.Print("Initializing SegmentsFile...")
	obj, err := NewFindSegmentsFile(directory, func(segmentFileName string) (obj interface{}, err error) {
		sis := &SegmentInfos{}
		err = sis.Read(directory, segmentFileName)
		if err != nil {
			return nil, err
		}
		log.Printf("Found %v segments...", len(sis.Segments))
		readers := make([]AtomicReader, len(sis.Segments))
		for i := len(sis.Segments) - 1; i >= 0; i-- {
			sr, err := NewSegmentReader(sis.Segments[i], termInfosIndexDivisor, store.IO_CONTEXT_READ)
			readers[i] = sr
			if err != nil {
				rs := make([]io.Closer, len(readers))
				for i, v := range readers {
					rs[i] = v
				}
				return nil, util.CloseWhileHandlingError(err, rs...)
			}
		}
		log.Printf("Obtained %v SegmentReaders.", len(readers))
		return newStandardDirectoryReader(directory, readers, *sis, termInfosIndexDivisor, false), nil
	}).run()
	if err != nil {
		return nil, err
	}
	return obj.(*StandardDirectoryReader).DirectoryReader, err
}
