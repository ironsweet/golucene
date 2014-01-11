package core_test

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/search"
	. "github.com/balzaczyy/golucene/test_framework"
	"github.com/balzaczyy/golucene/test_framework/analysis"
	. "github.com/balzaczyy/golucene/test_framework/util"
	. "github.com/balzaczyy/gounit"
	"os"
	"testing"
)

// Hook up custom test logic into Go's test runner.
func TestBefore(t *testing.T) {
	fmt.Printf("tests.codec: %v\n", os.Getenv("tests.codec"))

	index.DefaultSimilarity = func() index.Similarity {
		return search.NewDefaultSimilarity()
	}
	index.RegisterCodec(index.Lucene45Codec, index.Lucene42Codec)
	// This controls how suite-level rules are nested. It is important
	// that _all_ rules declared in testcase are executed in proper
	// order if they depend on each other.
	ClassRuleChain(ClassEnvRule)

	BeforeSuite(t)
}

func TestNegativeQueryBoost(t *testing.T) {
	Test(t, func(t *T) {
		q := search.NewTermQuery(index.NewTerm("foo", "bar"))
		q.SetBoost(-42)
		t.Assert(-42 == q.Boost())

		directory := NewDirectory()
		defer directory.Close()

		analyzer := analysis.NewMockAnalyzerWithRandom(Random())
		conf := NewIndexWriterConfig(TEST_VERSION_CURRENT, analyzer)

		writer, err := index.NewIndexWriter(directory, conf)
		if err != nil {
			t.Error(err)
		}
		defer writer.Close()

		d := index.NewDocument()
		d.Add(NewTextField("foo", "bar", true))
		writer.AddDocument(d.Fields())
		writer.Close() // ensure index is written

		reader, err := index.OpenDirectoryReader(directory)
		if err != nil {
			t.Error(err)
		}
		defer reader.Close()

		searcher := NewSearcher(reader)
		res, err := searcher.Search(q, nil, 1000)
		if err != nil {
			t.Error(err)
		}
		hits := res.ScoreDocs
		t.Assert(1 == len(hits))
		t.Assert2(hits[0].Score < 0, fmt.Sprintf("score is not negative: %v", hits[0].Score))

		explain, err := searcher.Explain(q, hits[0].Doc)
		if err != nil {
			t.Error(err)
		}
		t.Assert2(isSimilar(hits[0].Score, explain.Value(), 0.01), "score doesn't match explanation")
		t.Assert2(explain.IsMatch(), "explain doesn't think doc is a match")
	})
}

func isSimilar(f1, f2, delta float32) bool {
	diff := f1 - f2
	return diff > 0 && diff < delta || diff < 0 && -diff < delta
}

func TestAfter(t *testing.T) { AfterSuite(t) }
