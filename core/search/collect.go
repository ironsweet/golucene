package search

import (
	"container/heap"
	"fmt"
	"github.com/balzaczyy/golucene/core/index"
	"math"
)

/** Holds one hit in {@link TopDocs}. */
type ScoreDoc struct {
	/** The score of this document for the query. */
	Score float32
	/** A hit document's number.
	 * @see IndexSearcher#doc(int) */
	Doc int
	/** Only set by {@link TopDocs#merge} */
	shardIndex int
}

func newScoreDoc(doc int, score float32) *ScoreDoc {
	return newShardedScoreDoc(doc, score, -1)
}

func newShardedScoreDoc(doc int, score float32, shardIndex int) *ScoreDoc {
	return &ScoreDoc{score, doc, shardIndex}
}

func (d *ScoreDoc) String() string {
	return fmt.Sprintf("doc=%v score=%v shardIndex=%v", d.Doc, d.Score, d.shardIndex)
}

type PriorityQueue struct {
	items []interface{}
	less  func(i, j int) bool
}

func (pq PriorityQueue) Len() int            { return len(pq.items) }
func (pq PriorityQueue) Less(i, j int) bool  { return pq.less(i, j) }
func (pq PriorityQueue) Swap(i, j int)       { pq.items[i], pq.items[j] = pq.items[j], pq.items[i] }
func (pq *PriorityQueue) Push(x interface{}) { pq.items = append(pq.items, x) }
func (pq *PriorityQueue) Pop() interface{} {
	n := pq.Len()
	ans := pq.items[n-1]
	pq.items = pq.items[0 : n-1]
	return ans
}
func (pq *PriorityQueue) updateTop() interface{} {
	heap.Fix(pq, 0)
	return pq.items[0]
}

type TopDocs struct {
	TotalHits int
	ScoreDocs []*ScoreDoc
	maxScore  float64
}

type Collector interface {
	SetScorer(s Scorer)
	Collect(doc int) error
	SetNextReader(ctx *index.AtomicReaderContext)
	AcceptsDocsOutOfOrder() bool
}

// search/TopDocsCollector.java
/**
 * A base class for all collectors that return a {@link TopDocs} output. This
 * collector allows easy extension by providing a single constructor which
 * accepts a {@link PriorityQueue} as well as protected members for that
 * priority queue and a counter of the number of total hits.<br>
 * Extending classes can override any of the methods to provide their own
 * implementation, as well as avoid the use of the priority queue entirely by
 * passing null to {@link #TopDocsCollector(PriorityQueue)}. In that case
 * however, you might want to consider overriding all methods, in order to avoid
 * a NullPointerException.
 */
type TopDocsCollector interface {
	Collector
	/** Returns the top docs that were collected by this collector. */
	TopDocs() TopDocs
	/**
	 * Returns the documents in the rage [start .. start+howMany) that were
	 * collected by this collector. Note that if start >= pq.size(), an empty
	 * TopDocs is returned, and if pq.size() - start &lt; howMany, then only the
	 * available documents in [start .. pq.size()) are returned.<br>
	 * This method is useful to call in case pagination of search results is
	 * allowed by the search application, as well as it attempts to optimize the
	 * memory used by allocating only as much as requested by howMany.<br>
	 * <b>NOTE:</b> you cannot call this method more than once for each search
	 * execution. If you need to call it more than once, passing each time a
	 * different range, you should call {@link #topDocs()} and work with the
	 * returned {@link TopDocs} object, which will contain all the results this
	 * search execution collected.
	 */
	TopDocsRange(start, howMany int) TopDocs
}

type TopDocsCreator interface {
	/**
	 * Populates the results array with the ScoreDoc instances. This can be
	 * overridden in case a different ScoreDoc type should be returned.
	 */
	populateResults(results []*ScoreDoc, howMany int)
	/**
	 * Returns a {@link TopDocs} instance containing the given results. If
	 * <code>results</code> is null it means there are no results to return,
	 * either because there were 0 calls to collect() or because the arguments to
	 * topDocs were invalid.
	 */
	newTopDocs(results []*ScoreDoc, start int) TopDocs
	/** The number of valid PQ entries */
	topDocsSize() int
}

