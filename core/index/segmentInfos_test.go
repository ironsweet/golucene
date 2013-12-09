package index

import (
	"github.com/balzaczyy/golucene/core/store"
	"testing"
)

func TestLastCommitGeneration(t *testing.T) {
	d, err := store.OpenFSDirectory("../search/testdata/belfrysample")
	if err != nil {
		t.Error(err)
	}

	files, err := d.ListAll()
	if err != nil {
		t.Error(err)
	}
	if files != nil {
		genA := LastCommitGeneration(files)
		assertEquals(t, int64(1), genA)
	}
}

func assertEquals(t *testing.T, a, b interface{}) {
	if a != b {
		t.Errorf("Expected '%v', but '%v'", a, b)
	}
}
