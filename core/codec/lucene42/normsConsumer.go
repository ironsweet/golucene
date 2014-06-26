package lucene42

import (
	"github.com/balzaczyy/golucene/core/index/model"
)

/* Writer for Lucene42NormsFormat */
type NormsConsumer struct {
}

// TODO package private method
func NewNormsConsumer(state model.SegmentWriteState,
	dataCodec, dataExtension, metaCodec, metaExtension string,
	acceptableOverheadRatio float32) (*NormsConsumer, error) {
	panic("not implemented yet")
}

func (nc *NormsConsumer) AddNumericField(field *model.FieldInfo,
	f func() (interface{}, bool)) error {
	panic("not implemented yet")
}

func (nc *NormsConsumer) Close() error {
	panic("no timplemented yet")
}
