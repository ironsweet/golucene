package model

import (
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
)

// index/SegmentWriteState.java

/* Holder class for common parameters used during write. */
type SegmentWriteState struct {
	infoStream        util.InfoStream
	Directory         store.Directory
	SegmentInfo       *SegmentInfo
	FieldInfos        FieldInfos
	DelCountOnFlush   int
	SegUpdates        interface{} // BufferedUpdates
	LiveDocs          util.MutableBits
	SegmentSuffix     string
	termIndexInternal int
	Context           store.IOContext
}

func NewSegmentWriteState(infoStream util.InfoStream,
	dir store.Directory, segmentInfo *SegmentInfo,
	fieldInfos FieldInfos, termIndexInterval int,
	SegUpdates interface{}, ctx store.IOContext) *SegmentWriteState {

	return NewSegmentWriteState2(
		infoStream, dir, segmentInfo, fieldInfos, termIndexInterval,
		SegUpdates, ctx, "")
}

func NewSegmentWriteState2(infoStream util.InfoStream,
	dir store.Directory, segmentInfo *SegmentInfo,
	fieldInfos FieldInfos, termIndexInterval int,
	SegUpdates interface{}, ctx store.IOContext,
	segmentSuffix string) *SegmentWriteState {

	return &SegmentWriteState{
		infoStream, dir, segmentInfo, fieldInfos, 0,
		SegUpdates, nil, segmentSuffix, termIndexInterval, ctx}
}

/* Create a shallow copy of SegmentWriteState with a new segment suffix. */
func NewSegmentWriteStateFrom(state *SegmentWriteState,
	segmentSuffix string) *SegmentWriteState {

	return &SegmentWriteState{
		state.infoStream,
		state.Directory,
		state.SegmentInfo,
		state.FieldInfos,
		state.DelCountOnFlush,
		state.SegUpdates,
		nil,
		segmentSuffix,
		state.termIndexInternal,
		state.Context,
	}
}

// index/SegmentReadState.java

type SegmentReadState struct {
	Dir               store.Directory
	SegmentInfo       *SegmentInfo
	FieldInfos        FieldInfos
	Context           store.IOContext
	TermsIndexDivisor int
	SegmentSuffix     string
}

func NewSegmentReadState(dir store.Directory,
	info *SegmentInfo, fieldInfos FieldInfos,
	context store.IOContext, termsIndexDivisor int) SegmentReadState {

	return SegmentReadState{dir, info, fieldInfos, context, termsIndexDivisor, ""}
}
