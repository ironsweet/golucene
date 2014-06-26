package index

import (
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
)

// index/SegmentReadState.java

type SegmentReadState struct {
	dir               store.Directory
	segmentInfo       *model.SegmentInfo
	fieldInfos        model.FieldInfos
	context           store.IOContext
	termsIndexDivisor int
	segmentSuffix     string
}

func newSegmentReadState(dir store.Directory,
	info *model.SegmentInfo, fieldInfos model.FieldInfos,
	context store.IOContext, termsIndexDivisor int) SegmentReadState {

	return SegmentReadState{dir, info, fieldInfos, context, termsIndexDivisor, ""}
}
