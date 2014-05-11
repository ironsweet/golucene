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
	termsHashPerField *TermsHashPerField
	fieldInfo         model.FieldInfo
	hasProx           bool
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

/*
Walk through all unique text tokens (Posting instances) found in this
field and serialie them into a single RAM segment.
*/
func (w *FreqProxTermsWriterPerField) flush(fieldName string,
	consumer FieldsConsumer, state SegmentWriteState) error {
	panic("not implemented yet")
}
