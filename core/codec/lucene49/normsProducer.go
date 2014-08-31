package lucene49

import (
	"errors"
	"github.com/balzaczyy/golucene/core/codec"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"reflect"
	"sync"
)

// lucene49/Lucene49NormsProduer.java

type NormsEntry struct {
	format byte
	offset int64
}

type NormsProducer struct {
	sync.Locker

	norms   map[int]*NormsEntry
	data    store.IndexInput
	version int32

	instances map[int]NumericDocValues

	maxDoc       int
	ramBytesUsed int64 // atomic
}

func newLucene49NormsProducer(state SegmentReadState,
	dataCodec, dataExtension, metaCodec, metaExtension string) (np *NormsProducer, err error) {

	np = &NormsProducer{
		Locker:       new(sync.Mutex),
		maxDoc:       state.SegmentInfo.DocCount(),
		ramBytesUsed: util.ShallowSizeOfInstance(reflect.TypeOf(np)),
	}
	metaName := util.SegmentFileName(state.SegmentInfo.Name, state.SegmentSuffix, metaExtension)
	// read in the entries from the metadta file.
	var in store.ChecksumIndexInput
	if in, err = state.Dir.OpenChecksumInput(metaName, state.Context); err != nil {
		return nil, err
	}

	if err = func() error {
		var success = false
		defer func() {
			if success {
				err = util.Close(in)
			} else {
				util.CloseWhileSuppressingError(in)
			}
		}()

		if np.version, err = codec.CheckHeader(in, metaCodec, VERSION_START, VERSION_CURRENT); err != nil {
			return err
		}
		if err = np.readFields(in, state.FieldInfos); err != nil {
			return err
		}
		if _, err = codec.CheckFooter(in); err != nil {
			return err
		}
		success = true
		return nil
	}(); err != nil {
		return nil, err
	}

	dataName := util.SegmentFileName(state.SegmentInfo.Name, state.SegmentSuffix, dataExtension)
	if np.data, err = state.Dir.OpenInput(dataName, state.Context); err != nil {
		return nil, err
	}
	var success = false
	defer func() {
		if !success {
			util.CloseWhileSuppressingError(np.data)
		}
	}()

	var version2 int32
	if version2, err = codec.CheckHeader(np.data, dataCodec, VERSION_START, VERSION_CURRENT); err != nil {
		return nil, err
	}
	if version2 != np.version {
		return nil, errors.New("Format versions mismatch")
	}
	success = true

	return np, nil
}

func (np *NormsProducer) readFields(meta store.IndexInput, infos FieldInfos) error {
	panic("not implemented yet")
}

func (np *NormsProducer) Numeric(field *FieldInfo) (NumericDocValues, error) {
	np.Lock()
	defer np.Unlock()

	panic("not implemented yet")
}

func (np *NormsProducer) Binary(field *FieldInfo) (BinaryDocValues, error) {
	panic("not supported")
}

func (np *NormsProducer) Sorted(field *FieldInfo) (SortedDocValues, error) {
	panic("not supported")
}

func (np *NormsProducer) SortedSet(field *FieldInfo) (SortedSetDocValues, error) {
	panic("not supported")
}

func (np *NormsProducer) Close() error {
	return np.data.Close()
}