type abstractTopDocsCollector struct {
	Collector
	TopDocsCreator
	pq        *PriorityQueue // PriorityQueue
	TotalHits int
}

func newTopDocsCollector(self interface{}, pq *PriorityQueue) *abstractTopDocsCollector {
	return &abstractTopDocsCollector{
		Collector:      self.(Collector),
		TopDocsCreator: self.(TopDocsCreator),
		pq:             pq,
	}
}

func (c *abstractTopDocsCollector) AcceptsDocsOutOfOrder() bool {
	return false
}

func (c *abstractTopDocsCollector) populateResults(results []*ScoreDoc, howMany int) {
	for i := howMany - 1; i >= 0; i-- {
		results[i] = heap.Pop(c.pq).(*ScoreDoc)
	}
}

func (c *abstractTopDocsCollector) topDocsSize() int {
	// In case pq was populated with sentinel values, there might be less
	// results than pq.size(). Therefore return all results until either
	// pq.size() or totalHits.
	if n := c.pq.Len(); c.TotalHits >= n {
		return n
	}
	return c.TotalHits
}

func (c *abstractTopDocsCollector) TopDocs() TopDocs {
	// In case pq was populated with sentinel values, there might be less
	// results than pq.size(). Therefore return all results until either
	// pq.size() or totalHits.
	return c.TopDocsRange(0, c.topDocsSize())
}

func (c *abstractTopDocsCollector) TopDocsRange(start, howMany int) TopDocs {
	// In case pq was populated with sentinel values, there might be less
	// results than pq.size(). Therefore return all results until either
	// pq.size() or totalHits.
	size := c.topDocsSize()

	// Don't bother to throw an exception, just return an empty TopDocs in case
	// the parameters are invalid or out of range.
	// TODO: shouldn't we throw IAE if apps give bad params here so they dont
	// have sneaky silent bugs?
	if start < 0 || start >= size || howMany <= 0 {
		return c.newTopDocs(nil, start)
	}

	// We know that start < pqsize, so just fix howMany.
	if size-start < howMany {
		howMany = size - start
	}
	results := make([]*ScoreDoc, howMany)

	// pq's pop() returns the 'least' element in the queue, therefore need
	// to discard the first ones, until we reach the requested range.
	// Note that this loop will usually not be executed, since the common usage
	// should be that the caller asks for the last howMany results. However it's
	// needed here for completeness.
	for i := c.pq.Len() - start - howMany; i > 0; i-- {
		heap.Pop(c.pq)
	}

	// Get the requested results from pq.
	c.populateResults(results, howMany)

	return c.newTopDocs(results, start)
}

type TopScoreDocCollector struct {
	*abstractTopDocsCollector
	pqTop   *ScoreDoc
	docBase int
	scorer  Scorer
}

func newTocScoreDocCollector(numHits int) *TopScoreDocCollector {
	docs := make([]interface{}, numHits)
	for i, _ := range docs {
		docs[i] = newScoreDoc(math.MaxInt32, -math.MaxFloat32)
	}
	pq := &PriorityQueue{items: docs}
	pq.less = func(i, j int) bool {
		hitA := pq.items[i].(*ScoreDoc)
		hitB := pq.items[j].(*ScoreDoc)
		if hitA.Score == hitB.Score {
			return hitA.Doc > hitB.Doc
		}
		return hitA.Score < hitB.Score
	}
	heap.Init(pq)

	pqTop := heap.Pop(pq).(*ScoreDoc)
	heap.Push(pq, pqTop)
	c := &TopScoreDocCollector{pqTop: pqTop}
	c.abstractTopDocsCollector = newTopDocsCollector(c, pq)
	return c
}

