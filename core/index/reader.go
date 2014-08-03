package index

import (
	// "errors"
	"errors"
	"fmt"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	docu "github.com/balzaczyy/golucene/core/document"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/util"
	"io"
	"reflect"
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
	/** Expert: visits the fields of a stored document, for
	 *  custom processing/loading of each field.  If you
	 *  simply want to load all fields, use {@link
	 *  #document(int)}.  If you want to load a subset, use
	 *  {@link DocumentStoredFieldVisitor}.  */
	VisitDocument(docID int, visitor StoredFieldVisitor) error
	/**
	 * Returns the stored fields of the <code>n</code><sup>th</sup>
	 * <code>Document</code> in this index.  This is just
	 * sugar for using {@link DocumentStoredFieldVisitor}.
	 * <p>
	 * <b>NOTE:</b> for performance reasons, this method does not check if the
	 * requested document is deleted, and therefore asking for a deleted document
	 * may yield unspecified results. Usually this is not required, however you
	 * can test if the doc is deleted by checking the {@link
	 * Bits} returned from {@link MultiFields#getLiveDocs}.
	 *
	 * <b>NOTE:</b> only the content of a field is returned,
	 * if that field was stored during indexing.  Metadata
	 * like boost, omitNorm, IndexOptions, tokenized, etc.,
	 * are not preserved.
	 *
	 * @throws IOException if there is a low-level IO error
	 */
	// TODO: we need a separate StoredField, so that the
	// Document returned here contains that class not
	//model.IndexableField
	Document(docID int) (doc *docu.Document, err error)
	doClose() error
	Context() IndexReaderContext
	Leaves() []*AtomicReaderContext
	// Returns the number of documents containing the term. This method
	// returns 0 if the term of field does not exists. This method does
	// not take into account deleted documents that have not yet been
	// merged away.
	DocFreq(*Term) (int, error)
}

/* A custom listener that's invoked when the IndexReader is closed. */
type ReaderClosedListener interface {
	onClose(IndexReader)
}

type IndexReaderImplSPI interface {
	NumDocs() int
	MaxDoc() int
	VisitDocument(int, StoredFieldVisitor) error
	doClose() error
	Context() IndexReaderContext
	DocFreq(*Term) (int, error)
}

type IndexReaderImpl struct {
	IndexReaderImplSPI

	lock                      sync.Mutex
	closed                    bool
	closedByChild             bool
	refCount                  int32 // synchronized
	parentReaders             map[IndexReader]bool
	parentReadersLock         sync.RWMutex
	readerClosedListeners     map[ReaderClosedListener]bool
	readerClosedListenersLock sync.RWMutex
}

func newIndexReader(spi IndexReaderImplSPI) *IndexReaderImpl {
	return &IndexReaderImpl{
		IndexReaderImplSPI: spi,
		refCount:           1,
		parentReaders:      make(map[IndexReader]bool),
	}
}

