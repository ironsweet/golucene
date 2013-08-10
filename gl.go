package main

import (
	"github.com/balzaczyy/golucene/index"
	"github.com/balzaczyy/golucene/search"
	"github.com/balzaczyy/golucene/store"
	"log"
)

func main() {
	log.Print("Oepning FSDirectory...")
	// path := "search/testdata/belfrysample"
	path := "/private/tmp/kc/index/belfrysample"
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
	log.Print("Initializing IndexSearcher...")
	ss := search.NewIndexSearcher(r)
	log.Print("Searching...")
	n := ss.SearchTop(search.NewTermQuery(index.NewTerm("content", "bat")), 10)
	log.Print(n)
}
