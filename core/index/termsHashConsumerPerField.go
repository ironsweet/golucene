package index

import (
	. "github.com/balzaczyy/golucene/core/analysis/tokenattributes"
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/util"
)

type TermsHashConsumerPerField interface {
	start([]model.IndexableField, int) (bool, error)
	startField(model.IndexableField) error
	streamCount() int
	createPostingsArray(int) *ParallelPostingsArray
}

// index/TermVectorsConsumerPerField.java

type TermVectorsConsumerPerField struct {
	termsHashPerField *TermsHashPerField
	termsWriter       *TermVectorsConsumer
	fieldInfo         *model.FieldInfo
	docState          *docState
	fieldState        *FieldInvertState

	doVectors, doVectorPositions, doVectorOffsets, doVectorPayloads bool

	maxNumPostings int
	hasPayloads    bool // if enabled, and we actually saw any for this field
}

func newTermVectorsConsumerPerField(termsHashPerField *TermsHashPerField,
	termsWriter *TermVectorsConsumer, fieldInfo *model.FieldInfo) *TermVectorsConsumerPerField {
	return &TermVectorsConsumerPerField{
		termsHashPerField: termsHashPerField,
		termsWriter:       termsWriter,
		fieldInfo:         fieldInfo,
		docState:          termsHashPerField.docState,
		fieldState:        termsHashPerField.fieldState,
	}
}

func (c *TermVectorsConsumerPerField) streamCount() int { return 2 }

func (c *TermVectorsConsumerPerField) start(fields []model.IndexableField, count int) (bool, error) {
	c.doVectors = false
	c.doVectorPositions = false
	c.doVectorOffsets = false
	c.doVectorPayloads = false
	c.hasPayloads = false

	for _, field := range fields[:count] {
		t := field.FieldType()
		if t.Indexed() {
			if t.StoreTermVectors() {
				panic("not implemented yet")
			} else {
				assert2(!t.StoreTermVectorOffsets(),
					"cannot index term vector offsets when term vectors are not indexed (field='%v')",
					field.Name())
				assert2(!t.StoreTermVectorPositions(),
					"cannot index term vector positions when term vectors are not indexed (field='%v')",
					field.Name())
				assert2(!t.StoreTermVectorPayloads(),
					"cannot index term vector payloads when term vectors are not indexed (field='%v')",
					field.Name())
			}
		} else {
			panic("not implemented yet")
		}
	}

	if c.doVectors {
		c.termsWriter.hasVectors = true
		if c.termsHashPerField.bytesHash.Size() != 0 {
			// Only necessary if previous doc hit a non-aborting error
			// while writing vectors in this field:
			c.termsHashPerField.reset()
		}
	}

	// TODO: only if needed for performance
	// perThread.postingsCount = 0

	return c.doVectors, nil
}

func (c *TermVectorsConsumerPerField) finishDocument() error {
	panic("not implemented yet")
}

func (c *TermVectorsConsumerPerField) shrinkHash() {
	c.termsHashPerField.shrinkHash(c.maxNumPostings)
	c.maxNumPostings = 0
}

func (c *TermVectorsConsumerPerField) startField(f model.IndexableField) error {
	panic("not implemented yet")
}

func (c *TermVectorsConsumerPerField) createPostingsArray(size int) *ParallelPostingsArray {
	return newTermVectorsPostingArray(size)
}

type TermVectorsPostingArray struct {
	freqs         []int // How many times this term occurred in the current doc
	lastOffsets   []int // Last offset we saw
	lastPositions []int //Last position where this term occurred
}

func newTermVectorsPostingArray(size int) *ParallelPostingsArray {
	ans := new(TermVectorsPostingArray)
	return newParallelPostingsArray(ans, size)
}

func (arr *TermVectorsPostingArray) newInstance(size int) PostingsArray {
	return newTermVectorsPostingArray(size)
}

func (arr *TermVectorsPostingArray) copyTo(toArray PostingsArray, numToCopy int) {
	panic("not implemented yet")
}

func (arr *TermVectorsPostingArray) bytesPerPosting() int {
	return BYTES_PER_POSTING + 3*util.NUM_BYTES_INT
}

// TODO: break into separate freq and prox writers as codes; make
// separate container (tii/tis/skip/*) that can be configured as any
// number of files 1..N
type FreqProxTermsWriterPerField struct {
	parent            *FreqProxTermsWriter
	termsHashPerField *TermsHashPerField
	fieldInfo         *model.FieldInfo
	docState          *docState
	fieldState        *FieldInvertState
	hasFreq           bool
	hasProx           bool
	hasOffsets        bool
	payloadAttribute  PayloadAttribute
	offsetAttribute   OffsetAttribute
}

func newFreqProxTermsWriterPerField(termsHashPerField *TermsHashPerField,
	parent *FreqProxTermsWriter, fieldInfo *model.FieldInfo) *FreqProxTermsWriterPerField {
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

func (w *FreqProxTermsWriterPerField) start(fields []model.IndexableField, count int) (bool, error) {
	for _, field := range fields[:count] {
		if field.FieldType().Indexed() {
			return true, nil
		}
	}
	return false, nil
}

func (w *FreqProxTermsWriterPerField) startField(f model.IndexableField) error {
	atts := w.fieldState.attributeSource
	if atts.Has("PayloadAttribute") {
		w.payloadAttribute = atts.Get("PayloadAttribute").(PayloadAttribute)
	} else {
		w.payloadAttribute = nil
	}
	if w.hasOffsets {
		w.offsetAttribute = atts.Add("OffsetAttribute").(OffsetAttribute)
	} else {
		w.offsetAttribute = nil
	}
	return nil
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
	// fmt.Printf("PA init freqs=%v pos=%v offs=%v\n", writeFreqs, writeProx, writeOffsets)
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
