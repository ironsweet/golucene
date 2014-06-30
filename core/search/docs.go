package search

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/index"
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
func (s *abstractScorer) ScoreAndCollect(c Collector) (err error) {
	assert(s.spi.DocId() == -1) // not started
	c.SetScorer(s.spi.(Scorer))
	doc, err := s.spi.NextDoc()
	for doc != index.NO_MORE_DOCS && err == nil {
		c.Collect(doc)
		doc, err = s.spi.NextDoc()
	}
	return
}

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
