package main

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/search"
	. "github.com/balzaczyy/golucene/test_framework"
	"github.com/balzaczyy/golucene/test_framework/analysis"
	. "github.com/balzaczyy/golucene/test_framework/util"
	. "github.com/balzaczyy/gounit"
	"os"
)

func main() {
	fmt.Printf("tests_codec: %v\n", os.Getenv("tests_codec"))

	index.DefaultSimilarity = func() index.Similarity {
		return search.NewDefaultSimilarity()
	}
	// This controls how suite-level rules are nested. It is important
	// that _all_ rules declared in testcase are executed in proper
	// order if they depend on each other.
	ClassRuleChain(ClassEnvRule)

	BeforeSuite(nil)

	Test(nil, func(t *T) {
		q := search.NewTermQuery(index.NewTerm("foo", "bar"))
		q.SetBoost(-42)
		t.Assert(-42 == q.Boost())

		directory := NewDirectory()
		fmt.Println("Directory", directory)
		defer directory.Close()

		analyzer := analysis.NewMockAnalyzerWithRandom(Random())
		conf := NewIndexWriterConfig(TEST_VERSION_CURRENT, analyzer)

		writer, err := index.NewIndexWriter(directory, conf)
		t.Assert2(err == nil, "%v", err)
		t.Assert(writer != nil)

		d := index.NewDocument()
		d.Add(NewTextField("foo", "bar", true))
		err = writer.AddDocument(d.Fields())
		t.Assert2(err == nil, "%v", err)
		err = writer.Close() // ensure index is written
		t.Assert2(err == nil, "%v", err)

		reader, err := index.OpenDirectoryReader(directory)
		t.Assert2(err == nil, "%v", err)
		defer reader.Close()

		searcher := NewSearcher(reader)
		res, err := searcher.Search(q, nil, 1000)
		t.Assert2(err == nil, "%v", err)
		if err == nil {
			hits := res.ScoreDocs
			t.Assert2(1 == len(hits), "no hit")
			if len(hits) > 0 {
				t.Assert2(hits[0].Score < 0, fmt.Sprintf("score is not negative: %v", hits[0].Score))

				explain, err := searcher.Explain(q, hits[0].Doc)
				t.Assert2(err == nil, "%v", err)
				t.Assert2(isSimilar(hits[0].Score, explain.Value(), 0.01), "score doesn't match explanation")
				t.Assert2(explain.IsMatch(), "explain doesn't think doc is a match")
			}
		}
	})

	AfterSuite(nil)
}

func isSimilar(f1, f2, delta float32) bool {
	diff := f1 - f2
	return diff > 0 && diff < delta || diff < 0 && -diff < delta
}
