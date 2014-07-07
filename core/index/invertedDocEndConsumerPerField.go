package index

// import (
// 	"github.com/balzaczyy/golucene/core/index/model"
// )

// index/InvertedDocEndConsumerPerField.java

// type InvertedDocEndConsumerPerField interface {
// 	finish() error
// 	abort()
// }

// index/NormsConsumerPerField.java

// type NormsConsumerPerField struct {
// 	fieldInfo  *model.FieldInfo
// 	docState   *docState
// 	similarity Similarity
// 	fieldState *FieldInvertState
// 	consumer   *NumericDocValuesWriter
// }

// func newNormsConsumerPerField(docInverterPerField *DocInverterPerField,
// 	fieldInfo *model.FieldInfo, parent *NormsConsumer) *NormsConsumerPerField {
// 	return &NormsConsumerPerField{
// 		fieldInfo:  fieldInfo,
// 		docState:   docInverterPerField.docState,
// 		fieldState: docInverterPerField.fieldState,
// 		similarity: docInverterPerField.docState.similarity,
// 	}
// }

// func (nc *NormsConsumerPerField) finish() error {
// 	if nc.fieldInfo.IsIndexed() && !nc.fieldInfo.OmitsNorms() {
// 		if nc.consumer == nil {
// 			nc.fieldInfo.SetNormValueType(model.DOC_VALUES_TYPE_NUMERIC)
// 			nc.consumer = newNumericDocValuesWriter(nc.fieldInfo, nc.docState.docWriter._bytesUsed, false)
// 		}
// 		nc.consumer.addValue(nc.docState.docID, nc.similarity.ComputeNorm(nc.fieldState))
// 	}
// 	return nil
// }

// func (nc *NormsConsumerPerField) flush(state *model.SegmentWriteState,
// 	normsWriter DocValuesConsumer) error {

// 	docCount := state.SegmentInfo.DocCount()
// 	if nc.consumer == nil {
// 		return nil // nil type - not omitted but not written -
// 		// meaning the only docs that had
// 		// norms hit erros (but indexed=true is set...)
// 	}
// 	nc.consumer.finish(docCount)
// 	return nc.consumer.flush(state, normsWriter)
// }

// func (nc *NormsConsumerPerField) isEmpty() bool {
// 	return nc.consumer == nil
// }

// func (nc *NormsConsumerPerField) abort() {}
