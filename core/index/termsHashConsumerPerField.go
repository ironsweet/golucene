package index

import (
	// "fmt"
	. "github.com/balzaczyy/golucene/core/analysis/tokenattributes"
	"github.com/balzaczyy/golucene/core/codec"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/util"
)

// type TermsHashConsumerPerField interface {
// 	start([]IndexableField, int) (bool, error)
// 	finish() error
// 	startField(IndexableField) error
// 	newTerm(int) error
// 	streamCount() int
// 	createPostingsArray(int) *ParallelPostingsArray
// }

// index/TermVectorsConsumerPerField.java

type TermVectorsConsumerPerField struct {
	*TermsHashPerFieldImpl

	termVectorsPostingsArray *TermVectorsPostingArray

	termsWriter *TermVectorsConsumer

	doVectors, doVectorPositions, doVectorOffsets, doVectorPayloads bool

	payloadAttribute PayloadAttribute
	offsetAttribute  OffsetAttribute
	hasPayloads      bool // if enabled, and we actually saw any for this field
}

func newTermVectorsConsumerPerField(invertState *FieldInvertState,
	termsWriter *TermVectorsConsumer,
	fieldInfo *FieldInfo) *TermVectorsConsumerPerField {

	ans := &TermVectorsConsumerPerField{
		termsWriter: termsWriter,
	}
	ans.TermsHashPerFieldImpl = new(TermsHashPerFieldImpl)
	ans.TermsHashPerFieldImpl._constructor(
		ans, 2, invertState, termsWriter, nil, fieldInfo)
	return ans
}

