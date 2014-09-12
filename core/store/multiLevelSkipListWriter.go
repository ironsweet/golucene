package store

import (
	"github.com/balzaczyy/golucene/core/util"
)

/*
TODO: migrate original comment.

Note: this class was moved from package codec to store since it
caused cyclic dependency (store<->codec).
*/
type MultiLevelSkipListWriter struct {
	// number levels in this skip list
	numberOfSkipLevels int
	// the skip interval in ths list with level=0
	skipInterval int
	// skipInterval used for level > 0
	skipMultiplier int
	// for every skip level a different buffer is used
	skipBuffer []*RAMOutputStream
}

/* Creates a MultiLevelSkipListWriter. */
func NewMultiLevelSkipListWriter(skipInterval,
	skipMultiplier, maxSkipLevels, df int) *MultiLevelSkipListWriter {
	numberOfSkipLevels := 1
	// calculate the maximum number of skip levels for this document frequency
	if df > skipInterval {
		numberOfSkipLevels = 1 + util.Log(int64(df/skipInterval), skipMultiplier)
	}
	// make sure it does not exceed maxSkipLevels
	if numberOfSkipLevels > maxSkipLevels {
		numberOfSkipLevels = maxSkipLevels
	}
	return &MultiLevelSkipListWriter{
		skipInterval:       skipInterval,
		skipMultiplier:     skipMultiplier,
		numberOfSkipLevels: numberOfSkipLevels,
	}
}

/* Allocates internal skip buffers. */
func (w *MultiLevelSkipListWriter) init() {
	w.skipBuffer = make([]*RAMOutputStream, w.numberOfSkipLevels)
	for i, _ := range w.skipBuffer {
		w.skipBuffer[i] = NewRAMOutputStreamBuffer()
	}
}

/* Creates new buffers or empties the existing ones */
func (w *MultiLevelSkipListWriter) ResetSkip() {
	if w.skipBuffer == nil {
		w.init()
	} else {
		for _, v := range w.skipBuffer {
			v.Reset()
		}
	}
}

/*
Writes the current skip data to the buffers. The current document
frequency determines the max level is skip ddata is to be written to.
*/
func (w *MultiLevelSkipListWriter) BufferSkip(df int) error {
	panic("niy")
}

/* Writes the buffered skip lists to the given output. */
func (w *MultiLevelSkipListWriter) WriteSkip(output IndexOutput) (int64, error) {
	skipPointer := output.FilePointer()
	if len(w.skipBuffer) == 0 {
		return skipPointer, nil
	}

	for level := w.numberOfSkipLevels - 1; level > 0; level-- {
		if length := w.skipBuffer[level].FilePointer(); length > 0 {
			err := output.WriteVLong(length)
			if err != nil {
				return 0, err
			}
			err = w.skipBuffer[level].WriteTo(output)
			if err != nil {
				return 0, err
			}
		}
	}
	return skipPointer, w.skipBuffer[0].WriteTo(output)
}
