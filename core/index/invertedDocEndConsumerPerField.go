package index

import (
	"github.com/balzaczyy/golucene/core/index/model"
)

// index/InvertedDocEndConsumerPerField.java

type InvertedDocEndConsumerPerField interface {
	// finish() error
	abort()
}

// index/NormsConsumerPerField.java

type NormsConsumerPerField struct {
	fieldInfo  model.FieldInfo
	docState   *docState
	similarity Similarity
	fieldState *FieldInvertState
	consumer   *NumericDocValuesWriter
}

func newNormsConsumerPerField(docInverterPerField *DocInverterPerField,
	fieldInfo model.FieldInfo, parent *NormsConsumer) *NormsConsumerPerField {
	return &NormsConsumerPerField{
		fieldInfo:  fieldInfo,
		docState:   docInverterPerField.docState,
		fieldState: docInverterPerField.fieldState,
		similarity: docInverterPerField.docState.similarity,
	}
}

func (nc *NormsConsumerPerField) flush(state SegmentWriteState, normsWriter DocValuesConsumer) error {
	panic("not implemented yet")
}

func (nc *NormsConsumerPerField) isEmpty() bool {
	return nc.consumer == nil
}

func (nc *NormsConsumerPerField) abort() {}
