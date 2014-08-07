package lucene42

import (
	"github.com/balzaczyy/golucene/core/store"
	"testing"
)

func TestReadFieldInfos(t *testing.T) {
	path := "../../search/testdata/osx/belfrysample"
	d, err := store.OpenFSDirectory(path)
	if err != nil {
		t.Error(err)
	}
	ctx := store.NewIOContextBool(false)
	cd, err := store.NewCompoundFileDirectory(d, "_0.cfs", ctx, false)
	if err != nil {
		t.Error(err)
	}
	fis, err := Lucene42FieldInfosReader(cd, "_0", "", store.IO_CONTEXT_READONCE)
	if err != nil {
		t.Error(err)
	}
	if !fis.HasNorms || fis.HasDocValues {
		t.Errorf("hasNorms must be true and hasDocValues must be false, but found %v", fis)
	}
}
