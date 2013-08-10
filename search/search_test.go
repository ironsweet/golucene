package search

import (
	// "fmt"
	"github.com/balzaczyy/golucene/index"
	"github.com/balzaczyy/golucene/store"
	"testing"
)

func TestLastCommitGeneration(t *testing.T) {
	d, err := store.OpenFSDirectory("testdata/belfrysample")
	if err != nil {
		t.Error(err)
	}

	files, err := d.ListAll()
	if err != nil {
		t.Error(err)
	}
	if files != nil {
		genA := index.LastCommitGeneration(files)
		if genA != 1 {
			t.Error("Should be 1, but was %v", genA)
		}
	}
	// r, err := index.OpenDirectoryReader(d)
	// if err != nil {
	// 	t.Error(err)
	// }

	// ss := NewIndexSearcher(r)
	// assertEquals(t, 8, ss.SearchTop(NewTermQuery(index.NewTerm("content", "bat")), 10))
}

// func TestSingleSearch(t *testing.T) {
// 	ss := NewSearcher()
// 	ss.IncludeIndex("testdata/belfrysample")
// 	assertEquals(t, 8, ss.search("bat"))

// 	ss = NewSearcher()
// 	ss.IncludeIndex("testdata/usingworldtimepro")
// 	assertEquals(t, 16, ss.search("time"))
// }

func assertEquals(t *testing.T, a, b interface{}) {
	if a != b {
		t.Errorf("Expected '%v', but '%v'", a, b)
	}
}

// func TestFederatedSearch(t *testing.T) {
// 	ss := NewSearcher()
// 	ss.IncludeIndex("testdata/belfrysample")
// 	ss.IncludeIndex("testdata/usingworldtimepro")
// 	assertEquals(t, 17, ss.search("time"))
// }
