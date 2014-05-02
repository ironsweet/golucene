package index

// index/InvertedDocConsumerPerField.java

type InvertedDocConsumerPerField interface {
	// Called on hitting an aborting error
	abort()
}

type TermsHashPerField struct {
	consumer TermsHashConsumerPerField

	nextPerField *TermsHashPerField
}

func (h *TermsHashPerField) abort() {
	panic("not implemented yet")
}
