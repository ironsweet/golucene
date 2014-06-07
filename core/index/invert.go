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

/* Get total number of terms in this field. */
func (st *FieldInvertState) Length() int {
	return st.length
}

/* Get the number of terms with positionIncrement == 0. */
func (st *FieldInvertState) NumOverlap() int {
	return st.numOverlap
}

/*
Get boost value. This is the cumulative product of document boost and
field boost for all field instances sharing the same field name.
*/
func (st *FieldInvertState) Boost() float32 {
	return st.boost
}

/* Return the field's name */
func (st *FieldInvertState) Name() string {
	return st.name
}
