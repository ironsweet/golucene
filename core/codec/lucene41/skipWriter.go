package lucene41

import (
	"github.com/balzaczyy/golucene/core/codec"
	"github.com/balzaczyy/golucene/core/store"
)

type SkipWriter struct {
	*codec.MultiLevelSkipListWriter

	lastSkipDoc         []int
	lastSkipDocPointer  []int64
	lastSkipPosPointer  []int64
	lastSkipPayPointer  []int64
	lastPayloadByteUpto []int

	docOut store.IndexOutput
	posOut store.IndexOutput
	payOut store.IndexOutput
}

func NewSkipWriter(maxSkipLevels, blockSize, docCount int,
	docOut, posOut, payOut store.IndexOutput) *SkipWriter {
	ans := &SkipWriter{
		MultiLevelSkipListWriter: codec.NewMultiLevelSkipListWriter(blockSize, 8, maxSkipLevels, docCount),
		docOut:             docOut,
		posOut:             posOut,
		payOut:             payOut,
		lastSkipDoc:        make([]int, maxSkipLevels),
		lastSkipDocPointer: make([]int64, maxSkipLevels),
	}
	if posOut != nil {
		ans.lastSkipPosPointer = make([]int64, maxSkipLevels)
		if payOut != nil {
			ans.lastSkipPayPointer = make([]int64, maxSkipLevels)
		}
		ans.lastPayloadByteUpto = make([]int, maxSkipLevels)
	}
	return ans
}
