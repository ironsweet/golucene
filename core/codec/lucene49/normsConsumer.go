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
	iter func() func() (interface{}, bool)) (err error) {

	if err = nc.meta.WriteVInt(field.Number); err != nil {
		return
	}
	minValue, maxValue := int64(math.MaxInt64), int64(math.MinInt64)
	// TODO: more efficient?
	uniqueValues := newNormMap()

	count := int64(0)
	next := iter()
	for {
		nv, ok := next()
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

		if uniqueValues != nil && uniqueValues.add(v) && uniqueValues.size > 256 {
			uniqueValues = nil
		}

		count++
	}
	assert2(count == int64(nc.maxDoc),
		"illegal norms data for field %v, expected %v values, got %v",
		field.Name, nc.maxDoc, count)

	if uniqueValues != nil && uniqueValues.size == 1 {
		// 0 bpv
		if err = nc.meta.WriteByte(CONST_COMPRESSED); err != nil {
			return
		}
		if err = nc.meta.WriteLong(minValue); err != nil {
			return
		}
	} else if uniqueValues != nil {
		// small number of unique values; this is the typical case:
		// we only use bpv=1,2,4,8
		format := packed.PackedFormat(packed.PACKED_SINGLE_BLOCK)
		bitsPerValue := packed.BitsRequired(int64(uniqueValues.size) - 1)
		if bitsPerValue == 3 {
			bitsPerValue = 4
		} else if bitsPerValue > 4 {
			bitsPerValue = 8
		}

		if bitsPerValue == 8 && minValue >= 0 && maxValue <= 255 {
			if err = store.Stream(nc.meta).WriteByte(UNCOMPRESSED). // uncompressed []byte
										WriteLong(nc.data.FilePointer()).
										Close(); err != nil {
				return err
			}
			next = iter()
			for {
				nv, ok := next()
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
			if err = store.Stream(nc.meta).WriteByte(TABLE_COMPRESSED). // table-compressed
											WriteLong(nc.data.FilePointer()).
											Close(); err != nil {
				return err
			}
			if err = nc.data.WriteVInt(packed.VERSION_CURRENT); err != nil {
				return err
			}

			decode := uniqueValues.decodeTable()
			// upgrade to power of two sized array
			size := 1 << uint(bitsPerValue)
			if err = nc.data.WriteVInt(int32(size)); err != nil {
				return err
			}
			for _, v := range decode {
				if err = nc.data.WriteLong(v); err != nil {
					return err
				}
			}
			for i := len(decode); i < size; i++ {
				if err = nc.data.WriteLong(0); err != nil {
					return err
				}
			}

			if err = store.Stream(nc.data).WriteVInt(int32(format.Id())).
				WriteVInt(int32(bitsPerValue)).
				Close(); err != nil {
				return err
			}

			writer := packed.WriterNoHeader(nc.data, format, nc.maxDoc, bitsPerValue, packed.DEFAULT_BUFFER_SIZE)
			next = iter()
			for {
				nv, ok := next()
				if !ok {
					break
				}
				if err = writer.Add(int64(uniqueValues.ord(nv.(int64)))); err != nil {
					return err
				}
			}
			if err = writer.Finish(); err != nil {
				return err
			}
		}
	} else {
		panic("not implemented yet")
	}
	return nil
}

type Longs []int64

func (a Longs) Len() int           { return len(a) }
func (a Longs) Less(i, j int) bool { return a[i] < a[j] }
func (a Longs) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

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

/*
Specialized deduplication of long-ord for norms: 99.99999% of the
time this will be a single-byte range.
*/
type NormMap struct {
	// we use int16: at most we will add 257 values to this map before its rejected as too big above.
	singleByteRange []int16
	other           map[int64]int16
	size            int
}

func newNormMap() *NormMap {
	ans := &NormMap{
		singleByteRange: make([]int16, 256),
		other:           make(map[int64]int16),
	}
	for i, _ := range ans.singleByteRange {
		ans.singleByteRange[i] = -1
	}
	return ans
}

/* Adds an item to the mapping. Returns true if actually added. */
func (m *NormMap) add(l int64) bool {
	assert(m.size <= 256) // once we add > 256 values, we nullify the map in addNumericField and don't use this strategy
	if l >= math.MinInt8 && l <= math.MaxInt8 {
		index := int(l + 128)
		if previous := m.singleByteRange[index]; previous < 0 {
			m.singleByteRange[index] = int16(m.size)
			m.size++
			return true
		}
		return false
	}
	if _, ok := m.other[l]; !ok {
		m.other[l] = int16(m.size)
		m.size++
		return true
	}
	return false
}

/* Gets the ordinal for a previously added item. */
func (m *NormMap) ord(l int64) int {
	panic("niy")
}

/* Retrieves the ordinal table for previously added items. */
func (m *NormMap) decodeTable() []int64 {
	panic("niy")
}
