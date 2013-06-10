package search

import (
	"lucene/index"
	"lucene/util"
)

type Weight interface {
	ValueForNormalization() float32
	Normalize(norm, topLevelBoost float32) float32
	IsScoresDocsOutOfOrder() bool // usually false
	Scorer(ctx index.AtomicReaderContext, inOrder bool, topScorer bool, acceptDocs util.Bits) (sc Scorer, ok bool)
}
