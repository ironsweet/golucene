package search

import (
	"fmt"
	. "github.com/balzaczyy/golucene/core/index/model"
	// . "github.com/balzaczyy/golucene/core/search/model"
	"math"
)

// search/Scorer.java

type Scorer interface {
	DocsEnum
	IScorer
	// ScoreAndCollect(c Collector) error
}

type IScorer interface {
	Score() (float32, error)
}

type ScorerSPI interface {
	DocId() int
	NextDoc() (int, error)
}

type abstractScorer struct {
	spi ScorerSPI
	// index.DocsEnum
	// IScorer
	/** the Scorer's parent Weight. in some cases this may be null */
	// TODO can we clean this up?
	weight Weight
}

func newScorer(spi ScorerSPI, w Weight) *abstractScorer {
	return &abstractScorer{spi: spi, weight: w}
}

/** Scores and collects all matching documents.
 * @param collector The collector to which all matching documents are passed.
 */
// func (s *abstractScorer) ScoreAndCollect(c Collector) (err error) {
// 	assert(s.spi.DocId() == -1) // not started
// 	c.SetScorer(s.spi.(Scorer))
// 	doc, err := s.spi.NextDoc()
// 	for doc != NO_MORE_DOCS && err == nil {
// 		c.Collect(doc)
// 		doc, err = s.spi.NextDoc()
// 	}
// 	return
// }

func assert(ok bool) {
	if !ok {
		panic("assert fail")
	}
}

func assert2(ok bool, msg string, args ...interface{}) {
	if !ok {
		panic(fmt.Sprintf(msg, args...))
	}
}

// search/BulkScorer.java

/*
This class is used to score a range of documents at once, and is
returned by Weight.BulkScorer(). Only queries that have a more
optimized means of scoring across a range of documents need to
override this. Otherwise, a default implementation is wrapped around
the Scorer returned by Weight.Scorer().
*/
type BulkScorer interface {
	ScoreAndCollect(Collector) error
	BulkScorerImplSPI
}

type BulkScorerImplSPI interface {
	ScoreAndCollectUpto(Collector, int) (bool, error)
}

type BulkScorerImpl struct {
	spi BulkScorerImplSPI
}

func newBulkScorer(spi BulkScorerImplSPI) *BulkScorerImpl {
	return &BulkScorerImpl{spi}
}

func (bs *BulkScorerImpl) ScoreAndCollect(collector Collector) (err error) {
	assert(bs != nil)
	assert(bs.spi != nil)
	_, err = bs.spi.ScoreAndCollectUpto(collector, math.MaxInt32)
	return
}
