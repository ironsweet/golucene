package search

import (
	"github.com/balzaczyy/golucene/index"
	"math"
)

// IndexSearcher
type IndexSearcher struct {
	reader        *index.IndexReader
	readerContext index.IndexReaderContext
	leafContexts  []index.AtomicReaderContext
	similarity    Similarity
}

func NewIndexSearcher(r index.IndexReader) IndexSearcher {
	return NewIndexSearcherFromContext(r.Context())
}

func NewIndexSearcherFromContext(context index.IndexReaderContext) IndexSearcher {
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
	// TODO support concurrent search
	limit := ss.reader.MaxDoc()
	if limit == 0 {
		limit = 1
	}
	if nDocs > limit {
		nDocs = limit
	}
	collector := NewTopScoreDocCollector(nDocs, after, !w.IsScoresDocsOutOfOrder())
	ss.searchLWC(leaves, w, collector)
	return collector.TopDocs()
}

func (ss IndexSearcher) searchLWC(leaves []index.AtomicReaderContext, w Weight, c Collector) {
	// TODO: should we make this
	// threaded...?  the Collector could be sync'd?
	// always use single thread:
	for _, ctx := range leaves {
		c.SetNextReader(ctx)
		// GOTO: CollectionTerminatedException
		if scorer, ok := w.Scorer(ctx, !c.AcceptsDocsOutOfOrder(), true,
			ctx.Reader().LiveDocs()); ok {
			scorer.ScoreAndCollect(c)
			// GOTO: CollectionTerminatedException
		}
	}
}

func (ss IndexSearcher) TopReaderContext() *index.IndexReaderContext {
	return &ss.readerContext
}

func wrapFilter(q Query, f Filter) Query {
	if f == nil {
		return q
	}
	panic("FilteredQuery not supported yet")
}

func (ss IndexSearcher) createNormalizedWeight(q Query) Weight {
	q = rewrite(q, *(ss.reader))
	w := q.CreateWeight(ss)
	v := w.ValueForNormalization()
	norm := ss.similarity.queryNorm(v)
	if math.IsInf(norm, 1) || math.IsNaN(norm) {
		norm = 1.0
	}
	w.Normalize(norm, 1.0)
	return w
}

func rewrite(q Query, r index.IndexReader) Query {
	after := q.Rewrite(r)
	for after != q {
		q = after
		after = q.Rewrite(r)
	}
	return q
}

func (ss IndexSearcher) TermStatistics(term index.Term, context index.TermContext) TermStatistics {
	return NewTermStatistics(term.Bytes, int64(context.DocFreq), context.TotalTermFreq)
}

func (ss IndexSearcher) CollectionStatistics(field string) CollectionStatistics {
	terms := index.GetMultiTerms(*(ss.reader), field)
	if terms == nil {
		return NewCollectionStatistics(field, int64(ss.reader.MaxDoc()), 0, 0, 0)
	}
	return NewCollectionStatistics(field, int64(ss.reader.MaxDoc()), int64(terms.DocCount()), terms.SumTotalTermFreq(), terms.SumDocFreq())
}

type TermStatistics struct {
	Term                   []byte
	DocFreq, TotalTermFreq int64
}

func NewTermStatistics(term []byte, docFreq, totalTermFreq int64) TermStatistics {
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

type Similarity interface {
	queryNorm(valueForNormalization float32) float64
	computeWeight(queryBoost float32, collectionStats CollectionStatistics, termStats ...TermStatistics) SimWeight
	exactSimScorer(w SimWeight, ctx index.AtomicReaderContext) ExactSimScorer
}

type ExactSimScorer interface {
	Score(doc, freq int) float64
}

type SimWeight interface {
	ValueForNormalization() float32
	Normalize(norm float64, topLevelBoost float32) float32
}

type TFIDFSimilarity struct {
}

func (ts *TFIDFSimilarity) computeWeight(queryBoost float32, collectionStats CollectionStatistics, termStats ...TermStatistics) SimWeight {
	panic("not implemented yet")
}

func (ts *TFIDFSimilarity) exactSimScorer(w SimWeight, ctx index.AtomicReaderContext) ExactSimScorer {
	panic("not implemented yet")
}

type DefaultSimilarity struct {
	*TFIDFSimilarity
}

func (ds *DefaultSimilarity) queryNorm(sumOfSquaredWeights float32) float64 {
	return 1.0 / math.Sqrt(float64(sumOfSquaredWeights))
}

func NewDefaultSimilarity() Similarity {
	return &DefaultSimilarity{&TFIDFSimilarity{}}
}
