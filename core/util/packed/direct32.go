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
	ans.MutableImpl = newMutableImpl(valueCount, 32, ans)
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
