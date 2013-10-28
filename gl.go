package main

import (
	"github.com/balzaczyy/golucene/index"
	"github.com/balzaczyy/golucene/search"
	"github.com/balzaczyy/golucene/store"
	"log"
)

func main() {
	log.Print("Oepning FSDirectory...")
	// path := "src/github.com/balzaczyy/golucene/search/testdata/win8/belfrysample"
	path := "search/testdata/win8/belfrysample"
	// path := "/private/tmp/kc/index/belfrysample"
	// path := "c:/tmp/kc/index/belfrysample"
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
	log.Print("Initializing IndexSearcher...")
	ss := search.NewIndexSearcher(r)
	log.Print("Searching...")
	n, err := ss.SearchTop(search.NewTermQuery(index.NewTerm("content", "bat")), 10)
	if err != nil {
		panic(err)
	}
	log.Print(n)
}
