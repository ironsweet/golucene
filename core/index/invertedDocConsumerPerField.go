package index

import (
	ta "github.com/balzaczyy/golucene/core/analysis/tokenattributes"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/util"
)

// index/InvertedDocConsumerPerField.java

// type InvertedDocConsumerPerField interface {
// 	// Called once per field, and is given all IndexableField
// 	// occurrences for this field in the document. Return true if you
// 	// wish to see inverted tokens for these fields:
// 	start([]IndexableField, int) (bool, error)
// 	// Called before a field instance is being processed
// 	startField(IndexableField)
// 	// Called once per inverted token
// 	add() error
// 	// Called once per field per document, after all IndexableFields
// 	// are inverted
// 	finish() error
// 	// Called on hitting an aborting error
// 	abort()
// }

const HASH_INIT_SIZE = 4

type TermsHashPerField interface {
	reset()
	finish() error
	start(IndexableField, bool) bool
}

type TermsHashPerFieldSPI interface {
	// Called when postings array is initialized or resized.
	newPostingsArray()
	// Creates a new postings array of the specified size.
	createPostingsArray(int) *ParallelPostingsArray
}

type TermsHashPerFieldImpl struct {
	spi TermsHashPerFieldSPI

	termsHash TermsHash

	nextPerField TermsHashPerField
	docState     *docState
	fieldState   *FieldInvertState
	termAtt      ta.TermToBytesRefAttribute
	termBytesRef *util.BytesRef

	// Copied from our perThread
	intPool      *util.IntBlockPool
	bytePool     *util.ByteBlockPool
	termBytePool *util.ByteBlockPool

	streamCount   int
	numPostingInt int

	fieldInfo *FieldInfo

	bytesHash *util.BytesRefHash

	postingsArray *ParallelPostingsArray
	bytesUsed     util.Counter

	doNextCall bool

	intUptos     []int
	intUptoStart int
}

/*
streamCount: how many streams this field stores per term. E.g.
doc(+freq) is 1 stream, prox+offset is a second.

NOTE: due to Go's embedded inheritance, it has to be invoked after it
is initialized and embedded by child class.
*/
func (h *TermsHashPerFieldImpl) _constructor(spi TermsHashPerFieldSPI,
	streamCount int, fieldState *FieldInvertState,
	termsHash TermsHash, nextPerField TermsHashPerField,
	fieldInfo *FieldInfo) {

	termsHashImpl := termsHash.fields()

	h.spi = spi
	h.intPool = termsHashImpl.intPool
	h.bytePool = termsHashImpl.bytePool
	h.termBytePool = termsHashImpl.termBytePool
	h.docState = termsHashImpl.docState
	h.termsHash = termsHash
	h.bytesUsed = termsHashImpl.bytesUsed
	h.fieldState = fieldState
	h.streamCount = streamCount
	h.numPostingInt = 2 * streamCount
	h.fieldInfo = fieldInfo
	h.nextPerField = nextPerField
	byteStarts := newPostingsBytesStartArray(h, h.bytesUsed)
	h.bytesHash = util.NewBytesRefHash(termsHashImpl.termBytePool, HASH_INIT_SIZE, byteStarts)
}

// func (h *TermsHashPerField) shrinkHash(targetSize int) {
// 	// Fully free the bytesHash on each flush but keep the pool
// 	// untouched. bytesHash.clear will clear the BytesStartArray and
// 	// in turn the ParallelPostingsArray too
// 	h.bytesHash.Clear(false)
// }

func (h *TermsHashPerFieldImpl) reset() {
	h.bytesHash.Clear(false)
	if h.nextPerField != nil {
		h.nextPerField.reset()
	}
}

// func (h *TermsHashPerField) abort() {
// 	h.reset()
// 	if h.nextPerField != nil {
// 		h.nextPerField.abort()
// 	}
// }

