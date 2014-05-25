package index

import (
	"github.com/balzaczyy/golucene/core/index/model"
)

type DocFieldConsumerPerField interface {
	// Processes all occurrences of a single field
	processFields([]model.IndexableField, int) error
	abort()
	fieldInfo() *model.FieldInfo
}

// index/DocInverterPerField.java

type DocInverterPerField struct {
	_fieldInfo  *model.FieldInfo
	consumer    InvertedDocConsumerPerField
	endConsumer InvertedDocEndConsumerPerField
	docState    *docState
	fieldState  *FieldInvertState
}

func newDocInverterPerField(parent *DocInverter, fieldInfo *model.FieldInfo) *DocInverterPerField {
	ans := &DocInverterPerField{
		_fieldInfo: fieldInfo,
		docState:   parent.docState,
		fieldState: newFieldInvertState(fieldInfo.Name),
	}
	ans.consumer = parent.consumer.addField(ans, fieldInfo)
	ans.endConsumer = parent.endConsumer.addField(ans, fieldInfo)
	return ans
}

func (dipf *DocInverterPerField) abort() {
	defer dipf.endConsumer.abort()
	dipf.consumer.abort()
}

func (di *DocInverterPerField) processFields(fields []model.IndexableField, count int) error {
	di.fieldState.reset()

	doInvert, err := di.consumer.start(fields, count)
	if err != nil {
		return err
	}

	for i, field := range fields[:count] {
		fieldType := field.FieldType()

		// TODO FI: this should be "genericized" to querying consumer if
		// it wants to see this particular field tokenized.
		if fieldType.Indexed() && doInvert {
			panic("not implemented yet")
		}

		// LUCENE-2387: don't hang onto the field, so GC can recliam
		fields[i] = nil
	}

	err = di.consumer.finish()
	if err == nil {
		err = di.endConsumer.finish()
	}
	return err
}

func (dipf *DocInverterPerField) fieldInfo() *model.FieldInfo {
	return dipf._fieldInfo
}
