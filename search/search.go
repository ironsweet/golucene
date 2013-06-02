package lucene

import (
	"lucene/index"
)

// IndexSearcher
type IndexSearcher struct {
	reader        index.Reader
	readerContext index.ReaderContext
	leafContexts  []index.AtomicReaderContext
}

func NewIndexSearcher(r index.Reader) IndexSearcher {
	return NewIndexSearcherFromContext(r.Context())
}

func NewIndexSearcherFromContext(context index.ReaderContext) IndexSearcher {
	//assert context.isTopLevel: "IndexSearcher's ReaderContext must be topLevel for reader" + context.reader();
	return IndexSearcher{context.Reader(), context, context.Leaves()}
}
