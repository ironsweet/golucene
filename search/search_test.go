package search

import (
	"testing"
)

func TestIndexSearcher(t *testing.T) {
	d := FSDirectory.open("testdata/belfrysample")
	r := DirectoryReader.open(d)
	ss := NewIndexSearcher(r)
	assertEquals(t, 8, ss.Search("bat"))
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
