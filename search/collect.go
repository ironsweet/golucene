package search

import (
	"container/heap"
	"lucene/index"
	"math"
)

type ScoreDoc struct {
	score float64
	doc   int
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

type TopDocs struct {
	totalHits int
	scoreDocs []ScoreDoc
	maxScore  float64
}

type Collector interface {
	SetScorer(s Scorer)
	Collect(doc int)
	SetNextReader(ctx index.AtomicReaderContext)
	AcceptsDocsOutOfOrder() bool
}

type TopDocsCollector struct {
	Collector
	self      interface{}
	pq        *PriorityQueue // PriorityQueue
	TotalHits int
	// populateResults func(results []ScoreDoc, howMany int)
	// newTopDocs      func(results []ScoreDoc, start int) TopDocs
	// topDocsSize     func() int
	// TopDocs         func() TopDocs
}

func newTopDocsCollector(self interface{}, pq *PriorityQueue) *TopDocsCollector {
	return &TopDocsCollector{self: self, pq: pq}
}

func (c *TopDocsCollector) populateResults(results []ScoreDoc, howMany int) {
	for i := howMany - 1; i >= 0; i-- {
		results[i] = *(heap.Pop(c.pq).(*ScoreDoc))
	}
}

func (c *TopDocsCollector) newTopDocs(results []ScoreDoc, start int) TopDocs {
	if results == nil {
		return TopDocs{0, []ScoreDoc{}, math.NaN()}
	}
	return TopDocs{c.TotalHits, results, math.NaN()}
}

func (c *TopDocsCollector) topDocsSize() int {
	// In case pq was populated with sentinel values, there might be less
	// results than pq.size(). Therefore return all results until either
	// pq.size() or totalHits.
	if n := c.pq.Len(); c.TotalHits >= n {
		return n
	}
	return c.TotalHits
}

func (c *TopDocsCollector) TopDocs() TopDocs {
	// In case pq was populated with sentinel values, there might be less
	// results than pq.size(). Therefore return all results until either
	// pq.size() or totalHits.
	return c.TopDocsRange(0, c.topDocsSize())
}

func (c *TopDocsCollector) TopDocsRange(start, howMany int) TopDocs {
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
	results := make([]ScoreDoc, howMany)

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
	*TopDocsCollector
	pqTop   *ScoreDoc
	docBase int
	scorer  Scorer
}

func newTocScoreDocCollector(numHits int) *TopScoreDocCollector {
	docs := make([]interface{}, numHits)
	for i, _ := range docs {
		docs[i] = ScoreDoc{-math.MaxFloat32, math.MaxInt32}
	}
	pq := &PriorityQueue{items: docs}
	pq.less = func(i, j int) bool {
		hitA := pq.items[i].(*ScoreDoc)
		hitB := pq.items[j].(*ScoreDoc)
		if hitA.score == hitB.score {
			return hitA.doc > hitB.doc
		}
		return hitA.score < hitB.score
	}
	heap.Init(pq)

	pqTop := heap.Pop(pq).(*ScoreDoc)
	heap.Push(pq, pqTop)
	c := &TopScoreDocCollector{pqTop: pqTop}
	c.TopDocsCollector = newTopDocsCollector(c, pq)
	return c
}

func (c *TopScoreDocCollector) newTopDocs(results []ScoreDoc, start int) TopDocs {
	if results == nil {
		return TopDocs{0, []ScoreDoc{}, math.NaN()}
	}

	// We need to compute maxScore in order to set it in TopDocs. If start == 0,
	// it means the largest element is already in results, use its score as
	// maxScore. Otherwise pop everything else, until the largest element is
	// extracted and use its score as maxScore.
	maxScore := math.NaN()
	if start == 0 {
		maxScore = results[0].score
	} else {
		pq := c.TopDocsCollector.pq
		for i := pq.Len(); i > 1; i-- {
			heap.Pop(pq)
		}
		maxScore = heap.Pop(pq).(*ScoreDoc).score
	}

	return TopDocs{c.TopDocsCollector.TotalHits, results, maxScore}
}

func (c *TopScoreDocCollector) SetNextReader(ctx index.AtomicReaderContext) {
	c.docBase = ctx.DocBase
}

func NewTopScoreDocCollector(numHits int, after ScoreDoc, docsScoredInOrder bool) *TopDocsCollector {
	if numHits < 0 {
		panic("numHits must be > 0; please use TotalHitCountCollector if you just need the total hit count")
	}

	if docsScoredInOrder {
		return NewInOrderTopScoreDocCollector(numHits).TopScoreDocCollector.TopDocsCollector.Collector.(*TopDocsCollector)
		// TODO support paging
	} else {
		panic("not supported yet")
	}
}

type InOrderTopScoreDocCollector struct {
	*TopScoreDocCollector
}

func NewInOrderTopScoreDocCollector(numHits int) *InOrderTopScoreDocCollector {
	return &InOrderTopScoreDocCollector{newTocScoreDocCollector(numHits)}
}

func (c *InOrderTopScoreDocCollector) Collect(doc int) {
	score := c.scorer.Score()

	// This collector cannot handle these scores:
	// assert score != -math.MaxFloat64
	// assert !math.IsNaN(score)

	c.TotalHits++
	if score <= c.pqTop.score {
		// Since docs are returned in-order (i.e., increasing doc Id), a document
		// with equal score to pqTop.score cannot compete since HitQueue favors
		// documents with lower doc Ids. Therefore reject those docs too.
		return
	}
	c.pqTop.doc = doc + c.docBase
	c.pqTop.score = score
	heap.Pop(c.pq)
	heap.Push(c.pq, c.pqTop)
}

func (c *InOrderTopScoreDocCollector) AcceptsDocsOutOfOrder() bool {
	return false
}
