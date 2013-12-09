package search

import (
	"github.com/balzaczyy/golucene/index"
	"github.com/balzaczyy/golucene/util"
)

type Weight interface {
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
	Scorer(ctx index.AtomicReaderContext, inOrder bool, topScorer bool, acceptDocs util.Bits) (sc Scorer, err error)
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
