package index

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/store"
	"testing"
)

func TestLeaves(t *testing.T) {
	path := "../search/testdata/win8/belfrysample"
	d, err := store.OpenFSDirectory(path)
	if err != nil {
		t.Error(err)
	}
	r, err := OpenDirectoryReader(d)
	if err != nil {
		t.Error(err)
	}
	if len(r.Leaves()) != 1 {
		t.Error("Should have one sub reader.")
	}
	fmt.Println(r)
	r2 := r.Context().Reader()
	fmt.Println(r2)
	if len(r2.Leaves()) != 1 {
		t.Error("Should have one sub reader.")
	}
}