func (h *TermsHashPerFieldImpl) initReader(reader *ByteSliceReader, termId, stream int) {
	assert(stream < h.streamCount)
	intStart := h.postingsArray.intStarts[termId]
	ints := h.intPool.Buffers[intStart>>util.INT_BLOCK_SHIFT]
	upto := intStart & util.INT_BLOCK_MASK
	reader.init(h.bytePool,
		h.postingsArray.byteStarts[termId]+stream*util.FIRST_LEVEL_SIZE,
		ints[upto+stream])
}

/* Collapse the hash table & sort in-place; also sets sortedTermIDs to the results */
func (h *TermsHashPerFieldImpl) sortPostings(termComp func(a, b []byte) bool) []int {
	return h.bytesHash.Sort(termComp)
}

// func (h *TermsHashPerField) startField(f IndexableField) {
// 	h.termAtt = h.fieldState.attributeSource.Get("TermToBytesRefAttribute").(ta.TermToBytesRefAttribute)
// 	h.termBytesRef = h.termAtt.BytesRef()
// 	assert(h.termBytesRef != nil)
// 	h.consumer.startField(f)
// 	if h.nextPerField != nil {
// 		h.nextPerField.startField(f)
// 	}
// }

// func (h *TermsHashPerField) start(fields []IndexableField, count int) (bool, error) {
// 	var err error
// 	h.doCall, err = h.consumer.start(fields, count)
// 	if err != nil {
// 		return false, err
// 	}
// 	h.bytesHash.Reinit()
// 	if h.nextPerField != nil {
// 		h.doNextCall, err = h.nextPerField.start(fields, count)
// 		if err != nil {
// 			return false, err
// 		}
// 	}
// 	return h.doCall || h.doNextCall, nil
// }

/*
Secondary entry point (for 2nd & subsequent TermsHash), because token
text has already be "interned" into textStart, so we hash by textStart
*/
func (h *TermsHashPerFieldImpl) addFrom(textStart int) error {
	panic("not implemented yet")
}

/* Primary entry point (for first TermsHash) */
func (h *TermsHashPerFieldImpl) add() error {
	panic("not implemented yet")
	// // We are first in the chain so we must "intern" the term text into
	// // textStart address. Get the text & hash of this term.
	// termId, ok := h.bytesHash.Add(h.termBytesRef.Value, h.termAtt.FillBytesRef())
	// if !ok {
	// 	// Not enough room in current block. Just skip this term, to
	// 	// remain as robust as ossible during indexing. A TokenFilter can
	// 	// be inserted into the analyzer chain if other behavior is
	// 	// wanted (pruning the term to a prefix, returning an error, etc).
	// 	panic("not implemented yet")
	// 	return nil
	// }

	// if termId >= 0 { // new posting
	// 	h.bytesHash.ByteStart(termId)
	// 	// init stream slices
	// 	if h.numPostingInt+h.intPool.IntUpto > util.INT_BLOCK_SIZE {
	// 		h.intPool.NextBuffer()
	// 	}

	// 	if util.BYTE_BLOCK_SIZE-h.bytePool.ByteUpto < h.numPostingInt*util.FIRST_LEVEL_SIZE {
	// 		panic("not implemented yet")
	// 	}

	// 	h.intUptos = h.intPool.Buffer
	// 	h.intUptoStart = h.intPool.IntUpto
	// 	h.intPool.IntUpto += h.streamCount

	// 	h.postingsArray.intStarts[termId] = h.intUptoStart + h.intPool.IntOffset

	// 	for i := 0; i < h.streamCount; i++ {
	// 		upto := h.bytePool.NewSlice(util.FIRST_LEVEL_SIZE)
	// 		h.intUptos[h.intUptoStart+i] = upto + h.bytePool.ByteOffset
	// 	}
	// 	h.postingsArray.byteStarts[termId] = h.intUptos[h.intUptoStart]

	// 	err := h.consumer.newTerm(termId)
	// 	if err != nil {
	// 		return err
	// 	}
	// } else {
	// 	panic("not implemented yet")
	// }

	// if h.doNextCall {
	// 	return h.nextPerField.addFrom(h.postingsArray.textStarts[termId])
	// }
	// return nil
}

