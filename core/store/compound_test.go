package store

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	"testing"
)

func TestReadEntriesOSX(t *testing.T) {
	fmt.Println("TestReadEntriesOSX...")
	path := "../search/testdata/osx/belfrysample"
	d, err := OpenFSDirectory(path)
	if err != nil {
		t.Error(err)
	}
	ctx := NewIOContextBool(false)
	handle, err := d.CreateSlicer("_0.cfs", ctx)
	if err != nil {
		t.Error(err)
	}
	m, err := readEntries(handle, d, "_0.cfs")
	if err != nil {
		t.Error(err)
	}
	if len(m) != 9 {
		t.Errorf("Should have 9 entries.")
	}
	f := m[".fnm"]
	if f.offset != 31 || f.length != 541 {
		t.Errorf("'.fnm' (offset=31, length=541), now %v", f)
	}
	f = m["_Lucene41_0.tip"]
	if f.offset != 9820 || f.length != 252 {
		t.Errorf("'_Lucene41_0.tip' (offset=9820, length=242), now %v", f)
	}
}

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
