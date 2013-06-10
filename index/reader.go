package index

import (
	"fmt"
	"lucene/util"
)

type IndexReader struct {
	refCount      uint32
	closedByChild bool
	Context       func() IndexReaderContext
	MaxDoc        func() int
}

func newIndexReader() *IndexReader {
	return &IndexReader{refCount: 1}
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
	parent          *CompositeReaderContext
	isTopLevel      bool
	docBaseInParent int
	ordInParent     int
	Reader          func() IndexReader
	Leaves          func() []AtomicReaderContext
	Children        func() []IndexReaderContext
}

func newIndexReaderContext(parent *CompositeReaderContext, ordInParent, docBaseInParent int) *IndexReaderContext {
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
	ans.IndexReader = newIndexReader()
	ans.readerContext = newAtomicReaderContextFromReader(ans)
	return ans
}

func (r AtomicReader) Context() AtomicReaderContext {
	r.IndexReader.ensureOpen()
	return *(r.readerContext)
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
	super := newIndexReaderContext(parent, ord, docBase)
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

func (ctx *AtomicReaderContext) Reader() AtomicReader {
	return *(ctx.reader)
}

type CompositeReader struct {
	*IndexReader                                    // inherit IndexReader
	readerContext           *CompositeReaderContext // lazy load
	getSequentialSubReaders func() []IndexReader
}

func newCompositeReader() *CompositeReader {
	ans := &CompositeReader{}
	ans.IndexReader = newIndexReader()
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
	super := newIndexReaderContext(parent, ordInParent, docBaseInParent)
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

func (ctx *CompositeReaderContext) Reader() CompositeReader {
	return *(ctx.reader)
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
	reader IndexReader, ord, docBase int) IndexReaderContext {
	if ar, ok := reader.(*AtomicReader); ok {
		atomic := newAtomicReaderContext(parent, ar, ord, docBase, len(b.leaves), b.leafDocBase)
		b.leaves = append(b.leaves, atomic)
		b.leafDocBase += reader.MaxDoc()
		return atomic
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
		children[i] = *(b.build4(parent, r, i, newDocBase))
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
