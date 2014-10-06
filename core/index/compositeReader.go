package index

import (
	"bytes"
	"container/list"
	"fmt"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	"reflect"
)

type CompositeReaderSPI interface {
	getSequentialSubReaders() []IndexReader
}

type CompositeReader interface {
	IndexReader
	CompositeReaderSPI
}

type CompositeReaderImpl struct {
	*IndexReaderImpl
	CompositeReaderSPI
	readerContext *CompositeReaderContext // lazy load
}

func newCompositeReader(spi CompositeReaderSPI, self IndexReaderImplSPI) *CompositeReaderImpl {
	return &CompositeReaderImpl{
		IndexReaderImpl:    newIndexReader(self),
		CompositeReaderSPI: spi,
	}
}

func (r *CompositeReaderImpl) String() string {
	var buf bytes.Buffer
	class := reflect.TypeOf(r.IndexReaderImplSPI).Name()
	if class != "" {
		buf.WriteString(class)
	} else {
		buf.WriteString("CompositeReader")
	}
	buf.WriteString("(")
	subReaders := r.getSequentialSubReaders()
	if len(subReaders) > 0 {
		fmt.Fprintf(&buf, "%v", subReaders[0])
		for i, v := range subReaders {
			if i > 0 {
				fmt.Fprintf(&buf, " %v", v)
			}
		}
	}
	buf.WriteString(")")
	return buf.String()
}

func (r *CompositeReaderImpl) Context() IndexReaderContext {
	r.ensureOpen()
	// lazy init without thread safety for perf reasons: Building the readerContext twice does not hurt!
	if r.readerContext == nil {
		// log.Print("Obtaining context for: ", r)
		// assert getSequentialSubReaders() != null;
		r.readerContext = newCompositeReaderContext(r)
	}
	return r.readerContext
}

type CompositeReaderContext struct {
	*IndexReaderContextImpl
	children []IndexReaderContext
	leaves   *list.List // operated by builder
	reader   CompositeReader
}

func newCompositeReaderContext(r CompositeReader) *CompositeReaderContext {
	return newCompositeReaderContextBuilder(r).build()
}

func newCompositeReaderContext3(reader CompositeReader,
	children []IndexReaderContext, leaves *list.List) *CompositeReaderContext {
	return newCompositeReaderContext6(nil, reader, 0, 0, children, leaves)
}

func newCompositeReaderContext5(parent *CompositeReaderContext, reader CompositeReader,
	ordInParent, docBaseInParent int, children []IndexReaderContext) *CompositeReaderContext {
	return newCompositeReaderContext6(parent, reader, ordInParent, docBaseInParent, children, list.New())
}

func newCompositeReaderContext6(parent *CompositeReaderContext,
	reader CompositeReader,
	ordInParent, docBaseInParent int,
	children []IndexReaderContext,
	leaves *list.List) *CompositeReaderContext {
	ans := &CompositeReaderContext{}
	ans.IndexReaderContextImpl = newIndexReaderContext(parent, ordInParent, docBaseInParent)
	ans.children = children
	ans.leaves = leaves
	ans.reader = reader
	return ans
}

func (ctx *CompositeReaderContext) Leaves() []*AtomicReaderContext {
	assert2(ctx.isTopLevel, "This is not a top-level context.")
	assert(ctx.leaves != nil)
	ans := make([]*AtomicReaderContext, 0, ctx.leaves.Len())
	for e := ctx.leaves.Front(); e != nil; e = e.Next() {
		ans = append(ans, e.Value.(*AtomicReaderContext))
	}
	return ans
}

func (ctx *CompositeReaderContext) Children() []IndexReaderContext {
	return ctx.children
}

func (ctx *CompositeReaderContext) Reader() IndexReader {
	return ctx.reader
}

func (ctx *CompositeReaderContext) String() string {
	return fmt.Sprintf("CompositeReaderContext{%v %v %v}",
		ctx.IndexReaderContextImpl, ctx.children, ctx.reader)
}

type CompositeReaderContextBuilder struct {
	reader      CompositeReader
	leaves      *list.List
	leafDocBase int
}

func newCompositeReaderContextBuilder(r CompositeReader) CompositeReaderContextBuilder {
	return CompositeReaderContextBuilder{reader: r, leaves: list.New()}
}

func (b CompositeReaderContextBuilder) build() *CompositeReaderContext {
	return b.build4(nil, b.reader, 0, 0).(*CompositeReaderContext)
}

