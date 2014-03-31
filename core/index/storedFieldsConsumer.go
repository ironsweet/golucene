package index

import (
	"github.com/balzaczyy/golucene/core/util"
)

type StoredFieldsConsumer interface {
	// addField(docId int, field IndexableField, fieldInfo FieldInfo) error
	// flush(state SegmentWriteState) error
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

func (p *TwoStoredFieldsConsumers) abort() {
	panic("not implemented yet")
}

func (p *TwoStoredFieldsConsumers) finishDocument() error {
	panic("not implemented yet")
}

// index/StoredFieldsProcessor.java

/* This is a StoredFieldsConsumer that writes stored fields */
type StoredFieldsProcessor struct {
	docWriter *DocumentsWriterPerThread

	docState *docState
	codec    Codec
}

func newStoredFieldsProcessor(docWriter *DocumentsWriterPerThread) *StoredFieldsProcessor {
	return &StoredFieldsProcessor{
		docWriter: docWriter,
		docState:  docWriter.docState,
		codec:     docWriter.codec,
	}
}

func (p *StoredFieldsProcessor) abort() {
	panic("not implemented yet")
}

func (p *StoredFieldsProcessor) finishDocument() error {
	panic("not implemented yet")
}

// index/DocValuesProcessor.java

type DocValuesProcessor struct {
	bytesUsed util.Counter
}

func newDocValuesProcessor(bytesUsed util.Counter) *DocValuesProcessor {
	return &DocValuesProcessor{bytesUsed: bytesUsed}
}

func (p *DocValuesProcessor) abort() {
	panic("not implemented yet")
}

func (p *DocValuesProcessor) finishDocument() error {
	panic("not implemented yet")
}
