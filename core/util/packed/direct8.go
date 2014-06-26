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
	ans.MutableImpl = newMutableImpl(valueCount, 8, ans)
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
