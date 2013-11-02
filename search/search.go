package search

import (
	"fmt"
	"github.com/balzaczyy/golucene/index"
	"log"
	"math"
)

// IndexSearcher
type IndexSearcher struct {
	reader        index.IndexReader
	readerContext index.IndexReaderContext
	leafContexts  []index.AtomicReaderContext
	similarity    Similarity
}

func NewIndexSearcher(r index.IndexReader) IndexSearcher {
	log.Print("Initializing IndexSearcher from IndexReader: ", r)
	return NewIndexSearcherFromContext(r.Context())
}

func NewIndexSearcherFromContext(context index.IndexReaderContext) IndexSearcher {
	//assert context.isTopLevel: "IndexSearcher's ReaderContext must be topLevel for reader" + context.reader();
	defaultSimilarity := NewDefaultSimilarity()
	return IndexSearcher{context.Reader(), context, context.Leaves(), defaultSimilarity}
}

func (ss IndexSearcher) SearchTop(q Query, n int) (topDocs TopDocs, err error) {
	return ss.Search(q, nil, n)
}

func (ss IndexSearcher) Search(q Query, f Filter, n int) (topDocs TopDocs, err error) {
	w, err := ss.createNormalizedWeight(wrapFilter(q, f))
	if err != nil {
		return TopDocs{}, err
	}
	return ss.searchWSI(w, ScoreDoc{}, n), nil
}

/** Expert: Low-level search implementation.  Finds the top <code>n</code>
 * hits for <code>query</code>, applying <code>filter</code> if non-null.
 *
 * <p>Applications should usually call {@link IndexSearcher#search(Query,int)} or
 * {@link IndexSearcher#search(Query,Filter,int)} instead.
 * @throws BooleanQuery.TooManyClauses If a query would exceed
 *         {@link BooleanQuery#getMaxClauseCount()} clauses.
 */
func (ss IndexSearcher) searchWSI(w Weight, after ScoreDoc, nDocs int) TopDocs {
	// TODO support concurrent search
	return ss.searchLWSI(ss.leafContexts, w, after, nDocs)
}

/** Expert: Low-level search implementation.  Finds the top <code>n</code>
 * hits for <code>query</code>.
 *
 * <p>Applications should usually call {@link IndexSearcher#search(Query,int)} or
 * {@link IndexSearcher#search(Query,Filter,int)} instead.
 * @throws BooleanQuery.TooManyClauses If a query would exceed
 *         {@link BooleanQuery#getMaxClauseCount()} clauses.
 */
