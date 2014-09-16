package lucene49

import (
	"github.com/balzaczyy/golucene/core/codec"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"github.com/balzaczyy/golucene/core/util/packed"
	"math"
)

// lucene49/Lucene49NormsConsumer.java

const (
	DELTA_COMPRESSED = 0
	TABLE_COMPRESSED = 1
	CONST_COMPRESSED = 2
	UNCOMPRESSED     = 3
)

type NormsConsumer struct {
	data, meta store.IndexOutput
	maxDoc     int
}

func newLucene49NormsConsumer(state *SegmentWriteState,
	dataCodec, dataExtension, metaCodec, metaExtension string) (nc *NormsConsumer, err error) {

	assert(packed.PackedFormat(packed.PACKED_SINGLE_BLOCK).IsSupported(1))
	assert(packed.PackedFormat(packed.PACKED_SINGLE_BLOCK).IsSupported(2))
	assert(packed.PackedFormat(packed.PACKED_SINGLE_BLOCK).IsSupported(4))

	nc = &NormsConsumer{maxDoc: state.SegmentInfo.DocCount()}
	var success = false
	defer func() {
		if !success {
			util.CloseWhileSuppressingError(nc)
		}
	}()

	dataName := util.SegmentFileName(state.SegmentInfo.Name, state.SegmentSuffix, dataExtension)
	if nc.data, err = state.Directory.CreateOutput(dataName, state.Context); err != nil {
		return nil, err
	}
	if err = codec.WriteHeader(nc.data, dataCodec, VERSION_CURRENT); err != nil {
		return nil, err
	}
	metaName := util.SegmentFileName(state.SegmentInfo.Name, state.SegmentSuffix, metaExtension)
	if nc.meta, err = state.Directory.CreateOutput(metaName, state.Context); err != nil {
		return nil, err
	}
	if err = codec.WriteHeader(nc.meta, metaCodec, VERSION_CURRENT); err != nil {
		return nil, err
	}
	success = true
	return nc, nil
}

func (nc *NormsConsumer) AddNumericField(field *FieldInfo,
	f func() (interface{}, bool)) (err error) {

	if err = nc.meta.WriteVInt(field.Number); err != nil {
		return
	}
	minValue, maxValue := int64(math.MaxInt64), int64(math.MinInt64)
	// TODO: more efficient?
	uniqueValues := make(map[int64]bool)

	count := int64(0)
	for {
		nv, ok := f()
		if !ok {
			break
		}
		assert2(nv != nil, "illegal norms data for field %v, got null for value: %v", field.Name, count)
		v := nv.(int64)

		if v < minValue {
			minValue = v
		}
		if v > maxValue {
			maxValue = v
		}

		if uniqueValues != nil {
			_, ok := uniqueValues[v]
			uniqueValues[v] = true
			if !ok && len(uniqueValues) > 256 {
				uniqueValues = nil
			}
		}

		count++
	}
	assert2(count == int64(nc.maxDoc),
		"illegal norms data for field %v, expected %v values, got %v",
		field.Name, nc.maxDoc, count)

	if len(uniqueValues) == 1 {
		// 0 bpv
		if err = nc.meta.WriteByte(CONST_COMPRESSED); err != nil {
			return
		}
		if err = nc.meta.WriteLong(minValue); err != nil {
			return
		}
	} else if len(uniqueValues) > 0 {
		// small number of unique values; this is the typical case:
		// we only use bpv=1,2,4,8
		// format := packed.PACKED_SINGLE_BLOCK
		bitsPerValue := packed.BitsRequired(int64(len(uniqueValues)) - 1)
		if bitsPerValue == 3 {
			bitsPerValue = 4
		} else if bitsPerValue > 4 {
			bitsPerValue = 8
		}

		if bitsPerValue == 8 && minValue >= 0 && maxValue <= 255 {
			// uncompressed []byte
			if err = nc.meta.WriteByte(UNCOMPRESSED); err != nil {
				return err
			}
			if err = nc.meta.WriteLong(nc.data.FilePointer()); err != nil {
				return err
			}
			for {
				nv, ok := f()
				if !ok {
					break
				}
				n := byte(0)
				if nv != nil {
					n = byte(nv.(int64))
				}
				if err = nc.data.WriteByte(byte(n)); err != nil {
					return err
				}
			}
		} else {
			panic("not implemented yet")
		}
	} else {
		panic("not implemented yet")
	}
	return nil
}

func (nc *NormsConsumer) Close() (err error) {
	var success = false
	defer func() {
		if success {
			err = util.Close(nc.data, nc.meta)
		} else {
			util.CloseWhileSuppressingError(nc.data, nc.meta)
		}
	}()

	if nc.meta != nil {
		if err = nc.meta.WriteVInt(-1); err != nil { // write EOF marker
			return
		}
		if err = codec.WriteFooter(nc.meta); err != nil { // write checksum
			return
		}
	}
	if nc.data != nil {
		if err = codec.WriteFooter(nc.data); err != nil { // write checksum
			return
		}
	}
	success = true
	return nil
}
