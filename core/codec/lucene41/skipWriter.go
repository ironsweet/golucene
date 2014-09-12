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

	curDoc             int
	curDocPointer      int64
	curPosPointer      int64
	curPayPointer      int64
	curPosBufferUpto   int
	curPayloadByteUpto int
	fieldHasPositions  bool
	fieldHasOffsets    bool
	fieldHasPayloads   bool

	initialized bool
	lastDocFP   int64
	lastPosFP   int64
	lastPayFP   int64
}

func NewSkipWriter(maxSkipLevels, blockSize, docCount int,
	docOut, posOut, payOut store.IndexOutput) *SkipWriter {
	ans := &SkipWriter{
		docOut:             docOut,
		posOut:             posOut,
		payOut:             payOut,
		lastSkipDoc:        make([]int, maxSkipLevels),
		lastSkipDocPointer: make([]int64, maxSkipLevels),
	}
	ans.MultiLevelSkipListWriter = store.NewMultiLevelSkipListWriter(ans, blockSize, 8, maxSkipLevels, docCount)

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
	if !w.initialized {
		w.MultiLevelSkipListWriter.ResetSkip()
		for i, _ := range w.lastSkipDoc {
			w.lastSkipDoc[i] = 0
		}
		for i, _ := range w.lastSkipDocPointer {
			w.lastSkipDocPointer[i] = w.lastDocFP
		}
		if w.fieldHasPositions {
			for i, _ := range w.lastSkipPosPointer {
				w.lastSkipPosPointer[i] = w.lastPosFP
			}
			if w.fieldHasPayloads {
				for i, _ := range w.lastPayloadByteUpto {
					w.lastPayloadByteUpto[i] = 0
				}
			}
			if w.fieldHasOffsets || w.fieldHasPayloads {
				for i, _ := range w.lastSkipPayPointer {
					w.lastSkipPayPointer[i] = w.lastPayFP
				}
			}
		}
		w.initialized = true
	}
}

/* Sets the values for the current skip data. */
func (w *SkipWriter) BufferSkip(doc, numDocs int, posFP, payFP int64, posBufferUpto, payloadByteUpto int) error {
	w.initSkip()
	w.curDoc = doc
	w.curDocPointer = w.docOut.FilePointer()
	w.curPosPointer = posFP
	w.curPayPointer = payFP
	w.curPosBufferUpto = posBufferUpto
	w.curPayloadByteUpto = payloadByteUpto
	return w.MultiLevelSkipListWriter.BufferSkip(numDocs)
}

func (w *SkipWriter) WriteSkipData(level int, skipBuffer store.IndexOutput) error {
	delta := w.curDoc - w.lastSkipDoc[level]
	var err error
	if err = skipBuffer.WriteVInt(int32(delta)); err != nil {
		return err
	}
	w.lastSkipDoc[level] = w.curDoc

	if err = skipBuffer.WriteVInt(int32(w.curDocPointer - w.lastSkipDocPointer[level])); err != nil {
		return err
	}
	w.lastSkipDocPointer[level] = w.curDocPointer

	if w.fieldHasPositions {
		if err = skipBuffer.WriteVInt(int32(w.curPosPointer - w.lastSkipPosPointer[level])); err != nil {
			return err
		}
		w.lastSkipPosPointer[level] = w.curPosPointer
		if err = skipBuffer.WriteVInt(int32(w.curPosBufferUpto)); err != nil {
			return err
		}

		if w.fieldHasPayloads {
			if err = skipBuffer.WriteVInt(int32(w.curPayloadByteUpto)); err != nil {
				return err
			}
		}

		if w.fieldHasOffsets || w.fieldHasPayloads {
			if err = skipBuffer.WriteVInt(int32(w.curPayPointer - w.lastSkipPayPointer[level])); err != nil {
				return err
			}
			w.lastSkipPayPointer[level] = w.curPayPointer
		}
	}
	return nil
}
