package lucene41

import (
	"github.com/balzaczyy/golucene/core/store"
)

type SkipWriter struct {
	*store.MultiLevelSkipListWriter

	lastSkipDoc         []int
	lastSkipDocPointer  []int64
	lastSkipPosPointer  []int64
	lastSkipPayPointer  []int64
	lastPayloadByteUpto []int

	docOut store.IndexOutput
	posOut store.IndexOutput
	payOut store.IndexOutput

	fieldHasPositions bool
	fieldHasOffsets   bool
	fieldHasPayloads  bool

	initialized bool
	lastDocFP   int64
	lastPosFP   int64
	lastPayFP   int64
}

func NewSkipWriter(maxSkipLevels, blockSize, docCount int,
	docOut, posOut, payOut store.IndexOutput) *SkipWriter {
	ans := &SkipWriter{
		MultiLevelSkipListWriter: store.NewMultiLevelSkipListWriter(blockSize, 8, maxSkipLevels, docCount),
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

func (w *SkipWriter) SetField(fieldHasPositions, fieldHasOffsets, fieldHasPayloads bool) {
	w.fieldHasPositions = fieldHasPositions
	w.fieldHasOffsets = fieldHasOffsets
	w.fieldHasPayloads = fieldHasPayloads
}

func (w *SkipWriter) ResetSkip() {
	w.lastDocFP = w.docOut.FilePointer()
	if w.fieldHasPositions {
		w.lastPosFP = w.posOut.FilePointer()
		if w.fieldHasOffsets || w.fieldHasPayloads {
			w.lastPayFP = w.payOut.FilePointer()
		}
	}
	w.initialized = false
}

func (w *SkipWriter) initSkip() {
	panic("not implemented yet")
	// w.MultiLevelSkipListWriter.ResetSkip()
	// for i, _ := range w.lastSkipDoc {
	// 	w.lastSkipDoc[i] = 0
	// }
	// for i, _ := range w.lastSkipDocPointer {
	// 	w.lastSkipDocPointer[i] = w.docOut.FilePointer()
	// }
	// if w.fieldHasPositions {
	// 	for i, _ := range w.lastSkipPosPointer {
	// 		w.lastSkipPosPointer[i] = w.posOut.FilePointer()
	// 	}
	// 	if w.fieldHasPayloads {
	// 		for i, _ := range w.lastPayloadByteUpto {
	// 			w.lastPayloadByteUpto[i] = 0
	// 		}
	// 		if w.fieldHasOffsets || w.fieldHasPayloads {
	// 			for i, _ := range w.lastSkipPayPointer {
	// 				w.lastSkipPayPointer[i] = w.payOut.FilePointer()
	// 			}
	// 		}
	// 	}
	// }
}

/* Sets the values for the current skip data. */
func (w *SkipWriter) BufferSkip(doc, numDocs int, posFP, payFP int64, posBufferUpto, payloadByteUpto int) error {
	panic("not implemented yet")
}
