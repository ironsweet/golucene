package search

import (
	"fmt"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/util"
	"log"
	"math"
)

/* Define service that can be overrided */
type IndexSearcherSPI interface {
	CreateNormalizedWeight(Query) (Weight, error)
	Rewrite(Query) (Query, error)
	WrapFilter(Query, Filter) Query
	SearchLWC([]*index.AtomicReaderContext, Weight, Collector) error
}

// IndexSearcher
type IndexSearcher struct {
	spi           IndexSearcherSPI
	reader        index.IndexReader
	readerContext index.IndexReaderContext
	leafContexts  []*index.AtomicReaderContext
	similarity    Similarity
}

func NewIndexSearcher(r index.IndexReader) *IndexSearcher {
	// log.Print("Initializing IndexSearcher from IndexReader: ", r)
	return NewIndexSearcherFromContext(r.Context())
}

func NewIndexSearcherFromContext(context index.IndexReaderContext) *IndexSearcher {
	// assert2(context.isTopLevel, "IndexSearcher's ReaderContext must be topLevel for reader %v", context.reader())
	defaultSimilarity := NewDefaultSimilarity()
	ss := &IndexSearcher{nil, context.Reader(), context, context.Leaves(), defaultSimilarity}
	ss.spi = ss
	return ss
}

/* Expert: set the similarity implementation used by this IndexSearcher. */
func (ss *IndexSearcher) SetSimilarity(similarity Similarity) {
	ss.similarity = similarity
}

func (ss *IndexSearcher) SearchTop(q Query, n int) (topDocs TopDocs, err error) {
	return ss.Search(q, nil, n)
}

func (ss *IndexSearcher) Search(q Query, f Filter, n int) (topDocs TopDocs, err error) {
	w, err := ss.spi.CreateNormalizedWeight(ss.spi.WrapFilter(q, f))
	if err != nil {
		return TopDocs{}, err
	}
	return ss.searchWSI(w, nil, n), nil
}

/** Expert: Low-level search implementation.  Finds the top <code>n</code>
 * hits for <code>query</code>, applying <code>filter</code> if non-null.
 *
 * <p>Applications should usually call {@link IndexSearcher#search(Query,int)} or
 * {@link IndexSearcher#search(Query,Filter,int)} instead.
 * @throws BooleanQuery.TooManyClauses If a query would exceed
 *         {@link BooleanQuery#getMaxClauseCount()} clauses.
 */
