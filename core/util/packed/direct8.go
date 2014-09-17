// This file has been automatically generated, DO NOT EDIT

		package packed

		import (
			"github.com/balzaczyy/golucene/core/util"
		)

		// Direct wrapping of 8-bits values to a backing array.
type Direct8 struct {
	*MutableImpl
	values []byte
}

func newDirect8(valueCount int) *Direct8 {
	ans := &Direct8{
		values: make([]byte, valueCount),
	}
	ans.MutableImpl = newMutableImpl(ans, valueCount, 8)
  return ans
}

func newDirect8FromInput(version int32, in DataInput, valueCount int) (r PackedIntsReader, err error) {
	ans := newDirect8(valueCount)
	if err = in.ReadBytes(ans.values[:valueCount]); err == nil {
		// because packed ints have not always been byte-aligned
		remaining := PackedFormat(PACKED).ByteCount(version, int32(valueCount), 8) - 1*int64(valueCount)
		for i := int64(0); i < remaining; i++ {
			if _, err = in.ReadByte(); err != nil {
				break
			}
		}
	}
	return ans, err
}

func (d *Direct8) Get(index int) int64 {
	return int64(d.values[index]) & 0xFF
}

func (d *Direct8) Set(index int, value int64) {
	d.values[index] = byte(value)
}

func (d *Direct8) RamBytesUsed() int64 {
	return util.AlignObjectSize(
		util.NUM_BYTES_OBJECT_HEADER +
			2*util.NUM_BYTES_INT +
			util.NUM_BYTES_OBJECT_REF +
			util.SizeOf(d.values))
}

		func (d *Direct8) Clear() {
			for i, _ := range d.values {
				d.values[i] = 0
			}
		}

		func (d *Direct8) getBulk(index int, arr []int64) int {
			assert2(len(arr) > 0, "len must be > 0 (got %v)", len(arr))
			assert(index >= 0 && index < d.valueCount)

			gets := d.valueCount - index
			if len(arr) < gets {
				gets = len(arr)
			}
			for i, _ := range arr[:gets] {
				arr[i] = int64(d.values[index+i]) & 0xFF
			}
			return gets
		}

		func (d *Direct8) setBulk(index int, arr []int64) int {
			assert2(len(arr) > 0, "len must be > 0 (got %v)", len(arr))
			assert(index >= 0 && index < d.valueCount)

			sets := d.valueCount - index
			if len(arr) < sets {
				sets = len(arr)
			}
			for i, _ := range arr {
				d.values[index+i] = byte(arr[i])
			}
			return sets
		}

		func (d *Direct8) fill(from, to int, val int64) {
			assert(val == val & 0xFF)
			for i := from; i < to; i ++ {
				d.values[i] = byte(val)
			}
		}
				