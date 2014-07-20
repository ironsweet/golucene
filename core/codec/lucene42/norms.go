package lucene42

import (
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/util/packed"
)

// lucene42/Lucene42NormsFormat.java

/*
Lucene 4.2 score normalization format.

NOTE: this uses the same format as Lucene42DocValuesFormat Numeric
DocValues, but with different fiel extensions, and passing FASTEST
for uncompressed encoding: trading off space for performance.

Fields:
- .nvd: DocValues data
- .nvm: DocValues metadata
*/
type Lucene42NormsFormat struct {
	acceptableOverheadRatio float32
}

func NewLucene42NormsFormat() *Lucene42NormsFormat {
	// note: we choose FASTEST here (otherwise our norms are half as big
	// but 15% slower than previous lucene)
	return newLucene42NormsFormatWithOverhead(packed.PackedInts.FASTEST)
}

func newLucene42NormsFormatWithOverhead(acceptableOverheadRatio float32) *Lucene42NormsFormat {
	return &Lucene42NormsFormat{acceptableOverheadRatio}
}

func (f *Lucene42NormsFormat) NormsConsumer(state *SegmentWriteState) (w DocValuesConsumer, err error) {
	panic("this codec can only be used for reading")
}

func (f *Lucene42NormsFormat) NormsProducer(state SegmentReadState) (r DocValuesProducer, err error) {
	return newLucene42DocValuesProducer(state, "Lucene41NormsData", "nvd", "Lucene41NormsMetadata", "nvm")
}