func (ss *IndexSearcher) searchWSI(w Weight, after *ScoreDoc, nDocs int) TopDocs {
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
func (ss *IndexSearcher) searchLWSI(leaves []*index.AtomicReaderContext,
	w Weight, after *ScoreDoc, nDocs int) TopDocs {
	// single thread
	limit := ss.reader.MaxDoc()
	if limit == 0 {
		limit = 1
	}
	if nDocs > limit {
		nDocs = limit
	}
	collector := NewTopScoreDocCollector(nDocs, after, !w.IsScoresDocsOutOfOrder())
	ss.spi.SearchLWC(leaves, w, collector)
	return collector.TopDocs()
}

func (ss *IndexSearcher) SearchLWC(leaves []*index.AtomicReaderContext, w Weight, c Collector) (err error) {
	// TODO: should we make this
	// threaded...?  the Collector could be sync'd?
	// always use single thread:
	for _, ctx := range leaves { // search each subreader
		// TODO catch CollectionTerminatedException
		c.SetNextReader(ctx)

		scorer, err := w.BulkScorer(ctx, !c.AcceptsDocsOutOfOrder(),
			ctx.Reader().(index.AtomicReader).LiveDocs())
		if err != nil {
			return err
		}
		if scorer != nil {
			err = scorer.ScoreAndCollect(c)
		} // TODO catch CollectionTerminatedException
	}
	return
}

func (ss *IndexSearcher) WrapFilter(q Query, f Filter) Query {
	if f == nil {
		return q
	}
	panic("FilteredQuery not supported yet")
}

/*
Returns an Explanation that describes how doc scored against query.

This is intended to be used in developing Similiarity implemenations, and, for
good performance, should not be displayed with every hit. Computing an
explanation is as expensive as executing the query over the entire index.
*/
func (ss *IndexSearcher) Explain(query Query, doc int) (exp Explanation, err error) {
	w, err := ss.spi.CreateNormalizedWeight(query)
	if err == nil {
		return ss.explain(w, doc)
	}
	return
}

/*
Expert: low-level implementation method
Returns an Explanation that describes how doc scored against weight.

This is intended to be used in developing Similarity implementations, and, for
good performance, should not be displayed with every hit. Computing an
explanation is as expensive as executing the query over the entire index.

Applications should call explain(Query, int).
*/
func (ss *IndexSearcher) explain(weight Weight, doc int) (exp Explanation, err error) {
	n := index.SubIndex(doc, ss.leafContexts)
	ctx := ss.leafContexts[n]
	deBasedDoc := doc - ctx.DocBase
	return weight.Explain(ctx, deBasedDoc)
}

func (ss *IndexSearcher) CreateNormalizedWeight(q Query) (w Weight, err error) {
	q, err = ss.spi.Rewrite(q)
	if err != nil {
		return nil, err
	}
	log.Printf("After rewrite: %v", q)
	w, err = q.CreateWeight(ss)
	if err != nil {
		return nil, err
	}
	v := w.ValueForNormalization()
	norm := ss.similarity.QueryNorm(v)
	if math.IsInf(float64(norm), 1) || math.IsNaN(float64(norm)) {
		norm = 1.0
	}
	w.Normalize(norm, 1.0)
	return w, nil
}

func (ss *IndexSearcher) Rewrite(q Query) (Query, error) {
	log.Printf("Rewriting '%v'...", q)
	after := q.Rewrite(ss.reader)
	for after != q {
		q = after
		after = q.Rewrite(ss.reader)
	}
	return q, nil
}

// Returns this searhcers the top-level IndexReaderContext
func (ss *IndexSearcher) TopReaderContext() index.IndexReaderContext {
	return ss.readerContext
}

func (ss *IndexSearcher) String() string {
	return fmt.Sprintf("IndexSearcher(%v)", ss.reader)
}

func (ss *IndexSearcher) TermStatistics(term *index.Term, context *index.TermContext) TermStatistics {
	return NewTermStatistics(term.Bytes, int64(context.DocFreq), context.TotalTermFreq)
}

func (ss *IndexSearcher) CollectionStatistics(field string) CollectionStatistics {
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
	Score(doc int, freq float32) float32
	// Explain the score for a single document
	explain(int, Explanation) Explanation
}

type SimWeight interface {
	ValueForNormalization() float32
	Normalize(norm float32, topLevelBoost float32)
}

// search/similarities/TFIDFSimilarity.java

type ITFIDFSimilarity interface {
	/** Computes a score factor based on a term or phrase's frequency in a
	 * document.  This value is multiplied by the {@link #idf(long, long)}
	 * factor for each term in the query and these products are then summed to
	 * form the initial score for a document.
	 *
	 * <p>Terms and phrases repeated in a document indicate the topic of the
	 * document, so implementations of this method usually return larger values
	 * when <code>freq</code> is large, and smaller values when <code>freq</code>
	 * is small.
	 *
	 * @param freq the frequency of a term within a document
	 * @return a score factor based on a term's within-document frequency
	 */
	tf(freq float32) float32
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
	// Compute an index-time normalization value for this field instance.
	//
	// This value will be stored in a single byte lossy representation
	// by encodeNormValue().
	lengthNorm(*index.FieldInvertState) float32
	// Decodes a normalization factor stored in an index.
	decodeNormValue(norm int64) float32
	// Encodes a normalization factor for storage in an index.
	encodeNormValue(float32) int64
}

type TFIDFSimilarity struct {
	spi ITFIDFSimilarity
}

func newTFIDFSimilarity(spi ITFIDFSimilarity) *TFIDFSimilarity {
	return &TFIDFSimilarity{spi}
}

func (ts *TFIDFSimilarity) idfExplainTerm(collectionStats CollectionStatistics, termStats TermStatistics) Explanation {
	df, max := termStats.DocFreq, collectionStats.maxDoc
	idf := ts.spi.idf(df, max)
	return newExplanation(idf, fmt.Sprintf("idf(docFreq=%v, maxDocs=%v)", df, max))
}

func (ts *TFIDFSimilarity) idfExplainPhrase(collectionStats CollectionStatistics, termStats []TermStatistics) Explanation {
	details := make([]Explanation, len(termStats))
	var idf float32 = 0
	for i, stat := range termStats {
		details[i] = ts.idfExplainTerm(collectionStats, stat)
		idf += details[i].(*ExplanationImpl).value
	}
	ans := newExplanation(idf, fmt.Sprintf("idf(), sum of:"))
	ans.details = details
	return ans
}

func (ts *TFIDFSimilarity) ComputeNorm(state *index.FieldInvertState) int64 {
	return ts.spi.encodeNormValue(ts.spi.lengthNorm(state))
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

func (ts *TFIDFSimilarity) simScorer(stats SimWeight, ctx *index.AtomicReaderContext) (ss SimScorer, err error) {
	idfstats := stats.(*idfStats)
	ndv, err := ctx.Reader().(index.AtomicReader).NormValues(idfstats.field)
	if err != nil {
		return nil, err
	}
	return newTFIDFSimScorer(ts, idfstats, ndv), nil
}

type tfIDFSimScorer struct {
	owner       *TFIDFSimilarity
	stats       *idfStats
	weightValue float32
	norms       NumericDocValues
}

func newTFIDFSimScorer(owner *TFIDFSimilarity, stats *idfStats, norms NumericDocValues) *tfIDFSimScorer {
	return &tfIDFSimScorer{owner, stats, stats.value, norms}
}

func (ss *tfIDFSimScorer) Score(doc int, freq float32) float32 {
	raw := ss.owner.spi.tf(freq) * ss.weightValue // compute tf(f)*weight
	if ss.norms == nil {
		return raw
	}
	return raw * ss.owner.spi.decodeNormValue(ss.norms(doc)) // normalize for field
}

func (ss *tfIDFSimScorer) explain(doc int, freq Explanation) Explanation {
	return ss.owner.explainScore(doc, freq, ss.stats, ss.norms)
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
		queryWeight: idf.(*ExplanationImpl).value * queryBoost, // compute query weight
	}
}

func (stats *idfStats) ValueForNormalization() float32 {
	// TODO: (sorta LUCENE-1907) make non-static class and expose this squaring via a nice method to subclasses?
	return stats.queryWeight * stats.queryWeight // sum of squared weights
}

func (stats *idfStats) Normalize(queryNorm float32, topLevelBoost float32) {
	stats.queryNorm = queryNorm * topLevelBoost
	stats.queryWeight *= stats.queryNorm                                 // normalize query weight
	stats.value = stats.queryWeight * stats.idf.(*ExplanationImpl).value // idf for document
}

func (ss *TFIDFSimilarity) explainScore(doc int, freq Explanation,
	stats *idfStats, norms NumericDocValues) Explanation {

	// explain query weight
	boostExpl := newExplanation(stats.queryBoost, "boost")
	queryNormExpl := newExplanation(stats.queryNorm, "queryNorm")
	queryExpl := newExplanation(
		boostExpl.value*stats.idf.Value()*queryNormExpl.value,
		"queryWeight, product of:")
	if stats.queryBoost != 1 {
		queryExpl.addDetail(boostExpl)
	}
	queryExpl.addDetail(stats.idf)
	queryExpl.addDetail(queryNormExpl)

	// explain field weight
	tfExplanation := newExplanation(ss.spi.tf(freq.Value()),
		fmt.Sprintf("tf(freq=%v), with freq of:", freq.Value()))
	tfExplanation.addDetail(freq)
	fieldNorm := float32(1)
	if norms != nil {
		fieldNorm = ss.spi.decodeNormValue(norms(doc))
	}
	fieldNormExpl := newExplanation(fieldNorm, fmt.Sprintf("fieldNorm(doc=%v)", doc))
	fieldExpl := newExplanation(
		tfExplanation.value*stats.idf.Value()*fieldNormExpl.value,
		fmt.Sprintf("fieldWeight in %v, product of:", doc))
	fieldExpl.addDetail(tfExplanation)
	fieldExpl.addDetail(stats.idf)
	fieldExpl.addDetail(fieldNormExpl)

	if queryExpl.value == 1 {
		return fieldExpl
	}

	// combine them
	ans := newExplanation(queryExpl.value*fieldExpl.value,
		fmt.Sprintf("score(doc=%v,freq=%v), product of:", doc, freq))
	ans.addDetail(queryExpl)
	ans.addDetail(fieldExpl)
	return ans
}

// search/similarities/DefaultSimilarity.java

/** Cache of decoded bytes. */
var NORM_TABLE []float32 = buildNormTable()

func buildNormTable() []float32 {
	table := make([]float32, 256)
	for i, _ := range table {
		table[i] = util.Byte315ToFloat(byte(i))
	}
	return table
}

type DefaultSimilarity struct {
	*TFIDFSimilarity
	discountOverlaps bool
}

func NewDefaultSimilarity() *DefaultSimilarity {
	ans := &DefaultSimilarity{discountOverlaps: true}
	ans.TFIDFSimilarity = newTFIDFSimilarity(ans)
	return ans
}

func (ds *DefaultSimilarity) Coord(overlap, maxOverlap int) float32 {
	return float32(overlap) / float32(maxOverlap)
}

func (ds *DefaultSimilarity) QueryNorm(sumOfSquaredWeights float32) float32 {
	return float32(1.0 / math.Sqrt(float64(sumOfSquaredWeights)))
}

/*
Encodes a normalization factor for storage in an index.

The encoding uses a three-bit mantissa, a five-bit exponent, and the
zero-exponent point at 15, thus representing values from around
7x10^9 to 2x10^-9 with about one significant decimal digit of
accuracy. Zero is also represented. Negative numbers are rounded up
to zero. Values too large to represent are rounded down to the
largest representable value. Positive values too small to represent
are rounded up to the smallest positive representable value.
*/
func (ds *DefaultSimilarity) encodeNormValue(f float32) int64 {
	return int64(util.FloatToByte315(f))
}

func (ds *DefaultSimilarity) decodeNormValue(norm int64) float32 {
	return NORM_TABLE[int(norm&0xff)] // & 0xFF maps negative bytes to positive above 127
}

/*
Implemented as state.boost() * lengthNorm(numTerms), where numTerms
is FieldInvertState.length() if setDiscountOverlaps() is false, else
it's FieldInvertState.length() - FieldInvertState.numOverlap().
*/
func (ds *DefaultSimilarity) lengthNorm(state *index.FieldInvertState) float32 {
	var numTerms int
	if ds.discountOverlaps {
		numTerms = state.Length() - state.NumOverlap()
	} else {
		numTerms = state.Length()
	}
	return state.Boost() * float32(1.0/math.Sqrt(float64(numTerms)))
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
