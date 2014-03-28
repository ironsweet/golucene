package index

import (
	"github.com/balzaczyy/golucene/core/util"
)

// index/DocConsumer.java

type DocConsumer interface {
	processDocument(fieldInfos *FieldInfosBuilder) error
	finishDocument() error
	// flush(state *SegmentWriteState) error
	abort()
}

// index/DocFieldProcessor.java

/*
This is a DocConsumer that gathers all fields under the same name,
and calls per-field consumers to process field by field. This class
doesn't do any "real" work of its own: it just forwards the fields to
a DocFieldConsumer.
*/
type DocFieldProcessor struct {
	consumer       DocFieldConsumer
	storedConsumer StoredFieldsConsumer
	codec          Codec

	docState *docState

	bytesUsed util.Counter
}

func newDocFieldProcessor(docWriter *DocumentsWriterPerThread,
	consumer DocFieldConsumer, storedConsumer StoredFieldsConsumer) *DocFieldProcessor {

	return &DocFieldProcessor{
		docState:       docWriter.docState,
		codec:          docWriter.codec,
		bytesUsed:      docWriter.bytesUsed,
		consumer:       consumer,
		storedConsumer: storedConsumer,
	}
}

func (p *DocFieldProcessor) abort() {
	panic("not implemented yet")
}

func (p *DocFieldProcessor) processDocument(fieldInfos *FieldInfosBuilder) error {
	panic("not implemented yet")
}

func (p *DocFieldProcessor) finishDocument() (err error) {
	defer func() {
		err = mergeError(err, p.consumer.finishDocument())
	}()
	return p.storedConsumer.finishDocument()
}
