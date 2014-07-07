package index

// import (
// 	"github.com/balzaczyy/golucene/core/index/model"
// )

// type DocFieldConsumer interface {
// 	// Called when DWPT decides to create a new segment
// 	flush(fieldsToFlush map[string]DocFieldConsumerPerField, state *model.SegmentWriteState) error
// 	// Called when an aborting error is hit
// 	abort()
// 	startDocument()
// 	addField(fi *model.FieldInfo) DocFieldConsumerPerField
// 	finishDocument() error
// }

// /*
// This is a DocFieldCosumer that inverts each field, separately, from a
// Document, and accepts an InvertedTermsConsumer to process those terms.
// */
// type DocInverter struct {
// 	consumer    InvertedDocConsumer
// 	endConsumer InvertedDocEndConsumer
// 	docState    *docState
// }

// func newDocInverter(docState *docState, consumer InvertedDocConsumer,
// 	endConsumer InvertedDocEndConsumer) *DocInverter {

// 	return &DocInverter{consumer, endConsumer, docState}
// }

// func (di *DocInverter) flush(fieldsToFlush map[string]DocFieldConsumerPerField,
// 	state *model.SegmentWriteState) error {
// 	childFieldsToFlush := make(map[string]InvertedDocConsumerPerField)
// 	endChildFieldsToFlush := make(map[string]InvertedDocEndConsumerPerField)
// 	for k, v := range fieldsToFlush {
// 		perField := v.(*DocInverterPerField)
// 		childFieldsToFlush[k] = perField.consumer
// 		endChildFieldsToFlush[k] = perField.endConsumer
// 	}
// 	err := di.consumer.flush(childFieldsToFlush, state)
// 	if err == nil {
// 		err = di.endConsumer.flush(endChildFieldsToFlush, state)
// 	}
// 	return err
// }

// func (di *DocInverter) startDocument() {
// 	di.consumer.startDocument()
// 	di.endConsumer.startDocument()
// }

// func (di *DocInverter) finishDocument() error {
// 	err := di.consumer.finishDocument()
// 	if err == nil {
// 		di.endConsumer.finishDocument()
// 	}
// 	return err
// }

// func (di *DocInverter) abort() {
// 	defer di.endConsumer.abort()
// 	di.consumer.abort()
// }

// func (di *DocInverter) addField(fi *model.FieldInfo) DocFieldConsumerPerField {
// 	return newDocInverterPerField(di, fi)
// }
