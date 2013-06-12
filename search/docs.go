package search

import (
	"lucene/index"
)

type Scorer struct {
	*index.DocsEnum
	self   interface{}
	weight Weight
	Score  func() float64
}

func newScorer(self interface{}, docsEnum *index.DocsEnum, w Weight, score func() float64) *Scorer {
	return &Scorer{docsEnum, self, w, score}
}

func (s *Scorer) ScoreAndCollect(c Collector) {
	// assert docID() == -1; // not started
	c.SetScorer(*s)
	for {
		doc, more := s.DocsEnum.DocIdSetIterator.NextDoc()
		if !more {
			break
		}
		c.Collect(doc)
	}
}
