package search

import (
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/util"
)

// search/Weight.java

/*
Expert: calculate query weights and build query scorers.

The purpose of Weight is to ensure searching does not modify a Qurey,
so that a Query instance can be reused.
IndexSearcher dependent state of the query should reside in the Weight.
AtomicReader dependent state should reside in the Scorer.

Since Weight creates Scorer instances for a given AtomicReaderContext
(scorer()), callers must maintain the relationship between the
searcher's top-level IndexReaderCntext and context used to create a
Scorer.

A Weight is used in the following way:
	1. A Weight is constructed by a top-level query, given a
	IndexSearcher (Query.createWeight()).
	2. The valueForNormalizatin() method is called on the Weight to
	compute the query normalization factor Similarity.queryNorm() of
	the query clauses contained in the query.
	3. The query normlaization factor is passed to normalize(). At this
	point the weighting is complete.
	4. A Scorer is constructed by scorer().
*/
type Weight interface {
	// An explanation of the score computation for the named document.
	Explain(*index.AtomicReaderContext, int) (Explanation, error)
	/** The value for normalization of contained query clauses (e.g. sum of squared weights). */
	ValueForNormalization() float32
	/** Assigns the query normalization factor and boost from parent queries to this. */
	Normalize(norm float32, topLevelBoost float32)
	/**
	 * Returns a {@link Scorer} which scores documents in/out-of order according
	 * to <code>scoreDocsInOrder</code>.
	 * <p>
	 * <b>NOTE:</b> even if <code>scoreDocsInOrder</code> is false, it is
	 * recommended to check whether the returned <code>Scorer</code> indeed scores
	 * documents out of order (i.e., call {@link #scoresDocsOutOfOrder()}), as
	 * some <code>Scorer</code> implementations will always return documents
	 * in-order.<br>
	 * <b>NOTE:</b> null can be returned if no documents will be scored by this
	 * query.
	 */
	Scorer(ctx *index.AtomicReaderContext, inOrder bool, topScorer bool, acceptDocs util.Bits) (sc Scorer, err error)
	/**
	 * Returns true iff this implementation scores docs only out of order. This
	 * method is used in conjunction with {@link Collector}'s
	 * {@link Collector#acceptsDocsOutOfOrder() acceptsDocsOutOfOrder} and
	 * {@link #scorer(AtomicReaderContext, boolean, boolean, Bits)} to
	 * create a matching {@link Scorer} instance for a given {@link Collector}, or
	 * vice versa.
	 * <p>
	 * <b>NOTE:</b> the default implementation returns <code>false</code>, i.e.
	 * the <code>Scorer</code> scores documents in-order.
	 */
	IsScoresDocsOutOfOrder() bool // usually false
}
