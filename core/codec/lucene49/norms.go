package lucene49

import (
	"fmt"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
)

type Lucene49NormsFormat struct {
}

func (f *Lucene49NormsFormat) NormsConsumer(state *SegmentWriteState) (w DocValuesConsumer, err error) {
	return newLucene49NormsConsumer(state, DATA_CODEC, DATA_EXTENSION, METADATA_CODEC, METADATA_EXTENSION)
}

func (f *Lucene49NormsFormat) NormsProducer(state SegmentReadState) (r DocValuesProducer, err error) {
	return newLucene49NormsProducer(state, DATA_CODEC, DATA_EXTENSION, METADATA_CODEC, METADATA_EXTENSION)
}

const (
	DATA_CODEC         = "Lucene49NormsData"
	DATA_EXTENSION     = "nvd"
	METADATA_CODEC     = "Lucene49NormsMetadata"
	METADATA_EXTENSION = "nvm"
	VERSION_START      = 0
	VERSION_CURRENT    = VERSION_START
)

func assert(ok bool) {
	if !ok {
		panic("assert fail")
	}
}

func assert2(ok bool, msg string, args ...interface{}) {
	if !ok {
		panic(fmt.Sprintf(msg, args...))
	}
}
