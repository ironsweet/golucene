package index

import (
	"github.com/balzaczyy/golucene/core/util"
)

// index/DocInvertState.java

/*
Tracks the number and position / offset parameters of terms being
added to the index. The information collected in this class is also
used to calculate the normalization factor for a field
*/
type FieldInvertState struct {
	name             string
	position         int
	length           int
	numOverlap       int
	offset           int
	maxTermFrequency int
	uniqueTermCount  int
	boost            float32
	attributeSource  *util.AttributeSource
}

/* Creates FieldInvertState for the specified field name. */
func newFieldInvertState(name string) *FieldInvertState {
	return &FieldInvertState{name: name}
}

/* Re-initialize the state */
func (st *FieldInvertState) reset() {
	st.position = 0
	st.length = 0
	st.numOverlap = 0
	st.offset = 0
	st.maxTermFrequency = 0
	st.uniqueTermCount = 0
	st.boost = 1.0
	st.attributeSource = nil
}

/* Return the field's name */
func (st *FieldInvertState) Name() string {
	return st.name
}
