package search

import (
	"github.com/balzaczyy/golucene/index"
	"github.com/balzaczyy/golucene/util"
)

type Weight interface {
	ValueForNormalization() float32
	Normalize(norm float64, topLevelBoost float32) float32
	IsScoresDocsOutOfOrder() bool // usually false
	Scorer(ctx index.AtomicReaderContext, inOrder bool, topScorer bool, acceptDocs util.Bits) (sc Scorer, ok bool)
}
