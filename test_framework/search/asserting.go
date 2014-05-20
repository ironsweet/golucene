package search

import (
	"fmt"
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
	return &AssertingIndexSearcher{
		search.NewIndexSearcherFromContext(ctx),
		rand.New(rand.NewSource(random.Int63())),
	}
}

func (ss *AssertingIndexSearcher) CreateNormalizedWeight(query search.Query) (search.Weight, error) {
	panic("not implemented yet")
}

func (ss *AssertingIndexSearcher) Rewrite(original search.Query) (search.Query, error) {
	panic("not implemented yet")
}

func (ss *AssertingIndexSearcher) WrapFilter(original search.Query, filter search.Filter) search.Query {
	panic("not implemented yet")
}

func (ss *AssertingIndexSearcher) SearchLWC(leaves []*index.AtomicReaderContext,
	weight search.Weight, collector search.Collector) error {
	panic("not implemented yet")
}

func (ss *AssertingIndexSearcher) String() string {
	return fmt.Sprintf("AssertingIndexSearcher(%v)", ss.IndexSearcher)
}
