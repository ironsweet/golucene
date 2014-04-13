// This file has been automatically generated, DO NOT EDIT

		package packed

		import (
		)

		// Direct wrapping of 32-bits values to a backing array.
type Direct32 struct {
	PackedIntsReaderImpl
	values []int32
}

func newDirect32(valueCount int) *Direct32 {
	return &Direct32{
		PackedIntsReaderImpl: newPackedIntsReaderImpl(valueCount, 32),
		values: make([]int32, valueCount),
	}
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
