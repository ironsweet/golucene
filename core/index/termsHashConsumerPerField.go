package index

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/util"
)

type TermsHashConsumerPerField interface {
	streamCount() int
	createPostingsArray(int) *ParallelPostingsArray
}

// index/TermVectorsConsumerPerField.java

type TermVectorsConsumerPerField struct {
	termsHashPerField *TermsHashPerField
	termsWriter       *TermVectorsConsumer
	fieldInfo         model.FieldInfo
	docState          *docState
	fieldState        *FieldInvertState
}

func newTermVectorsConsumerPerField(termsHashPerField *TermsHashPerField,
	termsWriter *TermVectorsConsumer, fieldInfo model.FieldInfo) *TermVectorsConsumerPerField {
	return &TermVectorsConsumerPerField{
		termsHashPerField: termsHashPerField,
		termsWriter:       termsWriter,
		fieldInfo:         fieldInfo,
		docState:          termsHashPerField.docState,
		fieldState:        termsHashPerField.fieldState,
	}
}

func (c *TermVectorsConsumerPerField) streamCount() int { return 2 }

func (c *TermVectorsConsumerPerField) shrinkHash() {
	panic("not implemented yet")
}

func (c *TermVectorsConsumerPerField) createPostingsArray(size int) *ParallelPostingsArray {
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

func (w *FreqProxTermsWriterPerField) createPostingsArray(size int) *ParallelPostingsArray {
	return newFreqProxPostingsArray(size, w.hasFreq, w.hasProx, w.hasOffsets)
}

type FreqProxPostingsArray struct {
	termFreqs     []int // # times this term occurs in the current doc
	lastDocIDs    []int // Last docID where this term occurred
	lastDocCodes  []int // Code for prior doc
	lastPositions []int //Last position where this term occurred
	lastOffsets   []int // Last endOffsets where this term occurred
}

func newFreqProxPostingsArray(size int, writeFreqs, writeProx, writeOffsets bool) *ParallelPostingsArray {
	ans := new(FreqProxPostingsArray)
	if writeFreqs {
		ans.termFreqs = make([]int, size)
	}
	ans.lastDocIDs = make([]int, size)
	ans.lastDocCodes = make([]int, size)
	if writeProx {
		ans.lastPositions = make([]int, size)
		if writeOffsets {
			ans.lastOffsets = make([]int, size)
		}
	} else {
		assert(!writeOffsets)
	}
	fmt.Printf("PA init freqs=%v pos=%v offs=%v\n", writeFreqs, writeProx, writeOffsets)
	return newParallelPostingsArray(ans, size)
}

func (arr *FreqProxPostingsArray) newInstance(size int) PostingsArray {
	return newFreqProxPostingsArray(size, arr.termFreqs != nil,
		arr.lastPositions != nil, arr.lastOffsets != nil)
}

func (arr *FreqProxPostingsArray) copyTo(toArray PostingsArray, numToCopy int) {
	panic("not implemented yet")
}

func (arr *FreqProxPostingsArray) bytesPerPosting() int {
	bytes := BYTES_PER_POSTING + 2*util.NUM_BYTES_INT
	if arr.lastPositions != nil {
		bytes += util.NUM_BYTES_INT
	}
	if arr.lastOffsets != nil {
		bytes += util.NUM_BYTES_INT
	}
	if arr.termFreqs != nil {
		bytes += util.NUM_BYTES_INT
	}
	return bytes
}

/*
Walk through all unique text tokens (Posting instances) found in this
field and serialie them into a single RAM segment.
*/
func (w *FreqProxTermsWriterPerField) flush(fieldName string,
	consumer FieldsConsumer, state SegmentWriteState) error {
	panic("not implemented yet")
}
