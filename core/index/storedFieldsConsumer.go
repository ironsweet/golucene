package index

import (
	"github.com/balzaczyy/golucene/core/util"
)

type StoredFieldsConsumer interface {
	// addField(docId int, field IndexableField, fieldInfo FieldInfo) error
	flush(state SegmentWriteState) error
	abort()
	// startDocument() error
	finishDocument() error
}

/* Just switches between two DocFieldConsumers */
type TwoStoredFieldsConsumers struct {
	first  StoredFieldsConsumer
	second StoredFieldsConsumer
}

func newTwoStoredFieldsConsumers(first, second StoredFieldsConsumer) *TwoStoredFieldsConsumers {
	return &TwoStoredFieldsConsumers{first, second}
}

func (p *TwoStoredFieldsConsumers) flush(state SegmentWriteState) error {
	err := p.first.flush(state)
	if err == nil {
		err = p.second.flush(state)
	}
	return err
}

func (p *TwoStoredFieldsConsumers) abort() {
	p.first.abort()
	p.second.abort()
}

func (p *TwoStoredFieldsConsumers) finishDocument() error {
	panic("not implemented yet")
}

// index/StoredFieldsProcessor.java

/* This is a StoredFieldsConsumer that writes stored fields */
type StoredFieldsProcessor struct {
	fieldsWriter StoredFieldsWriter
	lastDocId    int

	docWriter *DocumentsWriterPerThread

	docState *docState
	codec    Codec

	numStoredFields int
	storedFields    []IndexableField
	fieldInfos      []FieldInfo
}

func newStoredFieldsProcessor(docWriter *DocumentsWriterPerThread) *StoredFieldsProcessor {
	return &StoredFieldsProcessor{
		docWriter: docWriter,
		docState:  docWriter.docState,
		codec:     docWriter.codec,
	}
}

func (p *StoredFieldsProcessor) reset() {
	p.numStoredFields = 0
	p.storedFields = nil
	p.fieldInfos = nil
}

func (p *StoredFieldsProcessor) flush(state SegmentWriteState) error {
	panic("not implemented yet")
}

func (p *StoredFieldsProcessor) abort() {
	p.reset()

	if p.fieldsWriter != nil {
		p.fieldsWriter.abort()
		p.fieldsWriter = nil
		p.lastDocId = 0
	}
}

func (p *StoredFieldsProcessor) finishDocument() error {
	panic("not implemented yet")
}

// index/DocValuesProcessor.java

type DocValuesProcessor struct {
	writers   map[string]DocValuesWriter
	bytesUsed util.Counter
}

func newDocValuesProcessor(bytesUsed util.Counter) *DocValuesProcessor {
	return &DocValuesProcessor{make(map[string]DocValuesWriter), bytesUsed}
}

func (p *DocValuesProcessor) finishDocument() error {
	panic("not implemented yet")
}

func (p *DocValuesProcessor) flush(state SegmentWriteState) error {
	panic("not implemented yet")
}

func (p *DocValuesProcessor) abort() {
	for _, writer := range p.writers {
		writer.abort()
	}
	p.writers = make(map[string]DocValuesWriter)
}
