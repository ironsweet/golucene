package lucene42

import (
	"github.com/balzaczyy/golucene/core/codec"
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
)

const NC_VERSION_GCD_COMPRESSION = 1
const NC_VERSION_CURRENT = NC_VERSION_GCD_COMPRESSION

/* Writer for Lucene42NormsFormat */
type NormsConsumer struct {
	data, meta              store.IndexOutput
	maxDoc                  int
	acceptableOverheadRatio float32
}

// TODO package private method
func NewNormsConsumer(state model.SegmentWriteState,
	dataCodec, dataExtension, metaCodec, metaExtension string,
	acceptableOverheadRatio float32) (*NormsConsumer, error) {

	ans := &NormsConsumer{
		acceptableOverheadRatio: acceptableOverheadRatio,
		maxDoc:                  state.SegmentInfo.DocCount(),
	}

	var success = false
	defer func() {
		if !success {
			util.CloseWhileSuppressingError(ans)
		}
	}()

	var err error
	dataName := util.SegmentFileName(state.SegmentInfo.Name, state.SegmentSuffix, dataExtension)
	ans.data, err = state.Directory.CreateOutput(dataName, state.Context)
	if err == nil {
		err = codec.WriteHeader(ans.data, dataCodec, NC_VERSION_CURRENT)
	}
	if err != nil {
		return nil, err
	}

	metaName := util.SegmentFileName(state.SegmentInfo.Name, state.SegmentSuffix, metaExtension)
	ans.meta, err = state.Directory.CreateOutput(metaName, state.Context)
	if err == nil {
		err = codec.WriteHeader(ans.meta, metaCodec, NC_VERSION_CURRENT)
	}
	if err != nil {
		return nil, err
	}
	success = true
	return ans, nil
}

func (nc *NormsConsumer) AddNumericField(field *model.FieldInfo,
	f func() (interface{}, bool)) error {
	panic("not implemented yet")
}

func (nc *NormsConsumer) Close() error {
	panic("no timplemented yet")
}
