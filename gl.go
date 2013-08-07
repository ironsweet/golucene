package main

import (
	"github.com/balzaczyy/golucene/index"
	"github.com/balzaczyy/golucene/search"
	"github.com/balzaczyy/golucene/store"
	"log"
)

func main() {
	log.Print("Oepning FSDirectory...")
	d, err := store.OpenFSDirectory("search/testdata/belfrysample")
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
