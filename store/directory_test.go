package store

import (
	"math/rand"
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
		context = NewIOContextForMerge(MergeInfo{randomNumDocs, int64(size), true, -1})
	case 4:
		context = NewIOContextForFlush(FlushInfo{randomNumDocs, int64(size)})
	default:
		context = IO_CONTEXT_DEFAULT
	}
	return context
}
