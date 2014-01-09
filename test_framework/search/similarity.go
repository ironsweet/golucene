package search

import (
	. "github.com/balzaczyy/golucene/core/search"
	"math/rand"
)

/*
Similarity implementation that randomizes Similarity implementations
per-field.

The choices are 'sticky', so the selected algorithm is ways used for
the same field.
*/
type RandomSimilarityProvider struct {
	*PerFieldSimilarityWrapper
}

func NewRandomSimilarityProvider(r *rand.Rand) *RandomSimilarityProvider {
	panic("not implemented yet")
}

func (rp *RandomSimilarityProvider) QueryNorm(valueForNormalization float32) float32 {
	panic("not implemented yet")
}
