// This file has been automatically generated, DO NOT EDIT

		package packed

		import (
		)

		// Direct wrapping of 64-bits values to a backing array.
type Direct64 struct {
	PackedIntsReaderImpl
	values []int64
}

func newDirect64(valueCount int) *Direct64 {
	return &Direct64{
		PackedIntsReaderImpl: newPackedIntsReaderImpl(valueCount, 64),
		values: make([]int64, valueCount),
	}
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
