package index

/* An in-place update to a DocValues field. */
type DocValuesUpdate struct {
	valueSizeInBytes func() int64
}

func (u *DocValuesUpdate) sizeInBytes() int {
	panic("not implemented yet")
}

func (u *DocValuesUpdate) String() string {
	panic("not implemented yet")
}

/* An in-place update to a binary DocValues field */
func newBinaryDocValuesUpdate(term *Term, field string, value []byte) *DocValuesUpdate {
	panic("not implemented yet")
}

/* An in-plsace update to a numeric DocValues field */
func newNumericDocValuesUpdate(term *Term, field string, value int64) *DocValuesUpdate {
	panic("not implemented yet")
}
