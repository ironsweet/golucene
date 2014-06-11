package index

import (
	"fmt"
	. "github.com/balzaczyy/golucene/core/analysis/tokenattributes"
	"github.com/balzaczyy/golucene/core/codec"
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/util"
)

type TermsHashConsumerPerField interface {
	start([]model.IndexableField, int) (bool, error)
	finish() error
	startField(model.IndexableField) error
	newTerm(int) error
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

	maxNumPostings   int
	payloadAttribute PayloadAttribute
	offsetAttribute  OffsetAttribute
	hasPayloads      bool // if enabled, and we actually saw any for this field
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

/*
Called once per field per document if term vectors are enabled, to
write the vectors to RAMOutputStream, which is then quickly flushed
to the real term vectors files in the Directory.
*/
func (c *TermVectorsConsumerPerField) finish() error {
	if !c.doVectors || c.termsHashPerField.bytesHash.Size() == 0 {
		return nil
	}
	c.termsWriter.addFieldToFlush(c)
	return nil
}

func (c *TermVectorsConsumerPerField) finishDocument() error {
	panic("not implemented yet")
}

func (c *TermVectorsConsumerPerField) shrinkHash() {
	c.termsHashPerField.shrinkHash(c.maxNumPostings)
	c.maxNumPostings = 0
}

func (c *TermVectorsConsumerPerField) startField(f model.IndexableField) error {
	atts := c.fieldState.attributeSource
	if c.doVectorOffsets {
		c.offsetAttribute = atts.Add("OffsetAttribute").(OffsetAttribute)
	} else {
		c.offsetAttribute = nil
	}
	if c.doVectorPayloads && atts.Has("PayloadAttribute") {
		c.payloadAttribute = atts.Get("PayloadAttribute").(PayloadAttribute)
	} else {
		c.payloadAttribute = nil
	}
	return nil
}

func (c *TermVectorsConsumerPerField) newTerm(termId int) error {
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
	hasPayloads       bool
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

func (w *FreqProxTermsWriterPerField) finish() error {
	if w.hasPayloads {
		panic("not implemented yet")
		// w.fieldInfo.SetStorePayloads()
	}
	return nil
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

func (w *FreqProxTermsWriterPerField) writeProx(termId, proxCode int) {
	assert(w.hasProx)
	var payload []byte
	if w.payloadAttribute != nil {
		payload = w.payloadAttribute.Payload()
	}

	if len(payload) > 0 {
		panic("not implemented yet")
	} else {
		w.termsHashPerField.writeVInt(1, proxCode<<1)
	}

	postings := w.termsHashPerField.postingsArray.PostingsArray.(*FreqProxPostingsArray)
	postings.lastPositions[termId] = w.fieldState.position
}

func (w *FreqProxTermsWriterPerField) writeOffsets(termId, offsetAccum int) {
	panic("not implemented yet")
}

func (w *FreqProxTermsWriterPerField) newTerm(termId int) error {
	// First time we're seeing this term since the last flush
	w.docState.testPoint("FreqProxTermsWriterPerField.newTerm start")

	postings := w.termsHashPerField.postingsArray.PostingsArray.(*FreqProxPostingsArray)
	postings.lastDocIDs[termId] = w.docState.docID
	if !w.hasFreq {
		postings.lastDocCodes[termId] = w.docState.docID
	} else {
		postings.lastDocCodes[termId] = w.docState.docID << 1
		postings.termFreqs[termId] = 1
		if w.hasProx {
			w.writeProx(termId, w.fieldState.position)
			if w.hasOffsets {
				w.writeOffsets(termId, w.fieldState.offset)
			}
		} else {
			assert(!w.hasOffsets)
		}
	}
	if 1 > w.fieldState.maxTermFrequency {
		w.fieldState.maxTermFrequency = 1
	}
	w.fieldState.uniqueTermCount++
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
	if !w.fieldInfo.IsIndexed() {
		return nil // nothing to flush, don't bother the codc with the unindexed field
	}

	termsConsumer, err := consumer.addField(w.fieldInfo)
	if err != nil {
		return err
	}
	termComp := termsConsumer.comparator()

	// CONFUSING: this.indexOptions holds the index options that were
	// current when we first saw this field. But it's posible this has
	// changed, e.g. when other documents are indexed that cause a
	// "downgrade" of the IndexOptions. So we must decode the in-RAM
	// buffer according to this.indexOptions, but then write the new
	// segment to the directory according to currentFieldIndexOptions:
	currentFieldIndexOptions := w.fieldInfo.IndexOptions()
	assert(int(currentFieldIndexOptions) != 0)

	writeTermFreq := int(currentFieldIndexOptions) >= int(model.INDEX_OPT_DOCS_AND_FREQS)
	writePositions := int(currentFieldIndexOptions) >= int(model.INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS)
	writeOffsets := int(currentFieldIndexOptions) >= int(model.INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS)

	readTermFreq := w.hasFreq
	readPositions := w.hasProx
	readOffsets := w.hasOffsets

	fmt.Printf("flush readTF=%v readPos=%v readOffs=%v\n",
		readTermFreq, readPositions, readOffsets)

	// Make sure FieldInfo.update is working correctly
	assert(!writeTermFreq || readTermFreq)
	assert(!writePositions || readPositions)
	assert(!writeOffsets || readOffsets)

	assert(!writeOffsets || writePositions)

	var segDeletes map[*Term]int
	if state.segDeletes != nil && len(state.segDeletes.terms) > 0 {
		segDeletes = state.segDeletes.terms
	}

	termIDs := w.termsHashPerField.sortPostings(termComp)
	numTerms := w.termsHashPerField.bytesHash.Size()
	text := new(util.BytesRef)
	postings2 := w.termsHashPerField.postingsArray
	postings := postings2.PostingsArray.(*FreqProxPostingsArray)
	freq := newByteSliceReader()
	prox := newByteSliceReader()

	visitedDocs := util.NewFixedBitSetOf(state.segmentInfo.DocCount())
	sumTotalTermFreq := int64(0)
	sumDocFreq := int64(0)

	protoTerm := NewEmptyTerm(fieldName)
	for i := 0; i < numTerms; i++ {
		termId := termIDs[i]
		fmt.Printf("term=%v\n", termId)
		// Get BytesRef
		textStart := postings2.textStarts[termId]
		w.termsHashPerField.bytePool.SetBytesRef(text, textStart)

		w.termsHashPerField.initReader(freq, termId, 0)
		if readPositions || readOffsets {
			w.termsHashPerField.initReader(prox, termId, 1)
		}

		// TODO: really TermsHashPerField shold take over most of this
		// loop, including merge sort of terms from multiple threads and
		// interacting with the TermsConsumer, only calling out to us
		// (passing us the DocConsumer) to handle delivery of docs/positions

		postingsConsumer, err := termsConsumer.startTerm(text.Value)
		if err != nil {
			return err
		}

		delDocLimit := 0
		if segDeletes != nil {
			protoTerm.Bytes = text.Value
			if docIDUpto, ok := segDeletes[protoTerm]; ok {
				delDocLimit = docIDUpto
			}
		}

		// Now termStates has numToMerge FieldMergeStates which call
		// share the same term. Now we must interleave the docID streams.
		docFreq := 0
		totalTermFreq := int64(0)
		docId := 0

		for {
			fmt.Println("  cycle")
			var termFreq int
			if freq.eof() {
				if postings.lastDocCodes[termId] != -1 {
					// return last doc
					docId = postings.lastDocIDs[termId]
					if readTermFreq {
						termFreq = postings.termFreqs[termId]
					} else {
						termFreq = -1
					}
					postings.lastDocCodes[termId] = -1
				} else {
					// EOF
					break
				}
			} else {
				code, err := freq.ReadVInt()
				if err != nil {
					return err
				}
				if !readTermFreq {
					docId += int(code)
					termFreq = -1
				} else {
					docId += int(uint(code) >> 1)
					if (code & 1) != 0 {
						termFreq = 1
					} else {
						n, err := freq.ReadVInt()
						if err != nil {
							return err
						}
						termFreq = int(n)
					}
				}

				assert(docId != postings.lastDocIDs[termId])
			}

			docFreq++
			assert2(docId < state.segmentInfo.DocCount(),
				"doc=%v maxDoc=%v", docId, state.segmentInfo.DocCount())

			// NOTE: we could check here if the docID was deleted, and skip
			// it. However, this is somewhat dangerous because it can yield
			// non-deterministic behavior since we may see the docID before
			// we see the term that caused it to be deleted. This would
			// mean some (but not all) of its postings may make it into the
			// index, which'd alter the docFreq for those terms. We could
			// fix this by doing two passes, i.e. first sweep marks all del
			// docs, and 2nd sweep does the real flush, but I suspect
			// that'd add too much time to flush.
			visitedDocs.Set(docId)
			err := postingsConsumer.StartDoc(docId,
				map[bool]int{true: termFreq, false: -1}[writeTermFreq])
			if err != nil {
				return err
			}
			if docId < delDocLimit {
				panic("not implemented yet")
			}

			totalTermFreq += int64(termFreq)

			// Carefully copy over the prox + payload info, changing the
			// format to match Lucene's segment format.

			if readPositions || readOffsets {
				panic("not implemented yet")
			}
			err = postingsConsumer.FinishDoc()
			if err != nil {
				return err
			}
		}
		err = termsConsumer.finishTerm(text.Value, codec.NewTermStats(docFreq,
			map[bool]int64{true: totalTermFreq, false: -1}[writeTermFreq]))
		if err != nil {
			return err
		}
		sumTotalTermFreq += int64(totalTermFreq)
		sumDocFreq += int64(docFreq)
	}

	return termsConsumer.finish(
		map[bool]int64{true: sumTotalTermFreq, false: -1}[writeTermFreq],
		sumDocFreq, visitedDocs.Cardinality())
}
