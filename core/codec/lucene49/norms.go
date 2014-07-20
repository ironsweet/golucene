package lucene49

import (
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
)

type Lucene49NormsFormat struct {
}

func (f *Lucene49NormsFormat) NormsConsumer(state *SegmentWriteState) (w DocValuesConsumer, err error) {
	panic("not implemented yet")
}

func (f *Lucene49NormsFormat) NormsProducer(state SegmentReadState) (r DocValuesProducer, err error) {
	panic("not implemented yet")
}
