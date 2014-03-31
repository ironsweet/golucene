package index

import (
	"github.com/balzaczyy/golucene/core/util"
)

// index/DocConsumer.java

type DocConsumer interface {
	processDocument(fieldInfos *FieldInfosBuilder) error
	finishDocument() error
	flush(state SegmentWriteState) error
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

	// Hash table for all fields ever seen
	fieldHash []*DocFieldProcessorPerField

	docState *docState

	bytesUsed util.Counter
}

func newDocFieldProcessor(docWriter *DocumentsWriterPerThread,
	consumer DocFieldConsumer, storedConsumer StoredFieldsConsumer) *DocFieldProcessor {

	return &DocFieldProcessor{
		fieldHash:      make([]*DocFieldProcessorPerField, 2),
		docState:       docWriter.docState,
		codec:          docWriter.codec,
		bytesUsed:      docWriter._bytesUsed,
		consumer:       consumer,
		storedConsumer: storedConsumer,
	}
}

func (p *DocFieldProcessor) flush(state SegmentWriteState) error {
	panic("not implemented yet")
}

func (p *DocFieldProcessor) abort() {
	for _, field := range p.fieldHash {
		for field != nil {
			next := field.next
			field.abort()
			field = next
		}
	}
	p.storedConsumer.abort()
	p.consumer.abort()
	// assert2(err == nil, err.Error())
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

// index/DocFieldProcessorPerField.java

/* Holds all per thread, per field state. */
type DocFieldProcessorPerField struct {
	consumer DocFieldConsumerPerField

	next *DocFieldProcessorPerField
}

func (f *DocFieldProcessorPerField) abort() {
	f.consumer.abort()
}
