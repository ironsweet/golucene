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
	next() TermsHashPerField
	reset()
	addFrom(int) error
	add() error
	finish() error
	start(IndexableField, bool) bool
}

type TermsHashPerFieldSPI interface {
	// Called when a term is seen for the first time.
	newTerm(int)
	// Called when a previously seen term is seen again.
	addTerm(int)
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

func (h *TermsHashPerFieldImpl) next() TermsHashPerField {
	return h.nextPerField
}

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

// Simpler version of Lucene's own method
func utf8ToString(iso8859_1_buf []byte) string {
	buf := make([]rune, len(iso8859_1_buf))
	for i, b := range iso8859_1_buf {
		buf[i] = rune(b)
	}
	return string(buf)
}

/*
Called once per inverted token. This is the primary entry point (for
first TermsHash); postings use this API.
*/
func (h *TermsHashPerFieldImpl) add() (err error) {
	h.termAtt.FillBytesRef()

	// We are first in the chain so we must "intern" the term text into
	// textStart address. Get the text & hash of this term.
	var termId int
	if termId, err = h.bytesHash.Add(h.termBytesRef.ToBytes()); err != nil {
		return
	}

	// fmt.Printf("add term=%v doc=%v termId=%v\n",
	// 	string(h.termBytesRef.Value), h.docState.docID, termId)

	if termId >= 0 { // new posting
		h.bytesHash.ByteStart(termId)
		// init stream slices
		if h.numPostingInt+h.intPool.IntUpto > util.INT_BLOCK_SIZE {
			h.intPool.NextBuffer()
		}

		if util.BYTE_BLOCK_SIZE-h.bytePool.ByteUpto < h.numPostingInt*util.FIRST_LEVEL_SIZE {
			h.bytePool.NextBuffer()
		}

		h.intUptos = h.intPool.Buffer
		h.intUptoStart = h.intPool.IntUpto
		h.intPool.IntUpto += h.streamCount

		h.postingsArray.intStarts[termId] = h.intUptoStart + h.intPool.IntOffset

		for i := 0; i < h.streamCount; i++ {
			upto := h.bytePool.NewSlice(util.FIRST_LEVEL_SIZE)
			h.intUptos[h.intUptoStart+i] = upto + h.bytePool.ByteOffset
		}
		h.postingsArray.byteStarts[termId] = h.intUptos[h.intUptoStart]

		h.spi.newTerm(termId)

	} else {
		termId = (-termId) - 1
		intStart := h.postingsArray.intStarts[termId]
		h.intUptos = h.intPool.Buffers[intStart>>util.INT_BLOCK_SHIFT]
		h.intUptoStart = intStart & util.INT_BLOCK_MASK
		h.spi.addTerm(termId)
	}

	if h.doNextCall {
		return h.nextPerField.addFrom(h.postingsArray.textStarts[termId])
	}
	return nil
}

func (h *TermsHashPerFieldImpl) writeByte(stream int, b byte) {
	upto := h.intUptos[h.intUptoStart+stream]
	bytes := h.bytePool.Buffers[upto>>util.BYTE_BLOCK_SHIFT]
	assert(bytes != nil)
	offset := upto & util.BYTE_BLOCK_MASK
	if bytes[offset] != 0 {
		// end of slice; allocate a new one
		offset = h.bytePool.AllocSlice(bytes, offset)
		bytes = h.bytePool.Buffer
		h.intUptos[h.intUptoStart+stream] = offset + h.bytePool.ByteOffset
	}
	bytes[offset] = b
	h.intUptos[h.intUptoStart+stream]++
}

func (h *TermsHashPerFieldImpl) writeVInt(stream, i int) {
	assert(stream < h.streamCount)
	for (i & ^0x7F) != 0 {
		h.writeByte(stream, byte((i&0x7F)|0x80))
		i = int(uint(i) >> 7)
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
		ss.perField.postingsArray = arr
		ss.perField.spi.newPostingsArray()
		ss.bytesUsed.AddAndGet(int64(arr.size * arr.bytesPerPosting()))
	}
	return ss.perField.postingsArray.textStarts
}

func (ss *PostingsBytesStartArray) Grow() []int {
	postingsArray := ss.perField.postingsArray
	oldSize := postingsArray.size
	postingsArray = postingsArray.grow()
	ss.perField.postingsArray = postingsArray
	ss.perField.spi.newPostingsArray()
	ss.bytesUsed.AddAndGet(int64(postingsArray.bytesPerPosting() * (postingsArray.size - oldSize)))
	return postingsArray.textStarts
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

func (arr *ParallelPostingsArray) bytesPerPosting() int {
	return BYTES_PER_POSTING
}

func (arr *ParallelPostingsArray) newInstance(size int) PostingsArray { // *ParallelPostingsArray
	ans := newParallelPostingsArray(nil, size)
	ans.PostingsArray = ans
	return ans
}

func (arr *ParallelPostingsArray) grow() *ParallelPostingsArray {
	newSize := util.Oversize(arr.size+1, arr.PostingsArray.bytesPerPosting())
	newArray := arr.PostingsArray.newInstance(newSize)
	arr.PostingsArray.copyTo(newArray, arr.size)
	return newArray.(*ParallelPostingsArray)
}

func (arr *ParallelPostingsArray) copyTo(toArray PostingsArray, numToCopy int) {
	to := toArray.(*ParallelPostingsArray)
	copy(to.textStarts[:numToCopy], arr.textStarts[:numToCopy])
	copy(to.intStarts[:numToCopy], arr.intStarts[:numToCopy])
	copy(to.byteStarts[:numToCopy], arr.byteStarts[:numToCopy])
}
