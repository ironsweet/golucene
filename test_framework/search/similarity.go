package search

import (
	"fmt"
	. "github.com/balzaczyy/golucene/core/search"
	"math/rand"
	"sync"
)

// search/RandomSimilarityProvider.java

var allSims = func() []Similarity {
	ans := make([]Similarity, 1)
	ans = append(ans, NewDefaultSimilarity())
	// ans = append(ans, newBM25Similarity())
	// for _, basicModel := range BASIC_MODELS {
	// 	for _, afterEffect := range AFTER_EFFECTS {
	// 		for _, normalization := range NORMALIZATIONS {
	// 			ans = append(ans, newDFRSimilarity(basicModel, afterEffect, normalization))
	// 		}
	// 	}
	// }
	// for _, distribution := range DISTRIBUTIONS {
	// 	for _, lambda := range LAMBDAS {
	// 		for _, normalization := range NORMALIZATIONS {
	// 			ans = append(ans, newIBSimilarity(ditribution, lambda, normalization))
	// 		}
	// 	}
	// }
	// ans = append(ans, newLMJelinekMercerSimilarity(0.1))
	// ans = append(ans, newLMJelinekMercerSimilarity(0.7))
	return ans
}()

/*
Similarity implementation that randomizes Similarity implementations
per-field.

The choices are 'sticky', so the selected algorithm is ways used for
the same field.
*/
type RandomSimilarityProvider struct {
	*PerFieldSimilarityWrapper
	sync.Locker
	defaultSim       *DefaultSimilarity
	knownSims        []Similarity
	previousMappings map[string]Similarity
	perFieldSeed     int
	coordType        int // 0 = no coord, 1 = coord, 2 = crazy coord
	shouldQueryNorm  bool
}

func NewRandomSimilarityProvider(r *rand.Rand) *RandomSimilarityProvider {
	sims := make([]Similarity, len(allSims))
	for i, v := range r.Perm(len(allSims)) {
		sims[i] = allSims[v]
	}
	return &RandomSimilarityProvider{
		PerFieldSimilarityWrapper: NewPerFieldSimilarityWrapper(func(name string) Similarity {
			panic("not implemented yet")
		}),
		Locker:           &sync.Mutex{},
		defaultSim:       NewDefaultSimilarity(),
		previousMappings: make(map[string]Similarity),
		perFieldSeed:     r.Int(),
		coordType:        r.Intn(3),
		shouldQueryNorm:  r.Intn(2) == 0,
		knownSims:        sims,
	}
}

func (rp *RandomSimilarityProvider) QueryNorm(valueForNormalization float32) float32 {
	panic("not implemented yet")
}

func (rp *RandomSimilarityProvider) String() string {
	rp.Lock() // synchronized
	defer rp.Unlock()
	var coordMethod string
	switch rp.coordType {
	case 0:
		coordMethod = "no"
	case 1:
		coordMethod = "yes"
	default:
		coordMethod = "crazy"
	}
	return fmt.Sprintf("RandomSimilarityProvider(queryNorm=%v,coord=%v): %v",
		rp.shouldQueryNorm, coordMethod, rp.previousMappings)
}
