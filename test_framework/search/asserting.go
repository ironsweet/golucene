package search

import (
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/search"
	"math/rand"
)

// search/AssertingIndexSearcher.java

/*
Helper class that adds some extra checks to ensure correct usage of
IndexSearcher and Weight.
*/
type AssertingIndexSearcher struct {
	*search.IndexSearcher
	random *rand.Rand
}

func NewAssertingIndexSearcher(random *rand.Rand, r index.IndexReader) *AssertingIndexSearcher {
	return &AssertingIndexSearcher{
		search.NewIndexSearcher(r),
		rand.New(rand.NewSource(random.Int63())),
	}
}

func NewAssertingIndexSearcherFromContext(random *rand.Rand, ctx index.IndexReaderContext) *AssertingIndexSearcher {
	panic("not implemented yet")
}
