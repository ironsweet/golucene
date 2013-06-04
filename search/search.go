package search

import (
	"lucene/index"
	"math"
)

// IndexSearcher
type IndexSearcher struct {
	reader        index.Reader
	readerContext index.ReaderContext
	leafContexts  []index.AtomicReaderContext
	similarity    Similarity
}

func NewIndexSearcher(r index.Reader) IndexSearcher {
	return NewIndexSearcherFromContext(r.Context())
}

func NewIndexSearcherFromContext(context index.ReaderContext) IndexSearcher {
	//assert context.isTopLevel: "IndexSearcher's ReaderContext must be topLevel for reader" + context.reader();
	return IndexSearcher{context.Reader(), context, context.Leaves()}
}

func (ss IndexSearcher) Search(q Query, f Filter, n int) TopDocs {
	return ss.Search(createNormalizedWeight(wrapFilter(q, f)), nil, n)
}

func (ss IndexSearcher) search(W Weight, after ScoreDoc, nDocs int) TopDocs {
	// TODO support concurrent search
	return TopDocs{}
}

func (ss IndexSearcher) TopReaderContext() {
	return ss.readerContext
}

func wrapFilter(q Query, f Filter) Query {
	if f == nil {
		return q
	}
	panic("FilteredQuery not supported yet")
}

func (ss IndexSearcher) createNormalizedWeight(q Query) Weight {
	q = rewrite(q, ss.reader)
	w := q.createWeight(ss)
	v := w.ValueForNormalization()
	norm := ss.similarity.queryNorm(v)
	if math.IsInf(norm, 1) || math.IsNaN(norm) {
		norm = 1.0
	}
	w.normalize(norm, 1.0)
	return w
}

func rewrite(q Query, r index.Reader) Query {
	after := q.rewrite(r)
	for after != q {
		q = after
		after = q.rewrite(r)
	}
	return q
}

func (ss IndexSearcher) termStatistics(term Term, context TermContext) TermStatistics {
	return NewTermStatistics(term.Bytes, context.DocFreq(), context.TotalTermFreq())
}

func (ss IndexSearcher) collectionStatistics(field string) CollectionStatistics {
	terms := index.GetTerms(ss.reader, field)
	if terms.iterator == nil {
		return NewCollectionStatistics(field, ss.reader.MaxDoc(), 0, 0, 0)
	}
	return NewCollectionStatistics(field, ss.reader.MaxDoc(), terms.DocCount(), terms.SumTotalTermFreq(), terms.SumDocFreq())
}

type ScoreDoc struct {
}

type TermStatistics struct {
	Term                   []byte
	DocFreq, TotalTermFreq long
}

func NewTermStatistics(term []byte, docFreq, totalTermFreq long) {
	// assert docFreq >= 0;
	// assert totalTermFreq == -1 || totalTermFreq >= docFreq; // #positions must be >= #postings
	return TermStatistics{term, docFreq, totalTermFreq}
}

type CollectionStatistics struct {
	field                                          string
	maxDoc, docCount, sumTotalTermFreq, sumDocFreq long
}

func NewCollectionStatistics(field string, maxDoc, docCount, sumTotalTermFreq, sumDocFreq long) CollectionStatistics {
	// assert maxDoc >= 0;
	// assert docCount >= -1 && docCount <= maxDoc; // #docs with field must be <= #docs
	// assert sumDocFreq == -1 || sumDocFreq >= docCount; // #postings must be >= #docs with field
	// assert sumTotalTermFreq == -1 || sumTotalTermFreq >= sumDocFreq; // #positions must be >= #postings
	return CollectionStatistics{field, maxDoc, docCount, sumTotalTermFreq, sumDocFreq}
}

type TopDocs struct {
	totalHits int
}

type Similarity struct {
}

func (sim Similartity) queryNorm(valueForNormalization) {
	return 1
}

type SimWeight interface {
	ValueForNormalization() float
	Normalize(norm, topLevelBoost float) float
}
