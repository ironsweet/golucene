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

	// Holds all fields seen in current doc
	_fields    []*DocFieldProcessorPerField
	fieldCount int

	// Hash table for all fields ever seen
	fieldHash       []*DocFieldProcessorPerField
	totalFieldCount int

	fieldGen int

	docState *docState

	bytesUsed util.Counter
}

func newDocFieldProcessor(docWriter *DocumentsWriterPerThread,
	consumer DocFieldConsumer, storedConsumer StoredFieldsConsumer) *DocFieldProcessor {

	assert(storedConsumer != nil)
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
	p.consumer.startDocument()
	p.storedConsumer.startDocument()

	p.fieldCount = 0

	thisFieldGen := p.fieldGen
	p.fieldGen++

	// Absorb any new fields first seen in this document. Also absort
	// any changes to fields we had already seen before (e.g. suddenly
	// turning on norms or vectors, etc.)

	for _, field := range p.docState.doc {
		fieldName := field.name()

		// Make sure we have a PerField allocated
		// TODO need a better hash code algorithm here
		hashPos := len(fieldName) / 2
		fp := p.fieldHash[hashPos]
		for fp != nil && fp.fieldInfo.Name != fieldName {
			fp = fp.next
		}

		if fp == nil {
			panic("not implemented yet")
		} else {
			panic("not implemented yet")
		}

		if thisFieldGen != fp.lastGen {
			panic("not implemented yet")
		}

		fp.addField(field)
		p.storedConsumer.addField(p.docState.docID, field, fp.fieldInfo)
	}
	return nil
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
	consumer  DocFieldConsumerPerField
	fieldInfo model.FieldInfo

	next    *DocFieldProcessorPerField
	lastGen int // -1
}

// func newDocFieldProcessorPerField(docFieldProcessor *DocFieldProcessor,
// 	fieldInfo model.FieldInfo) *DocFieldProcessorPerField {
// 	return &DocFieldProcessorPerField{
// 		consumer: docFieldProcessor.consumer.addField()
// 	}
// }

func (f *DocFieldProcessorPerField) addField(field IndexableField) {
	panic("not implemented yet")
}

func (f *DocFieldProcessorPerField) abort() {
	f.consumer.abort()
}