func (b CompositeReaderContextBuilder) build4(parent *CompositeReaderContext,
	reader IndexReader, ord, docBase int) IndexReaderContext {
	// log.Printf("Building context from %v(parent: %v, %v-%v)", reader, parent, ord, docBase)
	if ar, ok := reader.(AtomicReader); ok {
		// log.Print("AtomicReader is detected.")
		atomic := newAtomicReaderContext(parent, ar, ord, docBase, b.leaves.Len(), b.leafDocBase)
		b.leaves.PushBack(atomic)
		b.leafDocBase += reader.MaxDoc()
		return atomic
	}
	// log.Print("CompositeReader is detected: ", reader)
	cr := reader.(CompositeReader)
	sequentialSubReaders := cr.getSequentialSubReaders()
	// log.Printf("Found %v sub readers.", len(sequentialSubReaders))
	children := make([]IndexReaderContext, len(sequentialSubReaders))
	var newParent *CompositeReaderContext
	if parent == nil {
		newParent = newCompositeReaderContext3(cr, children, b.leaves)
	} else {
		newParent = newCompositeReaderContext5(parent, cr, ord, docBase, children)
	}
	newDocBase := 0
	for i, r := range sequentialSubReaders {
		children[i] = b.build4(newParent, r, i, newDocBase)
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

type BaseCompositeReaderSPI interface {
	IndexReaderImplSPI
	CompositeReaderSPI
}

type BaseCompositeReader struct {
	*CompositeReaderImpl
	subReaders []IndexReader
	starts     []int
	maxDoc     int
	numDocs    int

	subReadersList []IndexReader
}

func newBaseCompositeReader(spi BaseCompositeReaderSPI, readers []IndexReader) *BaseCompositeReader {
	// log.Printf("Initializing BaseCompositeReader with %v IndexReaders", len(readers))
	ans := &BaseCompositeReader{}
	ans.CompositeReaderImpl = newCompositeReader(spi, spi)
	ans.subReaders = readers
	ans.subReadersList = make([]IndexReader, len(readers))
	copy(ans.subReadersList, readers)
	ans.starts = make([]int, len(readers)+1) // build starts array
	var maxDoc, numDocs int
	for i, r := range readers {
		ans.starts[i] = maxDoc
		maxDoc += r.MaxDoc()                      // compute maxDocs
		if maxDoc < 0 || maxDoc > actualMaxDocs { // overflow
			panic(fmt.Sprintf(
				"Too many documents, composite IndexReaders cannot exceed %v",
				actualMaxDocs))
		}
		numDocs += r.NumDocs() // compute numDocs
		// log.Printf("Obtained %v docs (max %v)", numDocs, maxDoc)
		r.registerParentReader(ans)
	}
	ans.starts[len(readers)] = maxDoc
	ans.maxDoc = maxDoc
	ans.numDocs = numDocs
	// log.Print("Success")
	return ans
}

func (r *BaseCompositeReader) TermVectors(docID int) error {
	panic("not implemented yet")
	// r.ensureOpen()
	// i := readerIndex(docID)
	// return r.subReaders[i].TermVectors(docID - starts[i])
}

func (r *BaseCompositeReader) NumDocs() int {
	// Don't call ensureOpen() here (it could affect performance)
	return r.numDocs
}

func (r *BaseCompositeReader) MaxDoc() int {
	// Don't call ensureOpen() here (it could affect performance)
	return r.maxDoc
}

func (r *BaseCompositeReader) VisitDocument(docID int, visitor StoredFieldVisitor) error {
	r.ensureOpen()
	i := r.readerIndex(docID) // find subreader num
	return r.subReaders[i].VisitDocument(docID-r.starts[i], visitor)
}

func (r *BaseCompositeReader) DocFreq(term *Term) (int, error) {
	panic("not implemented yet")
}

func (r *BaseCompositeReader) TotalTermFreq(term *Term) int64 {
	panic("not implemented yet")
}

func (r *BaseCompositeReader) SumDocFreq(field string) int64 {
	panic("not implemented yet")
}

func (r *BaseCompositeReader) DocCount(field string) int {
	panic("not implemented yet")
}

func (r *BaseCompositeReader) SumTotalTermFreq(field string) int64 {
	panic("not implemented yet")
}

func (r *BaseCompositeReader) readerIndex(docID int) int {
	if docID < 0 || docID >= r.maxDoc {
		panic(fmt.Sprintf("docID must be [0, %v] (got docID=%v)", r.maxDoc, docID))
	}
	return subIndex(docID, r.starts)
}

func (r *BaseCompositeReader) readerBase(readerIndex int) int {
	panic("not implemented yet")
}

func (r *BaseCompositeReader) getSequentialSubReaders() []IndexReader {
	// log.Printf("Found %v sub readers.", len(r.subReadersList))
	return r.subReadersList
}
