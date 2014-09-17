package packed

import (
	"github.com/balzaczyy/golucene/core/util"
)

/* A reader which has all its values equal to 0 (bitsPerValue = 0). */
type NilReader struct {
	valueCount int
}

func newNilReader(valueCount int) *NilReader {
	return &NilReader{
		valueCount: valueCount,
	}
}

func (r *NilReader) Get(int) int64 { return 0 }

func (r *NilReader) getBulk(index int, arr []int64) int {
	length := len(arr)
	assert2(length > 0, "len must be > 0 (got %v)", length)
	assert(index >= 0 && index < r.valueCount)
	if r.valueCount-index < length {
		length = r.valueCount - index
	}
	for i, _ := range arr {
		arr[i] = 0
	}
	return length
}

func (r *NilReader) Size() int {
	return r.valueCount
}

func (r *NilReader) RamBytesUsed() int64 {
	return util.AlignObjectSize(util.NUM_BYTES_OBJECT_HEADER + util.NUM_BYTES_INT)
}