func (r *IndexReaderImpl) decRef() error {
	// only check refcount here (don't call ensureOpen()), so we can
	// still close the reader if it was made invalid by a child:
	assert2(r.refCount > 0, "this IndexReader is closed")

	rc := atomic.AddInt32(&r.refCount, -1)
	assert2(rc >= 0, "too many decRef calls: refCount is %v after decrement", rc)
	if rc == 0 {
		r.closed = true
		var err error
		defer func() {
			defer r.notifyReaderClosedListeners(err)
			r.reportCloseToParentReaders()
		}()
		return r.doClose()
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

func (r *IndexReaderImpl) notifyReaderClosedListeners(err error) {
	r.readerClosedListenersLock.RLock()
	defer r.readerClosedListenersLock.RUnlock()
	for listener, _ := range r.readerClosedListeners {
		func() {
			defer func() {
				if e := recover(); e != nil {
					err = mergeError(err, errors.New(fmt.Sprintf("%v", e)))
				}
			}()
			listener.onClose(r)
		}()
	}
	return
}

func (r *IndexReaderImpl) reportCloseToParentReaders() {
	r.parentReadersLock.RLock()
	defer r.parentReadersLock.RUnlock()
	for parent, _ := range r.parentReaders {
		if p, ok := parent.(*IndexReaderImpl); ok {
			p.closedByChild = true
			// cross memory barrier by a fake write:
			// FIXME do we need it in Go?
			atomic.AddInt32(&p.refCount, 0)
			// recurse:
			p.reportCloseToParentReaders()
		} else if p, ok := parent.(*BaseCompositeReader); ok {
			p.closedByChild = true
			// cross memory barrier by a fake write:
			// FIXME do we need it in Go?
			atomic.AddInt32(&p.refCount, 0)
			// recurse:
			p.reportCloseToParentReaders()
		} else {
			panic(fmt.Sprintf("Unknown IndexReader type: %v", reflect.TypeOf(parent).Name()))
		}
	}
}

/* Returns the number of deleted documents. */
func (r *IndexReaderImpl) numDeletedDocs() int {
	return r.MaxDoc() - r.NumDocs()
}

func (r *IndexReaderImpl) Document(docID int) (doc *docu.Document, err error) {
	visitor := docu.NewDocumentStoredFieldVisitor()
	if err = r.VisitDocument(docID, visitor); err != nil {
		return nil, err
	}
	return visitor.Document(), nil
}

/*
Returns true if any documents have been deleted. Implementers should
consider overriding this method if maxDoc() or numDocs() are not
constant-time operations.
*/
func (r *IndexReaderImpl) hasDeletions() bool {
	return r.numDeletedDocs() > 0
}

func (r *IndexReaderImpl) Close() error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if !r.closed {
		if err := r.decRef(); err != nil {
			return err
		}
		r.closed = true
	}
	return nil
}

func (r *IndexReaderImpl) Leaves() []*AtomicReaderContext {
	return r.Context().Leaves()
}

type IndexReaderContext interface {
	Reader() IndexReader
	Parent() *CompositeReaderContext
	Leaves() []*AtomicReaderContext
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

func (ctx *IndexReaderContextImpl) Parent() *CompositeReaderContext {
	return ctx.parent
}

type ARFieldsReader interface {
	Terms(field string) Terms
	Fields() Fields
	LiveDocs() util.Bits
	/** Returns {@link NumericDocValues} representing norms
	 *  for this field, or null if no {@link NumericDocValues}
	 *  were indexed. The returned instance should only be
	 *  used by a single thread. */
	NormValues(field string) (ndv NumericDocValues, err error)
}

type AtomicReader interface {
	IndexReader
	ARFieldsReader
}

type AtomicReaderImplSPI interface {
	IndexReaderImplSPI
	ARFieldsReader
}

type AtomicReaderImpl struct {
	*IndexReaderImpl
	ARFieldsReader

	readerContext *AtomicReaderContext
}

func newAtomicReader(spi AtomicReaderImplSPI) *AtomicReaderImpl {
	r := &AtomicReaderImpl{
		IndexReaderImpl: newIndexReader(spi),
		ARFieldsReader:  spi,
	}
	r.readerContext = newAtomicReaderContextFromReader(r)
	return r
}

func (r *AtomicReaderImpl) Context() IndexReaderContext {
	r.ensureOpen()
	return r.readerContext
}

func (r *AtomicReaderImpl) DocFreq(term *Term) (int, error) {
	if fields := r.Fields(); fields != nil {
		if terms := fields.Terms(term.Field); terms != nil {
			termsEnum := terms.Iterator(nil)
			ok, err := termsEnum.SeekExact(term.Bytes)
			if err != nil {
				return 0, err
			}
			if ok {
				return termsEnum.DocFreq()
			}
		}
	}
	return 0, nil
}

func (r *AtomicReaderImpl) TotalTermFreq(term *Term) (n int64, err error) {
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
	leaves       []*AtomicReaderContext
}

func (ctx *AtomicReaderContext) String() string {
	return fmt.Sprintf("AtomicReaderContext{%v ord=%v docBase=%v %v}",
		ctx.IndexReaderContextImpl, ctx.Ord, ctx.DocBase, ctx.reader)
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
		ans.leaves = []*AtomicReaderContext{ans}
	}
	return ans
}

func (ctx *AtomicReaderContext) Leaves() []*AtomicReaderContext {
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
