package store

import (
	"github.com/balzaczyy/golucene/core/codec"
	"github.com/balzaczyy/golucene/core/util"
	"math/rand"
	"testing"
)

func newTestIOContext(r *rand.Rand) IOContext {
	return newTestIOContextFrom(r, NewIOContextFromType(IOContextType(IO_CONTEXT_TYPE_DEFAULT)))
}

func newTestIOContextFrom(r *rand.Rand, oldContext IOContext) IOContext {
	randomNumDocs := r.Intn(4192)
	size := r.Intn(512) * randomNumDocs
	// ignore flushInfo and mergeInfo for now
	// Make a totally random IOContext:
	var context IOContext
	switch r.Intn(5) {
	case 0:
		context = IO_CONTEXT_DEFAULT
	case 1:
		context = IO_CONTEXT_READ
	case 2:
		context = IO_CONTEXT_READONCE
	case 3:
		context = NewIOContextForMerge(&MergeInfo{randomNumDocs, int64(size), true, -1})
	case 4:
		context = NewIOContextForFlush(&FlushInfo{randomNumDocs, int64(size)})
	default:
		context = IO_CONTEXT_DEFAULT
	}
	return context
}

func TestReadingFromSlicedIndexInputOSX(t *testing.T) {
	t.Logf("TestReadingFromSlicedIndexInputOSX...")
	path := "../search/testdata/osx/belfrysample"
	d, err := OpenFSDirectory(path)
	if err != nil {
		t.Error(err)
	}
	ctx := NewIOContextBool(false)
	cd, err := NewCompoundFileDirectory(d, "_0.cfs", ctx, false)
	name := util.SegmentFileName("_0", "Lucene41_0", "pos")
	posIn, err := cd.OpenInput(name, ctx)
	if err != nil {
		t.Error(err)
	}
	t.Log(posIn)
	codec.CheckHeader(posIn, "Lucene41PostingsWriterPos", 0, 0)
	// codec header mismatch: actual header=0 vs expected header=1071082519 (resource: SlicedIndexInput(SlicedIndexInput(_0_Lucene41_0.pos in SimpleFSIndexInput(path='/private/tmp/kc/index/belfrysample/_0.cfs')) in SimpleFSIndexInput(path='/private/tmp/kc/index/belfrysample/_0.cfs') slice=1461:3426))
}
