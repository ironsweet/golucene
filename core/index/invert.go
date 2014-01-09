package index

// index/DocInvertState.java

/*
Tracks the number and position / offset parameters of terms being
added to the index. The information collected in this class is also
used to calculate the normalization factor for a field
*/
type FieldInvertState struct{}

func (st *FieldInvertState) Name() string {
	panic("not implemented yet")
}
