package search

import (
	"github.com/balzaczyy/golucene/index"
)

type Scorer struct {
	self   interface{} // must be index.DocsEnum
	weight Weight
	Score  func() float64
}

func newScorer(self interface{}, w Weight, score func() float64) Scorer {
	return Scorer{self, w, score}
}

func (s *Scorer) ScoreAndCollect(c Collector) {
	// assert docID() == -1; // not started
	c.SetScorer(*s)
	for {
		doc, more := s.self.(index.DocIdSetIterator).NextDoc()
		if !more {
			break
		}
		c.Collect(doc)
	}
}
