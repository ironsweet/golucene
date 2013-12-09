package main

import (
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/search"
	test "github.com/balzaczyy/golucene/test_framework"
	"testing"
)

func TestNegativeQueryBoost(t *testing.T) {
	q := search.NewTermQuery(index.NewTerm("foo", "bar"))
	q.SetBoost(-42)
	assert(-42 == q.Boost())

	directory := test.NewDirectory()
	defer directory.Close()

	analyzer := newMockAnalyzer(random())
	conf := newIndexWriterConfig(test.TEST_VERSION_CURRENT, analyzer)

	writer := newIndexWriter(directory, conf)
	defer writer.Close()

	d := index.NewDocument()
	d.Add(test.NewTextField("foo", "bar", index.FIELD_STORE_YES))
	writer.AddDocument(d)
	writer.Close() // ensure index is written

	reader, err := index.OpenDirectoryReader(directory)
	if err != nil {
		t.Errorf(err)
	}
	defer reader.Close()

	searcher := test.NewSearcher(reader)
	hits := searcher.Search(q, nil, 1000).ScoreDocs
	assert(1 == len(hits))
	assert2(hits[0].Score < 0, fmt.Sprintf("score is not negative: %v", hits[0].Score))

	explain := searcher.Explain(q, hits[0].Doc)
	assert2(isSimilar(hits[0].Score, explain.Value(), 0.01), "score doesn't match explanation")
	assert2(explain.IsMatch(), "explain doesn't think doc is a match")
}

func assert(ok bool) {
	if !ok {
		panic("assert fail")
	}
}

func isSimilar(f1, f2, delta float32) {
	diff := f1 - f2
	return diff > 0 && diff < delta || diff < 0 && -diff < delta
}
