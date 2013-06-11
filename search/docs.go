package search

import (
	"lucene/index"
)

type DocIdSetIterator interface {
	DocId() int
	Freq() int
	NextDoc() (doc int, more bool)
}

type Scorer struct {
	*index.DocsEnum
	weight Weight
	Score  func() float64
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
