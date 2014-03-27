package index

type DocFieldConsumer interface {
}

/*
This is a DocFieldCosumer that inverts each field, separately, from a
Document, and accepts an InvertedTermsConsumer to process those terms.
*/
type DocInverter struct {
}

func newDocInverter(docState *docState, consumer InvertedDocConsumer,
	endConsumer InvertedDocEndConsumer) *DocInverter {

	panic("not implemented yet")
}
