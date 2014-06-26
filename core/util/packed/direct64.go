// This file has been automatically generated, DO NOT EDIT

		package packed

		import (
			"github.com/balzaczyy/golucene/core/util"
		)

		// Direct wrapping of 64-bits values to a backing array.
type Direct64 struct {
	*MutableImpl
	values []int64
}

func newDirect64(valueCount int) *Direct64 {
	ans := &Direct64{
		values: make([]int64, valueCount),
	}
	ans.MutableImpl = newMutableImpl(valueCount, 64, ans)
  return ans
}

func newDirect64FromInput(version int32, in DataInput, valueCount int) (r PackedIntsReader, err error) {
	ans := newDirect64(valueCount)
	for i, _ := range ans.values {
 		if ans.values[i], err = in.ReadLong(); err != nil {
			break
		}
	}
	if err == nil {
	}
	return ans, err
}

func (d *Direct64) Get(index int) int64 {
	return int64(d.values[index])
}

func (d *Direct64) Set(index int, value int64) {
	d.values[index] = int64(value)
}

func (d *Direct64) RamBytesUsed() int64 {
	return util.AlignObjectSize(
		util.NUM_BYTES_OBJECT_HEADER +
			2*util.NUM_BYTES_INT +
			util.NUM_BYTES_OBJECT_REF +
			util.SizeOf(d.values))
}
