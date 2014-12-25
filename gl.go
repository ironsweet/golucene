package main

import (
	"fmt"
	std "github.com/balzaczyy/golucene/analysis/standard"
	_ "github.com/balzaczyy/golucene/core/codec/lucene410"
	"github.com/balzaczyy/golucene/core/document"
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/search"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"os"
)

func main() {
	util.SetDefaultInfoStream(util.NewPrintStreamInfoStream(os.Stdout))
	index.DefaultSimilarity = func() index.Similarity {
		return search.NewDefaultSimilarity()
	}

	directory, _ := store.OpenFSDirectory("test_index")
	analyzer := std.NewStandardAnalyzer()
	conf := index.NewIndexWriterConfig(util.VERSION_LATEST, analyzer)
	writer, _ := index.NewIndexWriter(directory, conf)

	d := document.NewDocument()
	d.Add(document.NewTextFieldFromString("foo", "bar", document.STORE_YES))
	writer.AddDocument(d.Fields())
	writer.Close() // ensure index is written

	reader, _ := index.OpenDirectoryReader(directory)
	searcher := search.NewIndexSearcher(reader)

	q := search.NewTermQuery(index.NewTerm("foo", "bar"))
	res, _ := searcher.Search(q, nil, 1000)
	fmt.Printf("Found %v hit(s).\n", res.TotalHits)
	for _, hit := range res.ScoreDocs {
		fmt.Printf("Doc %v score: %v\n", hit.Doc, hit.Score)
		doc, _ := reader.Document(hit.Doc)
		fmt.Printf("foo -> %v\n", doc.Get("foo"))
	}

}
