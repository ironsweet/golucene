package main

import (
	"fmt"
	_ "github.com/balzaczyy/golucene/core/codec/lucene410"
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/search"
	"github.com/balzaczyy/golucene/core/store"
	"log"
)

func main() {
	log.Print("Oepning FSDirectory...")
	path := "core/search/testdata/win8/belfrysample"
	d, err := store.OpenFSDirectory(path)
	if err != nil {
		panic(err)
	}
	log.Print("Opening DirectoryReader...")
	r, err := index.OpenDirectoryReader(d)
	if err != nil {
		panic(err)
	}
	if r == nil {
		panic("DirectoryReader cannot be opened.")
	}
	if len(r.Leaves()) < 1 {
		panic("Should have one leaf.")
	}
	for _, ctx := range r.Leaves() {
		if ctx.Parent() != r.Context() {
			fmt.Println("DEBUG", ctx.Parent(), r.Context())
			panic("leaves not point to parent!")
		}
	}
	log.Print("Initializing IndexSearcher...")
	ss := search.NewIndexSearcher(r)
	log.Print("Searching...")
	docs, err := ss.SearchTop(search.NewTermQuery(index.NewTerm("content", "bat")), 10)
	if err != nil {
		panic(err)
	}
	log.Println("Hits:", docs.TotalHits)
	doc, err := r.Document(docs.ScoreDocs[0].Doc)
	if err != nil {
		panic(err)
	}
	log.Println("Hit 1's title: ", doc.Get("title"))
}
