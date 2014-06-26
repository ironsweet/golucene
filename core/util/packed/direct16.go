// This file has been automatically generated, DO NOT EDIT

		package packed

		import (
			"github.com/balzaczyy/golucene/core/util"
		)

		// Direct wrapping of 16-bits values to a backing array.
type Direct16 struct {
	*MutableImpl
	values []int16
}

func newDirect16(valueCount int) *Direct16 {
	ans := &Direct16{
		values: make([]int16, valueCount),
	}
	ans.MutableImpl = newMutableImpl(valueCount, 16, ans)
  return ans
}

func newDirect16FromInput(version int32, in DataInput, valueCount int) (r PackedIntsReader, err error) {
	ans := newDirect16(valueCount)
	for i, _ := range ans.values {
 		if ans.values[i], err = in.ReadShort(); err != nil {
			break
		}
	}
	if err == nil {
		// because packed ints have not always been byte-aligned
		remaining := PackedFormat(PACKED).ByteCount(version, int32(valueCount), 16) - 2*int64(valueCount)
		for i := int64(0); i < remaining; i++ {
			if _, err = in.ReadByte(); err != nil {
				break
			}
		}
	}
	return ans, err
}

func (d *Direct16) Get(index int) int64 {
	return int64(d.values[index]) & 0xFFFF
}

func (d *Direct16) Set(index int, value int64) {
	d.values[index] = int16(value)
}

func (d *Direct16) RamBytesUsed() int64 {
	return util.AlignObjectSize(
		util.NUM_BYTES_OBJECT_HEADER +
			2*util.NUM_BYTES_INT +
			util.NUM_BYTES_OBJECT_REF +
			util.SizeOf(d.values))
}
