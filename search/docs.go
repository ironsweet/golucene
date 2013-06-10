package search

type DocIdSetIterator interface {
	DocId() int
	Freq() int
	NextDoc() int
}

type DocsEnum struct {
	DocIdSetIterator
}

type Scorer struct {
	*DocsEnum
	weight Weight
	Score  func() float64
}

func (s *Scorer) ScoreAndCollect(c Collector) {
	// assert docID() == -1; // not started
	c.SetScorer(s)
	for {
		doc, more := s.DocsEnum.DocIdSetIterator.NextDoc()
		if !more {
			break
		}
		c.Collect(doc)
	}
}
