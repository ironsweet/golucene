package index

import (
	. "github.com/balzaczyy/golucene/core/analysis/tokenattributes"
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
	lastStartOffset  int
	lastPosition     int
	attributeSource  *util.AttributeSource

	offsetAttribute  OffsetAttribute
	posIncrAttribute PositionIncrementAttribute
	payloadAttribute PayloadAttribute
	termAttribute    TermToBytesRefAttribute
}

/* Creates FieldInvertState for the specified field name. */
func newFieldInvertState(name string) *FieldInvertState {
	return &FieldInvertState{name: name}
}

/* Re-initialize the state */
func (st *FieldInvertState) reset() {
	st.position = -1
	st.length = 0
	st.numOverlap = 0
	st.offset = 0
	st.maxTermFrequency = 0
	st.uniqueTermCount = 0
	st.boost = 1.0
	st.lastStartOffset = 0
	st.lastPosition = 0
}

/* Sets attributeSource to a new instance. */
func (st *FieldInvertState) setAttributeSource(attributeSource *util.AttributeSource) {
	if st.attributeSource != attributeSource {
		st.attributeSource = attributeSource
		st.termAttribute = attributeSource.Get("TermToBytesRefAttribute").(TermToBytesRefAttribute)
		st.posIncrAttribute = attributeSource.Add("PositionIncrementAttribute").(PositionIncrementAttribute)
		st.offsetAttribute = attributeSource.Add("OffsetAttribute").(OffsetAttribute)
		// st.payloadAttribute = attributeSource.Get("PayloadAttribute").(PayloadAttribute)
		st.payloadAttribute = nil
	}
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