func (ss IndexSearcher) searchLWSI(leaves []index.AtomicReaderContext, w Weight, after ScoreDoc, nDocs int) TopDocs {
	// single thread
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
	for _, ctx := range leaves { // search each subreader
		log.Print(ctx)
		c.SetNextReader(ctx)
		// GOTO: CollectionTerminatedException
		if scorer, ok := w.Scorer(ctx, !c.AcceptsDocsOutOfOrder(), true,
			ctx.Reader().(index.AtomicReader).LiveDocs()); ok {
			scorer.ScoreAndCollect(c)
			// GOTO: CollectionTerminatedException
		}
	}
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

func (ss IndexSearcher) createNormalizedWeight(q Query) (w Weight, err error) {
	q = rewrite(q, ss.reader)
	log.Printf("After rewrite: %v", q)
	w, err = q.CreateWeight(ss)
	if err != nil {
		return nil, err
	}
	v := w.ValueForNormalization()
	norm := ss.similarity.queryNorm(v)
	if math.IsInf(float64(norm), 1) || math.IsNaN(float64(norm)) {
		norm = 1.0
	}
	w.Normalize(norm, 1.0)
	return w, nil
}

func rewrite(q Query, r index.IndexReader) Query {
	log.Printf("Rewriting '%v'...", q)
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
	terms := index.GetMultiTerms(ss.reader, field)
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
	queryNorm(valueForNormalization float32) float32
	computeWeight(queryBoost float32, collectionStats CollectionStatistics, termStats ...TermStatistics) SimWeight
	simScorer(w SimWeight, ctx index.AtomicReaderContext) SimScorer
}

/**
 * API for scoring "sloppy" queries such as {@link TermQuery},
 * {@link SpanQuery}, and {@link PhraseQuery}.
 * <p>
 * Frequencies are floating-point values: an approximate
 * within-document frequency adjusted for "sloppiness" by
 * {@link SimScorer#computeSlopFactor(int)}.
 */
type SimScorer interface {
	/**
	 * Score a single document
	 * @param doc document id within the inverted index segment
	 * @param freq sloppy term frequency
	 * @return document's score
	 */
	Score(doc, freq int) float32
}

type SimWeight interface {
	ValueForNormalization() float32
	Normalize(norm float32, topLevelBoost float32)
}

// search/similarities/TFIDFSimilarity.java

type ITFIDFSimilarity interface {
	/** Computes a score factor based on a term's document frequency (the number
	 * of documents which contain the term).  This value is multiplied by the
	 * {@link #tf(float)} factor for each term in the query and these products are
	 * then summed to form the initial score for a document.
	 *
	 * <p>Terms that occur in fewer documents are better indicators of topic, so
	 * implementations of this method usually return larger values for rare terms,
	 * and smaller values for common terms.
	 *
	 * @param docFreq the number of documents which contain the term
	 * @param numDocs the total number of documents in the collection
	 * @return a score factor based on the term's document frequency
	 */
	idf(docFreq int64, numDocs int64) float32
}

type TFIDFSimilarity struct {
	ITFIDFSimilarity
}

func (ts *TFIDFSimilarity) idfExplainTerm(collectionStats CollectionStatistics, termStats TermStatistics) Explanation {
	df, max := termStats.DocFreq, collectionStats.maxDoc
	idf := ts.idf(df, max)
	return newExplanation(idf, fmt.Sprintf("idf(docFreq=%v, maxDocs=%v)", df, max))
}

func (ts *TFIDFSimilarity) idfExplainPhrase(collectionStats CollectionStatistics, termStats []TermStatistics) Explanation {
	details := make([]Explanation, len(termStats))
	var idf float32 = 0
	for i, stat := range termStats {
		details[i] = ts.idfExplainTerm(collectionStats, stat)
		idf += details[i].value
	}
	return newExplanation(idf, fmt.Sprintf("idf(), sum of:"))
}

func (ts *TFIDFSimilarity) computeWeight(queryBoost float32, collectionStats CollectionStatistics, termStats ...TermStatistics) SimWeight {
	var idf Explanation
	if len(termStats) == 1 {
		idf = ts.idfExplainTerm(collectionStats, termStats[0])
	} else {
		idf = ts.idfExplainPhrase(collectionStats, termStats)
	}
	return newIDFStats(collectionStats.field, idf, queryBoost)
}

func (ts *TFIDFSimilarity) simScorer(w SimWeight, ctx index.AtomicReaderContext) SimScorer {
	panic("not implemented yet")
}

/** Collection statistics for the TF-IDF model. The only statistic of interest
 * to this model is idf. */
type idfStats struct {
	field string
	/** The idf and its explanation */
	idf         Explanation
	queryNorm   float32
	queryWeight float32
	queryBoost  float32
	value       float32
}

func newIDFStats(field string, idf Explanation, queryBoost float32) *idfStats {
	// TODO: validate?
	return &idfStats{
		field:       field,
		idf:         idf,
		queryBoost:  queryBoost,
		queryWeight: idf.value * queryBoost, // compute query weight
	}
}

func (stats *idfStats) ValueForNormalization() float32 {
	// TODO: (sorta LUCENE-1907) make non-static class and expose this squaring via a nice method to subclasses?
	return stats.queryWeight * stats.queryWeight // sum of squared weights
}

func (stats *idfStats) Normalize(queryNorm float32, topLevelBoost float32) {
	stats.queryNorm = queryNorm * topLevelBoost
	stats.queryWeight *= stats.queryNorm              // normalize query weight
	stats.value = stats.queryWeight * stats.idf.value // idf for document
}

// search/Explanation.java
/** Expert: Describes the score computation for document and query. */
type Explanation struct {
	// the value of this node
	value float32
	// what it represents
	description string
}

func newExplanation(value float32, description string) Explanation {
	return Explanation{value: value, description: description}
}

// search/similarities/DefaultSimilarity.java

type DefaultSimilarity struct {
	*TFIDFSimilarity
	discountOverlaps bool
}

func NewDefaultSimilarity() Similarity {
	ans := &DefaultSimilarity{
		&TFIDFSimilarity{},
		true,
	}
	ans.ITFIDFSimilarity = ans
	return ans
}

func (ds *DefaultSimilarity) queryNorm(sumOfSquaredWeights float32) float32 {
	return 1.0 / float32(math.Sqrt(float64(sumOfSquaredWeights)))
}

func (ds *DefaultSimilarity) decodeNormValue(norm int64) float32 {
	panic("not implemented yet")
}

func (ds *DefaultSimilarity) tf(freq float32) float32 {
	return float32(math.Sqrt(float64(freq)))
}

func (ds *DefaultSimilarity) idf(docFreq int64, numDocs int64) float32 {
	return float32(math.Log(float64(numDocs)/float64(docFreq+1))) + 1.0
}

func (ds *DefaultSimilarity) String() string {
	return "DefaultSImilarity"
}
