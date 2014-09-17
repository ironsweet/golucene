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
	ans.MutableImpl = newMutableImpl(ans, valueCount, 64)
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

		func (d *Direct64) Clear() {
			for i, _ := range d.values {
				d.values[i] = 0
			}
		}

		func (d *Direct64) getBulk(index int, arr []int64) int {
			assert2(len(arr) > 0, "len must be > 0 (got %v)", len(arr))
			assert(index >= 0 && index < d.valueCount)

			gets := d.valueCount - index
			if len(arr) < gets {
				gets = len(arr)
			}
			for i, _ := range arr[:gets] {
				arr[i] = int64(d.values[index+i])
			}
			return gets
		}

		func (d *Direct64) setBulk(index int, arr []int64) int {
			assert2(len(arr) > 0, "len must be > 0 (got %v)", len(arr))
			assert(index >= 0 && index < d.valueCount)

			sets := d.valueCount - index
			if len(arr) < sets {
				sets = len(arr)
			}
			for i, _ := range arr {
				d.values[index+i] = (arr[i])
			}
			return sets
		}

		func (d *Direct64) fill(from, to int, val int64) {
			assert(val == val)
			for i := from; i < to; i ++ {
				d.values[i] = (val)
			}
		}
				