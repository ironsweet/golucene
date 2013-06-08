package index

import (
	"fmt"
	"lucene/util"
)

type Reader interface {
	Context() ReaderContext
	MaxDoc() int
}

type ReaderContext interface {
	Reader() Reader
	Leaves() []AtomicReaderContext
	Children() []IndexReaderContext
}

type IndexReader struct {
	refCount      uint32
	closedByChild bool
}

func NewIndexReader() *IndexReader {
	return &IndexReader{refCount: 1}
}

func (r *IndexReader) Context() ReaderContext {
	return nil
}

func (r *IndexReader) MaxDoc() int {
	return -1
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

type IndexReaderContext struct {
	parent                       *CompositeReaderContext
	isTopLevel                   bool
	docBaseInParent, ordInParent int
}

func newIndexReaderContext(parent *CompositeReaderContext, ordInParent, docBaseInParent int) IndexReaderContext {
	return IndexReaderContext{parent, parent == nil, docBaseInParent, ordInParent}
}

func (ctx *IndexReaderContext) Reader() Reader {
	panic("not implemented")
}

func (ctx *IndexReaderContext) Leaves() []AtomicReaderContext {
	panic("not implemented")
}

func (ctx *IndexReaderContext) Children() []IndexReaderContext {
	panic("not implemented")
}

type AtomicReader struct {
	*IndexReader  // inherit IndexReader
	readerContext *AtomicReaderContext
	Fields        func() Fields
	LiveDocs      func() util.Bits
}

func (r *AtomicReader) Context() ReaderContext {
	r.IndexReader.ensureOpen()
	return r.readerContext
}

func (r *AtomicReader) MaxDoc() int {
	return -1
}

type AtomicReaderContext struct {
	*IndexReaderContext // inherit IndexReaderContext
	Ord, DocBase        int
	reader              *AtomicReader
	leaves              []AtomicReaderContext
}

func (ctx *AtomicReaderContext) Reader() Reader {
	return ctx.reader
}

func (ctx *AtomicReaderContext) Leaves() []AtomicReaderContext {
	if !ctx.isTopLevel {
		panic("This is not a top-level context.")
	}
	// assert leaves != null
	return ctx.leaves
}

func (ctx *AtomicReaderContext) Children() []IndexReaderContext {
	return nil
}

func NewAtomicReaderContext(parent *CompositeReaderContext, reader *AtomicReader, ord, docBase, leafOrd, leafDocBase int) AtomicReaderContext {
	super := newIndexReaderContext(parent, ord, docBase)
	ans := AtomicReaderContext{&super, leafOrd, leafDocBase, reader, nil}
	if super.isTopLevel {
		ans.leaves = []AtomicReaderContext{ans}
	}
	return ans
}

type CompositeReader struct {
	*IndexReader                          // inherit IndexReader
	readerContext *CompositeReaderContext // lazy load
}

func getSequentialSubReaders(r *CompositeReader) []*IndexReader {
	panic("not implemented")
}

func (r *CompositeReader) Context() ReaderContext {
	r.IndexReader.ensureOpen()
	// lazy init without thread safety for perf reasons: Building the readerContext twice does not hurt!
	if r.readerContext == nil {
		// assert getSequentialSubReaders() != null;
		r.readerContext = newCompositeReaderContext(r)
	}
	return r.readerContext
}

func (r *CompositeReader) MaxDoc() int {
	return -1
}

type CompositeReaderContext struct {
	*IndexReaderContext // inherit IndexReaderContext
	children            []IndexReaderContext
	leaves              []AtomicReaderContext
	reader              *CompositeReader
}

func (ctx *CompositeReaderContext) Reader() Reader {
	return ctx.reader
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

func newCompositeReaderContext(r *CompositeReader) *CompositeReaderContext {
	return newCompositeReaderContextBuilder(r).build()
}

func newCompositeReaderContext3(reader *CompositeReader,
	children []IndexReaderContext, leaves []AtomicReaderContext) CompositeReaderContext {
	return newCompositeReaderContext6(nil, reader, 0, 0, children, leaves)
}

func newCompositeReaderContext5(parent *CompositeReaderContext, reader *CompositeReader,
	ordInParent, docBaseInParent int, children []IndexReaderContext) CompositeReaderContext {
	return newCompositeReaderContext6(parent, reader, ordInParent, docBaseInParent, children, nil)
}

func newCompositeReaderContext6(parent *CompositeReaderContext, reader *CompositeReader,
	ordInParent, docBaseInParent int, children []IndexReaderContext,
	leaves []AtomicReaderContext) CompositeReaderContext {
	super := newIndexReaderContext(parent, ordInParent, docBaseInParent)
	return CompositeReaderContext{&super, children, leaves, reader}
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
	return b.build4(nil, b.reader, 0, 0).(*CompositeReaderContext)
}

func (b CompositeReaderContextBuilder) build4(parent *CompositeReaderContext,
	reader Reader, ord, docBase int) ReaderContext {
	if ar, ok := reader.(*AtomicReader); ok {
		atomic := NewAtomicReaderContext(parent, ar, ord, docBase, len(b.leaves), b.leafDocBase)
		b.leaves = append(b.leaves, atomic)
		b.leafDocBase += reader.MaxDoc()
		return &atomic
	}
	cr := reader.(*CompositeReader)
	sequentialSubReaders := getSequentialSubReaders(cr)
	children := make([]IndexReaderContext, len(sequentialSubReaders))
	var newParent CompositeReaderContext
	if parent == nil {
		newParent = newCompositeReaderContext3(cr, children, b.leaves)
	} else {
		newParent = newCompositeReaderContext5(parent, cr, ord, docBase, children)
	}
	newDocBase := 0
	for i, r := range sequentialSubReaders {
		children[i] = *(b.build4(parent, r, i, newDocBase).(*IndexReaderContext))
		newDocBase = r.MaxDoc()
	}
	// assert newDocBase == cr.maxDoc()
	return &newParent
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
