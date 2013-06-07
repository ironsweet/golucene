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
	Similarity    Similarity
}

func NewIndexSearcher(r index.Reader) IndexSearcher {
	return NewIndexSearcherFromContext(r.Context())
}

func NewIndexSearcherFromContext(context index.ReaderContext) IndexSearcher {
	//assert context.isTopLevel: "IndexSearcher's ReaderContext must be topLevel for reader" + context.reader();
	defaultSimilarity := NewDefaultSimilarity()
	return IndexSearcher{context.Reader(), context, context.Leaves(), defaultSimilarity}
}

func (ss IndexSearcher) Search(q Query, f Filter, n int) TopDocs {
	return ss.searchWSI(ss.createNormalizedWeight(wrapFilter(q, f)), ScoreDoc{}, n)
}

func (ss IndexSearcher) searchWSI(w Weight, after ScoreDoc, nDocs int) TopDocs {
	// TODO support concurrent search
	return ss.searchLWSI(ss.leafContexts, w, after, nDocs)
}

func (ss IndexSearcher) searchLWSI(leaves []index.AtomicReaderContext,
	w Weight, after ScoreDoc, nDocs int) TopDocs {
	limit := ss.reader.MaxDoc()
	if limit == 0 {
		limit = 1
	}
	if nDocs > limit {
		nDocs = limit
	}
	collector := NewTopScoreDocCollector(nDocs, after, !w.IsScoresDocsOutOfOrder())
	searchLWC(leaves, w, collector)
	return collector.TopDocs
}

func (ss IndexSearcher) TopReaderContext() index.IndexReaderContext {
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
	after := q.Rewrite(r)
	for after != q {
		q = after
		after = q.Rewrite(r)
	}
	return q
}

func (ss IndexSearcher) TermStatistics(term index.Term, context index.TermContext) TermStatistics {
	return NewTermStatistics(term.Bytes, context.DocFreq, context.TotalTermFreq)
}

func (ss IndexSearcher) CollectionStatistics(field string) CollectionStatistics {
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
	DocFreq, TotalTermFreq int64
}

func NewTermStatistics(term []byte, docFreq, totalTermFreq int64) {
	// assert docFreq >= 0;
	// assert totalTermFreq == -1 || totalTermFreq >= docFreq; // #positions must be >= #postings
	return TermStatistics{term, docFreq, totalTermFreq}
}

type CollectionStatistics struct {
	field                                          string
	maxDoc, docCount, sumTotalTermFreq, sumDocFreq int64
}

func NewCollectionStatistics(field string, maxDoc, docCount, sumTotalTermFreq, sumDocFreq int64) CollectionStatistics {
	// assert maxDoc >= 0;
	// assert docCount >= -1 && docCount <= maxDoc; // #docs with field must be <= #docs
	// assert sumDocFreq == -1 || sumDocFreq >= docCount; // #postings must be >= #docs with field
	// assert sumTotalTermFreq == -1 || sumTotalTermFreq >= sumDocFreq; // #positions must be >= #postings
	return CollectionStatistics{field, maxDoc, docCount, sumTotalTermFreq, sumDocFreq}
}

type TopDocs struct {
	totalHits int
}

type Similarity interface {
	// queryNorm(valueForNormalization float32) float32
	computeWeight(queryBoost float32, collectionStats CollectionStatistics, termStats ...TermStatistics) SimWeight
}

type SimWeight interface {
	ValueForNormalization() float32
	Normalize(norm, topLevelBoost float32) float32
}

type TFIDFSimilarity struct {
}

type DefaultSimilarity struct {
	*TFIDFSimilarity
}

func NewDefaultSimilarity() Similarity {
	return &DefaultSimilarity{&TFIDFSimilarity{}}
}
