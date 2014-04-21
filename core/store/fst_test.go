package store

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	"testing"
)

func TestClone(t *testing.T) {
	fmt.Println("Testing Loading FST...")
	path := "../search/testdata/belfrysample"
	d, err := OpenFSDirectory(path)
	if err != nil {
		t.Error(err)
	}
	ctx := NewIOContextBool(false)
	in, err := d.OpenInput("_0_Lucene41_0.tip", ctx)
	if err != nil {
		t.Error(err)
	}
	version, err := codec.CheckHeader(in, "BLOCK_TREE_TERMS_INDEX", 0, 1)
	var indexDirOffset int64 = 0
	if version < 1 {
		indexDirOffset, err = in.ReadLong()
		if err != nil {
			t.Error(err)
		}
	} else { // >= 1
		in.Seek(in.Length() - 8)
		indexDirOffset, err = in.ReadLong()
		if err != nil {
			t.Error(err)
		}
	}
	fmt.Println("indexDirOffset:", indexDirOffset)
	in.Seek(indexDirOffset)

	indexStartFP, err := in.ReadVLong()
	if err != nil {
		t.Error(err)
	}
	fmt.Println("indexStartFP:", indexStartFP)

	fmt.Println("Before clone", in)
	clone := in.Clone()
	fmt.Println("After clone", clone)
	if _, ok := clone.(*SimpleFSIndexInput); !ok {
		t.Error("Clone() should return *SimpleFSIndexInput.")
	}
	clone.Seek(indexStartFP)
	fmt.Println("After clone.Seek()", clone)

	_, err = codec.CheckHeader(clone, "FST", 3, 4)
	if err != nil {
		t.Error(err)
	}
}
