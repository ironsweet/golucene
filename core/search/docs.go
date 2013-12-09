package search

import (
	"github.com/balzaczyy/golucene/index"
)

// search/Scorer.java

type Scorer interface {
	index.DocsEnum
	IScorer
	ScoreAndCollect(c Collector) error
}

type IScorer interface {
	Score() (value float64, err error)
}

type abstractScorer struct {
	index.DocsEnum
	IScorer
	/** the Scorer's parent Weight. in some cases this may be null */
	// TODO can we clean this up?
	weight Weight
}

func newScorer(self interface{}, w Weight) *abstractScorer {
	return &abstractScorer{self.(index.DocsEnum), self.(IScorer), w}
}

/** Scores and collects all matching documents.
 * @param collector The collector to which all matching documents are passed.
 */
func (s *abstractScorer) ScoreAndCollect(c Collector) (err error) {
	assert(s.DocId() == -1) // not started
	c.SetScorer(s)
	doc, err := s.NextDoc()
	for doc != index.NO_MORE_DOCS && err == nil {
		c.Collect(doc)
		doc, err = s.NextDoc()
	}
	return
}

func assert(ok bool) {
	if !ok {
		panic("assert fail")
	}
}
