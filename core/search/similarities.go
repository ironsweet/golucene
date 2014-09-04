package search

import (
	"github.com/balzaczyy/golucene/core/index"
)

// search/similarities/Similarity.java

/*
Similarity defines the components of Lucene scoring.

Expert: Scoring API.

This is a low-level API, you should only extend this API if you want
to implement an information retrieval model. If you are instead
looking for a convenient way to alter Lucene's scoring, consider
extending a high-level implementation such as TFIDFSimilarity, which
implements the vector space model with this API, or just tweaking the
default implementation: DefaultSimilarity.

Similarity determines how Lucene weights terms, and Lucene interacts
with this class at both index-time and query-time.

######Index-time

At indexing time, the indexer calls computeNorm(), allowing the
Similarity implementation to set a per-document value for the field
that will be later accessible via AtomicReader.NormValues(). Lucene
makes no assumption about what is in this norm, but it is most useful
for encoding length normalization information.

Implementations should carefully consider how the normalization is
encoded: while Lucene's classical TFIDFSimilarity encodes a
combination of index-time boost and length normalization information
with SmallFLoat into a single byte, this might not be suitble for all
purposes.

Many formulas require the use of average document length, which can
be computed via a combination of CollectionStatistics.SumTotalTermFreq()
and CollectionStatistics.MaxDoc() or CollectionStatistics.DocCount(),
depending upon whether the average should reflect field sparsity.

Additional scoring factors can be stored in named NumericDocValuesFields
and accessed at query-time with AtomicReader.NumericDocValues().

Finally, using index-time boosts (either via folding into the
normalization byte or via DocValues), is an inefficient way to boost
the scores of different fields if the boost will be the same for
every document, instead the Similarity can simply take a constant
boost parameter C, and PerFieldSimilarityWrapper can return different
instances with different boosts depending upon field name.

######Query-time

At query-time, Quries interact with the Similarity via these steps:

1. The computeWeight() method is called a single time, allowing the
implementation to compute any statistics (such as IDF, average
document length, etc) across the entire collection. The TermStatistics
and CollectionStatistics passed in already contain all of the raw
statistics involved, so a Similarity can freely use any combination
of statistics without causing any additional I/O. Lucene makes no
assumption about what is stored in the returned SimWeight object.
2. The query normalization process occurs a single time:
SimWeight.ValueForNormalization() is called for each query leaf node,
queryNorm() is called for the top-level query, and finally
SimWeight.Normalize() passes down the normalization value and any
top-level boosts (e.g. from enclosing BooleanQuerys).
3. For each sgment in the index, the Query creates a SimScorer. The
score() method is called for each matching document.

######Exlain-time
When IndexSearcher.explain() is called, queries consult the
Similarity's DocScorer for an explanation of how it computed its
score. The query passes in a the document id and an explanation of
how the frequency was computed.
*/
type Similarity interface {
	Coord(int, int) float32
	// Computes the normalization value for a query given the sum of
	// the normalized weights SimWeight.ValueForNormalization of each
	// of the query terms. This value is passed back to the weight
	// (SimWeight.normalize()) of each query term, to provide a  hook
	// to attempt to make scores from different queries comparable.
	QueryNorm(valueForNormalization float32) float32
	/*
		Computes the normalization value for a field, given the
		accumulated state of term processing for this field (see
		FieldInvertState).

		Matches in longer fields are less precise, so implementations
		of this method usually set smaller values when state.Lenght() is
		larger, and larger values when state.Lenght() is smaller.
	*/
	ComputeNorm(state *index.FieldInvertState) int64
	// Compute any collection-level weight (e.g. IDF, average document
	// length, etc) needed for scoring a query.
	computeWeight(queryBoost float32, collectionStats CollectionStatistics, termStats ...TermStatistics) SimWeight
	// Creates a new SimScorer to score matching documents from a
	// segment of the inverted index.
	simScorer(w SimWeight, ctx *index.AtomicReaderContext) (ss SimScorer, err error)
}

// similarities/PerFieldSimilarityWrapper

type PerFieldSimilarityWrapperSPI interface {
	Get(name string) Similarity
}

/*
Provides the ability to use a different Similarity for different
fields.

Subclasses should implement Get() to return an appropriate Similarity
(for example, using field-specific parameter values) for the field.
*/
type PerFieldSimilarityWrapper struct {
	spi PerFieldSimilarityWrapperSPI
}

func NewPerFieldSimilarityWrapper(spi PerFieldSimilarityWrapperSPI) *PerFieldSimilarityWrapper {
	return &PerFieldSimilarityWrapper{spi: spi}
}

func (wrapper *PerFieldSimilarityWrapper) ComputeNorm(state *index.FieldInvertState) int64 {
	return wrapper.spi.Get(state.Name()).ComputeNorm(state)
}

func (wrapper *PerFieldSimilarityWrapper) computeWeight(queryBoost float32,
	collectionStats CollectionStatistics, termStats ...TermStatistics) SimWeight {
	sim := wrapper.spi.Get(collectionStats.field)
	return &PerFieldSimWeight{sim, sim.computeWeight(queryBoost, collectionStats, termStats...)}
}

func (wrapper *PerFieldSimilarityWrapper) simScorer(w SimWeight, ctx *index.AtomicReaderContext) (ss SimScorer, err error) {
	panic("not implemented yet")
}

type PerFieldSimWeight struct {
	delegate       Similarity
	delegateWeight SimWeight
}

func (w *PerFieldSimWeight) ValueForNormalization() float32 {
	return w.delegateWeight.ValueForNormalization()
}

func (w *PerFieldSimWeight) Normalize(queryNorm, topLevelBoost float32) {
	w.delegateWeight.Normalize(queryNorm, topLevelBoost)
}
