package index

import (
	"github.com/balzaczyy/golucene/core/index/model"
)

type TermsHashConsumerPerField interface {
	streamCount() int
}

// index/TermVectorsConsumerPerField.java

type TermVectorsConsumerPerField struct {
	termsHashPerField *TermsHashPerField
}

func (c *TermVectorsConsumerPerField) streamCount() int { return 2 }

func (c *TermVectorsConsumerPerField) shrinkHash() {
	panic("not implemented yet")
}

// TODO: break into separate freq and prox writers as codes; make
// separate container (tii/tis/skip/*) that can be configured as any
// number of files 1..N
type FreqProxTermsWriterPerField struct {
	parent            *FreqProxTermsWriter
	termsHashPerField *TermsHashPerField
	fieldInfo         model.FieldInfo
	docState          *docState
	fieldState        *FieldInvertState
	hasFreq           bool
	hasProx           bool
	hasOffsets        bool
}

func newFreqProxTermsWriterPerField(termsHashPerField *TermsHashPerField,
	parent *FreqProxTermsWriter, fieldInfo model.FieldInfo) *FreqProxTermsWriterPerField {
	ans := &FreqProxTermsWriterPerField{
		termsHashPerField: termsHashPerField,
		parent:            parent,
		fieldInfo:         fieldInfo,
		docState:          termsHashPerField.docState,
		fieldState:        termsHashPerField.fieldState,
	}
	ans.setIndexOptions(fieldInfo.IndexOptions())
	return ans
}

func (w *FreqProxTermsWriterPerField) streamCount() int {
	if !w.hasProx {
		return 1
	}
	return 2
}

/* Called after flush */
func (w *FreqProxTermsWriterPerField) reset() {
	panic("not implemented yet")
}

func (w *FreqProxTermsWriterPerField) setIndexOptions(indexOptions model.IndexOptions) {
	if n := int(indexOptions); n > 0 {
		w.hasFreq = n >= int(model.INDEX_OPT_DOCS_AND_FREQS)
		w.hasProx = n >= int(model.INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS)
		w.hasOffsets = n >= int(model.INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS)
	} else {
		// field could later be updated with indexed=true, so set everything on
		w.hasFreq = true
		w.hasProx = true
		w.hasOffsets = true
	}
}

/*
Walk through all unique text tokens (Posting instances) found in this
field and serialie them into a single RAM segment.
*/
func (w *FreqProxTermsWriterPerField) flush(fieldName string,
	consumer FieldsConsumer, state SegmentWriteState) error {
	panic("not implemented yet")
}
