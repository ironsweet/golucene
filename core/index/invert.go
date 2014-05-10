package index

// index/DocInvertState.java

/*
Tracks the number and position / offset parameters of terms being
added to the index. The information collected in this class is also
used to calculate the normalization factor for a field
*/
type FieldInvertState struct {
	name string
}

/* Creates FieldInvertState for the specified field name. */
func newFieldInvertState(name string) *FieldInvertState {
	return &FieldInvertState{name: name}
}

/* Return the field's name */
func (st *FieldInvertState) Name() string {
	return st.name
}
