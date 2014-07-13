package store

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	"testing"
)

func TestCheckHeaderWin8(t *testing.T) {
	fmt.Println("TestCheckHeaderWin8...")
	path := "../search/testdata/win8/belfrysample"
	d, err := OpenFSDirectory(path)
	if err != nil {
		t.Error(err)
	}
	ctx := NewIOContextBool(false)
	cd, err := NewCompoundFileDirectory(d, "_0.cfs", ctx, false)
	if err != nil {
		t.Error(err)
	}
	r, err := cd.OpenInput("_0_Lucene41_0.pos", ctx)
	_, err = codec.CheckHeader(r, "Lucene41PostingsWriterPos", 0, 0)
	if err != nil {
		t.Error(err)
	}
}
