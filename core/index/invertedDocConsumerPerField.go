package index

import (
	"github.com/balzaczyy/golucene/core/util"
)

// index/InvertedDocConsumerPerField.java

type InvertedDocConsumerPerField interface {
	// Called on hitting an aborting error
	abort()
}

type TermsHashPerField struct {
	consumer TermsHashConsumerPerField

	termsHash *TermsHash

	nextPerField *TermsHashPerField

	bytesHash *util.BytesRefHash
}

func (h *TermsHashPerField) shrinkHash(targetSize int) {
	panic("not implemented yet")
}

func (h *TermsHashPerField) reset() {
	panic("not implemented yet")
}

func (h *TermsHashPerField) abort() {
	panic("not implemented yet")
}
