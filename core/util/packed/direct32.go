// This file has been automatically generated, DO NOT EDIT

		package packed

		import (
			"github.com/balzaczyy/golucene/core/util"
		)

		// Direct wrapping of 32-bits values to a backing array.
type Direct32 struct {
	*MutableImpl
	values []int32
}

func newDirect32(valueCount int) *Direct32 {
	ans := &Direct32{
		values: make([]int32, valueCount),
	}
	ans.MutableImpl = newMutableImpl(ans, valueCount, 32)
  return ans
}

func newDirect32FromInput(version int32, in DataInput, valueCount int) (r PackedIntsReader, err error) {
	ans := newDirect32(valueCount)
	for i, _ := range ans.values {
 		if ans.values[i], err = in.ReadInt(); err != nil {
			break
		}
	}
	if err == nil {
		// because packed ints have not always been byte-aligned
		remaining := PackedFormat(PACKED).ByteCount(version, int32(valueCount), 32) - 4*int64(valueCount)
		for i := int64(0); i < remaining; i++ {
			if _, err = in.ReadByte(); err != nil {
				break
			}
		}
	}
	return ans, err
}

func (d *Direct32) Get(index int) int64 {
	return int64(d.values[index]) & 0xFFFFFFFF
}

func (d *Direct32) Set(index int, value int64) {
	d.values[index] = int32(value)
}

func (d *Direct32) RamBytesUsed() int64 {
	return util.AlignObjectSize(
		util.NUM_BYTES_OBJECT_HEADER +
			2*util.NUM_BYTES_INT +
			util.NUM_BYTES_OBJECT_REF +
			util.SizeOf(d.values))
}

		func (d *Direct32) Clear() {
			for i, _ := range d.values {
				d.values[i] = 0
			}
		}

		func (d *Direct32) getBulk(index int, arr []int64) int {
			assert2(len(arr) > 0, "len must be > 0 (got %v)", len(arr))
			assert(index >= 0 && index < d.valueCount)

			gets := d.valueCount - index
			if len(arr) < gets {
				gets = len(arr)
			}
			for i, _ := range arr[:gets] {
				arr[i] = int64(d.values[index+i]) & 0xFFFFFFFF
			}
			return gets
		}

		func (d *Direct32) setBulk(index int, arr []int64) int {
			assert2(len(arr) > 0, "len must be > 0 (got %v)", len(arr))
			assert(index >= 0 && index < d.valueCount)

			sets := d.valueCount - index
			if len(arr) < sets {
				sets = len(arr)
			}
			for i, _ := range arr {
				d.values[index+i] = int32(arr[i])
			}
			return sets
		}

		func (d *Direct32) fill(from, to int, val int64) {
			assert(val == val & 0xFFFFFFFF)
			for i := from; i < to; i ++ {
				d.values[i] = int32(val)
			}
		}
				