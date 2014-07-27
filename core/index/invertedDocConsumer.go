package index

import (
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/util"
)

// type InvertedDocConsumer interface {
// 	// Abort (called after hitting abort error)
// 	abort()
// 	addField(*DocInverterPerField, *model.FieldInfo) InvertedDocConsumerPerField
// 	// Flush a new segment
// 	flush(map[string]InvertedDocConsumerPerField, *model.SegmentWriteState) error
// 	startDocument()
// 	finishDocument() error
// }

/*
This class is passed each token produced by the analyzer on each
field during indexing, and it stores these tokens in a hash table,
and allocates separate byte streams per token. Consumers of this
class, eg FreqProxTermsWriter and TermVectorsConsumer, write their
own byte streams under each term.
*/
type TermsHash interface {
	startDocument()
	finishDocument() error
	abort()
	reset()
	setTermBytePool(*util.ByteBlockPool)
	flush(map[string]TermsHashPerField, *SegmentWriteState) error

	fields() *TermsHashImpl // workaround abstract class

	TermsHashImplSPI
}

type TermsHashImplSPI interface {
	addField(*FieldInvertState, *FieldInfo) TermsHashPerField
}

type TermsHashImpl struct {
	spi TermsHashImplSPI

	nextTermsHash TermsHash

	intPool      *util.IntBlockPool
	bytePool     *util.ByteBlockPool
	termBytePool *util.ByteBlockPool
	bytesUsed    util.Counter

	docState *docState

	trackAllocations bool
}

func newTermsHash(spi TermsHashImplSPI,
	docWriter *DocumentsWriterPerThread,
	trackAllocations bool, nextTermsHash TermsHash) *TermsHashImpl {

	ans := &TermsHashImpl{
		spi:              spi,
		docState:         docWriter.docState,
		trackAllocations: trackAllocations,
		nextTermsHash:    nextTermsHash,
		intPool:          util.NewIntBlockPool(docWriter.intBlockAllocator),
		bytePool:         util.NewByteBlockPool(docWriter.byteBlockAllocator),
	}
	if trackAllocations {
		ans.bytesUsed = docWriter._bytesUsed
	} else {
		ans.bytesUsed = util.NewCounter()
	}
	if nextTermsHash != nil {
		ans.termBytePool = ans.bytePool
		nextTermsHash.setTermBytePool(ans.bytePool)
	}
	return ans
}

func (h *TermsHashImpl) fields() *TermsHashImpl {
	return h
}

func (hash *TermsHashImpl) setTermBytePool(p *util.ByteBlockPool) {
	hash.termBytePool = p
}

func (hash *TermsHashImpl) abort() {
	defer func() {
		if hash.nextTermsHash != nil {
			hash.nextTermsHash.abort()
		}
	}()
	hash.reset()
}

/* Clear all state */
func (hash *TermsHashImpl) reset() {
	// we don't reuse so we drop everything and don't fill with 0
	hash.intPool.Reset(false, false)
	hash.bytePool.Reset(false, false)
}

func (hash *TermsHashImpl) flush(fieldsToFlush map[string]TermsHashPerField,
	state *SegmentWriteState) error {

	if hash.nextTermsHash != nil {
		nextChildFields := make(map[string]TermsHashPerField)
		for k, v := range fieldsToFlush {
			nextChildFields[k] = v.next()
		}
		return hash.nextTermsHash.flush(nextChildFields, state)
	}
	return nil
}

// func (h *TermsHash) addField(docInverterPerField *DocInverterPerField,
// 	fieldInfo *model.FieldInfo) InvertedDocConsumerPerField {
// 	return newTermsHashPerField(docInverterPerField, h, h.nextTermsHash, fieldInfo)
// }

func (h *TermsHashImpl) finishDocument() error {
	if h.nextTermsHash != nil {
		return h.nextTermsHash.finishDocument()
	}
	return nil
}

func (h *TermsHashImpl) startDocument() {
	if h.nextTermsHash != nil {
		h.nextTermsHash.startDocument()
	}
}
