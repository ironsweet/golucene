package search

import (
	_ "github.com/balzaczyy/golucene/core/codec/lucene42"
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/store"
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
}

func TestKeywordSearch(t *testing.T) {
	d, err := store.OpenFSDirectory("testdata/belfrysample")
	if err != nil {
		t.Error(err)
	}
	r, err := index.OpenDirectoryReader(d)
	if err != nil {
		t.Error(err)
	}
	if r == nil {
		t.Error("DirectoryReader cannot be opened.")
	}
	if len(r.Leaves()) < 1 {
		t.Error("Should have one leaf.")
	}
	ss := NewIndexSearcher(r)
	docs, err := ss.SearchTop(NewTermQuery(index.NewTerm("content", "bat")), 10)
	if err != nil {
		t.Error(err)
	}

	assertEquals(t, 8, docs.TotalHits)
	doc, err := r.Document(docs.ScoreDocs[0].Doc)
	if err != nil {
		t.Error(err)
	}
	assertEquals(t, "Bat recycling", doc.Get("title"))
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
