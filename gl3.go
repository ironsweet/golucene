package main

import (
	"fmt"
	std "github.com/balzaczyy/golucene/analysis/standard"
	docu "github.com/balzaczyy/golucene/core/document"
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/search"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
)

func main() {
	index.DefaultSimilarity = func() index.Similarity {
		return search.NewDefaultSimilarity()
	}

	q := search.NewTermQuery(index.NewTerm("foo", "bar"))
	q.SetBoost(-42)
	assert(q.Boost() == -42)

	directory, err := store.OpenFSDirectory(".gltest")
	assert(err == nil)
	assert(directory != nil)
	fmt.Println("Directory", directory)
	defer directory.Close()

	analyzer := std.NewStandardAnalyzer(util.VERSION_45)
	conf := index.NewIndexWriterConfig(util.VERSION_45, analyzer)

	writer, err := index.NewIndexWriter(directory, conf)
	assert2(err == nil, "%v", err)

	d := docu.NewDocument()
	d.Add(docu.NewTextField("foo", "bar", docu.STORE_NO))
	err = writer.AddDocument(d.Fields())
	assert(err == nil)
	err = writer.Close() // ensure index is written
	assert(err == nil)

	reader, err := index.OpenDirectoryReader(directory)
	assert(err == nil)
	defer reader.Close()

	searcher := search.NewIndexSearcher(reader)
	res, err := searcher.Search(q, nil, 1000)
	assert(err == nil)
	hits := res.ScoreDocs
	assert2(len(hits) == 1, "Expect 1 hits, but %v only.", len(hits))
	assert2(hits[0].Score < 0, "score is not negative: %v", hits[0].Score)

	explain, err := searcher.Explain(q, hits[0].Doc)
	assert(err == nil)
	assert2(isSimilar(hits[0].Score, explain.Value(), 0.01), "score doesn't match explanation")
	assert2(explain.IsMatch(), "explain doesn't think doc is a match")
}

func isSimilar(f1, f2, delta float32) bool {
	diff := f1 - f2
	return diff > 0 && diff < delta || diff < 0 && -diff < delta
}

func assert(ok bool) {
	if !ok {
		panic("assert fail")
	}
}

func assert2(ok bool, msg string, args ...interface{}) {
	if !ok {
		panic(fmt.Sprintf(msg, args...))
	}
}
