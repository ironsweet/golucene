package index

import (
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
)

// index/DocConsumer.java

type DocConsumer interface {
	processDocument(fieldInfos *model.FieldInfosBuilder) error
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
	fieldHash       []*DocFieldProcessorPerField
	totalFieldCount int

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
	childFields := make(map[string]DocFieldConsumerPerField)
	for _, f := range p.fields() {
		childFields[f.fieldInfo().Name] = f
	}

	err := p.storedConsumer.flush(state)
	if err != nil {
		return err
	}
	err = p.consumer.flush(childFields, state)
	if err != nil {
		return err
	}

	// Impotant to save after asking consumer to flush so consumer can
	// alter the FieldInfo if necessary. E.g., FreqProxTermsWriter does
	// this with FieldInfo.storePayload.
	infosWriter := p.codec.FieldInfosFormat().FieldInfosWriter()
	return infosWriter(state.directory, state.segmentInfo.Name,
		state.fieldInfos, store.IO_CONTEXT_DEFAULT)
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

func (p *DocFieldProcessor) fields() []DocFieldConsumerPerField {
	var fields []DocFieldConsumerPerField
	for _, field := range p.fieldHash {
		for field != nil {
			fields = append(fields, field.consumer)
			field = field.next
		}
	}
	assert(len(fields) == p.totalFieldCount)
	return fields
}

func (p *DocFieldProcessor) processDocument(fieldInfos *model.FieldInfosBuilder) error {
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
