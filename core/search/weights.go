package search

import (
	"github.com/balzaczyy/golucene/core/index"
	. "github.com/balzaczyy/golucene/core/search/model"
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
	// Scorer(*index.AtomicReaderContext, util.Bits) (Scorer, error)
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
	BulkScorer(*index.AtomicReaderContext, bool, util.Bits) (BulkScorer, error)
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

type WeightImplSPI interface {
	Scorer(*index.AtomicReaderContext, util.Bits) (Scorer, error)
}

type WeightImpl struct {
	spi WeightImplSPI
}

func newWeightImpl(spi WeightImplSPI) *WeightImpl {
	return &WeightImpl{spi}
}

func (w *WeightImpl) BulkScorer(ctx *index.AtomicReaderContext,
	scoreDocsInOrder bool, acceptDoc util.Bits) (bs BulkScorer, err error) {

	var scorer Scorer
	if scorer, err = w.spi.Scorer(ctx, acceptDoc); err != nil {
		return nil, err
	} else if scorer == nil {
		// no docs match
		return nil, nil
	}

	// this impl always scores docs in order, so we can ignore scoreDocsInOrder:
	return newDefaultScorer(scorer), nil
}

/* Just wraps a Scorer and performs top scoring using it. */
type DefaultBulkScorer struct {
	*BulkScorerImpl
	scorer Scorer
}

func newDefaultScorer(scorer Scorer) *DefaultBulkScorer {
	assert(scorer != nil)
	ans := &DefaultBulkScorer{scorer: scorer}
	ans.BulkScorerImpl = newBulkScorer(ans)
	return ans
}

func (s *DefaultBulkScorer) ScoreAndCollectUpto(collector Collector, max int) (ok bool, err error) {
	collector.SetScorer(s.scorer)
	if max == NO_MORE_DOCS {
		return false, s.scoreAll(collector, s.scorer)
	}
	doc := s.scorer.DocId()
	if doc < 0 {
		if doc, err = s.scorer.NextDoc(); err != nil {
			return false, err
		}
	}
	return s.scoreRange(collector, s.scorer, doc, max)
}

func (s *DefaultBulkScorer) scoreRange(collector Collector,
	scorer Scorer, currentDoc, end int) (bool, error) {

	var err error
	for currentDoc < end && err == nil {
		if err = collector.Collect(currentDoc); err == nil {
			currentDoc, err = scorer.NextDoc()
		}
	}
	return currentDoc != NO_MORE_DOCS, err
}

func (s *DefaultBulkScorer) scoreAll(collector Collector, scorer Scorer) (err error) {
	var doc int
	for doc, err = scorer.NextDoc(); doc != NO_MORE_DOCS && err == nil; doc, err = scorer.NextDoc() {
		err = collector.Collect(doc)
	}
	return
}
