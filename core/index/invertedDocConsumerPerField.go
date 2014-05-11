package index

import (
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/util"
)

// index/InvertedDocConsumerPerField.java

type InvertedDocConsumerPerField interface {
	// Called on hitting an aborting error
	abort()
}

const HASH_INIT_SIZE = 4

type TermsHashPerField struct {
	consumer TermsHashConsumerPerField

	termsHash *TermsHash

	nextPerField *TermsHashPerField
	docState     *docState
	fieldState   *FieldInvertState

	// Copied from our perThread
	intPool      *util.IntBlockPool
	bytePool     *util.ByteBlockPool
	termBytePool *util.ByteBlockPool

	streamCount   int
	numPostingInt int

	fieldInfo model.FieldInfo

	bytesHash *util.BytesRefHash

	bytesUsed util.Counter
}

func newTermsHashPerField(docInverterPerField *DocInverterPerField,
	termsHash *TermsHash, nextTermsHash *TermsHash,
	fieldInfo model.FieldInfo) *TermsHashPerField {

	ans := &TermsHashPerField{
		intPool:      termsHash.intPool,
		bytePool:     termsHash.bytePool,
		termBytePool: termsHash.termBytePool,
		docState:     termsHash.docState,
		termsHash:    termsHash,
		bytesUsed:    termsHash.bytesUsed,
		fieldState:   docInverterPerField.fieldState,
		fieldInfo:    fieldInfo,
	}
	ans.consumer = termsHash.consumer.addField(ans, fieldInfo)
	byteStarts := newPostingsBytesStartArray(ans, termsHash.bytesUsed)
	ans.bytesHash = util.NewBytesRefHash(termsHash.termBytePool, HASH_INIT_SIZE, byteStarts)
	ans.streamCount = ans.consumer.streamCount()
	ans.numPostingInt = 2 * ans.streamCount
	if nextTermsHash != nil {
		ans.nextPerField = nextTermsHash.addField(docInverterPerField, fieldInfo).(*TermsHashPerField)
	}
	return ans
}

func (h *TermsHashPerField) shrinkHash(targetSize int) {
	panic("not implemented yet")
}

func (h *TermsHashPerField) reset() {
	panic("not implemented yet")
}

func (h *TermsHashPerField) abort() {
	panic("not implemented yet")
}

type PostingsBytesStartArray struct {
	perField  *TermsHashPerField
	bytesUsed util.Counter
}

func newPostingsBytesStartArray(perField *TermsHashPerField,
	bytesUsed util.Counter) *PostingsBytesStartArray {
	return &PostingsBytesStartArray{perField, bytesUsed}
}