func (c *TermVectorsConsumerPerField) start(field IndexableField, first bool) bool {
	t := field.FieldType()
	assert(t.Indexed())

	if first {

		if c.bytesHash.Size() != 0 {
			// only necessary if previous doc hit a non-aborting error
			// while writing vectors in this field:
			c.reset()
		}

		c.bytesHash.Reinit()

		c.hasPayloads = false

		if c.doVectors = t.StoreTermVectors(); c.doVectors {
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

	if c.doVectors {
		panic("not implemented yet")
	}

	return c.doVectors
}

// /*
// Called once per field per document if term vectors are enabled, to
// write the vectors to RAMOutputStream, which is then quickly flushed
// to the real term vectors files in the Directory.
// */
// func (c *TermVectorsConsumerPerField) finish() error {
// 	if !c.doVectors || c.termsHashPerField.bytesHash.Size() == 0 {
// 		return nil
// 	}
// 	c.termsWriter.addFieldToFlush(c)
// 	return nil
// }

func (c *TermVectorsConsumerPerField) finishDocument() error {
	panic("not implemented yet")
}

// func (c *TermVectorsConsumerPerField) shrinkHash() {
// 	c.termsHashPerField.shrinkHash(c.maxNumPostings)
// 	c.maxNumPostings = 0
// }

// func (c *TermVectorsConsumerPerField) startField(f IndexableField) error {
// 	atts := c.fieldState.attributeSource
// 	if c.doVectorOffsets {
// 		c.offsetAttribute = atts.Add("OffsetAttribute").(OffsetAttribute)
// 	} else {
// 		c.offsetAttribute = nil
// 	}
// 	if c.doVectorPayloads && atts.Has("PayloadAttribute") {
// 		c.payloadAttribute = atts.Get("PayloadAttribute").(PayloadAttribute)
// 	} else {
// 		c.payloadAttribute = nil
// 	}
// 	return nil
// }

func (c *TermVectorsConsumerPerField) newTerm(termId int) {
	panic("not implemented yet")
}

func (c *TermVectorsConsumerPerField) addTerm(termid int) {
	panic("not implemented yet")
}

func (c *TermVectorsConsumerPerField) newPostingsArray() {
	if c.postingsArray != nil {
		c.termVectorsPostingsArray = c.postingsArray.PostingsArray.(*TermVectorsPostingArray)
	} else {
		c.termVectorsPostingsArray = nil
	}
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
	*TermsHashPerFieldImpl

	freqProxPostingsArray *FreqProxPostingsArray

	// parent            *FreqProxTermsWriter
	// termsHashPerField *TermsHashPerField
	// fieldInfo         *FieldInfo
	// docState          *docState
	// fieldState        *FieldInvertState

	hasFreq          bool
	hasProx          bool
	hasOffsets       bool
	hasPayloads      bool
	payloadAttribute PayloadAttribute
	offsetAttribute  OffsetAttribute

	sawPayloads bool // true if any token had a payload in the current segment
}

func newFreqProxTermsWriterPerField(invertState *FieldInvertState,
	termsHash TermsHash, fieldInfo *FieldInfo,
	nextPerField TermsHashPerField) *FreqProxTermsWriterPerField {

	indexOptions := fieldInfo.IndexOptions()
	assert(int(indexOptions) != 0)
	hasProx := indexOptions >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS
	ans := &FreqProxTermsWriterPerField{
		hasFreq:    indexOptions >= INDEX_OPT_DOCS_AND_FREQS,
		hasProx:    hasProx,
		hasOffsets: indexOptions >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS,
	}
	streamCount := map[bool]int{true: 2, false: 1}[hasProx]
	ans.TermsHashPerFieldImpl = new(TermsHashPerFieldImpl)
	ans.TermsHashPerFieldImpl._constructor(
		ans, streamCount, invertState,
		termsHash, nextPerField, fieldInfo,
	)
	return ans
}

// func (w *FreqProxTermsWriterPerField) streamCount() int {
// 	if !w.hasProx {
// 		return 1
// 	}
// 	return 2
// }

func (w *FreqProxTermsWriterPerField) finish() error {
	err := w.TermsHashPerFieldImpl.finish()
	if err == nil && w.sawPayloads {
		panic("not implemented yet")
		// w.fieldInfo.SetStorePayloads()
	}
	return err
}

/* Called after flush */
// func (w *FreqProxTermsWriterPerField) reset() {
// 	// record, up front, whether our in-RAM format will be
// 	// with or without term freqs:
// 	w.setIndexOptions(w.fieldInfo.IndexOptions())
// 	w.payloadAttribute = nil
// }

// func (w *FreqProxTermsWriterPerField) setIndexOptions(indexOptions IndexOptions) {
// 	if n := int(indexOptions); n > 0 {
// 		w.hasFreq = n >= int(INDEX_OPT_DOCS_AND_FREQS)
// 		w.hasProx = n >= int(INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS)
// 		w.hasOffsets = n >= int(INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS)
// 	} else {
// 		// field could later be updated with indexed=true, so set everything on
// 		w.hasFreq = true
// 		w.hasProx = true
// 		w.hasOffsets = true
// 	}
// }

// func (w *FreqProxTermsWriterPerField) start(fields []IndexableField, count int) (bool, error) {
// 	for _, field := range fields[:count] {
// 		if field.FieldType().Indexed() {
// 			return true, nil
// 		}
// 	}
// 	return false, nil
// }

func (w *FreqProxTermsWriterPerField) start(f IndexableField, first bool) bool {
	w.TermsHashPerFieldImpl.start(f, first)
	w.payloadAttribute = w.fieldState.payloadAttribute
	w.offsetAttribute = w.fieldState.offsetAttribute
	return true
}

func (w *FreqProxTermsWriterPerField) writeProx(termId, proxCode int) {
	if w.payloadAttribute == nil {
		w.writeVInt(1, proxCode<<1)
	} else {
		payload := w.payloadAttribute.Payload()
		if len(payload) > 0 {
			panic("not implemented yet")
		} else {
			w.writeVInt(1, proxCode<<1)
		}
	}

	assert(w.postingsArray.PostingsArray == w.freqProxPostingsArray)
	w.freqProxPostingsArray.lastPositions[termId] = w.fieldState.position
}

func (w *FreqProxTermsWriterPerField) writeOffsets(termId, offsetAccum int) {
	panic("not implemented yet")
}

func (w *FreqProxTermsWriterPerField) newTerm(termId int) {
	// First time we're seeing this term since the last flush
	w.docState.testPoint("FreqProxTermsWriterPerField.newTerm start")

	postings := w.freqProxPostingsArray
	assert(postings != nil)

	postings.lastDocIDs[termId] = w.docState.docID
	if !w.hasFreq {
		assert(postings.termFreqs == nil)
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
}

func (w *FreqProxTermsWriterPerField) addTerm(termId int) {
	w.docState.testPoint("FreqProxTermsWriterPerField.addTerm start")

	postings := w.freqProxPostingsArray

	assert(!w.hasFreq || postings.termFreqs[termId] > 0)

	if !w.hasFreq {
		panic("not implemented yet")
	} else if w.docState.docID != postings.lastDocIDs[termId] {
		assert2(w.docState.docID > postings.lastDocIDs[termId],
			"id: %v postings ID: %v termID: %v",
			w.docState.docID, postings.lastDocIDs[termId], termId)
		// Term not yet seen in the current doc but previously seen in
		// other doc(s) since the last flush

		// Now that we know doc freq for previous doc, write it & lastDocCode
		if 1 == postings.termFreqs[termId] {
			w.writeVInt(0, postings.lastDocCodes[termId]|1)
		} else {
			w.writeVInt(0, postings.lastDocCodes[termId])
			w.writeVInt(0, postings.termFreqs[termId])
		}

		// Init freq for the current document
		postings.termFreqs[termId] = 1
		if w.fieldState.maxTermFrequency < 1 {
			w.fieldState.maxTermFrequency = 1
		}
		postings.lastDocCodes[termId] = (w.docState.docID - postings.lastDocIDs[termId]) << 1
		postings.lastDocIDs[termId] = w.docState.docID
		if w.hasProx {
			w.writeProx(termId, w.fieldState.position)
			if w.hasOffsets {
				panic("niy")
			}
		} else {
			assert(!w.hasOffsets)
		}
		w.fieldState.uniqueTermCount++
	} else {
		postings.termFreqs[termId]++
		if n := postings.termFreqs[termId]; n > w.fieldState.maxTermFrequency {
			w.fieldState.maxTermFrequency = n
		}
		if w.hasProx {
			w.writeProx(termId, w.fieldState.position-postings.lastPositions[termId])
			if w.hasOffsets {
				w.writeOffsets(termId, w.fieldState.offset)
			}
		}
	}
}

func (w *FreqProxTermsWriterPerField) newPostingsArray() {
	if arr := w.postingsArray; arr != nil {
		w.freqProxPostingsArray = arr.PostingsArray.(*FreqProxPostingsArray)
	} else {
		w.freqProxPostingsArray = nil
	}
}

func (w *FreqProxTermsWriterPerField) createPostingsArray(size int) *ParallelPostingsArray {
	indexOptions := w.fieldInfo.IndexOptions()
	assert(indexOptions != 0)
	hasFreq := indexOptions >= INDEX_OPT_DOCS_AND_FREQS
	hasProx := indexOptions >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS
	hasOffsets := indexOptions >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
	return newFreqProxPostingsArray(size, hasFreq, hasProx, hasOffsets)
}

type FreqProxPostingsArray struct {
	*ParallelPostingsArray
	termFreqs     []int // # times this term occurs in the current doc
	lastDocIDs    []int // Last docID where this term occurred
	lastDocCodes  []int // Code for prior doc
	lastPositions []int //Last position where this term occurred
	lastOffsets   []int // Last endOffsets where this term occurred
}

func newFreqProxPostingsArray(size int, writeFreqs, writeProx, writeOffsets bool) *ParallelPostingsArray {
	ans := new(FreqProxPostingsArray)
	ans.ParallelPostingsArray = newParallelPostingsArray(ans, size)
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
	return ans.ParallelPostingsArray
}

func (arr *FreqProxPostingsArray) newInstance(size int) PostingsArray {
	return newFreqProxPostingsArray(size, arr.termFreqs != nil,
		arr.lastPositions != nil, arr.lastOffsets != nil)
}

func (arr *FreqProxPostingsArray) copyTo(toArray PostingsArray, numToCopy int) {
	to, ok := toArray.(*ParallelPostingsArray).PostingsArray.(*FreqProxPostingsArray)
	assert(ok)

	arr.ParallelPostingsArray.copyTo(toArray, numToCopy)

	copy(to.lastDocIDs[:numToCopy], arr.lastDocIDs[:numToCopy])
	copy(to.lastDocCodes[:numToCopy], arr.lastDocCodes[:numToCopy])
	if arr.lastPositions != nil {
		assert(to.lastPositions != nil)
		copy(to.lastPositions[:numToCopy], arr.lastPositions[:numToCopy])
	}
	if arr.lastOffsets != nil {
		assert(to.lastOffsets != nil)
		copy(to.lastOffsets[:numToCopy], arr.lastOffsets[:numToCopy])
	}
	if arr.termFreqs != nil {
		assert(to.termFreqs != nil)
		copy(to.termFreqs[:numToCopy], arr.termFreqs[:numToCopy])
	}
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
	consumer FieldsConsumer, state *SegmentWriteState) error {
	if !w.fieldInfo.IsIndexed() {
		return nil // nothing to flush, don't bother the codc with the unindexed field
	}

	termsConsumer, err := consumer.AddField(w.fieldInfo)
	if err != nil {
		return err
	}
	termComp := termsConsumer.Comparator()

	// CONFUSING: this.indexOptions holds the index options that were
	// current when we first saw this field. But it's posible this has
	// changed, e.g. when other documents are indexed that cause a
	// "downgrade" of the IndexOptions. So we must decode the in-RAM
	// buffer according to this.indexOptions, but then write the new
	// segment to the directory according to currentFieldIndexOptions:
	currentFieldIndexOptions := w.fieldInfo.IndexOptions()
	assert(int(currentFieldIndexOptions) != 0)

	writeTermFreq := int(currentFieldIndexOptions) >= int(INDEX_OPT_DOCS_AND_FREQS)
	writePositions := int(currentFieldIndexOptions) >= int(INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS)
	writeOffsets := int(currentFieldIndexOptions) >= int(INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS)

	readTermFreq := w.hasFreq
	readPositions := w.hasProx
	readOffsets := w.hasOffsets

	// fmt.Printf("flush readTF=%v readPos=%v readOffs=%v\n",
	// 	readTermFreq, readPositions, readOffsets)

	// Make sure FieldInfo.update is working correctly
	assert(!writeTermFreq || readTermFreq)
	assert(!writePositions || readPositions)
	assert(!writeOffsets || readOffsets)

	assert(!writeOffsets || writePositions)

	var segUpdates map[*Term]int
	if state.SegUpdates != nil && len(state.SegUpdates.(*BufferedUpdates).terms) > 0 {
		segUpdates = state.SegUpdates.(*BufferedUpdates).terms
	}

	termIDs := w.sortPostings(termComp)
	numTerms := w.bytesHash.Size()
	text := new(util.BytesRef)
	postings := w.freqProxPostingsArray
	freq := newByteSliceReader()
	prox := newByteSliceReader()

	visitedDocs := util.NewFixedBitSetOf(state.SegmentInfo.DocCount())
	sumTotalTermFreq := int64(0)
	sumDocFreq := int64(0)

	protoTerm := NewEmptyTerm(fieldName)
	for i := 0; i < numTerms; i++ {
		termId := termIDs[i]
		// fmt.Printf("term=%v\n", termId)
		// Get BytesRef
		textStart := postings.textStarts[termId]
		w.bytePool.SetBytesRef(text, textStart)

		w.initReader(freq, termId, 0)
		if readPositions || readOffsets {
			w.initReader(prox, termId, 1)
		}

		// TODO: really TermsHashPerField shold take over most of this
		// loop, including merge sort of terms from multiple threads and
		// interacting with the TermsConsumer, only calling out to us
		// (passing us the DocConsumer) to handle delivery of docs/positions

		postingsConsumer, err := termsConsumer.StartTerm(text.ToBytes())
		if err != nil {
			return err
		}

		delDocLimit := 0
		if segUpdates != nil {
			protoTerm.Bytes = text.ToBytes()
			if docIDUpto, ok := segUpdates[protoTerm]; ok {
				delDocLimit = docIDUpto
			}
		}

		// Now termStates has numToMerge FieldMergeStates which call
		// share the same term. Now we must interleave the docID streams.
		docFreq := 0
		totalTermFreq := int64(0)
		docId := 0

		for {
			// fmt.Println("  cycle")
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
			assert2(docId < state.SegmentInfo.DocCount(),
				"doc=%v maxDoc=%v", docId, state.SegmentInfo.DocCount())

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
				// we did record positions (& maybe payload) and/or offsets
				position := 0
				// offset := 0
				for j := 0; j < termFreq; j++ {
					var thisPayload []byte

					if readPositions {
						code, err := prox.ReadVInt()
						if err != nil {
							return err
						}
						position += int(uint(code) >> 1)

						if (code & 1) != 0 {
							panic("not implemented yet")
						}

						if readOffsets {
							panic("not implemented yet")
						} else if writePositions {
							err = postingsConsumer.AddPosition(position, thisPayload, -1, -1)
							if err != nil {
								return err
							}
						}
					}
				}
			}
			err = postingsConsumer.FinishDoc()
			if err != nil {
				return err
			}
		}
		err = termsConsumer.FinishTerm(text.ToBytes(), codec.NewTermStats(docFreq,
			map[bool]int64{true: totalTermFreq, false: -1}[writeTermFreq]))
		if err != nil {
			return err
		}
		sumTotalTermFreq += int64(totalTermFreq)
		sumDocFreq += int64(docFreq)
	}

	return termsConsumer.Finish(
		map[bool]int64{true: sumTotalTermFreq, false: -1}[writeTermFreq],
		sumDocFreq, visitedDocs.Cardinality())
}
