package index

import (
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/util"
)

/* Default general purpose indexing chain, which handles indexing all types of fields */
type DefaultIndexChain struct {
	bytesUsed  util.Counter
	docState   *docState
	docWriter  *DocumentsWriterPerThread
	fieldInfos *FieldInfosBuilder

	// Writes postings and term vectors:
	termsHash TermsHash
}

func newDefaultIndexingChain(docWriter *DocumentsWriterPerThread) *DefaultIndexChain {
	termVectorsWriter := newTermVectorsConsumer(docWriter)
	return &DefaultIndexChain{
		docWriter:  docWriter,
		fieldInfos: docWriter.fieldInfos,
		docState:   docWriter.docState,
		bytesUsed:  docWriter._bytesUsed,
		termsHash:  newFreqProxTermsWriter(docWriter, termVectorsWriter),
	}
}

func (c *DefaultIndexChain) flush(state *SegmentWriteState) error {
	panic("not implemented yet")
}

func (c *DefaultIndexChain) abort() {
	panic("not implemented yet")
}

func (c *DefaultIndexChain) processDocument() error {
	panic("not implemented yet")
}
