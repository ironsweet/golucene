package model

import (
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"strconv"
	"strings"
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
	termIndexInterval int
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

	ans := &SegmentWriteState{
		infoStream:        infoStream,
		Directory:         dir,
		SegmentInfo:       segmentInfo,
		FieldInfos:        fieldInfos,
		SegUpdates:        SegUpdates,
		SegmentSuffix:     segmentSuffix,
		termIndexInterval: termIndexInterval,
		Context:           ctx,
	}
	assert(ans.assertSegmentSuffix(segmentSuffix))
	return ans
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
		state.termIndexInterval,
		state.Context,
	}
}

func (s *SegmentWriteState) assertSegmentSuffix(segmentSuffix string) bool {
	if len(segmentSuffix) == 0 {
		return true
	}
	numParts := len(strings.SplitN(segmentSuffix, "_", 3))
	if numParts == 2 {
		return true
	}
	if numParts == 1 {
		_, err := strconv.ParseInt(segmentSuffix, 36, 64)
		assert(err == nil)
		return true
	}
	return false
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
