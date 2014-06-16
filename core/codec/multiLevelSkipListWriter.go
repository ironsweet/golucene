package codec

import (
	"github.com/balzaczyy/golucene/core/util"
)

type MultiLevelSkipListWriter struct {
	// number levels in this skip list
	numberOfSkipLevels int
	// the skip interval in ths list with level=0
	skipInterval int
	// skipInterval used for level > 0
	skipMultiplier int
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
