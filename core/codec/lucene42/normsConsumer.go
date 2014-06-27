package lucene42

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"github.com/balzaczyy/golucene/core/util/packed"
	"math"
)

const NC_VERSION_GCD_COMPRESSION = 1
const NC_VERSION_CURRENT = NC_VERSION_GCD_COMPRESSION

const NUMBER = 0

const UNCOMPRESSED = 2

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
	err := nc.meta.WriteVInt(field.Number)
	if err == nil {
		err = nc.meta.WriteByte(NUMBER)
		if err == nil {
			err = nc.meta.WriteLong(nc.data.FilePointer())
		}
	}
	if err != nil {
		return err
	}
	minValue := int64(math.MaxInt64)
	maxValue := int64(math.MinInt64)
	gcd := int64(0)
	// TODO: more efficient?
	uniqueValues := make(map[int64]bool)

	count := int64(0)
	for {
		nv, ok := f()
		if !ok {
			break
		}
		assert(nv != nil)
		v := nv.(int64)

		if gcd != 1 {
			if v < math.MinInt64/2 || v > math.MaxInt64/2 {
				// in that case v - minValue might overflow and make the GCD
				// computation return wrong results. Since these extreme
				// values are unlikely, we just discard GCD computation for them
				gcd = 1
			} else if count != 0 { // minValue needs to be set first
				gcd = util.Gcd(gcd, v-minValue)
			}
		}

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
	assert(count == int64(nc.maxDoc))

	if uniqueValues != nil {
		// small number of unique values
		bitsPerValue := packed.BitsRequired(int64(len(uniqueValues)) - 1)
		formatAndBits := packed.FastestFormatAndBits(nc.maxDoc, bitsPerValue, nc.acceptableOverheadRatio)
		if formatAndBits.BitsPerValue == 8 && minValue >= math.MinInt8 && maxValue <= math.MaxInt8 {
			err = nc.meta.WriteByte(UNCOMPRESSED) // uncompressed
			if err != nil {
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
				err = nc.data.WriteByte(byte(n))
				if err != nil {
					return err
				}
			}
		} else {
			panic("not implemented yet")
		}
	} else if gcd != 0 && gcd != 1 {
		panic("not implemented yet")
	} else {
		panic("not implemented yet")
	}
	return nil
}

func assert(ok bool) {
	assert2(ok, "assert fail")
}

func assert2(ok bool, msg string, args ...interface{}) {
	if !ok {
		panic(fmt.Sprintf(msg, args...))
	}
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
		err = nc.meta.WriteVInt(-1) // write EOF marker
	}
	success = err == nil
	return
}
