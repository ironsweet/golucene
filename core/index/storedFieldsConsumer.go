package index

import (
	"github.com/balzaczyy/golucene/core/util"
)

type StoredFieldsConsumer interface {
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

func (p *DocValuesProcessor) finishDocument() error {
	panic("not implemented yet")
}