func (h *TermsHashPerFieldImpl) writeByte(stream int, b byte) {
	upto := h.intUptos[h.intUptoStart+stream]
	bytes := h.bytePool.Buffers[upto>>util.BYTE_BLOCK_SHIFT]
	assert(bytes != nil)
	offset := upto & util.BYTE_BLOCK_MASK
	if bytes[offset] != 0 {
		// end of slice; allocate a new one
		panic("not implemented yet")
	}
	bytes[offset] = b
	h.intUptos[h.intUptoStart+stream]++
}

func (h *TermsHashPerFieldImpl) writeVInt(stream, i int) {
	assert(stream < h.streamCount)
	for (i & ^0x7F) != 0 {
		h.writeByte(stream, byte((i&0x7F)|0x80))
	}
	h.writeByte(stream, byte(i))
}

func (h *TermsHashPerFieldImpl) finish() error {
	if h.nextPerField != nil {
		return h.nextPerField.finish()
	}
	return nil
}

/*
Start adding a new field instance; first is true if this is the first
time this field name was seen in the document.
*/
func (h *TermsHashPerFieldImpl) start(field IndexableField, first bool) bool {
	if h.termAtt = h.fieldState.termAttribute; h.termAtt != nil {
		// EmptyTokenStream can have nil term att
		h.termBytesRef = h.termAtt.BytesRef()
	}
	if h.nextPerField != nil {
		h.doNextCall = h.nextPerField.start(field, first)
	}
	return true
}

type PostingsBytesStartArray struct {
	perField  *TermsHashPerFieldImpl
	bytesUsed util.Counter
}

func newPostingsBytesStartArray(perField *TermsHashPerFieldImpl,
	bytesUsed util.Counter) *PostingsBytesStartArray {
	return &PostingsBytesStartArray{perField, bytesUsed}
}

func (ss *PostingsBytesStartArray) Init() []int {
	if ss.perField.postingsArray == nil {
		arr := ss.perField.spi.createPostingsArray(2)
		ss.perField.spi.newPostingsArray()
		ss.bytesUsed.AddAndGet(int64(arr.size * arr.bytesPerPosting()))
		ss.perField.postingsArray = arr
	}
	return ss.perField.postingsArray.textStarts
}

func (ss *PostingsBytesStartArray) Grow() []int {
	panic("not implemented yet")
}

func (ss *PostingsBytesStartArray) Clear() []int {
	if arr := ss.perField.postingsArray; arr != nil {
		ss.bytesUsed.AddAndGet(-int64(arr.size * arr.bytesPerPosting()))
		ss.perField.postingsArray = nil
		ss.perField.spi.newPostingsArray()
	}
	return nil
}

func (ss *PostingsBytesStartArray) BytesUsed() util.Counter {
	return ss.bytesUsed
}

// index/ParallelPostingsArray.java

const BYTES_PER_POSTING = 3 * util.NUM_BYTES_INT

type PostingsArray interface {
	bytesPerPosting() int
	newInstance(size int) PostingsArray
	copyTo(toArray PostingsArray, numToCopy int)
}

type ParallelPostingsArray struct {
	PostingsArray
	size       int
	textStarts []int
	intStarts  []int
	byteStarts []int
}

func newParallelPostingsArray(spi PostingsArray, size int) *ParallelPostingsArray {
	return &ParallelPostingsArray{
		PostingsArray: spi,
		size:          size,
		textStarts:    make([]int, size),
		intStarts:     make([]int, size),
		byteStarts:    make([]int, size),
	}
}

func (arr *ParallelPostingsArray) grow() *ParallelPostingsArray {
	panic("not implemented yet")
}
