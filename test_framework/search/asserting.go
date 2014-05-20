package search

import (
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/search"
	"math/rand"
	"reflect"
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
	random2 := rand.New(rand.NewSource(random.Int63()))
	return &AssertingIndexSearcher{rewrite(random2, search.NewIndexSearcher(r)), random2}
}

func NewAssertingIndexSearcherFromContext(random *rand.Rand, ctx index.IndexReaderContext) *AssertingIndexSearcher {
	random2 := rand.New(rand.NewSource(random.Int63()))
	return &AssertingIndexSearcher{rewrite(random2, search.NewIndexSearcherFromContext(ctx)), random2}
}

func rewrite(random *rand.Rand, ss *search.IndexSearcher) *search.IndexSearcher {
	v := reflect.ValueOf(ss)
	m1 := v.MethodByName("createNormalizedWeight")
	m1.Set(reflect.MakeFunc(m1.Type(), func(in []reflect.Value) []reflect.Value {
		panic("not implemented yet")
	}))
	m2 := v.MethodByName("rewrite")
	m2.Set(reflect.MakeFunc(m2.Type(), func(in []reflect.Value) []reflect.Value {
		panic("not implemented yet")
	}))
	m3 := v.MethodByName("searchLWSI")
	m3.Set(reflect.MakeFunc(m3.Type(), func(in []reflect.Value) []reflect.Value {
		panic("not implemented yet")
	}))
	return ss
}
