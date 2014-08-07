package index

import (
	_ "github.com/balzaczyy/golucene/core/codec/lucene42"
	"github.com/balzaczyy/golucene/core/store"
	"testing"
)

func TestInitTerm(t *testing.T) {
	path := "../search/testdata/win8/belfrysample"
	d, err := store.OpenFSDirectory(path)
	if err != nil {
		t.Error(err)
	}
	r, err := OpenDirectoryReader(d)
	if err != nil {
		t.Error(err)
	}
	leaves := r.Context().Leaves()
	if len(leaves) != 1 {
		t.Error("Should have one leaf.")
	}
	ctx := leaves[0]
	fields := ctx.reader.Fields()
	terms := fields.Terms("content")
	termsEnum := terms.Iterator(nil)

	if len(termsEnum.Term()) != 0 {
		t.Error("Initial term should has zero length.")
	}

	ok, err := termsEnum.SeekExact(NewTerm("content", "bat").Bytes)
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Error("SeekExact should return true.")
	}
}