func (c *TopScoreDocCollector) newTopDocs(results []*ScoreDoc, start int) TopDocs {
	if results == nil {
		return TopDocs{0, []*ScoreDoc{}, math.NaN()}
	}

	// We need to compute maxScore in order to set it in TopDocs. If start == 0,
	// it means the largest element is already in results, use its score as
	// maxScore. Otherwise pop everything else, until the largest element is
	// extracted and use its score as maxScore.
	maxScore := math.NaN()
	if start == 0 {
		maxScore = float64(results[0].Score)
	} else {
		pq := c.pq
		for i := pq.Len(); i > 1; i-- {
			heap.Pop(pq)
		}
		maxScore = float64(heap.Pop(pq).(ScoreDoc).Score)
	}

	return TopDocs{c.TotalHits, results, maxScore}
}

func (c *TopScoreDocCollector) SetNextReader(ctx *index.AtomicReaderContext) {
	c.docBase = ctx.DocBase
}

func (c *TopScoreDocCollector) SetScorer(scorer Scorer) {
	c.scorer = scorer
}

func NewTopScoreDocCollector(numHits int, after *ScoreDoc, docsScoredInOrder bool) TopDocsCollector {
	if numHits < 0 {
		panic("numHits must be > 0; please use TotalHitCountCollector if you just need the total hit count")
	}

	if docsScoredInOrder {
		if after == nil {
			return newInOrderTopScoreDocCollector(numHits)
		}
		panic("not implemented yet")
		// TODO support paging
	} else {
		if after == nil {
			return newOutOfOrderTopScoreDocCollector(numHits)
		}
		panic("not implemented yet")
	}
}

// Assumes docs are scored in order.
type InOrderTopScoreDocCollector struct {
	*TopScoreDocCollector
}

func newInOrderTopScoreDocCollector(numHits int) *InOrderTopScoreDocCollector {
	return &InOrderTopScoreDocCollector{newTocScoreDocCollector(numHits)}
}

func (c *InOrderTopScoreDocCollector) Collect(doc int) (err error) {
	score, err := c.scorer.Score()
	if err != nil {
		return err
	}

	// This collector cannot handle these scores:
	assert(score != -math.MaxFloat32)
	assert(!math.IsNaN(float64(score)))

	c.TotalHits++
	if score <= c.pqTop.Score {
		// Since docs are returned in-order (i.e., increasing doc Id), a document
		// with equal score to pqTop.score cannot compete since HitQueue favors
		// documents with lower doc Ids. Therefore reject those docs too.
		return
	}
	c.pqTop.Doc = doc + c.docBase
	c.pqTop.Score = float32(score)
	c.pqTop = c.pq.updateTop().(*ScoreDoc)
	return
}

func (c *InOrderTopScoreDocCollector) AcceptsDocsOutOfOrder() bool {
	return false
}

type OutOfOrderTopScoreDocCollector struct {
	*TopScoreDocCollector
}

func newOutOfOrderTopScoreDocCollector(numHits int) *OutOfOrderTopScoreDocCollector {
	return &OutOfOrderTopScoreDocCollector{
		TopScoreDocCollector: newTocScoreDocCollector(numHits),
	}
}

func (c *OutOfOrderTopScoreDocCollector) Collect(doc int) (err error) {
	var score float32
	if score, err = c.scorer.Score(); err != nil {
		return err
	}

	// This collector cannot handle NaN
	assert(!math.IsNaN(float64(score)))

	c.TotalHits++
	if score < c.pqTop.Score {
		// Doesn't compete w/ bottom entry in queue
		return nil
	}
	doc += c.docBase
	if score == c.pqTop.Score && doc > c.pqTop.Doc {
		// Break tie in score by doc ID:
		return nil
	}
	c.pqTop.Doc = doc
	c.pqTop.Score = score
	c.pqTop = c.pq.updateTop().(*ScoreDoc)
	return nil
}

func (c *OutOfOrderTopScoreDocCollector) AcceptsDocsOutOfOrder() bool {
	return true
}
