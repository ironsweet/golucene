package test_framework

import (
	"github.com/balzaczyy/golucene/core/analysis"
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/search"
	"github.com/balzaczyy/golucene/core/util"
	"github.com/balzaczyy/golucene/test_framework/store"
	"math/rand"
	"time"
)

// -----------------------------------------------------------------
// Test facilities and facades for subclasses.
// -----------------------------------------------------------------

// -----------------------------------------------------------------
// Truly immutable fields and constants, initialized once and valid
// for all suites ever since.
// -----------------------------------------------------------------

// Use this constant then creating Analyzers and any other version-dependent
// stuff. NOTE: Change this when developmenet starts for new Lucene version:
const TEST_VERSION_CURRENT = util.VERSION_45

// L500
/*
Note it's different from Lucene's Randomized Test Runner.

There is an overhead connected with getting the Random for a particular
context and thread. It is better to cache this Random locally if tight loops
with multiple invocations are present or create a derivative local Random for
millions of calls like this:

		r := rand.New(rand.NewSource(99))
		// tight loop with many invocations.

*/
func Random() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
}

// 500
// Create a new index writer config with random defaults
func NewIndexWriterConfig(v util.Version, a analysis.Analyzer) *index.IndexWriterConfig {
	panic("not implemented yet")
}

// L900
func NewDirectory() *store.BaseDirectoryWrapper {
	panic("not implemented yet")
}

// L1064
func NewTextField(name, value string, stored bool) *index.Field {
	flag := index.TEXT_FIELD_TYPE_STORED
	if !stored {
		flag = index.TEXT_FIELD_TYPE_NOT_STORED
	}
	return NewField(Random(), name, value, flag)
}

func NewField(r *rand.Rand, name, value string, typ *index.FieldType) *index.Field {
	panic("not implemented yet")
}

// L1305
// Create a new searcher over the reader. This searcher might randomly use threads
func NewSearcher(r index.IndexReader) *search.IndexSearcher {
	panic("not implemented yet")
}
