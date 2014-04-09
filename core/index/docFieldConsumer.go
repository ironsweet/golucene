package index

type DocFieldConsumer interface {
	// Called when DWPT decides to create a new segment
	flush(fieldsToFlush map[string]DocFieldConsumerPerField, state SegmentWriteState) error
	// Called when an aborting error is hit
	abort()
	startDocument()
	// addField(fi FieldInfo) DocFieldConsumerPerField
	finishDocument() error
}

/*
This is a DocFieldCosumer that inverts each field, separately, from a
Document, and accepts an InvertedTermsConsumer to process those terms.
*/
type DocInverter struct {
	consumer    InvertedDocConsumer
	endConsumer InvertedDocEndConsumer
	docState    *docState
}

func newDocInverter(docState *docState, consumer InvertedDocConsumer,
	endConsumer InvertedDocEndConsumer) *DocInverter {

	return &DocInverter{consumer, endConsumer, docState}
}

func (di *DocInverter) flush(fieldsToFlush map[string]DocFieldConsumerPerField, state SegmentWriteState) error {
	panic("not implemented yet")
}

func (di *DocInverter) startDocument() {
	// di.consumer.startDocument()
	// di.endConsumer.startDocument()
	panic("not implemented yet")
}

func (di *DocInverter) finishDocument() error {
	panic("not implemented yet")
}

func (di *DocInverter) abort() {
	defer di.endConsumer.abort()
	di.consumer.abort()

}
