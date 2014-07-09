package model

import (
	"math"
)

const NO_MORE_DOCS = math.MaxInt32

type DocIdSetIterator interface {
	/**
	 * Returns the following:
	 * <ul>
	 * <li>-1 or {@link #NO_MORE_DOCS} if {@link #nextDoc()} or
	 * {@link #advance(int)} were not called yet.
	 * <li>{@link #NO_MORE_DOCS} if the iterator has exhausted.
	 * <li>Otherwise it should return the doc ID it is currently on.
	 * </ul>
	 * <p>
	 *
	 * @since 2.9
	 */
	DocId() int
	/**
	 * Advances to the next document in the set and returns the doc it is
	 * currently on, or {@link #NO_MORE_DOCS} if there are no more docs in the
	 * set.<br>
	 *
	 * <b>NOTE:</b> after the iterator has exhausted you should not call this
	 * method, as it may result in unpredicted behavior.
	 *
	 * @since 2.9
	 */
	NextDoc() (doc int, err error)
	/**
	 * Advances to the first beyond the current whose document number is greater
	 * than or equal to <i>target</i>, and returns the document number itself.
	 * Exhausts the iterator and returns {@link #NO_MORE_DOCS} if <i>target</i>
	 * is greater than the highest document number in the set.
	 * <p>
	 * The behavior of this method is <b>undefined</b> when called with
	 * <code> target &le; current</code>, or after the iterator has exhausted.
	 * Both cases may result in unpredicted behavior.
	 * <p>
	 * When <code> target &gt; current</code> it behaves as if written:
	 *
	 * <pre class="prettyprint">
	 * int advance(int target) {
	 *   int doc;
	 *   while ((doc = nextDoc()) &lt; target) {
	 *   }
	 *   return doc;
	 * }
	 * </pre>
	 *
	 * Some implementations are considerably more efficient than that.
	 * <p>
	 * <b>NOTE:</b> this method may be called with {@link #NO_MORE_DOCS} for
	 * efficiency by some Scorers. If your implementation cannot efficiently
	 * determine that it should exhaust, it is recommended that you check for that
	 * value in each call to this method.
	 * <p>
	 *
	 * @since 2.9
	 */
	Advance(target int) (doc int, err error)
	/**
	 * Returns the estimated cost of this {@link DocIdSetIterator}.
	 * <p>
	 * This is generally an upper bound of the number of documents this iterator
	 * might match, but may be a rough heuristic, hardcoded value, or otherwise
	 * completely inaccurate.
	 */
	// Cost() int64
}
