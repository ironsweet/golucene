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
