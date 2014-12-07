package core_test

import (
	"fmt"
	std "github.com/balzaczyy/golucene/analysis/standard"
	_ "github.com/balzaczyy/golucene/core/codec/lucene410"
	docu "github.com/balzaczyy/golucene/core/document"
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/search"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	// . "github.com/balzaczyy/golucene/test_framework"
	// "github.com/balzaczyy/golucene/test_framework/analysis"
	// . "github.com/balzaczyy/golucene/test_framework/util"
	. "github.com/balzaczyy/gounit"
	"os"
	"testing"
)

// Hook up custom test logic into Go's test runner.
func TestBefore(t *testing.T) {
	fmt.Printf("tests_codec: %v\n", os.Getenv("tests_codec"))

	// util.SetDefaultInfoStream(util.NewPrintStreamInfoStream(os.Stdout))
	index.DefaultSimilarity = func() index.Similarity {
		return search.NewDefaultSimilarity()
	}
	// This controls how suite-level rules are nested. It is important
	// that _all_ rules declared in testcase are executed in proper
	// order if they depend on each other.
	// ClassRuleChain(ClassEnvRule)

	// BeforeSuite(t)
}

func TestBasicIndexAndSearch(t *testing.T) {
	q := search.NewTermQuery(index.NewTerm("foo", "bar"))
	q.SetBoost(-42)

	os.RemoveAll(".gltest")

	directory, err := store.OpenFSDirectory(".gltest")
	It(t).Should("has no error: %v", err).Assert(err == nil)
	It(t).Should("has valid directory").Assert(directory != nil)
	fmt.Println("Directory", directory)
	defer directory.Close()

	analyzer := std.NewStandardAnalyzer()
	conf := index.NewIndexWriterConfig(util.VERSION_LATEST, analyzer)

	writer, err := index.NewIndexWriter(directory, conf)
	It(t).Should("has no error: %v", err).Assert(err == nil)

	d := docu.NewDocument()
	d.Add(docu.NewTextFieldFromString("foo", "bar", docu.STORE_YES))
	err = writer.AddDocument(d.Fields())
	It(t).Should("has no error: %v", err).Assert(err == nil)
	err = writer.Close() // ensure index is written
	It(t).Should("has no error: %v", err).Assert(err == nil)

	reader, err := index.OpenDirectoryReader(directory)
	It(t).Should("has no error: %v", err).Assert(err == nil)
	defer reader.Close()

	searcher := search.NewIndexSearcher(reader)
	res, err := searcher.Search(q, nil, 1000)
	It(t).Should("has no error: %v", err).Assert(err == nil)
	hits := res.ScoreDocs
	It(t).Should("expect 1 hits, but %v only.", len(hits)).Assert(len(hits) == 1)
	It(t).Should("expect score to be negative (got %v)", hits[0].Score).Verify(hits[0].Score < 0)

	explain, err := searcher.Explain(q, hits[0].Doc)
	It(t).Should("has no error: %v", err).Assert(err == nil)
	It(t).Should("score doesn't match explanation (%v vs %v)", hits[0].Score, explain.Value()).Verify(isSimilar(hits[0].Score, explain.Value(), 0.001))
	It(t).Should("explain doesn't think doc is a match").Verify(explain.IsMatch())
}

// func TestNegativeQueryBoost(t *testing.T) {
// 	Test(t, func(t *T) {
// 		q := search.NewTermQuery(index.NewTerm("foo", "bar"))
// 		q.SetBoost(-42)
// 		t.Assert(-42 == q.Boost())

// 		directory := NewDirectory()
// 		defer directory.Close()

// 		analyzer := analysis.NewMockAnalyzerWithRandom(Random())
// 		conf := NewIndexWriterConfig(TEST_VERSION_CURRENT, analyzer)

// 		writer, err := index.NewIndexWriter(directory, conf)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 		defer writer.Close()

// 		d := docu.NewDocument()
// 		d.Add(NewTextField("foo", "bar", true))
// 		writer.AddDocument(d.Fields())
// 		writer.Close() // ensure index is written

// 		reader, err := index.OpenDirectoryReader(directory)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 		defer reader.Close()

// 		searcher := NewSearcher(reader)
// 		res, err := searcher.Search(q, nil, 1000)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 		hits := res.ScoreDocs
// 		t.Assert(1 == len(hits))
// 		t.Assert2(hits[0].Score < 0, fmt.Sprintf("score is not negative: %v", hits[0].Score))

// 		explain, err := searcher.Explain(q, hits[0].Doc)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 		t.Assert2(isSimilar(hits[0].Score, explain.Value(), 0.001), "score doesn't match explanation")
// 		t.Assert2(explain.IsMatch(), "explain doesn't think doc is a match")
// 	})
// }

func isSimilar(f1, f2, delta float32) bool {
	diff := f1 - f2
	return diff >= 0 && diff < delta || diff < 0 && -diff < delta
}

func TestAfter(t *testing.T) {
	// AfterSuite(t)
}
