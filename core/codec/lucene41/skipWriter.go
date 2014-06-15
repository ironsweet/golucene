package lucene41

import (
	"github.com/balzaczyy/golucene/core/store"
)

type SkipWriter struct {
}

func NewSkipWriter(maxSipLevels, blockSize, docCount int,
	docOut, posOut, payOut store.IndexOutput) *SkipWriter {
	panic("not implemented yet")
}
