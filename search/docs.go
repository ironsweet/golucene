package search

import (
	"github.com/balzaczyy/golucene/index"
)

// search/Scorer.java

type Scorer struct {
	index.IDocsEnum
	IScorer
	/** the Scorer's parent Weight. in some cases this may be null */
	// TODO can we clean this up?
	weight Weight
}

type IScorer interface {
	Score() float32
}

func newScorer(self interface{}, w Weight) *Scorer {
	return &Scorer{self.(index.IDocsEnum), self.(IScorer), w}
}

/** Scores and collects all matching documents.
 * @param collector The collector to which all matching documents are passed.
 */
func (s *Scorer) ScoreAndCollect(c Collector) (err error) {
	assert(s.DocId() == -1) // not started
	c.SetScorer(s)
	doc := 0
	for err == nil {
		doc, err = s.NextDoc()
		if doc == index.NO_MORE_DOCS {
			break
		} else {
			c.Collect(doc)
		}
	}
	return
}

func assert(ok bool) {
	if !ok {
		panic("assert fail")
	}
}
